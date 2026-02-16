# 用户注册功能规格

## ADDED Requirements

### Requirement: 用户可以通过邮箱和密码注册

系统 SHALL 允许用户通过提供邮箱、用户名、密码和姓名进行注册。系统 MUST 验证输入的有效性，使用 bcrypt 哈希存储密码，并在注册成功后自动登录用户。

#### Scenario: 成功注册新用户

- **GIVEN** 系统中不存在邮箱 "newuser@example.com"
- **WHEN** 用户提交注册请求，包含：
  - email: "newuser@example.com"
  - username: "johndoe"
  - password: "SecurePass123!"
  - name: "John Doe"
- **THEN** 系统创建新用户记录
- **AND** 密码使用 bcrypt 哈希存储（不存储明文）
- **AND** 返回 HTTP 201 状态码
- **AND** 返回用户信息和访问令牌

#### Scenario: 邮箱已被注册

- **GIVEN** 系统中已存在邮箱 "existing@example.com"
- **WHEN** 用户提交注册请求，包含：
  - email: "existing@example.com"
  - username: "newuser"
  - password: "SecurePass123!"
  - name: "New User"
- **THEN** 返回 HTTP 409 Conflict 状态码
- **AND** 返回错误信息 "邮箱已被注册"

#### Scenario: 用户名已被使用

- **GIVEN** 系统中已存在用户名 "existinguser"
- **WHEN** 用户提交注册请求，包含：
  - email: "new@example.com"
  - username: "existinguser"
  - password: "SecurePass123!"
  - name: "New User"
- **THEN** 返回 HTTP 409 Conflict 状态码
- **AND** 返回错误信息 "用户名已被使用"

### Requirement: 系统验证注册输入格式

系统 MUST 验证所有注册字段符合格式要求。

#### Scenario: 邮箱格式无效

- **WHEN** 用户提交注册请求，包含：
  - email: "invalid-email"
  - username: "testuser"
  - password: "SecurePass123!"
  - name: "Test User"
- **THEN** 返回 HTTP 400 Bad Request 状态码
- **AND** 返回错误信息 "邮箱格式无效"

#### Scenario: 密码强度不足

- **WHEN** 用户提交注册请求，包含：
  - email: "test@example.com"
  - username: "testuser"
  - password: "weak"
  - name: "Test User"
- **THEN** 返回 HTTP 400 Bad Request 状态码
- **AND** 返回错误信息 "密码必须至少8个字符，包含大小写字母和数字"

#### Scenario: 用户名格式无效

- **WHEN** 用户提交注册请求，包含：
  - email: "test@example.com"
  - username: "invalid user!"
  - password: "SecurePass123!"
  - name: "Test User"
- **THEN** 返回 HTTP 400 Bad Request 状态码
- **AND** 返回错误信息 "用户名只能包含字母、数字和下划线"

#### Scenario: 缺少必填字段

- **WHEN** 用户提交注册请求，缺少 email 字段
- **THEN** 返回 HTTP 400 Bad Request 状态码
- **AND** 返回错误信息 "邮箱为必填项"

### Requirement: 用户名格式限制

系统 SHALL 只允许用户名包含字母、数字、下划线和连字符，长度为 3-50 个字符。

#### Scenario: 用户名过短

- **WHEN** 用户提交注册请求，包含 username: "ab"
- **THEN** 返回 HTTP 400 Bad Request 状态码
- **AND** 返回错误信息 "用户名长度必须为3-50个字符"

#### Scenario: 用户名包含特殊字符

- **WHEN** 用户提交注册请求，包含 username: "test@user"
- **THEN** 返回 HTTP 400 Bad Request 状态码
- **AND** 返回错误信息 "用户名只能包含字母、数字、下划线和连字符"

### Requirement: 新用户默认角色

系统 SHALL 为新注册用户分配默认角色 "member"。

#### Scenario: 新用户角色设置

- **WHEN** 用户成功注册
- **THEN** 用户记录的 role 字段为 "member"
- **AND** 用户 NOT 具有管理员权限
