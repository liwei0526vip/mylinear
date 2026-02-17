# Workspace 管理规格

## ADDED Requirements

### Requirement: 获取当前工作区

系统 SHALL 允许已认证用户获取其所属工作区的详细信息。

#### Scenario: 成功获取工作区

- **GIVEN** 用户已认证且属于某个工作区
- **WHEN** 用户请求 `GET /api/v1/workspaces/:id`
- **THEN** 返回 HTTP 200 状态码
- **AND** 返回工作区详情（id, name, slug, logo_url, settings, created_at, updated_at）

#### Scenario: 用户不属于该工作区

- **GIVEN** 用户已认证但不属于请求的工作区
- **WHEN** 用户请求 `GET /api/v1/workspaces/:id`
- **THEN** 返回 HTTP 403 Forbidden 状态码
- **AND** 返回错误信息 "无权访问该工作区"

#### Scenario: 工作区不存在

- **GIVEN** 用户已认证
- **WHEN** 用户请求 `GET /api/v1/workspaces/:id` 且工作区不存在
- **THEN** 返回 HTTP 404 Not Found 状态码

### Requirement: 更新工作区

系统 SHALL 允许 Admin 角色用户更新工作区基本信息。

#### Scenario: Admin 更新工作区名称

- **GIVEN** 用户角色为 "admin" 或 "global_admin"
- **WHEN** 用户请求 `PUT /api/v1/workspaces/:id` 包含 `{ "name": "新名称" }`
- **THEN** 返回 HTTP 200 状态码
- **AND** 工作区名称被更新
- **AND** 返回更新后的工作区详情

#### Scenario: Admin 更新工作区 Logo

- **GIVEN** 用户角色为 "admin" 或 "global_admin"
- **WHEN** 用户请求 `PUT /api/v1/workspaces/:id` 包含 `{ "logo_url": "https://..." }`
- **THEN** 返回 HTTP 200 状态码
- **AND** 工作区 Logo URL 被更新

#### Scenario: Member 用户无法更新工作区

- **GIVEN** 用户角色为 "member"
- **WHEN** 用户请求 `PUT /api/v1/workspaces/:id`
- **THEN** 返回 HTTP 403 Forbidden 状态码
- **AND** 返回错误信息 "权限不足"

#### Scenario: 更新不存在的字段被忽略

- **GIVEN** 用户角色为 "admin"
- **WHEN** 用户请求 `PUT /api/v1/workspaces/:id` 包含 `{ "name": "新名称", "unknown_field": "value" }`
- **THEN** 返回 HTTP 200 状态码
- **AND** 只更新 name 字段
- **AND** unknown_field 被忽略

### Requirement: 工作区 Slug 唯一性

系统 SHALL 确保工作区 Slug 在全局唯一。

#### Scenario: Slug 已存在

- **GIVEN** 已存在 Slug 为 "my-workspace" 的工作区
- **WHEN** 用户请求更新另一个工作区的 Slug 为 "my-workspace"
- **THEN** 返回 HTTP 409 Conflict 状态码
- **AND** 返回错误信息 "Slug 已被使用"

### Requirement: 获取工作区统计信息

系统 SHALL 允许 Admin 用户获取工作区统计信息。

#### Scenario: 获取工作区统计

- **GIVEN** 用户角色为 "admin" 或 "global_admin"
- **WHEN** 用户请求 `GET /api/v1/workspaces/:id/stats`
- **THEN** 返回 HTTP 200 状态码
- **AND** 返回统计信息（teams_count, members_count, issues_count）
