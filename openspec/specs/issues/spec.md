## MODIFIED Requirements

### Requirement: Issue 模型

Issue 模型 MUST 定义 Issue 实体的完整结构，支持丰富的关联关系。

#### Scenario: Issue 模型字段

- **WHEN** 定义 `Issue` 结构体
- **THEN** MUST 包含字段：`ID`、`TeamID`（FK）、`Number`（int, 团队内序号）、`Title`、`Description`（*string）、`StatusID`（FK）、`Priority`（int）、`AssigneeID`（*UUID）、`ProjectID`（*UUID）、`MilestoneID`（*UUID）、`CycleID`（*UUID）、`ParentID`（*UUID）、`Estimate`（*int）、`DueDate`（*time.Time）、`SLADueAt`（*time.Time）、`Labels`（pq.StringArray）、`Position`（float64, 排序用）、`CreatedByID`（FK）、`CreatedAt`、`UpdatedAt`、`CompletedAt`、`CancelledAt`

#### Scenario: Issue 优先级常量

- **WHEN** 定义 `Priority` 常量
- **THEN** MUST 支持 `PriorityNone`（0）、`PriorityUrgent`（1）、`PriorityHigh`（2）、`PriorityMedium`（3）、`PriorityLow`（4）

#### Scenario: Issue 关联关系

- **WHEN** 定义 `Issue` 关联
- **THEN** MUST 支持 `Team`、`Status`、`Assignee`、`Project`、`Milestone`、`Cycle`、`Parent`、`CreatedBy` 的 belongs-to 关系，以及 `Children []Issue`、`Comments []Comment`、`Attachments []Attachment`、`Subscribers []IssueSubscription` 的 has-many 关系

---

## ADDED Requirements

### Requirement: IssueSubscription 模型

IssueSubscription 模型 MUST 定义 Issue 订阅关系。

#### Scenario: IssueSubscription 模型字段

- **WHEN** 定义 `IssueSubscription` 结构体
- **THEN** MUST 包含字段：`IssueID`（UUID, 复合主键之一）、`UserID`（UUID, 复合主键之一）、`CreatedAt`（time.Time）

#### Scenario: IssueSubscription 关联关系

- **WHEN** 定义 `IssueSubscription` 关联
- **THEN** MUST 支持 `Issue` 和 `User` 的 belongs-to 关系

#### Scenario: IssueSubscription 表名

- **WHEN** 定义 `IssueSubscription` 的 `TableName()` 方法
- **THEN** MUST 返回 `issue_subscriptions`

#### Scenario: IssueSubscription 复合主键

- **WHEN** 创建 `issue_subscriptions` 表
- **THEN** MUST 使用 (`issue_id`, `user_id`) 作为复合主键

---

## ADDED Requirements

### Requirement: Issue 创建

系统 SHALL 支持在指定团队下创建 Issue，自动生成唯一标识符。

#### Scenario: 创建 Issue 基础字段

- **WHEN** 用户通过 `POST /api/v1/teams/:teamId/issues` 创建 Issue
- **THEN** 系统 SHALL 创建 Issue 记录，包含 `title`（必填）、`description`（可选，Markdown 格式）
- **AND** 系统 SHALL 返回完整的 Issue 对象（包含生成的 ID 和标识符）

#### Scenario: Issue 标识符自动生成

- **WHEN** 在团队下创建新 Issue
- **THEN** 系统 SHALL 自动生成 `{Team.Key}-{Number}` 格式的标识符（如 `ENG-123`）
- **AND** `Number` SHALL 在团队内自增且唯一

#### Scenario: 创建时关联工作流状态

- **WHEN** 创建 Issue 时未指定 `statusId`
- **THEN** 系统 SHALL 自动关联团队的默认工作流状态（`is_default = true`）

#### Scenario: 创建时关联可选字段

- **WHEN** 用户创建 Issue 时指定可选字段
- **THEN** 系统 SHALL 支持关联：`priority`（优先级）、`assigneeId`（负责人）、`labelIds`（标签数组）、`projectId`（项目）、`dueDate`（截止日期）、`estimate`（预估工作量）

#### Scenario: 自动订阅创建者

- **WHEN** 用户创建 Issue
- **THEN** 系统 SHALL 自动将创建者添加为订阅者

---

### Requirement: Issue 查询

系统 SHALL 支持多种方式查询 Issue 列表和详情。

#### Scenario: 获取单个 Issue 详情

