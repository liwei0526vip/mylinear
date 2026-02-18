# C08 — 评论与活动流 技术设计

## Context

### 背景

Issue 详情面板需要两个核心协作能力：
1. **评论系统**：团队成员可在 Issue 下讨论问题
2. **活动流**：追踪 Issue 的所有变更历史

### 现有状态

| 组件 | 状态 | 说明 |
|------|------|------|
| `Comment` 模型 | ✅ 已存在 | 包含 ID、IssueID、ParentID、UserID、Body、EditedAt 等字段 |
| `IssueStatusHistory` 模型 | ✅ 已存在 | 包含 FromStatusID、ToStatusID、ChangedByID、ChangedAt |
| `comments` 表 | ✅ 已存在 | C02 已创建迁移 |
| `issue_status_history` 表 | ✅ 已存在 | C02 已创建迁移 |
| Comment CRUD API | ❌ 未实现 | 本 Change 实现 |
| Activity 模型 | ❌ 不存在 | 本 Change 新增 |

### 约束

- 遵循 TDD 开发原则：先写测试，再写实现
- 使用真实数据库进行集成测试，拒绝 Mocks
- 前端使用 shadcn/ui 组件，遵循 Linear UI 规范

---

## Goals / Non-Goals

### Goals

1. 实现评论 CRUD API，支持嵌套回复和 @mention 解析
2. 实现活动流系统，记录 Issue 的所有变更事件
3. 在 Issue 状态变更时写入 `issue_status_history` 表
4. 前端 Issue 详情面板集成评论区和活动时间线

### Non-Goals

- 通知推送（C09 实现）
- 实时 WebSocket 推送（C41 实现）
- 富文本编辑器增强如图片上传（C19 实现）
- 评论表情反应（Phase 4 实现）

---

## Decisions

### D1: 活动模型设计

**决策**：新增 `activities` 表，使用 JSONB 存储活动详情

**方案**：
```go
type Activity struct {
    ID        uuid.UUID              `gorm:"type:uuid;primary_key"`
    IssueID   uuid.UUID              `gorm:"type:uuid;not null;index"`
    Type      ActivityType           `gorm:"type:varchar(50);not null;index"`
    ActorID   uuid.UUID              `gorm:"type:uuid;not null;index"`
    Payload   datatypes.JSON         `gorm:"type:jsonb"`
    CreatedAt time.Time              `gorm:"not null;default:now();index"`

    // 关联
    Issue *Issue `gorm:"foreignKey:IssueID"`
    Actor *User  `gorm:"foreignKey:ActorID"`
}
```

**活动类型定义**：
```go
type ActivityType string

const (
    ActivityIssueCreated       ActivityType = "issue_created"
    ActivityTitleChanged       ActivityType = "title_changed"
    ActivityDescriptionChanged ActivityType = "description_changed"
    ActivityStatusChanged      ActivityType = "status_changed"
    ActivityPriorityChanged    ActivityType = "priority_changed"
    ActivityAssigneeChanged    ActivityType = "assignee_changed"
    ActivityDueDateChanged     ActivityType = "due_date_changed"
    ActivityProjectChanged     ActivityType = "project_changed"
    ActivityLabelsChanged      ActivityType = "labels_changed"
    ActivityCommentAdded       ActivityType = "comment_added"
)
```

**理由**：
- JSONB 存储灵活：不同活动类型的 payload 结构不同，JSONB 避免了大量 nullable 字段
- PostgreSQL JSONB 性能优秀：支持索引和高效查询
- 扩展性强：新增活动类型只需添加常量，无需修改表结构

**替代方案**：
- ❌ 每种活动类型单独建表：过于复杂，查询时需要 UNION
- ❌ 使用 TEXT 存储 JSON：无法利用 PostgreSQL 的 JSONB 索引

---

### D2: @mention 解析策略

**决策**：使用正则表达式提取 `@username`，查询 User 表匹配

**方案**：
```go
var mentionRegex = regexp.MustCompile(`@([a-zA-Z0-9_]+)`)

func ParseMentions(body string) []string {
    matches := mentionRegex.FindAllStringSubmatch(body, -1)
    usernames := make([]string, 0, len(matches))
    for _, m := range matches {
        usernames = append(usernames, m[1])
    }
    return usernames
}
```

**处理流程**：
1. 创建评论时，解析 body 中的 @mentions
2. 查询 User 表验证 username 是否存在
3. 将被提及的用户自动添加为 Issue 订阅者
4. 为 C09 通知系统准备数据（当前仅记录，不推送）

**理由**：
- 简单高效：正则匹配足够处理常见场景
- 延迟验证：先解析 username，再查询数据库验证

