# 用户个人资料管理规格

## ADDED Requirements

### Requirement: 用户可以查看自己的资料

系统 SHALL 允许已认证用户获取自己的完整资料信息。

#### Scenario: 获取当前用户资料

- **GIVEN** 用户已认证
- **WHEN** 用户请求 GET /api/v1/users/me
- **THEN** 返回 HTTP 200 OK 状态码
- **AND** 返回用户信息，包含：
  - id: 用户 UUID
  - email: 用户邮箱
  - username: 用户名
  - name: 全名
  - avatar_url: 头像 URL（可为 null）
  - role: 角色
  - created_at: 创建时间
  - updated_at: 更新时间
- **AND** 不返回 password_hash 字段

### Requirement: 用户可以更新自己的资料

系统 SHALL 允许已认证用户更新自己的姓名、用户名和邮箱。

#### Scenario: 更新用户名和全名

- **GIVEN** 用户已认证
- **WHEN** 用户请求 PATCH /api/v1/users/me，包含：
  - name: "New Name"
  - username: "newusername"
- **THEN** 返回 HTTP 200 OK 状态码
- **AND** 返回更新后的用户信息
- **AND** 数据库中用户记录已更新

#### Scenario: 更新邮箱

- **GIVEN** 用户已认证，当前邮箱为 "old@example.com"
- **WHEN** 用户请求 PATCH /api/v1/users/me，包含：
  - email: "new@example.com"
- **THEN** 返回 HTTP 200 OK 状态码
- **AND** 返回更新后的用户信息
- **AND** 新邮箱在系统中唯一

#### Scenario: 更新邮箱已被使用

- **GIVEN** 系统中已存在邮箱 "existing@example.com"
- **WHEN** 用户请求 PATCH /api/v1/users/me，包含：
  - email: "existing@example.com"
- **THEN** 返回 HTTP 409 Conflict 状态码
- **AND** 返回错误信息 "邮箱已被使用"

#### Scenario: 更新用户名已被使用

- **GIVEN** 系统中已存在用户名 "existinguser"
- **WHEN** 用户请求 PATCH /api/v1/users/me，包含：
  - username: "existinguser"
- **THEN** 返回 HTTP 409 Conflict 状态码
- **AND** 返回错误信息 "用户名已被使用"

### Requirement: 用户可以上传头像

系统 SHALL 允许已认证用户上传头像图片到 MinIO 存储。

#### Scenario: 成功上传头像

- **GIVEN** 用户已认证
- **WHEN** 用户请求 POST /api/v1/users/me/avatar，上传图片文件（JPEG, < 5MB）
- **THEN** 返回 HTTP 200 OK 状态码
- **AND** 返回新的 avatar_url
- **AND** 图片存储在 MinIO 的 avatars bucket 中
- **AND** 用户记录的 avatar_url 字段已更新

#### Scenario: 上传文件过大

- **WHEN** 用户上传超过 5MB 的图片文件
- **THEN** 返回 HTTP 400 Bad Request 状态码
- **AND** 返回错误信息 "文件大小不能超过 5MB"

#### Scenario: 上传不支持的文件类型

- **WHEN** 用户上传非图片文件（如 .txt, .pdf）
- **THEN** 返回 HTTP 400 Bad Request 状态码
- **AND** 返回错误信息 "只支持 JPG、PNG、GIF、WebP 格式的图片"

#### Scenario: 上传文件损坏

- **WHEN** 用户上传的文件声称是图片但实际不是有效图片
- **THEN** 返回 HTTP 400 Bad Request 状态码
- **AND** 返回错误信息 "无效的图片文件"

### Requirement: 未认证用户无法访问资料端点

系统 MUST 拒绝未认证用户访问资料相关端点。

#### Scenario: 未认证访问资料

- **WHEN** 未认证用户请求 GET /api/v1/users/me
- **THEN** 返回 HTTP 401 Unauthorized 状态码

#### Scenario: 未认证更新资料

- **WHEN** 未认证用户请求 PATCH /api/v1/users/me
- **THEN** 返回 HTTP 401 Unauthorized 状态码

#### Scenario: 未认证上传头像

- **WHEN** 未认证用户请求 POST /api/v1/users/me/avatar
- **THEN** 返回 HTTP 401 Unauthorized 状态码
