## ADDED Requirements

### Requirement: Project 模型

Project 模型 MUST 定义项目实体的完整结构，支持团队级项目管理。

#### Scenario: Project 模型字段

- **WHEN** 定义 `Project` 结构体
- **THEN** MUST 包含字段：`ID`（UUID）、`TeamID`（UUID, FK）、`Name`（string）、`Description`（*string, Markdown）、`Status`（string）、`LeadID`（*UUID, FK）、`StartDate`（*time.Time）、`TargetDate`（*time.Time）、`CreatedByID`（UUID, FK）、`CreatedAt`（time.Time）、`UpdatedAt`（time.Time）、`CompletedAt`（*time.Time）、`CancelledAt`（*time.Time）

#### Scenario: Project 状态常量

- **WHEN** 定义 `ProjectStatus` 常量
- **THEN** MUST 支持 `ProjectStatusPlanned`（planned）、`ProjectStatusInProgress`（in_progress）、`ProjectStatusPaused`（paused）、`ProjectStatusCompleted`（completed）、`ProjectStatusCancelled`（cancelled）

#### Scenario: Project 关联关系

- **WHEN** 定义 `Project` 关联
- **THEN** MUST 支持 `Team`、`Lead`、`CreatedBy` 的 belongs-to 关系

---

### Requirement: Project 创建

系统 SHALL 支持在指定团队下创建项目。

#### Scenario: 创建 Project 基础字段

- **WHEN** 用户通过 `POST /api/v1/teams/:teamId/projects` 创建项目
- **THEN** 系统 SHALL 创建 Project 记录，包含 `name`（必填，1-100 字符）、`description`（可选，Markdown 格式）、`leadId`（可选，负责人 ID）、`startDate`（可选）、`targetDate`（可选）
- **AND** 系统 SHALL 返回完整的 Project 对象（包含生成的 ID）

#### Scenario: Project 默认状态

- **WHEN** 创建项目时未指定 `status`
- **THEN** 系统 SHALL 自动设置状态为 `planned`

#### Scenario: 验证负责人

- **WHEN** 创建项目时指定 `leadId`
- **THEN** 系统 SHALL 验证该用户是团队成员
- **AND** 若验证失败 SHALL 返回 400 Bad Request

#### Scenario: 验证日期逻辑

- **WHEN** 创建项目时同时指定 `startDate` 和 `targetDate`
- **THEN** 系统 SHALL 验证 `targetDate` >= `startDate`
- **AND** 若验证失败 SHALL 返回 400 Bad Request

---

### Requirement: Project 查询

系统 SHALL 支持多种方式查询项目列表和详情。

#### Scenario: 获取团队项目列表

- **WHEN** 用户通过 `GET /api/v1/teams/:teamId/projects` 查询项目列表
- **THEN** 系统 SHALL 返回该团队下所有项目（按 `updatedAt` 倒序）
- **AND** 系统 SHALL 支持按 `status` 过滤

#### Scenario: 获取单个项目详情

- **WHEN** 用户通过 `GET /api/v1/projects/:id` 请求项目详情
- **THEN** 系统 SHALL 返回项目完整信息，包含关联的 `Team`、`Lead`、关联 Issue 统计

#### Scenario: 分页查询

- **WHEN** 用户查询项目列表时指定 `page` 和 `pageSize`
- **THEN** 系统 SHALL 返回分页结果，包含 `total`（总数）、`items`（当前页数据）
- **AND** 默认 `pageSize` SHALL 为 50，最大 100

---

### Requirement: Project 更新

系统 SHALL 支持更新项目的各个字段。

#### Scenario: 更新基础字段

- **WHEN** 用户通过 `PUT /api/v1/projects/:id` 更新项目
- **THEN** 系统 SHALL 支持更新：`name`、`description`、`leadId`、`startDate`、`targetDate`

#### Scenario: 更新状态

- **WHEN** 用户更新项目的 `status`
- **THEN** 系统 SHALL 验证状态值为有效枚举（planned/in_progress/paused/completed/cancelled）
- **AND** 若目标状态为 `completed`，系统 SHALL 设置 `completedAt`
- **AND** 若目标状态为 `cancelled`，系统 SHALL 设置 `cancelledAt`

#### Scenario: 更新负责人

- **WHEN** 用户更新项目的 `leadId`
- **THEN** 系统 SHALL 验证新负责人是团队成员

---

### Requirement: Project 删除

系统 SHALL 支持软删除项目，保留历史记录。

#### Scenario: 软删除 Project

- **WHEN** 用户通过 `DELETE /api/v1/projects/:id` 删除项目
- **THEN** 系统 SHALL 设置 `deletedAt` 字段（软删除）
- **AND** 项目不再出现在默认查询结果中
- **AND** 关联的 Issue 的 `projectId` 字段 SHALL 保持不变（Issue 保留历史关联）

#### Scenario: 恢复已删除 Project

- **WHEN** 管理员通过 `POST /api/v1/projects/:id/restore` 恢复项目
- **THEN** 系统 SHALL 清除 `deletedAt` 字段

---

### Requirement: Project 进度统计

系统 SHALL 支持自动计算项目进度。

#### Scenario: 获取项目进度

- **WHEN** 用户通过 `GET /api/v1/projects/:id/progress` 请求项目进度
- **THEN** 系统 SHALL 返回进度统计，包含：
  - `totalIssues`（关联 Issue 总数）
  - `completedIssues`（已完成 Issue 数）
  - `cancelledIssues`（已取消 Issue 数）
  - `progressPercentage`（完成百分比，0-100）

#### Scenario: 进度计算规则

- **WHEN** 系统计算项目进度
- **THEN** 系统 SHALL 基于 Issue 的工作流状态类型（`completed`/`cancelled`）统计
- **AND** `progressPercentage` SHALL = `completedIssues / (totalIssues - cancelledIssues) * 100`
- **AND** 若 `totalIssues - cancelledIssues = 0`，`progressPercentage` SHALL 为 0

---

### Requirement: Project 关联 Issue

系统 SHALL 支持查询项目关联的 Issue 列表。

#### Scenario: 获取项目 Issue 列表

- **WHEN** 用户通过 `GET /api/v1/projects/:id/issues` 请求项目 Issue 列表
- **THEN** 系统 SHALL 返回该项目关联的所有 Issue
- **AND** 系统 SHALL 支持按 `status`、`priority`、`assigneeId` 过滤
- **AND** 系统 SHALL 支持分页

#### Scenario: Issue 列表排序

- **WHEN** 用户查询项目 Issue 列表
- **THEN** 系统 SHALL 默认按 `position` 升序排序
- **AND** 系统 SHALL 支持按 `priority`、`createdAt`、`updatedAt` 排序

---

### Requirement: 权限控制

Project 操作 SHALL 受权限控制。

#### Scenario: 团队成员创建 Project

- **WHEN** 团队成员在团队下创建项目
- **THEN** 系统 SHALL 允许操作

#### Scenario: 非成员访问私有团队 Project

- **WHEN** 非团队成员尝试访问私有团队的项目
- **THEN** 系统 SHALL 返回 403 Forbidden

#### Scenario: 项目更新权限

- **WHEN** 用户尝试更新项目
- **THEN** 系统 SHALL 验证用户是团队成员
- **AND** 仅 Admin/Member/ProjectLead 可更新项目

#### Scenario: 项目删除权限

- **WHEN** 用户尝试删除项目
- **THEN** 系统 SHALL 验证用户是团队 Admin
