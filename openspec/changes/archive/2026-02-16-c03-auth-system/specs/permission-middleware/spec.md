# 权限中间件规格

## ADDED Requirements

### Requirement: 系统支持基于角色的权限控制

系统 SHALL 支持基于用户角色（Admin / Member）的访问控制。

#### Scenario: Admin 用户访问管理端点

- **GIVEN** 用户角色为 "admin" 或 "global_admin"
- **WHEN** 用户访问需要 Admin 权限的端点
- **THEN** 请求被允许继续处理

#### Scenario: Member 用户访问管理端点被拒绝

- **GIVEN** 用户角色为 "member"
- **WHEN** 用户访问需要 Admin 权限的端点
- **THEN** 返回 HTTP 403 Forbidden 状态码
- **AND** 返回错误信息 "权限不足"

### Requirement: 权限中间件可配置

系统 SHALL 提供可配置的权限中间件，支持指定所需角色列表。

#### Scenario: RequireRole 中间件

- **GIVEN** 路由配置了 `RequireRole("admin", "global_admin")` 中间件
- **WHEN** 角色为 "member" 的用户访问该路由
- **THEN** 返回 HTTP 403 Forbidden 状态码

#### Scenario: RequireAdmin 中间件

- **GIVEN** 路由配置了 `RequireAdmin()` 中间件
- **WHEN** 角色为 "admin" 的用户访问该路由
- **THEN** 请求被允许继续处理

#### Scenario: RequireGlobalAdmin 中间件

- **GIVEN** 路由配置了 `RequireGlobalAdmin()` 中间件
- **WHEN** 角色为 "admin"（非 global_admin）的用户访问该路由
- **THEN** 返回 HTTP 403 Forbidden 状态码
- **AND** 只有 "global_admin" 角色可以通过

### Requirement: 权限检查必须在认证之后

系统 MUST 确保权限中间件在认证中间件之后执行。

#### Scenario: 未认证用户访问权限保护端点

- **GIVEN** 路由同时配置了 Auth 和 Permission 中间件
- **WHEN** 未认证用户访问该路由
- **THEN** 返回 HTTP 401 Unauthorized 状态码（由 Auth 中间件返回）
- **AND** Permission 中间件不执行

### Requirement: 角色层级

系统 SHALL 定义角色层级：global_admin > admin > member > guest。

#### Scenario: GlobalAdmin 拥有所有权限

- **GIVEN** 用户角色为 "global_admin"
- **WHEN** 用户访问需要 "admin" 权限的端点
- **THEN** 请求被允许继续处理

#### Scenario: Admin 不能访问 GlobalAdmin 端点

- **GIVEN** 用户角色为 "admin"
- **WHEN** 用户访问需要 "global_admin" 权限的端点
- **THEN** 返回 HTTP 403 Forbidden 状态码

### Requirement: 从上下文获取用户角色

系统 SHALL 提供便捷方法从请求上下文获取用户角色。

#### Scenario: GetCurrentUserRole 函数

- **GIVEN** 请求已通过认证中间件
- **WHEN** 调用 `GetCurrentUserRole(c)` 函数
- **THEN** 返回当前用户的 Role 枚举值

#### Scenario: IsAdmin 函数

- **GIVEN** 请求已通过认证中间件，用户角色为 "admin"
- **WHEN** 调用 `IsAdmin(c)` 函数
- **THEN** 返回 true
