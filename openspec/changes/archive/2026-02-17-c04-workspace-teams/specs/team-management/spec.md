# Team 管理规格

## ADDED Requirements

### Requirement: 获取团队列表

系统 SHALL 允许已认证用户获取其有权限访问的团队列表。

#### Scenario: 获取工作区内团队列表

- **GIVEN** 用户已认证且属于某个工作区
- **WHEN** 用户请求 `GET /api/v1/teams?workspace_id=:wid`
- **THEN** 返回 HTTP 200 状态码
- **AND** 返回用户可访问的团队列表（公开团队 + 用户所属的私有团队）

#### Scenario: 分页查询团队

- **GIVEN** 用户已认证
- **WHEN** 用户请求 `GET /api/v1/teams?workspace_id=:wid&page=1&page_size=20`
- **THEN** 返回 HTTP 200 状态码
- **AND** 返回分页结果（items, total, page, page_size）

### Requirement: 创建团队

系统 SHALL 允许 Admin 角色用户创建新团队。

#### Scenario: Admin 创建团队

- **GIVEN** 用户角色为 "admin" 或 "global_admin"
- **WHEN** 用户请求 `POST /api/v1/teams` 包含 `{ "name": "Engineering", "key": "ENG", "workspace_id": "..." }`
- **THEN** 返回 HTTP 201 Created 状态码
- **AND** 返回创建的团队详情
- **AND** 创建者自动成为团队的 Team Owner

#### Scenario: 团队标识符 Key 重复

- **GIVEN** 工作区内已存在 Key 为 "ENG" 的团队
- **WHEN** 用户请求 `POST /api/v1/teams` 包含 `{ "name": "New Team", "key": "ENG", "workspace_id": "..." }`
- **THEN** 返回 HTTP 409 Conflict 状态码
- **AND** 返回错误信息 "团队标识符已存在"

#### Scenario: 团队标识符格式校验

- **GIVEN** 用户角色为 "admin"
- **WHEN** 用户请求 `POST /api/v1/teams` 包含 `{ "name": "Team", "key": "eng-lower", "workspace_id": "..." }`
- **THEN** 返回 HTTP 400 Bad Request 状态码
- **AND** 返回错误信息 "团队标识符必须为大写字母和数字，长度 2-10 位"

#### Scenario: Member 用户无法创建团队

- **GIVEN** 用户角色为 "member"
- **WHEN** 用户请求 `POST /api/v1/teams`
- **THEN** 返回 HTTP 403 Forbidden 状态码

### Requirement: 获取团队详情

系统 SHALL 允许已认证用户获取其有权限访问的团队详情。

#### Scenario: 成功获取团队详情

- **GIVEN** 用户是团队成员或团队为公开团队
- **WHEN** 用户请求 `GET /api/v1/teams/:id`
- **THEN** 返回 HTTP 200 状态码
- **AND** 返回团队详情（id, workspace_id, name, key, icon_url, timezone, is_private, created_at, updated_at）

#### Scenario: 无权访问私有团队

- **GIVEN** 用户不是私有团队成员
- **WHEN** 用户请求 `GET /api/v1/teams/:id`（私有团队）
- **THEN** 返回 HTTP 403 Forbidden 状态码
- **AND** 返回错误信息 "无权访问该团队"

#### Scenario: 团队不存在

- **GIVEN** 用户已认证
- **WHEN** 用户请求 `GET /api/v1/teams/:id` 且团队不存在
- **THEN** 返回 HTTP 404 Not Found 状态码

### Requirement: 更新团队

系统 SHALL 允许 Team Owner 或 Admin 用户更新团队信息。

#### Scenario: Team Owner 更新团队名称

- **GIVEN** 用户是该团队的 Team Owner
- **WHEN** 用户请求 `PUT /api/v1/teams/:id` 包含 `{ "name": "New Name" }`
- **THEN** 返回 HTTP 200 状态码
- **AND** 团队名称被更新

#### Scenario: Team Owner 更新团队 Key

- **GIVEN** 用户是该团队的 Team Owner
- **WHEN** 用户请求 `PUT /api/v1/teams/:id` 包含 `{ "key": "NEW" }`
- **THEN** 返回 HTTP 200 状态码
- **AND** 团队 Key 被更新
- **AND** 后续创建的 Issue ID 使用新 Key（如 NEW-124）

#### Scenario: Admin 更新任意团队

