> **TDD 开发原则**：严格遵循 Red-Green-Refactor 循环
> - 🔴 Red：先写失败的测试
> - 🟢 Green：写最少的代码让测试通过
> - 🔵 Refactor：重构代码（保持测试通过）

## 1. 数据库迁移

- [x] 1.1 创建 `notifications` 表迁移文件（包含索引）
- [x] 1.2 创建 `notification_preferences` 表迁移文件（包含唯一约束）
- [x] 1.3 执行迁移，验证表结构正确

---

## 2. Notification 模型与 Store 层

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

### 2.1 Model 定义

- [x] 2.1.1 🔴 编写 Notification 模型测试（验证字段、TableName、关联关系）
- [x] 2.1.2 🟢 实现 Notification 模型（`internal/model/notification.go`）
- [x] 2.1.3 🔴 编写 NotificationType 枚举测试（验证所有类型值）
- [x] 2.1.4 🟢 实现 NotificationType 枚举

### 2.2 NotificationStore CRUD

- [x] 2.2.1 🔴 编写 NotificationStore.Create 测试（表格驱动：正常创建、字段验证）
- [x] 2.2.2 🟢 实现 NotificationStore.Create
- [x] 2.2.3 🔴 编写 NotificationStore.GetByID 测试（表格驱动：存在、不存在）
- [x] 2.2.4 🟢 实现 NotificationStore.GetByID
- [x] 2.2.5 🔴 编写 NotificationStore.List 测试（表格驱动：分页、按用户过滤、按已读状态过滤、按类型过滤）
- [x] 2.2.6 🟢 实现 NotificationStore.List
- [x] 2.2.7 🔴 编写 NotificationStore.CountUnread 测试（表格驱动：有未读、无未读、多用户隔离）
- [x] 2.2.8 🟢 实现 NotificationStore.CountUnread
- [x] 2.2.9 🔴 编写 NotificationStore.MarkAsRead 测试（表格驱动：单条标记、批量标记、标记全部）
- [x] 2.2.10 🟢 实现 NotificationStore.MarkAsRead、MarkAllAsRead、MarkBatchAsRead

### 2.3 NotificationPreferenceStore

- [x] 2.3.1 🔴 编写 NotificationPreferenceStore.GetByUser 测试（表格驱动：有配置、无配置返回默认值）
- [x] 2.3.2 🟢 实现 NotificationPreferenceStore.GetByUser
- [x] 2.3.3 🔴 编写 NotificationPreferenceStore.Upsert 测试（表格驱动：新建、更新、批量更新）
- [x] 2.3.4 🟢 实现 NotificationPreferenceStore.Upsert、BatchUpsert
- [x] 2.3.5 🔴 编写 NotificationPreferenceStore.IsEnabled 测试（表格驱动：启用、禁用、无配置默认启用）
- [x] 2.3.6 🟢 实现 NotificationPreferenceStore.IsEnabled

---

## 3. Notification Service 层

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

### 3.1 通知创建服务

- [x] 3.1.1 🔴 编写 NotificationService.CreateNotification 测试（表格驱动：Issue指派通知、@mention通知、订阅变更通知）
- [x] 3.1.2 🟢 实现 NotificationService.CreateNotification
- [x] 3.1.3 🔴 编写 NotificationService.NotifyIssueAssigned 测试（表格驱动：正常指派、指派给自己不通知、取消指派不通知）
- [x] 3.1.4 🟢 实现 NotificationService.NotifyIssueAssigned
- [x] 3.1.5 🔴 编写 NotificationService.NotifyIssueMentioned 测试（表格驱动：单个mention、多个mention、自己mention自己不通知、无效username忽略）
- [x] 3.1.6 🟢 实现 NotificationService.NotifyIssueMentioned（含 @mention 解析逻辑）
- [x] 3.1.7 🔴 编写 NotificationService.NotifySubscribers 测试（表格驱动：状态变更、新评论、优先级变更、排除操作者本人）
- [x] 3.1.8 🟢 实现 NotificationService.NotifySubscribers

### 3.2 通知查询服务