**替代方案**：
- ❌ 使用 Markdown AST 解析：复杂度高，收益不大
- ❌ 前端解析后传给后端：前端信任问题，后端仍需验证

---

### D3: 评论嵌套回复深度

**决策**：支持无限层级嵌套，但前端 UI 仅展示 2 层

**理由**：
- 数据模型支持无限层级（通过 `parent_id` 递归）
- Linear 的 UI 设计：深层回复扁平化展示，避免嵌套过深
- 可在后续版本按需调整前端展示策略

**Payload 示例**：
```json
{
  "old_value": "原始标题",
  "new_value": "新标题"
}
```

---

### D4: 活动记录触发时机

**决策**：在 Service 层统一触发，不使用数据库触发器

**方案**：
```go
func (s *IssueService) UpdateIssue(ctx context.Context, id uuid.UUID, req UpdateIssueRequest) error {
    // 1. 获取当前 Issue
    issue, err := s.store.GetIssueByID(ctx, id)
    // ...

    // 2. 检测变更并记录活动
    if req.Title != nil && issue.Title != *req.Title {
        s.activityService.Record(ctx, Activity{
            IssueID: id,
            Type:    ActivityTitleChanged,
            ActorID: userID,
            Payload: json.RawMessage(`{"old":"`+issue.Title+`","new":"`+*req.Title+`"}`),
        })
    }

    // 3. 更新 Issue
    return s.store.UpdateIssue(ctx, id, req)
}
```

**理由**：
- Go 代码可控性强：易于测试、调试和扩展
- 避免数据库触发器：保持数据库简单，减少隐式行为
- 业务逻辑集中：所有活动记录逻辑在 Service 层

**替代方案**：
- ❌ 使用 PostgreSQL 触发器：难以测试，逻辑分散
- ❌ 使用 GORM 钩子：与 Service 层逻辑耦合不清晰

---

## Data Model Changes

### 新增：`activities` 表

```sql
CREATE TABLE activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    actor_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_activities_issue_id ON activities(issue_id);
CREATE INDEX idx_activities_type ON activities(type);
CREATE INDEX idx_activities_actor_id ON activities(actor_id);
CREATE INDEX idx_activities_created_at ON activities(created_at DESC);
```

### 已有：`comments` 表（无需修改）

字段已满足需求：`id`、`issue_id`、`parent_id`、`user_id`、`body`、`created_at`、`updated_at`、`edited_at`

### 已有：`issue_status_history` 表（无需修改）

字段已满足需求：`id`、`issue_id`、`from_status_id`、`to_status_id`、`changed_by_id`、`changed_at`

---

## API Design

### 评论 API

| 方法 | 端点 | 说明 |
|------|------|------|
| POST | `/api/v1/issues/:issueId/comments` | 创建评论 |
| GET | `/api/v1/issues/:issueId/comments` | 获取评论列表 |
| PUT | `/api/v1/comments/:id` | 更新评论 |
| DELETE | `/api/v1/comments/:id` | 删除评论 |

#### POST /api/v1/issues/:issueId/comments

**Request:**
```json
{
  "body": "这是一个评论，@alice 请看一下",
  "parent_id": "uuid-of-parent-comment"  // 可选，回复时提供
}
```

**Response (201):**
```json
{
  "id": "uuid",
  "issue_id": "uuid",
  "parent_id": null,
  "user": {
    "id": "uuid",
    "name": "张三",
    "username": "zhangsan",
    "avatar_url": "..."
  },
  "body": "这是一个评论，@alice 请看一下",
  "created_at": "2026-02-18T10:00:00Z",
  "updated_at": "2026-02-18T10:00:00Z",
  "edited_at": null,
  "mentions": [
    {"id": "uuid", "username": "alice", "name": "Alice"}
  ]
}
```

#### GET /api/v1/issues/:issueId/comments

**Query Parameters:**
- `page` (int, default: 1)
- `page_size` (int, default: 50, max: 100)
- `sort` (string, default: `created_at`, options: `created_at`, `-created_at`)

**Response (200):**
```json
{
  "items": [
    {
      "id": "uuid",
      "issue_id": "uuid",
      "parent_id": null,
      "user": {...},
      "body": "评论内容",
      "created_at": "...",
      "updated_at": "...",
      "edited_at": null,
      "replies": [
        {
          "id": "uuid",
          "parent_id": "uuid",
          "user": {...},
          "body": "回复内容",
          "created_at": "...",
          "replies": []
        }
      ]
    }
  ],
  "total": 10,
  "page": 1,
  "page_size": 50
}
```

