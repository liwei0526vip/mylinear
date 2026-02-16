## ADDED Requirements

### Requirement: GORM 模型文件结构

所有 GORM 模型 MUST 定义在 `server/internal/model/` 目录下，每个表对应一个 Go 文件。

#### Scenario: 模型文件命名

- **WHEN** 创建 GORM 模型文件
- **THEN** 文件名 MUST 使用 snake_case（如 `workspace.go`、`team.go`、`issue.go`）

#### Scenario: 模型文件组织

- **WHEN** 创建多个相关模型
- **THEN** 可以在同一文件中定义紧密关联的模型（如 `team.go` 中同时定义 `Team` 和 `TeamMember`）

---

### Requirement: GORM 模型基础结构

所有模型 MUST 遵循统一的基础结构，包含 ID、时间戳等标准字段。

#### Scenario: 基础模型字段

- **WHEN** 定义 GORM 模型
- **THEN** MUST 包含 `ID`（`uuid.UUID`）、`CreatedAt`（`time.Time`）、`UpdatedAt`（`time.Time`）字段

#### Scenario: UUID 类型

- **WHEN** 定义 UUID 字段
- **THEN** MUST 使用 `github.com/google/uuid` 包的 `uuid.UUID` 类型

#### Scenario: GORM 标签规范

- **WHEN** 定义模型字段
- **THEN** MUST 使用 GORM 结构标签指定数据库约束（`gorm:"column:xxx;type:uuid;primary_key"`）

---

### Requirement: Workspace 模型

Workspace 模型 MUST 定义工作区实体的完整结构。

#### Scenario: Workspace 模型字段

- **WHEN** 定义 `Workspace` 结构体
- **THEN** MUST 包含字段：`ID`（UUID）、`Name`（string）、`Slug`（string, unique）、`LogoURL`（*string）、`Settings`（JSONB/map）、`CreatedAt`、`UpdatedAt`

#### Scenario: Workspace 关联关系

- **WHEN** 定义 `Workspace` 关联
- **THEN** MUST 定义 `Teams []Team` 和 `Users []User` 的 has-many 关系

---

### Requirement: Team 模型

Team 模型 MUST 定义团队实体的完整结构，支持嵌套团队。

#### Scenario: Team 模型字段

- **WHEN** 定义 `Team` 结构体
- **THEN** MUST 包含字段：`ID`、`WorkspaceID`（FK）、`ParentID`（*UUID, nullable）、`Name`、`Key`（string, unique）、`IconURL`、`Timezone`、`IsPrivate`（bool）、`CycleSettings`（JSONB）、`WorkflowSettings`（JSONB）、`CreatedAt`、`UpdatedAt`

#### Scenario: Team 自引用关联

- **WHEN** 定义 `Team` 自引用
- **THEN** MUST 支持 `Parent *Team` 和 `Children []Team` 的嵌套关系

---

### Requirement: User 模型

User 模型 MUST 定义用户实体的完整结构，支持多角色。

#### Scenario: User 模型字段

- **WHEN** 定义 `User` 结构体
- **THEN** MUST 包含字段：`ID`、`WorkspaceID`（FK）、`Email`（string, unique）、`Name`、`Username`（string, unique）、`AvatarURL`、`PasswordHash`、`Role`（enum）、`Settings`（JSONB）、`CreatedAt`、`UpdatedAt`

#### Scenario: User 角色枚举

- **WHEN** 定义 `Role` 类型
- **THEN** MUST 支持 `RoleGlobalAdmin`、`RoleAdmin`、`RoleMember`、`RoleGuest` 四种角色

---

### Requirement: Issue 模型

Issue 模型 MUST 定义 Issue 实体的完整结构，支持丰富的关联关系。

#### Scenario: Issue 模型字段

