# Token 刷新功能规格

## ADDED Requirements

### Requirement: 用户可以刷新 Access Token

系统 SHALL 允许用户使用有效的 refresh token 获取新的 access token 和 refresh token。

#### Scenario: 成功刷新 Token

- **GIVEN** 用户持有有效的 refresh_token
- **WHEN** 用户提交刷新请求，包含：
  - refresh_token: "valid_refresh_token"
- **THEN** 返回 HTTP 200 OK 状态码
- **AND** 返回新的 access_token
- **AND** 返回新的 refresh_token
- **AND** 新的 access_token 有效期为 15 分钟
- **AND** 新的 refresh_token 有效期为 7 天

#### Scenario: 旧的 Refresh Token 失效

- **GIVEN** 用户使用 refresh_token 刷新成功
- **WHEN** 用户再次使用同一个 refresh_token 尝试刷新
- **THEN** 返回 HTTP 401 Unauthorized 状态码
- **AND** 返回错误信息 "Token 已失效"

### Requirement: Token 刷新验证

系统 MUST 验证 refresh token 的有效性。

#### Scenario: Refresh Token 已过期

- **GIVEN** 用户的 refresh_token 已超过 7 天有效期
- **WHEN** 用户提交刷新请求
- **THEN** 返回 HTTP 401 Unauthorized 状态码
- **AND** 返回错误信息 "Token 已过期"

#### Scenario: Refresh Token 格式无效

- **WHEN** 用户提交刷新请求，包含：
  - refresh_token: "invalid_token_format"
- **THEN** 返回 HTTP 401 Unauthorized 状态码
- **AND** 返回错误信息 "Token 格式无效"

#### Scenario: 使用 Access Token 刷新

- **WHEN** 用户提交刷新请求，使用 access_token 作为 refresh_token
- **THEN** 返回 HTTP 401 Unauthorized 状态码
- **AND** 返回错误信息 "无效的 Token 类型"

### Requirement: Token 黑名单机制

系统 SHALL 支持将 refresh token 加入黑名单以使其失效。

#### Scenario: Token 在黑名单中

- **GIVEN** refresh_token 已被加入 Redis 黑名单
- **WHEN** 用户使用该 refresh_token 尝试刷新
- **THEN** 返回 HTTP 401 Unauthorized 状态码
- **AND** 返回错误信息 "Token 已失效"

#### Scenario: 用户登出后 Token 失效

- **GIVEN** 用户执行了登出操作
- **WHEN** 用户使用登出前的 refresh_token 尝试刷新
- **THEN** 返回 HTTP 401 Unauthorized 状态码
- **AND** 返回错误信息 "Token 已失效"

### Requirement: 刷新时 Token 轮换

系统 SHOULD 在每次刷新时生成新的 refresh token，使旧的失效。

#### Scenario: Token 轮换机制

- **GIVEN** 用户持有 refresh_token_A
- **WHEN** 用户使用 refresh_token_A 刷新成功
- **THEN** 系统返回 refresh_token_B
- **AND** refresh_token_A 被加入黑名单
- **AND** 后续使用 refresh_token_A 将被拒绝