### 活动流 API

| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/v1/issues/:issueId/activities` | 获取活动时间线 |

#### GET /api/v1/issues/:issueId/activities

**Query Parameters:**
- `page` (int, default: 1)
- `page_size` (int, default: 50, max: 100)
- `types` (string[], optional) - 过滤活动类型

**Response (200):**
```json
{
  "items": [
    {
      "id": "uuid",
      "type": "status_changed",
      "actor": {
        "id": "uuid",
        "name": "张三",
        "username": "zhangsan",
        "avatar_url": "..."
      },
      "payload": {
        "old_status": {"id": "uuid", "name": "Todo", "color": "#..."},
        "new_status": {"id": "uuid", "name": "In Progress", "color": "#..."}
      },
      "created_at": "2026-02-18T10:00:00Z"
    },
    {
      "id": "uuid",
      "type": "comment_added",
      "actor": {...},
      "payload": {
        "comment_id": "uuid",
        "comment_preview": "评论内容预览（前 100 字符）..."
      },
      "created_at": "2026-02-18T09:30:00Z"
    }
  ],
  "total": 25,
  "page": 1,
  "page_size": 50
}
```

---

## UI Design

### 评论区设计

参考 Linear UI 规范（`docs/竞品分析/11-UI-UX设计规范.md`）：

**布局结构**：
```
┌─────────────────────────────────────────────────────────┐
│ Comments                                    Sort ▼     │
├─────────────────────────────────────────────────────────┤
│ 👤 张三  2小时前                                         │
│ 这是评论内容，@alice 请看一下                           │
│                     Reply · Edited · Delete             │
├─────────────────────────────────────────────────────────┤
│   👤 Alice  1小时前                                     │
│   这是回复内容                                          │
│                       Reply · Delete                    │
├─────────────────────────────────────────────────────────┤
│ 👤 李四  30分钟前                                       │
│ 另一条评论                                              │
│                     Reply · Delete                      │
├─────────────────────────────────────────────────────────┤
│ ┌─────────────────────────────────────────────────────┐ │
│ │ Leave a comment...                                  │ │
│ │                                                     │ │
│ │ [Markdown 预览] [附件]              [Comment]       │ │
│ └─────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

**组件拆分**：
- `CommentSection.tsx` - 评论区容器
- `CommentList.tsx` - 评论列表
- `CommentItem.tsx` - 单条评论（含嵌套回复）
- `CommentInput.tsx` - 评论输入框

### 活动时间线设计

**布局结构**：
```
┌─────────────────────────────────────────────────────────┐
│ Activity                                                │
├─────────────────────────────────────────────────────────┤
│ 🔵 张三 changed status from Todo to In Progress        │
│    2 hours ago                                          │
├─────────────────────────────────────────────────────────┤
│ 💬 Alice commented                                      │
│    "这是评论预览..."                                    │
│    3 hours ago                                          │
├─────────────────────────────────────────────────────────┤
│ ✏️ 李四 updated title                                   │
│    "新标题" ← "旧标题"                                  │
│    5 hours ago                                          │
└─────────────────────────────────────────────────────────┘
```

**组件拆分**：
- `ActivityTimeline.tsx` - 活动时间线容器
- `ActivityItem.tsx` - 单条活动

---

## Risks / Trade-offs

### R1: 活动记录性能

**风险**：高频 Issue 更新可能产生大量活动记录

**缓解措施**：
- 活动表按 `created_at` 降序索引，查询高效
- 分页查询，避免一次加载过多
- 后续可考虑活动归档策略（如：超过 1 年的活动归档）

### R2: @mention 解析准确性

**风险**：正则解析可能误匹配（如邮箱地址 `user@example.com`）

**缓解措施**：
- 正则限定 `@` 后为 `[a-zA-Z0-9_]+`，排除含 `.` 的情况
- 解析后查询数据库验证 username 存在性
- 前端高亮显示已识别的 @mention

### R3: 评论嵌套深度

**风险**：无限层级嵌套可能导致 UI 展示复杂

**缓解措施**：
- 前端限制展示深度为 2 层，更深层级"查看回复"展开
- API 返回树形结构，前端按需渲染
- 可在后续版本添加深度限制配置

---

## Open Questions

1. **活动记录保留策略**：是否需要活动归档/清理机制？
   - 暂定：不实现，保留所有活动记录

2. **评论排序**：默认按时间正序还是倒序？
   - 暂定：正序（Linear 默认行为）

3. **评论编辑权限**：仅作者可编辑，还是管理员也可编辑？
   - 暂定：仅作者可编辑自己的评论
