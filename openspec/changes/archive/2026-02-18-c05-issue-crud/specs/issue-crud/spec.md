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
