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
