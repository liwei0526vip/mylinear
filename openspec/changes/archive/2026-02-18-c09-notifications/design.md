## Context

通知系统是 MVP 阶段的核心协作功能，需要与已有的 Activity Stream（C08）和 Issue CRUD（C05）深度集成。当前系统已具备：
- 用户认证与权限体系（C03）
- Issue 核心模型与 CRUD（C05）
- Activity Stream 活动记录（C08）
- Comment 评论功能（C08）

本 change 需要在上述基础上构建通知推送和收件箱功能。

## Goals / Non-Goals

**Goals:**
- 实现应用内通知收件箱（列表、已读/未读、批量操作）
- 实现通知触发机制（指派、@mention、订阅变更）
- 预留通知配置扩展能力（为后续 Phase 邮件/IM 通知做准备）
- 与现有 Activity Stream 集成，复用活动类型

**Non-Goals:**
- 暂不实现邮件通知（Phase 3）
- 暂不实现 IM 通知推送（Phase 4）
- 暂不实现 WebSocket 实时推送（Phase 2 考虑）
- 暂不实现桌面/移动端推送通知

## Decisions

### D1: 通知数据模型

**决策**：使用单表 `notifications` 存储通知，通过 `type` 字段区分通知类型，`resource_type` + `resource_id` 关联资源。

**表结构**：
```sql
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,           -- 通知类型
    title VARCHAR(255) NOT NULL,         -- 通知标题
    body TEXT,                           -- 通知内容
    resource_type VARCHAR(50),           -- 关联资源类型（issue, comment, project）
    resource_id UUID,                    -- 关联资源 ID
    read_at TIMESTAMPTZ,                 -- 已读时间（null 表示未读）
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_user_read ON notifications(user_id, read_at) WHERE read_at IS NULL;
```

**通知类型枚举**：
| Type | 说明 | 触发场景 |
|------|------|----------|
| `issue_assigned` | Issue 被指派给你 | Issue.assigneeId 变更为当前用户 |
| `issue_mentioned` | 在评论中被 @mention | 评论内容包含 @username |
| `issue_status_changed` | 订阅的 Issue 状态变更 | Issue.statusId 变更，用户已订阅 |
| `issue_commented` | 订阅的 Issue 有新评论 | 新评论创建，用户已订阅 |
| `issue_priority_changed` | 订阅的 Issue 优先级变更 | Issue.priority 变更，用户已订阅 |

**替代方案**：
- 多态关联（Polymorphic）：使用 `resource_type` + `resource_id` 字段。**选择理由**：灵活且符合现有模式（参考 Attachment 表设计）。
- 独立表：每种通知类型单独建表。**放弃理由**：复杂度高，查询聚合困难。

### D2: 通知触发机制

**决策**：在 Service 层通过显式调用触发通知，而非使用事件总线。

**实现方式**：
```go
// NotificationService 提供 Notify 接口
type NotificationService interface {
    NotifyIssueAssigned(ctx context.Context, issueID, assigneeID, actorID uuid.UUID) error
    NotifyIssueMentioned(ctx context.Context, issueID, mentionedUserID, actorID uuid.UUID, commentPreview string) error
    NotifySubscribers(ctx context.Context, issueID, actorID uuid.UUID, activityType activity.ActivityType) error
}

// IssueService 在更新时调用通知服务
func (s *issueService) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*Issue, error) {
    // ... 更新逻辑
    if input.AssigneeID != nil && *input.AssigneeID != oldAssigneeID {
        s.notificationService.NotifyIssueAssigned(ctx, id, *input.AssigneeID, actorID)
    }
}
```

**替代方案**：
- 事件总线（Event Bus）：发布/订阅模式。**放弃理由**：MVP 阶段过度设计，增加复杂度。
- 数据库触发器：在 Activity 表插入时触发。**放弃理由**：业务逻辑应在应用层，便于测试和维护。

### D3: 订阅机制

**决策**：复用 C05 已实现的订阅机制，通过 `issue_subscribers` 中间表管理。