- **WHEN** 用户通过 `GET /api/v1/issues/:id` 请求 Issue 详情
- **THEN** 系统 SHALL 返回 Issue 完整信息，包含关联的 `Team`、`Status`、`Assignee`、`Labels`、`Project` 等关联数据

#### Scenario: 列表查询基础过滤

- **WHEN** 用户通过 `GET /api/v1/teams/:teamId/issues` 查询 Issue 列表
- **THEN** 系统 SHALL 支持按以下条件过滤：
  - `statusId` - 工作流状态
  - `priority` - 优先级
  - `assigneeId` - 负责人
  - `labelIds` - 标签（支持多选）
  - `projectId` - 项目
  - `cycleId` - 迭代

#### Scenario: 列表查询排序

- **WHEN** 用户查询 Issue 列表
- **THEN** 系统 SHALL 支持按以下字段排序：
  - `priority` - 优先级
  - `createdAt` - 创建时间
  - `updatedAt` - 更新时间
  - `dueDate` - 截止日期
  - `position` - 自定义排序

#### Scenario: 分页查询

- **WHEN** 用户查询 Issue 列表时指定 `page` 和 `pageSize`
- **THEN** 系统 SHALL 返回分页结果，包含 `total`（总数）、`items`（当前页数据）
- **AND** 默认 `pageSize` SHALL 为 50，最大 100

---

### Requirement: Issue 更新

系统 SHALL 支持更新 Issue 的各个字段。

#### Scenario: 更新基础字段

- **WHEN** 用户通过 `PUT /api/v1/issues/:id` 更新 Issue
- **THEN** 系统 SHALL 支持更新：`title`、`description`、`priority`、`dueDate`、`estimate`

#### Scenario: 更新状态

- **WHEN** 用户更新 Issue 的 `statusId`
- **THEN** 系统 SHALL 验证目标状态属于同一团队
- **AND** 系统 SHALL 记录状态变更历史到 `issue_status_history` 表
- **AND** 若目标状态类型为 `completed`，系统 SHALL 设置 `completedAt`
- **AND** 若目标状态类型为 `cancelled`，系统 SHALL 设置 `cancelledAt`

#### Scenario: 更新负责人

- **WHEN** 用户更新 Issue 的 `assigneeId`
- **THEN** 系统 SHALL 验证被分配用户是团队成员
- **AND** 新负责人 SHALL 自动订阅该 Issue

#### Scenario: 更新标签

- **WHEN** 用户更新 Issue 的 `labelIds`
- **THEN** 系统 SHALL 替换原有标签关联（全量更新）

#### Scenario: 更新项目和迭代

- **WHEN** 用户更新 Issue 的 `projectId` 或 `cycleId`
- **THEN** 系统 SHALL 验证目标资源存在且可访问

---

### Requirement: Issue 删除与归档

系统 SHALL 支持软删除 Issue，保留历史记录。

#### Scenario: 软删除 Issue

- **WHEN** 用户通过 `DELETE /api/v1/issues/:id` 删除 Issue
- **THEN** 系统 SHALL 设置 `deletedAt` 字段（软删除）
- **AND** Issue 不再出现在默认查询结果中

#### Scenario: 恢复已删除 Issue

- **WHEN** 管理员通过 `POST /api/v1/issues/:id/restore` 恢复 Issue
- **THEN** 系统 SHALL 清除 `deletedAt` 字段

---

### Requirement: Issue 拖拽排序

系统 SHALL 支持通过 position 字段实现自定义排序。

#### Scenario: 更新排序位置

- **WHEN** 用户通过 `PUT /api/v1/issues/:id/position` 更新 Issue 位置
- **AND** 请求体包含 `position`（新位置）或 `afterId`（排在某 Issue 之后）
- **THEN** 系统 SHALL 更新 Issue 的 `position` 字段
- **AND** 系统 SHALL 自动重算受影响的其他 Issue 的 position 值（避免冲突）

#### Scenario: 跨状态拖拽

- **WHEN** 用户将 Issue 拖拽到不同状态列
- **THEN** 系统 SHALL 同时更新 `statusId` 和 `position`
- **AND** 触发状态变更的完整流程（记录历史、设置时间戳）

---

### Requirement: Issue 批量查询

系统 SHALL 支持跨团队的 Issue 查询。

#### Scenario: 查询用户负责的 Issue

- **WHEN** 用户通过 `GET /api/v1/issues?assigneeId=me` 查询
- **THEN** 系统 SHALL 返回当前用户负责的所有 Issue（跨团队）

