# C03 — 用户认证与权限系统

## Why

用户认证是所有业务功能的前置条件。没有认证系统，用户无法登录、无法关联操作记录、无法实现权限隔离。这是 MVP 阶段的核心基础设施，必须在 Workspace、Teams、Issues 等业务模块之前完成。

此 change 依赖 C02（数据库模型），User 表已在 C02 中定义，本 change 实现完整的认证流程和权限控制。

## What Changes

### 后端 API

- 新增用户注册 API（`POST /api/v1/auth/register`）：邮箱 + 密码注册，密码使用 bcrypt 哈希
- 新增用户登录 API（`POST /api/v1/auth/login`）：返回 JWT access token 和 refresh token
- 新增 Token 刷新 API（`POST /api/v1/auth/refresh`）：使用 refresh token 换取新的 access token
- 新增用户 Profile API（`GET/PATCH /api/v1/users/me`）：获取和更新当前用户信息
- 新增头像上传 API（`POST /api/v1/users/me/avatar`）：上传头像到 MinIO
- 新增认证中间件：验证 JWT 并注入用户信息到请求上下文
- 新增权限中间件：检查用户角色（Admin / Member）

### 前端 UI

- 新增登录页面：邮箱 + 密码登录表单
- 新增注册页面：邮箱 + 用户名 + 密码注册表单
- 新增 Profile 设置页：头像、邮箱、全名、用户名管理
- 实现 JWT 存储（localStorage）和自动刷新机制
- 实现路由守卫：未登录用户重定向到登录页

### 配置与基础设施

- 新增 JWT 相关配置项（密钥、过期时间）
- 新增 MinIO 头像存储 bucket 初始化

## Capabilities

### New Capabilities

- `auth-registration`：用户注册功能（邮箱 + 密码，bcrypt 哈希，用户名设置）
- `auth-login`：用户登录功能（JWT 签发，access token + refresh token）
- `auth-token-refresh`：Token 刷新功能（refresh token 换取新 access token）
- `auth-middleware`：认证中间件（JWT 验证，用户上下文注入）
- `user-profile`：用户个人资料管理（头像上传、邮箱、全名、用户名）
- `permission-middleware`：权限中间件（角色检查：Admin / Member）

### Modified Capabilities

- `backend-scaffold`：扩展配置结构体，增加 JWT 相关配置项

## Impact

### 代码变更

| 目录/文件 | 变更类型 | 说明 |
|-----------|----------|------|
| `server/internal/config/config.go` | 修改 | 新增 JWT 配置项 |
| `server/internal/handler/auth.go` | 新增 | 认证相关 HTTP 处理器 |
| `server/internal/handler/user.go` | 新增 | 用户 Profile 处理器 |
| `server/internal/service/auth.go` | 新增 | 认证业务逻辑（密码哈希、JWT 签发） |
| `server/internal/service/user.go` | 新增 | 用户业务逻辑 |
| `server/internal/store/user.go` | 新增 | 用户数据访问层 |
| `server/internal/middleware/auth.go` | 新增 | JWT 认证中间件 |
| `server/internal/middleware/permission.go` | 新增 | 角色权限中间件 |
| `web/src/pages/Login.tsx` | 新增 | 登录页面 |
| `web/src/pages/Register.tsx` | 新增 | 注册页面 |
| `web/src/pages/Profile.tsx` | 新增 | Profile 设置页 |
| `web/src/stores/authStore.ts` | 新增 | 认证状态管理 |
| `web/src/api/auth.ts` | 新增 | 认证 API 调用 |

### API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/register` | 用户注册 |
| POST | `/api/v1/auth/login` | 用户登录 |
| POST | `/api/v1/auth/refresh` | Token 刷新 |
| GET | `/api/v1/users/me` | 获取当前用户信息 |
| PATCH | `/api/v1/users/me` | 更新当前用户信息 |
| POST | `/api/v1/users/me/avatar` | 上传头像 |

### 依赖关系

```
C02 (数据库模型) ──► C03 (认证系统)
                         │
                         ▼
                   C04 (Workspace 与 Teams)
```

### 安全考虑

- 密码使用 bcrypt 哈希（cost=12）
- JWT access token 有效期 15 分钟
- JWT refresh token 有效期 7 天
- 敏感端点需要认证中间件保护

### 对应路线图功能项

- #86 用户系统（注册/登录、JWT）
- #87 用户个人资料（Profile）
- #91 角色权限（Admin / Member）
