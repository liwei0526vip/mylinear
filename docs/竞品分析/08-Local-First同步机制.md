## 八、Local-First 同步机制

### 8.1 架构概览

```
┌─────────────────────────────────────────────────────────────────────┐
│                           Client                                     │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────────┐  │
│  │   React UI  │◄──►│   MobX      │◄──►│   IndexedDB             │  │
│  │             │    │   Store     │    │   (本地数据持久化)        │  │
│  └─────────────┘    └──────┬──────┘    └───────────┬─────────────┘  │
│                            │                        │                │
│                            │   Optimistic Update    │                │
│                            │   (即时 UI 响应)        │                │
│                            ▼                        ▼                │
│                     ┌──────────────────────────────────┐            │
│                     │        Sync Engine               │            │
│                     │  (增量同步 + 冲突解决)             │            │
│                     └──────────────┬───────────────────┘            │
└────────────────────────────────────┼────────────────────────────────┘
                                     │
                              WebSocket
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────┐
│                          Server                                      │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────────┐  │
│  │  WebSocket  │───►│   Sync      │───►│   PostgreSQL            │  │
│  │   Server    │    │   Service   │    │   (数据存储)             │  │
│  └─────────────┘    └─────────────┘    └─────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

### 8.2 IndexedDB Schema 设计

```typescript
// IndexedDB 数据库结构
const DB_NAME = 'linear-offline';
const DB_VERSION = 1;

// 对象存储（Object Stores）
const stores = {
  // Issue 存储
  issues: {
    keyPath: 'id',
    indexes: [
      'team_id', 'project_id', 'cycle_id', 'assignee_id',
      'status_id', 'parent_id', 'number', 'updated_at'
    ]
  },

  // Project 存储
  projects: {
    keyPath: 'id',
    indexes: ['workspace_id', 'team_id', 'status']
  },

  // 同步元数据
  syncMetadata: {
    keyPath: 'entity_type', // 'issues', 'projects', etc.
  },

  // 操作队列（离线时暂存）
  operationQueue: {
    keyPath: 'id',
    autoIncrement: true,
    indexes: ['timestamp', 'status']
  }
};

// 初始化数据库
function initDB(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(DB_NAME, DB_VERSION);

    request.onupgradeneeded = (event) => {
      const db = (event.target as IDBOpenDBRequest).result;

      // 创建 Issue 存储
      const issueStore = db.createObjectStore('issues', { keyPath: 'id' });
      issueStore.createIndex('team_id', 'team_id', { unique: false });
      issueStore.createIndex('project_id', 'project_id', { unique: false });
      issueStore.createIndex('updated_at', 'updated_at', { unique: false });

      // 创建同步元数据存储
      db.createObjectStore('syncMetadata', { keyPath: 'entity_type' });

      // 创建操作队列存储
      const opStore = db.createObjectStore('operationQueue', {
        keyPath: 'id',
        autoIncrement: true
      });
      opStore.createIndex('timestamp', 'timestamp', { unique: false });
    };

    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error);
  });
}
```

### 8.3 增量同步算法

```typescript
// 同步引擎
class SyncEngine {
  private ws: WebSocket | null = null;
  private db: IDBDatabase;
  private syncVersion: Record<string, string> = {}; // 实体类型 → 版本号

  // 初始化同步
  async initialize(): Promise<void> {
    // 1. 加载本地同步元数据
    await this.loadSyncMetadata();

    // 2. 连接 WebSocket
    this.connectWebSocket();

    // 3. 执行初始同步
    await this.performInitialSync();
  }

  // 加载同步元数据
  private async loadSyncMetadata(): Promise<void> {
    const tx = this.db.transaction('syncMetadata', 'readonly');
    const store = tx.objectStore('syncMetadata');

    const entities = ['issues', 'projects', 'cycles', 'teams'];
    for (const entity of entities) {
      const metadata = await store.get(entity);
      if (metadata) {
        this.syncVersion[entity] = metadata.last_sync_version;
      }
    }
  }