#### Scenario: 查询用户创建的 Issue

- **WHEN** 用户通过 `GET /api/v1/issues?createdById=me` 查询
- **THEN** 系统 SHALL 返回当前用户创建的所有 Issue（跨团队）

#### Scenario: 查询用户订阅的 Issue

- **WHEN** 用户通过 `GET /api/v1/issues?subscribed=me` 查询
- **THEN** 系统 SHALL 返回当前用户订阅的所有 Issue

---

### Requirement: 权限控制

Issue 操作 SHALL 受权限控制。

#### Scenario: 团队成员创建 Issue

- **WHEN** 团队成员在团队下创建 Issue
- **THEN** 系统 SHALL 允许操作

#### Scenario: 非成员访问私有团队 Issue

- **WHEN** 非团队成员尝试访问私有团队的 Issue
- **THEN** 系统 SHALL 返回 403 Forbidden

#### Scenario: Guest 用户限制

- **WHEN** Guest 用户尝试删除 Issue
- **THEN** 系统 SHALL 返回 403 Forbidden（仅 Admin/Member 可删除）

---

## ADDED Requirements

### Requirement: Issue 订阅机制

系统 SHALL 支持 Issue 订阅功能，用户可追踪关注的 Issue 变更。

#### Scenario: 订阅 Issue

- **WHEN** 用户通过 `POST /api/v1/issues/:id/subscribe` 订阅 Issue
- **THEN** 系统 SHALL 在 `issue_subscriptions` 表创建订阅记录
- **AND** 用户后续 SHALL 收到该 Issue 的变更通知

#### Scenario: 取消订阅 Issue

- **WHEN** 用户通过 `DELETE /api/v1/issues/:id/subscribe` 取消订阅
- **THEN** 系统 SHALL 删除对应的订阅记录
- **AND** 用户不再收到该 Issue 的变更通知（除非重新订阅）

#### Scenario: 重复订阅幂等

- **WHEN** 用户重复订阅同一 Issue
- **THEN** 系统 SHALL 返回成功（幂等操作，不创建重复记录）

#### Scenario: 查询订阅者列表

- **WHEN** 用户通过 `GET /api/v1/issues/:id/subscribers` 查询订阅者
- **THEN** 系统 SHALL 返回该 Issue 的所有订阅者列表（包含用户基本信息）

---

### Requirement: 自动订阅规则

系统 SHALL 根据用户行为自动管理订阅关系。

#### Scenario: 创建者自动订阅

- **WHEN** 用户创建 Issue
- **THEN** 系统 SHALL 自动将创建者添加为订阅者

#### Scenario: 负责人自动订阅

- **WHEN** Issue 被分配给用户
- **THEN** 系统 SHALL 自动将该用户添加为订阅者（若未订阅）

#### Scenario: 评论者自动订阅

- **WHEN** 用户在 Issue 下发表评论
- **THEN** 系统 SHALL 自动将该用户添加为订阅者（若未订阅）

#### Scenario: @mention 自动订阅

- **WHEN** Issue 描述或评论中 @mention 某用户
- **THEN** 系统 SHALL 自动将被提及的用户添加为订阅者

---

### Requirement: 订阅与通知联动

订阅状态 SHALL 影响通知推送行为。

#### Scenario: 订阅者接收通知

- **WHEN** Issue 发生变更（状态、负责人、评论等）
- **THEN** 系统 SHALL 向所有订阅者推送通知

#### Scenario: 取消订阅后停止通知

- **WHEN** 用户取消订阅 Issue
- **THEN** 系统 SHALL 停止向该用户推送该 Issue 的后续通知

#### Scenario: 负责人变更时通知

- **WHEN** Issue 负责人变更
- **THEN** 系统 SHALL 通知原负责人（若仍订阅）和新负责人

---

### Requirement: 订阅数据模型

订阅关系 SHALL 通过独立数据表管理。

#### Scenario: 订阅表结构

- **WHEN** 系统初始化
- **THEN** `issue_subscriptions` 表 SHALL 包含字段：`issue_id`（FK）、`user_id`（FK）、`created_at`
- **AND** 复合主键 SHALL 为 (`issue_id`, `user_id`)

#### Scenario: 级联删除

- **WHEN** Issue 被删除
- **THEN** 相关订阅记录 SHALL 自动删除（级联）

#### Scenario: 用户删除时清理

- **WHEN** 用户被删除
- **THEN** 该用户的所有订阅记录 SHALL 自动删除（级联）