- **WHEN** 定义 `Issue` 结构体
- **THEN** MUST 包含字段：`ID`、`TeamID`（FK）、`Number`（int, 团队内序号）、`Title`、`Description`（*string）、`StatusID`（FK）、`Priority`（int）、`AssigneeID`（*UUID）、`ProjectID`（*UUID）、`MilestoneID`（*UUID）、`CycleID`（*UUID）、`ParentID`（*UUID）、`Estimate`（*int）、`DueDate`（*time.Time）、`SLADueAt`（*time.Time）、`Labels`（pq.StringArray）、`CreatedByID`（FK）、`CreatedAt`、`UpdatedAt`、`CompletedAt`、`CancelledAt`

#### Scenario: Issue 优先级常量

- **WHEN** 定义 `Priority` 常量
- **THEN** MUST 支持 `PriorityNone`（0）、`PriorityUrgent`（1）、`PriorityHigh`（2）、`PriorityMedium`（3）、`PriorityLow`（4）

#### Scenario: Issue 关联关系

- **WHEN** 定义 `Issue` 关联
- **THEN** MUST 支持 `Team`、`Status`、`Assignee`、`Project`、`Milestone`、`Cycle`、`Parent`、`CreatedBy` 的 belongs-to 关系，以及 `Children []Issue`、`Comments []Comment`、`Attachments []Attachment` 的 has-many 关系

---

### Requirement: WorkflowState 模型

WorkflowState 模型 MUST 定义工作流状态实体，支持 5 种状态类型。

#### Scenario: WorkflowState 模型字段

- **WHEN** 定义 `WorkflowState` 结构体
- **THEN** MUST 包含字段：`ID`、`TeamID`（FK）、`Name`、`Type`（enum）、`Color`、`Position`（float64）、`IsDefault`（bool）、`CreatedAt`

#### Scenario: WorkflowState 类型枚举

- **WHEN** 定义 `StateType` 类型
- **THEN** MUST 支持 `StateTypeBacklog`、`StateTypeUnstarted`、`StateTypeStarted`、`StateTypeCompleted`、`StateTypeCancelled` 五种类型

---

### Requirement: JSONB 字段处理

JSONB 类型字段 MUST 使用适当的数据结构存储和查询。

#### Scenario: Settings 字段类型

- **WHEN** 定义 `Settings` JSONB 字段
- **THEN** SHOULD 使用 `datatypes.JSON`（gorm.io/datatypes）或 `map[string]interface{}` 类型

#### Scenario: JSONB 查询支持

- **WHEN** 查询 JSONB 字段
- **THEN** GORM 模型 MUST 支持通过 `db.Where("settings->>'key' = ?", value)` 语法查询

---

### Requirement: 数组字段处理

PostgreSQL 数组类型字段 MUST 使用 `pq` 库处理。

#### Scenario: UUID 数组字段

- **WHEN** 定义 UUID 数组字段（如 `Labels`）
- **THEN** MUST 使用 `github.com/lib/pq` 的 `pq.StringArray` 类型（存储为字符串数组，业务层转换）

#### Scenario: 数组查询支持

- **WHEN** 查询数组字段
- **THEN** GORM 模型 MUST 支持通过 `db.Where("labels && ?", pq.Array([]string{labelID}))` 语法查询（重叠）

---

### Requirement: 软删除支持

需要保留历史记录的模型 SHOULD 支持软删除。

#### Scenario: 软删除字段

- **WHEN** 定义需要软删除的模型
- **THEN** MUST 包含 `DeletedAt gorm.DeletedAt` 字段

#### Scenario: 软删除查询

- **WHEN** 查询软删除模型
- **THEN** GORM 自动排除已删除记录，可通过 `Unscoped()` 包含已删除记录

---

### Requirement: 模型表名约定

所有模型 MUST 显式定义表名，确保命名一致性。

#### Scenario: TableName 方法

- **WHEN** 定义 GORM 模型
- **THEN** MUST 实现 `TableName() string` 方法，返回复数形式的蛇形表名（如 `workspaces`、`team_members`）

#### Scenario: 默认表名

- **WHEN** 未实现 `TableName()` 方法
- **THEN** GORM 使用结构体名称的蛇形复数形式（推荐显式定义以避免歧义）