- **GIVEN** 用户角色为 "admin" 或 "global_admin"
- **WHEN** 用户请求 `PUT /api/v1/teams/:id`
- **THEN** 返回 HTTP 200 状态码
- **AND** 团队信息被更新

#### Scenario: 普通成员无法更新团队

- **GIVEN** 用户是该团队的普通成员（非 Owner）
- **WHEN** 用户请求 `PUT /api/v1/teams/:id`
- **THEN** 返回 HTTP 403 Forbidden 状态码

### Requirement: 删除团队

系统 SHALL 允许 Team Owner 或 Admin 用户删除团队。

#### Scenario: Team Owner 删除团队

- **GIVEN** 用户是该团队的 Team Owner
- **WHEN** 用户请求 `DELETE /api/v1/teams/:id`
- **THEN** 返回 HTTP 204 No Content 状态码
- **AND** 团队被软删除（标记为已删除）

#### Scenario: 删除有关联 Issue 的团队

- **GIVEN** 团队下存在 Issue
- **WHEN** 用户请求 `DELETE /api/v1/teams/:id`
- **THEN** 返回 HTTP 400 Bad Request 状态码
- **AND** 返回错误信息 "团队下存在 Issue，无法删除"

### Requirement: 团队成员列表

系统 SHALL 允许已认证用户获取团队成员列表。

#### Scenario: 获取团队成员列表

- **GIVEN** 用户是团队成员或团队为公开团队
- **WHEN** 用户请求 `GET /api/v1/teams/:id/members`
- **THEN** 返回 HTTP 200 状态码
- **AND** 返回成员列表（user_id, role, joined_at, user 详情）

### Requirement: 添加团队成员

系统 SHALL 允许 Team Owner 或 Admin 用户添加团队成员。

#### Scenario: Team Owner 添加成员

- **GIVEN** 用户是该团队的 Team Owner
- **WHEN** 用户请求 `POST /api/v1/teams/:id/members` 包含 `{ "user_id": "...", "role": "member" }`
- **THEN** 返回 HTTP 201 Created 状态码
- **AND** 用户被添加为团队成员

#### Scenario: 添加已存在的成员

- **GIVEN** 用户已是该团队成员
- **WHEN** 用户请求 `POST /api/v1/teams/:id/members` 包含该用户
- **THEN** 返回 HTTP 409 Conflict 状态码
- **AND** 返回错误信息 "用户已是团队成员"

#### Scenario: 添加不存在的用户

- **GIVEN** user_id 对应的用户不存在
- **WHEN** 用户请求 `POST /api/v1/teams/:id/members`
- **THEN** 返回 HTTP 404 Not Found 状态码
- **AND** 返回错误信息 "用户不存在"

### Requirement: 移除团队成员

系统 SHALL 允许 Team Owner 或 Admin 用户移除团队成员。

#### Scenario: Team Owner 移除成员

- **GIVEN** 用户是该团队的 Team Owner
- **WHEN** 用户请求 `DELETE /api/v1/teams/:id/members/:uid`
- **THEN** 返回 HTTP 204 No Content 状态码
- **AND** 用户被从团队移除

#### Scenario: 移除最后一个 Team Owner

- **GIVEN** 团队只有一个 Team Owner
- **WHEN** 用户请求移除该 Team Owner
- **THEN** 返回 HTTP 400 Bad Request 状态码
- **AND** 返回错误信息 "团队必须至少有一个 Owner"

#### Scenario: 移除不存在的成员

- **GIVEN** 用户不是该团队成员
- **WHEN** 用户请求 `DELETE /api/v1/teams/:id/members/:uid`
- **THEN** 返回 HTTP 404 Not Found 状态码

### Requirement: 更新成员角色

系统 SHALL 允许 Team Owner 或 Admin 用户更新团队成员角色。

#### Scenario: Team Owner 提升成员为 Owner

- **GIVEN** 用户是该团队的 Team Owner
- **WHEN** 用户请求 `PUT /api/v1/teams/:id/members/:uid` 包含 `{ "role": "owner" }`
- **THEN** 返回 HTTP 200 状态码
- **AND** 成员角色被更新为 Owner

#### Scenario: 自己降级自己

- **GIVEN** 用户是该团队唯一的 Team Owner
- **WHEN** 用户请求将自己降级为 Member
- **THEN** 返回 HTTP 400 Bad Request 状态码
- **AND** 返回错误信息 "团队必须至少有一个 Owner"
