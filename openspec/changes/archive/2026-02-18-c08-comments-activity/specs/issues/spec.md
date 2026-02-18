## MODIFIED Requirements

### Requirement: Issue 更新

系统 SHALL 支持更新 Issue 的各个字段，并自动记录变更活动。

#### Scenario: 更新基础字段

- **WHEN** 用户通过 `PUT /api/v1/issues/:id` 更新 Issue
- **THEN** 系统 SHALL 支持更新：`title`、`description`、`priority`、`dueDate`、`estimate`

#### Scenario: 更新状态

- **WHEN** 用户更新 Issue 的 `statusId`
- **THEN** 系统 SHALL 验证目标状态属于同一团队
- **AND** 系统 SHALL 记录状态变更历史到 `issue_status_history` 表
- **AND** 系统 SHALL 创建 `status_changed` 类型的活动记录
- **AND** 若目标状态类型为 `completed`，系统 SHALL 设置 `completedAt`
- **AND** 若目标状态类型为 `cancelled`，系统 SHALL 设置 `cancelledAt`

#### Scenario: 更新负责人

- **WHEN** 用户更新 Issue 的 `assigneeId`
- **THEN** 系统 SHALL 验证被分配用户是团队成员
- **AND** 新负责人 SHALL 自动订阅该 Issue
- **AND** 系统 SHALL 创建 `assignee_changed` 类型的活动记录

#### Scenario: 更新标签

- **WHEN** 用户更新 Issue 的 `labelIds`
- **THEN** 系统 SHALL 替换原有标签关联（全量更新）
- **AND** 系统 SHALL 创建 `labels_changed` 类型的活动记录

#### Scenario: 更新项目和迭代

- **WHEN** 用户更新 Issue 的 `projectId` 或 `cycleId`
- **THEN** 系统 SHALL 验证目标资源存在且可访问
- **AND** 系统 SHALL 创建对应的活动记录

#### Scenario: 更新标题时记录活动

- **WHEN** 用户更新 Issue 的 `title`
- **THEN** 系统 SHALL 创建 `title_changed` 类型的活动记录

#### Scenario: 更新描述时记录活动

- **WHEN** 用户更新 Issue 的 `description`
- **THEN** 系统 SHALL 创建 `description_changed` 类型的活动记录

#### Scenario: 更新优先级时记录活动

- **WHEN** 用户更新 Issue 的 `priority`
- **THEN** 系统 SHALL 创建 `priority_changed` 类型的活动记录

#### Scenario: 更新截止日期时记录活动

- **WHEN** 用户更新 Issue 的 `dueDate`
- **THEN** 系统 SHALL 创建 `due_date_changed` 类型的活动记录

---

## ADDED Requirements

### Requirement: Issue 创建时记录活动

系统 SHALL 在 Issue 创建时自动记录活动。

#### Scenario: 创建 Issue 时记录

- **WHEN** 用户创建 Issue
- **THEN** 系统 SHALL 创建 `issue_created` 类型的活动记录
- **AND** `actor_id` SHALL 为创建者
- **AND** `issue_id` SHALL 为新创建的 Issue
