# 认证中间件规格

## ADDED Requirements

### Requirement: 受保护端点需要 JWT 认证

系统 MUST 对需要认证的端点验证 JWT access token，验证通过后将用户信息注入请求上下文。

#### Scenario: 有效 Token 访问受保护端点

- **GIVEN** 用户持有有效的 access_token
- **WHEN** 用户访问受保护端点，请求头包含：
  - Authorization: "Bearer valid_access_token"
- **THEN** 请求被允许继续处理
- **AND** 请求上下文中包含用户信息（user_id, email, role）

#### Scenario: 缺少 Authorization 头

- **WHEN** 用户访问受保护端点，请求头不包含 Authorization
- **THEN** 返回 HTTP 401 Unauthorized 状态码
- **AND** 返回错误信息 "未提供认证令牌"

#### Scenario: Authorization 头格式错误

- **WHEN** 用户访问受保护端点，请求头包含：
  - Authorization: "InvalidFormat token"
- **THEN** 返回 HTTP 401 Unauthorized 状态码
- **AND** 返回错误信息 "认证令牌格式无效"

#### Scenario: Token 格式无效

- **WHEN** 用户访问受保护端点，请求头包含：
  - Authorization: "Bearer invalid_token"
- **THEN** 返回 HTTP 401 Unauthorized 状态码
- **AND** 返回错误信息 "认证令牌无效"

### Requirement: Token 过期处理

系统 MUST 拒绝已过期的 access token。

#### Scenario: Access Token 已过期

- **GIVEN** 用户的 access_token 已超过 15 分钟有效期
- **WHEN** 用户使用该 token 访问受保护端点
- **THEN** 返回 HTTP 401 Unauthorized 状态码
- **AND** 返回错误信息 "认证令牌已过期"
- **AND** 响应头包含 `X-Token-Expired: true`

### Requirement: Token 签名验证

系统 MUST 验证 JWT 签名以防止伪造。

#### Scenario: Token 签名无效

- **GIVEN** access_token 被篡改或签名不正确
- **WHEN** 用户使用该 token 访问受保护端点
- **THEN** 返回 HTTP 401 Unauthorized 状态码
- **AND** 返回错误信息 "认证令牌签名无效"

### Requirement: 用户状态验证

系统 SHOULD 在认证时验证用户账户状态。

#### Scenario: 用户不存在

- **GIVEN** access_token 中的用户 ID 在数据库中不存在
- **WHEN** 用户使用该 token 访问受保护端点
- **THEN** 返回 HTTP 401 Unauthorized 状态码
- **AND** 返回错误信息 "用户不存在"

### Requirement: 从上下文获取用户信息

系统 SHALL 在认证成功后通过上下文提供用户信息访问方法。

#### Scenario: Handler 获取当前用户

- **GIVEN** 请求已通过认证中间件
- **WHEN** Handler 调用 `GetCurrentUser(c)` 函数
- **THEN** 返回当前登录用户的 User 结构体
- **AND** 包含用户 ID、邮箱、角色等信息

#### Scenario: 未认证请求获取用户

- **GIVEN** 请求未通过认证中间件
- **WHEN** Handler 调用 `GetCurrentUser(c)` 函数
- **THEN** 返回 nil
