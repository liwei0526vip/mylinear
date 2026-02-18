## ADDED Requirements

### Requirement: Notification 模型

Notification 模型 MUST 定义通知记录的完整结构。

#### Scenario: Notification 模型字段

- **WHEN** 定义 `Notification` 结构体
- **THEN** MUST 包含字段：`ID`（UUID）、`UserID`（FK）、`Type`（NotificationType）、`Title`（string）、`Body`（string, nullable）、`ResourceType`（string, nullable）、`ResourceID`（UUID, nullable）、`ReadAt`（time, nullable）、`CreatedAt`

#### Scenario: NotificationType 枚举

- **WHEN** 定义 `NotificationType` 类型
- **THEN** MUST 支持：`issue_assigned`、`issue_mentioned`、`issue_status_changed`、`issue_commented`、`issue_priority_changed`

#### Scenario: Notification 关联关系

- **WHEN** 定义 `Notification` 关联
- **THEN** MUST 支持 `User` 的 belongs-to 关系

#### Scenario: Notification 表名

- **WHEN** 定义 `Notification` 的 `TableName()` 方法
- **THEN** MUST 返回 `notifications`

---

### Requirement: 通知创建 - Issue 指派

系统 SHALL 在 Issue 被指派给用户时创建通知。

#### Scenario: Issue 指派创建通知

- **WHEN** Issue 的 `assigneeId` 被设置为一个用户
- **THEN** 系统 SHALL 为该用户创建 `issue_assigned` 类型的通知
- **AND** `title` SHALL 为 "你被指派了一个 Issue"
- **AND** `body` SHALL 包含 Issue 标题
- **AND** `resource_type` SHALL 为 "issue"
- **AND** `resource_id` SHALL 为 Issue ID

#### Scenario: 指派给自己不创建通知

- **WHEN** 用户将 Issue 指派给自己
- **THEN** 系统 SHALL NOT 创建通知

#### Scenario: 取消指派不创建通知

- **WHEN** Issue 的 `assigneeId` 被设置为 null
- **THEN** 系统 SHALL NOT 创建通知

---

### Requirement: 通知创建 - @Mention

系统 SHALL 在评论中用户被 @mention 时创建通知。

#### Scenario: @mention 创建通知

- **WHEN** 评论内容包含 `@username` 格式的提及
- **AND** 该 username 对应一个有效用户
- **THEN** 系统 SHALL 为被提及的用户创建 `issue_mentioned` 类型的通知
- **AND** `title` SHALL 为 "你在评论中被提及"
- **AND** `body` SHALL 包含评论内容预览（前 100 字符）

#### Scenario: 自己 @mention 自己不创建通知

- **WHEN** 评论者在自己评论中 @mention 自己
- **THEN** 系统 SHALL NOT 创建通知

#### Scenario: 多个 @mention 创建多条通知

- **WHEN** 评论内容包含多个 `@username` 提及
- **THEN** 系统 SHALL 为每个被提及的用户分别创建通知

#### Scenario: 无效 username 不创建通知

- **WHEN** 评论内容包含 `@invalid_user` 格式
- **AND** 该 username 不存在
- **THEN** 系统 SHALL 忽略该提及，不创建通知

---

### Requirement: 通知创建 - 订阅变更

系统 SHALL 在用户订阅的 Issue 发生变更时创建通知。

#### Scenario: 状态变更通知订阅者

- **WHEN** Issue 的 `statusId` 发生变更
- **THEN** 系统 SHALL 为所有订阅该 Issue 的用户创建 `issue_status_changed` 类型的通知
- **AND** 通知 SHALL 排除变更操作者本人

#### Scenario: 新评论通知订阅者

- **WHEN** Issue 收到新评论
- **THEN** 系统 SHALL 为所有订阅该 Issue 的用户创建 `issue_commented` 类型的通知
- **AND** 通知 SHALL 排除评论者本人

#### Scenario: 优先级变更通知订阅者

- **WHEN** Issue 的 `priority` 发生变更
- **THEN** 系统 SHALL 为所有订阅该 Issue 的用户创建 `issue_priority_changed` 类型的通知
- **AND** 通知 SHALL 排除变更操作者本人

---

### Requirement: 通知查询 API

系统 SHALL 支持查询用户的通知列表。

#### Scenario: 获取通知列表

- **WHEN** 用户通过 `GET /api/v1/notifications` 查询通知
- **THEN** 系统 SHALL 返回当前用户的通知列表
- **AND** 默认按 `created_at` 倒序排列（最新在前）

#### Scenario: 分页查询

- **WHEN** 用户指定 `page` 和 `page_size` 参数
- **THEN** 系统 SHALL 返回分页结果，包含 `total`（总数）、`items`（当前页）
- **AND** 默认 `page_size` SHALL 为 20，最大 100

#### Scenario: 按已读状态过滤

- **WHEN** 用户指定 `read=false` 参数
- **THEN** 系统 SHALL 仅返回未读通知

#### Scenario: 按类型过滤

- **WHEN** 用户指定 `type=issue_assigned,issue_mentioned` 参数
- **THEN** 系统 SHALL 仅返回指定类型的通知

#### Scenario: 未读通知优先

- **WHEN** 用户未指定排序参数
- **THEN** 未读通知 SHALL 排在已读通知之前

---

### Requirement: 未读数量 API

系统 SHALL 支持查询用户的未读通知数量。

#### Scenario: 获取未读数量

- **WHEN** 用户通过 `GET /api/v1/notifications/unread-count` 查询
- **THEN** 系统 SHALL 返回 `{"count": <number>}` 格式的响应

#### Scenario: 仅统计当前用户

- **WHEN** 查询未读数量
- **THEN** 系统 SHALL 仅统计当前用户的通知

---

### Requirement: 标记已读 API

系统 SHALL 支持将通知标记为已读。

#### Scenario: 标记单条已读

- **WHEN** 用户通过 `POST /api/v1/notifications/:id/read` 标记已读
- **THEN** 系统 SHALL 将该通知的 `read_at` 设置为当前时间
- **AND** 返回更新后的通知对象

#### Scenario: 标记不存在的通知

- **WHEN** 用户尝试标记不存在的通知
- **THEN** 系统 SHALL 返回 404 Not Found

#### Scenario: 标记他人通知

- **WHEN** 用户尝试标记属于其他用户的通知
- **THEN** 系统 SHALL 返回 404 Not Found（不泄露存在性）

#### Scenario: 标记全部已读

- **WHEN** 用户通过 `POST /api/v1/notifications/read-all` 标记全部已读
- **THEN** 系统 SHALL 将当前用户所有未读通知的 `read_at` 设置为当前时间
- **AND** 返回 `{"updated_count": <number>}` 格式的响应

#### Scenario: 批量标记已读

- **WHEN** 用户通过 `POST /api/v1/notifications/batch-read` 并提供 `ids` 数组
- **THEN** 系统 SHALL 将指定 ID 的通知标记为已读
- **AND** 仅处理属于当前用户的通知

---

### Requirement: 通知资源跳转

通知 SHALL 支持跳转到关联资源。

#### Scenario: 跳转到 Issue

- **WHEN** 通知的 `resource_type` 为 "issue"
- **THEN** 前端 SHALL 支持点击通知跳转到 Issue 详情页

#### Scenario: 包含资源基本信息

- **WHEN** 返回通知列表
- **THEN** 通知 SHALL 包含关联资源的 `resource_id` 和 `resource_type`
- **AND** 如有关联 Issue，SHALL 包含 `issue_number`（如 "ENG-123"）和 `issue_title`
