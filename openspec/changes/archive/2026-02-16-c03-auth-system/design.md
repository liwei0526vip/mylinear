# C03 — 用户认证与权限系统 技术设计

## Context

### 当前状态

- C01 项目脚手架已完成：Gin 框架、配置加载、健康检查端点
- C02 数据库模型已完成：User 表已定义（含 `password_hash`、`role` 字段）
- 现有配置系统不支持 JWT 相关配置
- 无认证中间件、无权限检查机制

### 约束

- **宪法约束**：测试先行（TDD）、表格驱动测试、拒绝 Mocks（使用真实数据库进行集成测试）
- **技术约束**：使用 Go 标准库优先，避免过度抽象
- **安全约束**：密码使用 bcrypt，JWT 使用 HS256 签名

### 利益相关者

- 后端开发：需要认证中间件保护 API
- 前端开发：需要登录/注册 UI 和 JWT 存储

## Goals / Non-Goals

**Goals:**

1. 实现邮箱 + 密码的用户注册和登录
2. 实现 JWT 认证机制（access token + refresh token）
3. 实现基于角色的权限控制（Admin / Member）
4. 实现用户 Profile 管理（头像、邮箱、全名、用户名）
5. 提供可复用的认证中间件

**Non-Goals:**

- SSO / LDAP 集成（Phase 4）
- OAuth 第三方登录（Phase 3）
- 双因素认证（2FA）
- 密码重置功能（需要邮件服务，后续实现）
- 多设备 Session 管理

## Decisions

### D1: JWT 库选择

**选择**: `github.com/golang-jwt/jwt/v5`

**理由**:
- 社区广泛使用，维护活跃
- API 简洁，符合 Go 哲学
- 支持 HS256、RS256 等多种算法

**替代方案**:
- `github.com/dgrijalva/jwt-go`：已停止维护，存在安全问题
- `github.com/lestrrat-go/jwx`：功能更丰富但更复杂，不符合简单性原则

### D2: Token 策略

**选择**: 双 Token 机制（Access Token + Refresh Token）

| Token 类型 | 有效期 | 存储位置 | 用途 |
|-----------|--------|---------|------|
| Access Token | 15 分钟 | 内存（前端 Zustand store） | API 认证 |
| Refresh Token | 7 天 | localStorage + Redis 黑名单 | 刷新 Access Token |

**理由**:
- 短期 Access Token 降低泄露风险
- Refresh Token 支持长期登录体验
- Redis 黑名单支持 Token 主动失效（用户登出、密码修改）

**替代方案**:
- 单 Token 长有效期：安全性差
- Session + Cookie：不适合 SPA + 私有部署场景

### D3: 密码哈希

**选择**: bcrypt，cost = 12

**理由**:
- 行业标准，内置盐值
- cost 可调整，适应计算能力增长
- Go 标准库 `golang.org/x/crypto/bcrypt`

**替代方案**:
- Argon2：更安全但配置复杂，bcrypt 已足够
- SHA256 + 盐：不推荐，易受 GPU 攻击

### D4: 依赖注入方式

**选择**: 构造函数注入

```go
type AuthService struct {
    db    *gorm.DB
    redis *redis.Client
    cfg   *config.Config
}

func NewAuthService(db *gorm.DB, redis *redis.Client, cfg *config.Config) *AuthService {
    return &AuthService{db: db, redis: redis, cfg: cfg}
}
```

**理由**:
- 符合宪法"禁止全局变量"约束
- 显式依赖，易于测试
- 简单直接，无需依赖注入框架

**替代方案**:
- Wire / Dig 等依赖注入框架：增加复杂度，不符合简单性原则

### D5: 头像存储

**选择**: MinIO（S3 兼容）

**理由**:
- 已在 C01 规划中
- 支持私有部署
- S3 API 标准，未来可迁移到其他 S3 服务

**实现**:
- 头像存储路径: `avatars/{user_id}/{uuid}.{ext}`
- 支持格式: jpg, jpeg, png, gif, webp
- 最大文件: 5MB

### D6: API 响应格式

**选择**: 统一 JSON 响应结构

```go
// 成功响应
type Response struct {
    Data interface{} `json:"data"`
}

// 错误响应
type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message,omitempty"`
}
```

**理由**:
- 结构简单，符合 Linear API 风格
- 前端易于处理

### D7: 前端状态管理

**选择**: Zustand + localStorage 持久化

**理由**:
- 已在技术栈中确定
- 轻量级，API 简洁
- 支持持久化中间件

## Architecture

### 后端分层架构

```
┌─────────────────────────────────────────────────────────────┐
│                        HTTP Handler                          │
│  (auth.go, user.go)                                          │
│  - 请求解析与验证                                              │
│  - 响应序列化                                                  │
└───────────────────────────┬─────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────┐
│                        Service Layer                         │
│  (auth_service.go, user_service.go)                          │
│  - 业务逻辑                                                   │
│  - 密码哈希/验证                                               │
│  - JWT 签发/验证                                               │
└───────────────────────────┬─────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────┐
│                        Store Layer                           │
│  (user_store.go)                                             │
│  - 数据库操作                                                  │
│  - Redis 操作                                                 │
└─────────────────────────────────────────────────────────────┘
```

### 中间件流程

```
Request ──► Auth Middleware ──► Permission Middleware ──► Handler
                │                       │
                ▼                       ▼
          验证 JWT Token         检查用户角色
          注入 User 到 Context   拒绝无权限请求