- [x] 3.2.1 🔴 编写 NotificationService.ListNotifications 测试（表格驱动：分页、过滤、未读优先排序）
- [x] 3.2.2 🟢 实现 NotificationService.ListNotifications
- [x] 3.2.3 🔴 编写 NotificationService.GetUnreadCount 测试
- [x] 3.2.4 🟢 实现 NotificationService.GetUnreadCount

### 3.3 标记已读服务

- [x] 3.3.1 🔴 编写 NotificationService.MarkAsRead 测试（表格驱动：单条标记、标记他人通知返回错误）
- [x] 3.3.2 🟢 实现 NotificationService.MarkAsRead
- [x] 3.3.3 🔴 编写 NotificationService.MarkAllAsRead 测试
- [x] 3.3.4 🟢 实现 NotificationService.MarkAllAsRead
- [x] 3.3.5 🔴 编写 NotificationService.MarkBatchAsRead 测试
- [x] 3.3.6 🟢 实现 NotificationService.MarkBatchAsRead

### 3.4 通知配置服务

- [x] 3.4.1 🔴 编写 NotificationPreferenceService.GetPreferences 测试（表格驱动：有配置、无配置返回默认值）
- [x] 3.4.2 🟢 实现 NotificationPreferenceService.GetPreferences
- [x] 3.4.3 🔴 编写 NotificationPreferenceService.UpdatePreferences 测试（表格驱动：单个更新、批量更新）
- [x] 3.4.4 🟢 实现 NotificationPreferenceService.UpdatePreferences

---

## 4. Notification Handler 层

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

### 4.1 通知 API Handler

- [x] 4.1.1 🔴 编写 Handler.ListNotifications 测试（表格驱动：正常获取、分页参数、过滤参数、未授权）
- [x] 4.1.2 🟢 实现 Handler.ListNotifications（`GET /api/v1/notifications`）
- [x] 4.1.3 🔴 编写 Handler.GetUnreadCount 测试（表格驱动：正常获取、未授权）
- [x] 4.1.4 🟢 实现 Handler.GetUnreadCount（`GET /api/v1/notifications/unread-count`）
- [x] 4.1.5 🔴 编写 Handler.MarkAsRead 测试（表格驱动：正常标记、不存在、他人通知、未授权）
- [x] 4.1.6 🟢 实现 Handler.MarkAsRead（`POST /api/v1/notifications/:id/read`）
- [x] 4.1.7 🔴 编写 Handler.MarkAllAsRead 测试（表格驱动：正常标记、无未读、未授权）
- [x] 4.1.8 🟢 实现 Handler.MarkAllAsRead（`POST /api/v1/notifications/read-all`）
- [x] 4.1.9 🔴 编写 Handler.MarkBatchAsRead 测试（表格驱动：正常批量、部分不存在、未授权）
- [x] 4.1.10 🟢 实现 Handler.MarkBatchAsRead（`POST /api/v1/notifications/batch-read`）

### 4.2 通知配置 API Handler

- [x] 4.2.1 🔴 编写 Handler.GetPreferences 测试（表格驱动：正常获取、按渠道过滤、未授权）
- [x] 4.2.2 🟢 实现 Handler.GetPreferences（`GET /api/v1/notification-preferences`）
- [x] 4.2.3 🔴 编写 Handler.UpdatePreferences 测试（表格驱动：单个更新、批量更新、MVP仅支持in_app渠道、未授权）
- [x] 4.2.4 🟢 实现 Handler.UpdatePreferences（`PUT /api/v1/notification-preferences`）

### 4.3 路由注册

- [x] 4.3.1 🔴 编写路由注册测试（验证所有端点已注册且需要认证）
- [x] 4.3.2 🟢 注册通知相关路由到 Gin Router

---

## 5. 集成点扩展

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

### 5.1 IssueService 集成

- [x] 5.1.1 🔴 编写 IssueService.Update 通知触发测试（表格驱动：指派通知、状态变更通知订阅者、优先级变更通知订阅者）
- [x] 5.1.2 🟢 扩展 IssueService.Update 添加通知触发逻辑
- [x] 5.1.3 🔴 编写 IssueService.Create 自动订阅测试（验证创建者自动订阅）
- [x] 5.1.4 🟢 确认 IssueService.Create 已实现创建者自动订阅（C05 已实现）