**订阅规则**：
- Issue 创建者自动订阅
- 被指派人自动订阅
- 评论者自动订阅
- 用户可手动取消订阅（Unsubscribe）

**通知过滤**：
```go
// 只通知已订阅且非操作者本人的用户
func (s *notificationService) NotifySubscribers(ctx context.Context, issueID, actorID uuid.UUID, activityType activity.ActivityType) error {
    subscribers := s.getSubscribers(ctx, issueID)
    for _, sub := range subscribers {
        if sub.UserID == actorID {
            continue // 不通知操作者本人
        }
        s.createNotification(ctx, sub.UserID, issueID, activityType)
    }
}
```

### D4: API 设计

**决策**：遵循 RESTful 风格，提供标准 CRUD 和批量操作。

**端点设计**：
| Method | Path | 说明 |
|--------|------|------|
| GET | `/api/v1/notifications` | 获取通知列表（支持分页、过滤） |
| POST | `/api/v1/notifications/:id/read` | 标记单条已读 |
| POST | `/api/v1/notifications/read-all` | 标记全部已读 |
| POST | `/api/v1/notifications/batch-read` | 批量标记已读 |
| GET | `/api/v1/notifications/unread-count` | 获取未读数量 |
| GET | `/api/v1/notification-preferences` | 获取通知配置 |
| PUT | `/api/v1/notification-preferences` | 更新通知配置 |

**查询参数**：
```
GET /api/v1/notifications?read=false&type=issue_assigned&page=1&page_size=20
```

### D5: 前端架构

**决策**：使用 Zustand 管理通知状态，与侧边栏导航集成显示未读计数。

**状态结构**：
```typescript
interface NotificationState {
  notifications: Notification[];
  unreadCount: number;
  isLoading: boolean;
  fetchNotifications: (params?: QueryParams) => Promise<void>;
  markAsRead: (id: string) => Promise<void>;
  markAllAsRead: () => Promise<void>;
  fetchUnreadCount: () => Promise<void>;
}
```

**UI 组件**：
- `InboxPage`：通知收件箱主页面（三栏布局的中间内容区）
- `NotificationList`：通知列表组件
- `NotificationItem`：单条通知组件（支持点击跳转到资源）
- `NotificationBadge`：侧边栏未读计数徽章

### D6: 通知配置模型（预留）

**决策**：使用 `notification_preferences` 表存储用户偏好，为后续 Phase 扩展预留。

**表结构**：
```sql
CREATE TABLE notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel VARCHAR(20) NOT NULL DEFAULT 'in_app',  -- in_app, email, slack
    type VARCHAR(50) NOT NULL,                       -- 通知类型
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, channel, type)
);
```

**MVP 范围**：仅支持 `in_app` 渠道，默认全部启用。

## Risks / Trade-offs

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| @mention 解析复杂度高 | 评论中解析 @username 需要处理边界情况 | 使用正则表达式 + 用户名白名单校验 |
| 通知风暴 | 短时间内大量变更可能产生大量通知 | 同一 Issue 同类型变更合并（后续优化） |
| 性能问题 | 通知表增长快，查询变慢 | 添加索引，定期归档旧通知 |
| 实时性不足 | 用户需刷新页面才能看到新通知 | MVP 阶段可接受，Phase 2 引入 WebSocket |

## Migration Plan

1. **数据库迁移**：创建 `notifications` 和 `notification_preferences` 表
2. **后端实现**：按 Model → Store → Service → Handler 顺序实现
3. **集成点扩展**：修改 IssueService 和 CommentService 添加通知触发
4. **前端实现**：创建 Inbox 页面和通知组件
5. **端到端测试**：验证通知创建、显示、标记已读流程

## Open Questions

- [ ] 是否需要在通知列表中显示通知预览内容？（建议：是，显示 title + body 摘要）
- [ ] 未读通知是否需要按时间倒序排列？（建议：是，最新在前）
- [ ] 通知是否支持删除功能？（建议：MVP 不支持，仅标记已读）
