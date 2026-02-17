# Labels Spec

## 1. Data Model

### 1.1 `labels` Table

| Field         | Type      | Description                                                |
| :------------ | :-------- | :--------------------------------------------------------- |
| `id`          | UUID      | 主键，UUIDv4                                               |
| `team_id`     | UUID      | 关联 Teams 表的外键，允许为 NULL (NULL = Workspace 级标签) |
| `name`        | String    | 标签名称，范围内唯一 (Team 前或 Workspace 级)              |
| `color`       | String    | Hex 颜色代码                                               |
| `description` | String    | 可选描述                                                   |
| `created_at`  | Timestamp | 创建时间                                                   |
| `updated_at`  | Timestamp | 更新时间                                                   |

## 2. Requirements

### 2.1 Label Creation (POST)

**Feature**: 创建新标签（支持 Team 级或 Workspace 级）。

-   **R1**: 系统 **MUST** 支持创建 **Team** 级别的标签（`team_id` 非空）。
-   **R2**: 系统 **MUST** 支持创建 **Workspace** 级别的标签（`team_id` 为空）。
-   **R3**: `name` **MUST** 在其作用域内唯一。

**Scenario: Workspace 标签创建**
-   **GIVEN** 一个具有 Workspace 权限的 Admin 用户
-   **WHEN** 该 Admin 请求 `POST /api/v1/labels`，Payload 为 `{ name: "Bug", color: "#FF0000" }`
-   **THEN** 系统创建一个 team_id 为空的全局标签。

**Scenario: Team 标签创建**
-   **GIVEN** 一个已存在的 Team `T1`
-   **WHEN** 用户请求 `POST /api/v1/teams/T1/labels`
-   **THEN** 系统创建一个 team_id 为 `T1` 的团队标签。

### 2.2 Label Retrieval (GET)

**Feature**: 获取标签列表。

-   **R1**: `GET /api/v1/teams/:team_id/labels` **MUST** 返回以下两类标签：
    1.  该 Team 的标签（`team_id` 匹配）。
    2.  Workspace 全局标签（`team_id` 为空）。
-   **R2**: `GET /api/v1/labels` (无 Team 参数) **MUST** 仅返回 Workspace 全局标签。

**Scenario: 团队标签列表**
-   **GIVEN** Team `T1` 拥有标签 "Feature"
-   **AND** Workspace 拥有全局标签 "High Priority"
-   **WHEN** 用户请求 `GET /api/v1/teams/T1/labels`
-   **THEN** 返回列表中包含 "Feature" 和 "High Priority"。

### 2.3 Label Deletion (DELETE)

**Feature**: 删除标签。

-   **R1**: 删除标签 **MUST NOT** 级联删除相关的 Issue，但 **SHOULD** 解除 Issue 的该标签关联。
-   **R2**: 仅 Workspace Admin 可删除全局标签。

**Scenario: 删除标签**
-   **GIVEN** 存在标签 "Deprecated"
-   **WHEN** Admin 删除该标签
-   **THEN** 数据库中该标签被移除
-   **AND** 原本标记为 "Deprecated" 的 Issue 不再显示该标签。