  // 执行初始同步
  private async performInitialSync(): Promise<void> {
    // 发送同步请求，带上本地版本号
    this.sendMessage({
      type: 'sync_request',
      versions: this.syncVersion
    });
  }

  // 处理同步响应
  private async handleSyncResponse(data: SyncResponse): Promise<void> {
    const { entity_type, changes, new_version } = data;

    const tx = this.db.transaction(entity_type, 'readwrite');
    const store = tx.objectStore(entity_type);

    // 应用变更
    for (const change of changes) {
      switch (change.operation) {
        case 'create':
        case 'update':
          await store.put(change.entity);
          break;
        case 'delete':
          await store.delete(change.entity_id);
          break;
      }
    }

    // 更新同步版本
    this.syncVersion[entity_type] = new_version;
    await this.saveSyncMetadata(entity_type, new_version);

    // 通知 UI 更新
    this.notifyUI(entity_type, changes);
  }

  // 处理本地操作（乐观更新）
  async performLocalOperation(operation: Operation): Promise<void> {
    // 1. 立即更新本地存储
    await this.applyOperationLocally(operation);

    // 2. 立即更新 UI（乐观）
    this.notifyUI(operation.entity_type, [operation]);

    // 3. 如果在线，发送到服务器
    if (this.isOnline()) {
      this.sendMessage({
        type: 'mutation',
        operation: operation
      });
    } else {
      // 4. 如果离线，加入操作队列
      await this.queueOperation(operation);
    }
  }

  // 离线操作队列
  private async queueOperation(operation: Operation): Promise<void> {
    const tx = this.db.transaction('operationQueue', 'readwrite');
    const store = tx.objectStore('operationQueue');

    await store.add({
      ...operation,
      timestamp: Date.now(),
      status: 'pending'
    });
  }

  // 重新上线时同步离线操作
  async syncOfflineOperations(): Promise<void> {
    const tx = this.db.transaction('operationQueue', 'readwrite');
    const store = tx.objectStore('operationQueue');
    const index = store.index('status');

    const pendingOps = await index.getAll('pending');

    for (const op of pendingOps) {
      this.sendMessage({
        type: 'mutation',
        operation: op
      });
    }
  }
}
```

### 8.4 冲突解决策略

```typescript
// 冲突类型
type ConflictType = 'update_update' | 'update_delete';

// 冲突解决策略
class ConflictResolver {
  // 基于时间戳的 Last-Write-Wins
  resolveByTimestamp(local: Entity, remote: Entity): Entity {
    if (remote.updated_at > local.updated_at) {
      return remote;
    }
    return local;
  }

  // 字段级合并
  resolveByFieldMerge(local: Entity, remote: Entity, base: Entity): Entity {
    const resolved: Entity = { ...local };

    for (const key of Object.keys(remote)) {
      // 如果远程字段与基准不同，说明远程有修改
      if (remote[key] !== base[key]) {
        // 如果本地字段与基准相同，使用远程值
        if (local[key] === base[key]) {
          resolved[key] = remote[key];
        }
        // 否则保留本地值（Last-Write-Wins）
      }
    }

    return resolved;
  }

  // 版本向量冲突检测
  detectConflict(local: Entity, remote: Entity): boolean {
    // 如果版本不连续，说明存在并发修改
    return Math.abs(local.version - remote.version) > 1;
  }
}
```

### 8.5 WebSocket 消息格式

```typescript
// 消息类型定义
interface WebSocketMessage {
  type: 'sync_request' | 'sync_response' | 'mutation' | 'mutation_result' | 'push';
}

// 同步请求
interface SyncRequest extends WebSocketMessage {
  type: 'sync_request';
  versions: Record<string, string>; // 实体类型 → 最后同步版本
}

// 同步响应
interface SyncResponse extends WebSocketMessage {
  type: 'sync_response';
  entity_type: string;
  changes: Change[];
  new_version: string;
}

interface Change {
  operation: 'create' | 'update' | 'delete';
  entity_id: string;
  entity?: Entity;
  version: string;
  timestamp: string;
}

// 变更推送
interface PushNotification extends WebSocketMessage {
  type: 'push';
  entity_type: string;
  change: Change;
}
```

---