```

### Token 刷新流程

```
┌─────────┐     POST /auth/refresh     ┌─────────┐
│ Client  │ ──────────────────────────►│ Server  │
└─────────┘     {refresh_token}        └────┬────┘
     ▲                                      │
     │         {access_token,               │
     │          refresh_token}              │
     └──────────────────────────────────────┘
                     │
        ┌────────────┴────────────┐
        ▼                         ▼
   验证 refresh_token        检查 Redis 黑名单
   签发新 token pair        (如存在则拒绝)
```

## Data Model Changes

### 配置扩展

在 `config/config.go` 中新增：

```go
type Config struct {
    // ... 现有字段 ...

    // JWT 配置
    JWTSecret          string
    JWTAccessExpiry    time.Duration  // 默认 15 分钟
    JWTRefreshExpiry   time.Duration  // 默认 7 天
}
```

### Redis 数据结构

| Key Pattern | 类型 | TTL | 说明 |
|-------------|------|-----|------|
| `refresh_blacklist:{token_id}` | String | 7 天 | Refresh Token 黑名单 |
| `user_sessions:{user_id}` | Set | 7 天 | 用户活跃 Session 列表（可选） |

## API Design

### 认证 API

#### POST /api/v1/auth/register

**Request:**
```json
{
  "email": "user@example.com",
  "username": "johndoe",
  "password": "SecurePass123!",
  "name": "John Doe"
}
```

**Response (201):**
```json
{
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "username": "johndoe",
      "name": "John Doe"
    },
    "access_token": "eyJ...",
    "refresh_token": "eyJ..."
  }
}
```

#### POST /api/v1/auth/login

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response (200):**
```json
{
  "data": {
    "user": {...},
    "access_token": "eyJ...",
    "refresh_token": "eyJ..."
  }
}
```

#### POST /api/v1/auth/refresh

**Request:**
```json
{
  "refresh_token": "eyJ..."
}
```

**Response (200):**
```json
{
  "data": {
    "access_token": "eyJ...",
    "refresh_token": "eyJ..."
  }
}
```

### 用户 API

#### GET /api/v1/users/me

**Headers:** `Authorization: Bearer {access_token}`

**Response (200):**
```json
{
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "johndoe",
    "name": "John Doe",
    "avatar_url": "https://...",
    "role": "member",
    "created_at": "2026-02-16T00:00:00Z"
  }
}
```

#### PATCH /api/v1/users/me

**Request:**
```json
{
  "name": "John Smith",
  "username": "johnsmith",
  "email": "johnsmith@example.com"
}
```

**Response (200):**
```json
{
  "data": { /* 更新后的用户信息 */ }
}
```

#### POST /api/v1/users/me/avatar

**Request:** `multipart/form-data`
- `avatar`: 图片文件

**Response (200):**
```json
{
  "data": {
    "avatar_url": "https://minio/avatars/uuid/uuid.jpg"
  }
}
```

## Frontend Design

### 路由结构

```
/login          - 登录页（公开）
/register       - 注册页（公开）
/               - 主应用（需认证）
/settings/profile - Profile 设置页（需认证）
```

### Auth Store (Zustand)

```typescript
interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;

  login: (email: string, password: string) => Promise<void>;
  register: (data: RegisterData) => Promise<void>;
  logout: () => void;
  refreshTokens: () => Promise<void>;
  setUser: (user: User) => void;
}
```

### Token 自动刷新

- Axios 拦截器捕获 401 响应
- 使用 refresh_token 获取新 access_token
- 重试原请求
- 刷新失败则跳转登录页

## Risks / Trade-offs

### R1: JWT 无法主动失效

**风险**: Access Token 在有效期内无法撤销

**缓解措施**:
- Access Token 有效期短（15 分钟）
- 敏感操作（修改密码）后将 Refresh Token 加入黑名单
- Phase 2 可考虑引入 Token 版本号机制

### R2: Refresh Token 泄露风险

**风险**: Refresh Token 泄露可导致长期账户被盗

**缓解措施**:
- 前端使用 localStorage 存储（XSS 风险）
- 后端支持 Refresh Token 轮换（每次刷新生成新 token）
- 提供"登出所有设备"功能（清除所有 refresh token）

### R3: 头像上传安全

**风险**: 恶意文件上传

**缓解措施**:
- 服务端验证文件类型（magic number）
- 限制文件大小（5MB）
- 重命名文件，不使用原始文件名
- 图片处理后存储（可选，Phase 2）

### R4: 密码强度

**风险**: 弱密码导致账户被盗

**缓解措施**:
- 最小长度 8 字符
- 必须包含大小写字母和数字
- Phase 2 可增加密码强度检测和常见密码黑名单

## Migration Plan

### 部署步骤

1. 部署后端变更（无数据库迁移）
2. 部署前端变更
3. 验证登录/注册流程

### 回滚策略

- 后端回滚： revert 代码，无数据库变更影响
- 前端回滚： revert 代码

## Open Questions

1. **用户名唯一性范围**：当前设计为全局唯一，是否需要支持工作区内唯一？（当前决策：全局唯一）
2. **首次注册是否自动创建 Workspace**：当前设计需要单独创建，是否需要引导创建？（决策：Phase 1 保持分离，Phase 2 优化 onboarding）
3. **MinIO bucket 初始化时机**：在 C01 还是 C03？（决策：C03 头像上传时按需创建）
