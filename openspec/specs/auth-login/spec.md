# 用户登录功能规格

## ADDED Requirements

### Requirement: 用户可以通过邮箱和密码登录

系统 SHALL 允许已注册用户通过邮箱和密码进行身份验证，验证成功后签发 JWT access token 和 refresh token。

#### Scenario: 成功登录

- **GIVEN** 系统中存在用户，邮箱为 "user@example.com"，密码为 "SecurePass123!"
- **WHEN** 用户提交登录请求，包含：
  - email: "user@example.com"
  - password: "SecurePass123!"
- **THEN** 返回 HTTP 200 OK 状态码
- **AND** 返回用户信息（不包含密码哈希）
- **AND** 返回有效的 access_token
- **AND** 返回有效的 refresh_token

#### Scenario: 邮箱不存在

- **GIVEN** 系统中不存在邮箱 "nonexistent@example.com"
- **WHEN** 用户提交登录请求，包含：
  - email: "nonexistent@example.com"
  - password: "AnyPassword123!"
- **THEN** 返回 HTTP 401 Unauthorized 状态码
- **AND** 返回错误信息 "邮箱或密码错误"
- **AND** 不返回任何 token

#### Scenario: 密码错误

- **GIVEN** 系统中存在用户，邮箱为 "user@example.com"
- **WHEN** 用户提交登录请求，包含：
  - email: "user@example.com"
  - password: "WrongPassword123!"
- **THEN** 返回 HTTP 401 Unauthorized 状态码
- **AND** 返回错误信息 "邮箱或密码错误"
- **AND** 不返回任何 token

### Requirement: 登录返回 Token 格式

系统 MUST 在成功登录后返回符合 JWT 标准的 access token 和 refresh token。

#### Scenario: Access Token 包含必要声明

- **WHEN** 用户成功登录
- **THEN** access_token 的 payload 包含：
  - sub: 用户 UUID
  - email: 用户邮箱
  - role: 用户角色
  - exp: 过期时间戳（15分钟后）
  - iat: 签发时间戳
  - jti: Token 唯一标识符

#### Scenario: Refresh Token 包含必要声明

- **WHEN** 用户成功登录
- **THEN** refresh_token 的 payload 包含：
  - sub: 用户 UUID
  - type: "refresh"
  - exp: 过期时间戳（7天后）
  - iat: 签发时间戳
  - jti: Token 唯一标识符

### Requirement: 缺少必填字段

系统 MUST 验证登录请求包含所有必填字段。

#### Scenario: 缺少邮箱

- **WHEN** 用户提交登录请求，缺少 email 字段
- **THEN** 返回 HTTP 400 Bad Request 状态码
- **AND** 返回错误信息 "邮箱为必填项"

#### Scenario: 缺少密码

- **WHEN** 用户提交登录请求，缺少 password 字段
- **THEN** 返回 HTTP 400 Bad Request 状态码
- **AND** 返回错误信息 "密码为必填项"
