## ADDED Requirements

### Requirement: Activity 模型

Activity 模型 MUST 定义活动记录的完整结构，使用 JSONB 存储活动详情。

#### Scenario: Activity 模型字段

- **WHEN** 定义 `Activity` 结构体
- **THEN** MUST 包含字段：`ID`（UUID）、`IssueID`（FK）、`Type`（ActivityType）、`ActorID`（FK）、`Payload`（JSONB）、`CreatedAt`

#### Scenario: ActivityType 枚举

- **WHEN** 定义 `ActivityType` 类型
- **THEN** MUST 支持：`issue_created`、`title_changed`、`description_changed`、`status_changed`、`priority_changed`、`assignee_changed`、`due_date_changed`、`project_changed`、`labels_changed`、`comment_added`

#### Scenario: Activity 关联关系

- **WHEN** 定义 `Activity` 关联
- **THEN** MUST 支持 `Issue`、`Actor`（User）的 belongs-to 关系

#### Scenario: Activity 表名

- **WHEN** 定义 `Activity` 的 `TableName()` 方法
- **THEN** MUST 返回 `activities`

---

### Requirement: 活动记录创建

系统 SHALL 在 Issue 变更时自动创建活动记录。

#### Scenario: Issue 创建时记录

- **WHEN** 用户创建 Issue
- **THEN** 系统 SHALL 创建 `issue_created` 类型的活动记录
- **AND** `actor_id` SHALL 为创建者

#### Scenario: 标题变更时记录

- **WHEN** 用户更新 Issue 的 `title`
- **THEN** 系统 SHALL 创建 `title_changed` 类型的活动记录
- **AND** `payload` SHALL 包含 `old_value` 和 `new_value`

#### Scenario: 状态变更时记录

- **WHEN** 用户更新 Issue 的 `statusId`
- **THEN** 系统 SHALL 创建 `status_changed` 类型的活动记录
- **AND** `payload` SHALL 包含 `old_status` 和 `new_status` 对象

#### Scenario: 负责人变更时记录

- **WHEN** 用户更新 Issue 的 `assigneeId`
- **THEN** 系统 SHALL 创建 `assignee_changed` 类型的活动记录
- **AND** `payload` SHALL 包含 `old_assignee` 和 `new_assignee` 对象（可为 null）

#### Scenario: 评论添加时记录

- **WHEN** 用户在 Issue 下创建评论
- **THEN** 系统 SHALL 创建 `comment_added` 类型的活动记录
- **AND** `payload` SHALL 包含 `comment_id` 和 `comment_preview`（前 100 字符）

#### Scenario: 批量变更时单条记录

- **WHEN** 用户一次更新 Issue 的多个字段
- **THEN** 系统 SHALL 为每个变更字段创建独立的活动记录

---

### Requirement: 活动查询

系统 SHALL 支持查询 Issue 的活动时间线。

#### Scenario: 获取活动列表

- **WHEN** 用户通过 `GET /api/v1/issues/:issueId/activities` 查询活动
- **THEN** 系统 SHALL 返回该 Issue 的所有活动记录
- **AND** 默认按 `created_at` 倒序排列（最新在前）

#### Scenario: 分页查询

- **WHEN** 用户指定 `page` 和 `page_size` 参数
- **THEN** 系统 SHALL 返回分页结果，包含 `total`（总数）、`items`（当前页）
- **AND** 默认 `page_size` SHALL 为 50，最大 100

#### Scenario: 按类型过滤

- **WHEN** 用户指定 `types` 参数（如 `types=status_changed,comment_added`）
- **THEN** 系统 SHALL 仅返回指定类型的活动

#### Scenario: 包含 actor 信息

- **WHEN** 返回活动列表
- **THEN** 每条活动 SHALL 包含 `actor` 对象（`id`、`name`、`username`、`avatar_url`）

---

### Requirement: 状态历史记录

系统 SHALL 在 Issue 状态变更时写入 `issue_status_history` 表。

#### Scenario: 状态变更写入历史

- **WHEN** 用户更新 Issue 的 `statusId`
- **THEN** 系统 SHALL 写入 `issue_status_history` 表
- **AND** 记录 `from_status_id`（可为 null，初始状态）、`to_status_id`、`changed_by_id`、`changed_at`

#### Scenario: 初始状态无 from_status_id

- **WHEN** Issue 创建时的初始状态
- **THEN** `from_status_id` SHALL 为 null
- **AND** `to_status_id` SHALL 为默认状态

#### Scenario: 历史记录查询

- **WHEN** 需要计算 Time in Status
- **THEN** 系统 SHALL 可通过 `issue_status_history` 表查询状态停留时间

---

### Requirement: Activity Payload 结构

不同活动类型的 Payload SHALL 遵循预定义结构。

#### Scenario: title_changed payload

- **WHEN** 活动类型为 `title_changed`
- **THEN** `payload` SHALL 为 `{"old_value": "原标题", "new_value": "新标题"}`

#### Scenario: status_changed payload

- **WHEN** 活动类型为 `status_changed`
- **THEN** `payload` SHALL 为 `{"old_status": {"id": "...", "name": "...", "color": "..."}, "new_status": {...}}`

#### Scenario: assignee_changed payload

- **WHEN** 活动类型为 `assignee_changed`
- **THEN** `payload` SHALL 为 `{"old_assignee": {...} | null, "new_assignee": {...} | null}`

#### Scenario: comment_added payload

- **WHEN** 活动类型为 `comment_added`
- **THEN** `payload` SHALL 为 `{"comment_id": "...", "comment_preview": "评论前100字符..."}`

#### Scenario: labels_changed payload

- **WHEN** 活动类型为 `labels_changed`
- **THEN** `payload` SHALL 为 `{"added": [...], "removed": [...]}`

---

### Requirement: 权限控制

活动查询 SHALL 受权限控制。

#### Scenario: 团队成员可查询

- **WHEN** 团队成员查询 Issue 的活动
- **THEN** 系统 SHALL 允许操作

#### Scenario: 非成员无法查询私有团队

- **WHEN** 非团队成员查询私有团队 Issue 的活动
- **THEN** 系统 SHALL 返回 403 Forbidden
