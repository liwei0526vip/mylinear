# Workflow States Spec

## 1. Data Model

### 1.1 `workflow_states` Table

| Field         | Type      | Description                                                        |
| :------------ | :-------- | :----------------------------------------------------------------- |
| `id`          | UUID      | 主键，UUIDv4                                                       |
| `team_id`     | UUID      | 关联 Teams 表的外键，非空                                          |
| `name`        | String    | 状态名称，Team 内唯一                                              |
| `type`        | String    | 枚举值: `backlog`, `unstarted`, `started`, `completed`, `canceled` |
| `color`       | String    | Hex 颜色代码 (例如 `#F2C94C`)                                      |
| `position`    | Float     | 排序字段，默认间隔 1000.0                                          |
| `description` | String    | 可选描述                                                           |
| `created_at`  | Timestamp | 创建时间                                                           |
| `updated_at`  | Timestamp | 更新时间                                                           |

## 2. Requirements

### 2.1 State Creation (POST)

**Feature**: 为团队创建新的工作流状态。

-   **R1**: 系统 **MUST** 校验 `name` 在指定 `team_id` 内的唯一性。
-   **R2**: `type` **MUST** 是 5 个允许的枚举值之一。
-   **R3**: 如果未提供 `position`，系统 **SHOULD** 计算新值，使其排在同类型状态组的末尾。

**Scenario: 成功创建**
-   **GIVEN** 存在 ID 为 `T1` 的 Team
-   **WHEN** 管理员请求 `POST /api/v1/teams/T1/workflow-states`，Payload 为 `{ name: "Review", type: "started", color: "#FF0000" }`
-   **THEN** 系统创建该状态
-   **AND** 返回 201 Created 及状态对象。

**Scenario: 名称重复**
-   **GIVEN** Team `T1` 已有名为 "Todo" 的状态
-   **WHEN** 管理员请求在 `T1` 中再创建一个 "Todo" 状态
-   **THEN** 系统阻止创建
-   **AND** 返回 409 Conflict。

### 2.2 Default State Initialization

**Feature**: 团队创建时自动预置默认状态。

-   **R1**: 当新 Team 创建时（扩展 C04 逻辑），系统 **MUST** 自动创建以下状态：
    1.  `Backlog` (type: `backlog`)
    2.  `Todo` (type: `unstarted`)
    3.  `In Progress` (type: `started`)
    4.  `Done` (type: `completed`)
    5.  `Canceled` (type: `canceled`)

**Scenario: 自动预置**
-   **GIVEN** 存在 Workspace
-   **WHEN** 管理员创建一个新 Team "Engineering"
-   **THEN** 系统创建该 Team
-   **AND** 系统为该 Team 创建 5 个默认工作流状态。

### 2.3 State Deletion (DELETE)

**Feature**: 删除工作流状态。

-   **R1**: 如果该状态是团队中该 `type` 的最后一个状态，系统 **MUST NOT** 允许删除。
-   **R2**: 如果有 Issue 当前处于该状态，系统 **MUST NOT** 允许删除 (Phase 1 简化处理：直接阻止；Phase 2 可增加迁移逻辑)。

**Scenario: 阻止删除类型下的唯一状态**
-   **GIVEN** Team `T1` 中 `unstarted` 类型下只有一个 "Todo" 状态
-   **WHEN** 管理员尝试删除 "Todo"
-   **THEN** 系统拒绝请求
-   **AND** 返回 400 Bad Request，错误信息 "Cannot delete the last state of type 'unstarted'"。

**Scenario: 阻止删除有关联 Issue 的状态**
-   **GIVEN** Team `T1` 的 "In Progress" 状态下有 5 个活跃 Issue
-   **WHEN** 管理员尝试删除 "In Progress"
-   **THEN** 系统拒绝请求
-   **AND** 返回 400 Bad Request，错误信息 "Cannot delete state with assigned issues"。

### 2.4 State Reordering (PUT)

**Feature**: 更新状态顺序。

-   **R1**: 系统 **MUST** 允许通过更新 `position` 字段来重排状态。
-   **R2**: API **SHOULD** 接受新的 `position` 浮点数值。

**Scenario: 重排状态**
-   **GIVEN** State A (pos 1000) 和 State B (pos 2000)
-   **WHEN** 管理员更新 State B 的 position 为 500
-   **THEN** State B 在列表中排在 State A 之前。