### 5.2 CommentService 集成

- [x] 5.2.1 🔴 编写 CommentService.Create @mention 解析测试（表格驱动：单个mention、多个mention、无效username、评论者自己不通知）
- [x] 5.2.2 🟢 实现 @mention 解析工具函数（`internal/service/mention.go`）
- [x] 5.2.3 🔴 编写 CommentService.Create 通知触发测试（表格驱动：mention通知、评论通知订阅者、评论者自动订阅）
- [x] 5.2.4 🟢 扩展 CommentService.Create 添加通知触发逻辑

---

## 6. 前端实现

### 6.1 API 层

- [x] 6.1.1 创建 `src/api/notifications.ts`（通知列表、未读数量、标记已读 API）
- [x] 6.1.2 创建 `src/api/notification-preferences.ts`（通知配置 API）

### 6.2 Zustand Store

- [x] 6.2.1 创建 `src/stores/notificationStore.ts`（notifications、unreadCount、fetchNotifications、markAsRead、markAllAsRead）
- [x] 6.2.2 实现 fetchUnreadCount 方法
- [x] 6.2.3 实现 markAsRead 和 markAllAsRead 方法

### 6.3 通知组件

- [x] 6.3.1 创建 `src/components/notifications/NotificationItem.tsx`（单条通知组件，支持点击跳转）
- [x] 6.3.2 创建 `src/components/notifications/NotificationList.tsx`（通知列表组件，支持滚动加载）
- [x] 6.3.3 创建 `src/components/notifications/NotificationBadge.tsx`（侧边栏未读计数徽章）
- [x] 6.3.4 创建 `src/components/notifications/NotificationItemSkeleton.tsx`（加载骨架屏）

### 6.4 Inbox 页面

- [x] 6.4.1 创建 `src/pages/inbox/InboxPage.tsx`（通知收件箱主页面）
- [x] 6.4.2 实现通知列表展示（未读/已读状态、时间显示）
- [x] 6.4.3 实现批量操作工具栏（全部已读、批量已读）
- [x] 6.4.4 实现空状态展示

### 6.5 路由与导航集成

- [x] 6.5.1 添加 Inbox 路由到 React Router
- [x] 6.5.2 侧边栏导航添加 Inbox 入口与未读徽章
- [x] 6.5.3 实现点击通知跳转到对应 Issue 详情页

---

## 7. 端到端验证

### 7.1 API 集成测试

- [x] 7.1.1 🔴 编写 E2E 测试：Issue 指派 → 生成通知 → 查询通知列表
- [x] 7.1.2 🔴 编写 E2E 测试：评论 @mention → 生成通知 → 标记已读
- [x] 7.1.3 🔴 编写 E2E 测试：Issue 状态变更 → 通知订阅者 → 未读计数更新
- [x] 7.1.4 🔴 编写 E2E 测试：通知配置禁用 → 不生成通知

### 7.2 前端功能验证

- [x] 7.2.1 验证 Inbox 页面正常显示通知列表
- [x] 7.2.2 验证未读计数在侧边栏正确显示
- [x] 7.2.3 验证点击通知跳转到 Issue 详情页
- [x] 7.2.4 验证标记已读功能正常工作

---

## 任务统计

| 分组                 | 任务数 | 预估工时 |
| -------------------- | ------ | -------- |
| 1. 数据库迁移        | 3      | 1h       |
| 2. Model 与 Store 层 | 16     | 4h       |
| 3. Service 层        | 16     | 4h       |
| 4. Handler 层        | 12     | 3h       |
| 5. 集成点扩展        | 8      | 2h       |
| 6. 前端实现          | 14     | 4h       |
| 7. 端到端验证        | 8      | 2h       |
| **总计**             | **77** | **~20h** |

> 后端 TDD 任务数：55（含测试任务 28 个）
> 前端任务数：14
> 验证任务数：8
