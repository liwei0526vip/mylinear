# C03 — 用户认证与权限系统 实现任务清单

> **TDD 开发原则**：严格遵循 Red-Green-Refactor 循环
> - 🔴 Red：先写失败的测试
> - 🟢 Green：写最少的代码让测试通过
> - 🔵 Refactor：重构代码（保持测试通过）

---

## 1. 基础设施与配置

### 1.1 添加依赖

- [x] 1.1 添加 JWT 和 bcrypt 依赖到 go.mod（golang-jwt/jwt/v5, x/crypto/bcrypt）
- [x] 1.2 更新 .env.example 添加 JWT 配置示例

### 1.2 配置扩展（TDD）

- [x] 1.3 🔴 编写 config_test.go：测试 JWT 配置项加载（表格驱动，含默认值和环境变量场景）
- [x] 1.4 🟢 扩展 config.go，添加 JWTSecret、JWTAccessExpiry、JWTRefreshExpiry 字段和加载逻辑
- [x] 1.5 🔴 编写 config_test.go：测试生产环境必须设置 JWT_SECRET（表格驱动）
- [x] 1.6 🟢 实现生产环境 JWT_SECRET 检查逻辑

---

## 2. 用户数据访问层（Store）

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

### 2.1 UserStore 接口与基础结构

- [x] 2.1 🔴 编写 store/user_test.go：测试 UserStore 接口定义存在
- [x] 2.2 🟢 创建 store/user.go，定义 UserStore 接口和 userStore 结构体

### 2.2 CreateUser 方法（TDD）

- [x] 2.3 🔴 编写 CreateUser 测试：成功创建用户（表格驱动）
- [x] 2.4 🟢 实现 CreateUser 方法：插入用户记录
- [x] 2.5 🔴 编写 CreateUser 测试：邮箱重复时返回错误（表格驱动）
- [x] 2.6 🟢 实现 CreateUser 方法：邮箱唯一性检查
- [x] 2.7 🔴 编写 CreateUser 测试：用户名重复时返回错误（表格驱动）
- [x] 2.8 🟢 实现 CreateUser 方法：用户名唯一性检查

### 2.3 GetUserByEmail 方法（TDD）

- [x] 2.9 🔴 编写 GetUserByEmail 测试：成功获取用户（表格驱动）
- [x] 2.10 🟢 实现 GetUserByEmail 方法
- [x] 2.11 🔴 编写 GetUserByEmail 测试：用户不存在时返回错误（表格驱动）
- [x] 2.12 🟢 实现 GetUserByEmail 方法：处理用户不存在场景

### 2.4 GetUserByID 方法（TDD）

- [x] 2.13 🔴 编写 GetUserByID 测试：成功获取用户（表格驱动）
- [x] 2.14 🟢 实现 GetUserByID 方法
- [x] 2.15 🔴 编写 GetUserByID 测试：用户不存在时返回错误（表格驱动）
- [x] 2.16 🟢 实现 GetUserByID 方法：处理用户不存在场景

### 2.5 GetUserByUsername 方法（TDD）

- [x] 2.17 🔴 编写 GetUserByUsername 测试：成功获取用户（表格驱动）
- [x] 2.18 🟢 实现 GetUserByUsername 方法
- [x] 2.19 🔴 编写 GetUserByUsername 测试：用户不存在时返回错误（表格驱动）
- [x] 2.20 🟢 实现 GetUserByUsername 方法：处理用户不存在场景

### 2.6 UpdateUser 方法（TDD）

- [x] 2.21 🔴 编写 UpdateUser 测试：成功更新用户信息（表格驱动）
- [x] 2.22 🟢 实现 UpdateUser 方法
- [x] 2.23 🔴 编写 UpdateUser 测试：更新邮箱时检查唯一性（表格驱动）
- [x] 2.24 🟢 实现 UpdateUser 方法：邮箱唯一性检查
- [x] 2.25 🔴 编写 UpdateUser 测试：更新用户名时检查唯一性（表格驱动）
- [x] 2.26 🟢 实现 UpdateUser 方法：用户名唯一性检查
- [x] 2.27 🔴 编写 UpdateUser 测试：更新头像 URL（表格驱动）
- [x] 2.28 🟢 实现 UpdateUser 方法：头像 URL 更新逻辑

---

## 3. JWT 服务

### 3.1 JWTService 结构体

- [x] 3.1 🔴 编写 service/jwt_test.go：测试 JWTService 结构体定义
- [x] 3.2 🟢 创建 service/jwt.go，定义 JWTService 结构体和构造函数

### 3.2 GenerateAccessToken 方法（TDD）

- [x] 3.3 🔴 编写 GenerateAccessToken 测试：成功生成 token（表格驱动）
- [x] 3.4 🟢 实现 GenerateAccessToken 方法
- [x] 3.5 🔴 编写 GenerateAccessToken 测试：token 包含正确的 claims（sub, email, role, exp, iat, jti）
- [x] 3.6 🟢 实现 token claims 设置逻辑

### 3.3 GenerateRefreshToken 方法（TDD）

- [x] 3.7 🔴 编写 GenerateRefreshToken 测试：成功生成 token（表格驱动）
- [x] 3.8 🟢 实现 GenerateRefreshToken 方法
- [x] 3.9 🔴 编写 GenerateRefreshToken 测试：token 包含正确的 claims（sub, type, exp, iat, jti）
- [x] 3.10 🟢 实现 refresh token claims 设置逻辑

### 3.4 ValidateToken 方法（TDD）

- [x] 3.11 🔴 编写 ValidateToken 测试：有效 token 返回 claims（表格驱动）
- [x] 3.12 🟢 实现 ValidateToken 方法
- [x] 3.13 🔴 编写 ValidateToken 测试：过期 token 返回错误（表格驱动）
- [x] 3.14 🟢 实现 token 过期检查逻辑
- [x] 3.15 🔴 编写 ValidateToken 测试：签名无效返回错误（表格驱动）
- [x] 3.16 🟢 实现 token 签名验证逻辑
- [x] 3.17 🔴 编写 ValidateToken 测试：格式无效返回错误（表格驱动）
- [x] 3.18 🟢 实现 token 格式验证逻辑

### 3.5 GetTokenClaims 方法（TDD）

- [x] 3.19 🔴 编写 GetTokenClaims 测试：成功解析 claims（表格驱动）
- [x] 3.20 🟢 实现 GetTokenClaims 方法

---

## 4. 认证服务

> 使用真实数据库和 Redis 进行集成测试

### 4.1 AuthService 结构体

- [x] 4.1 🔴 编写 service/auth_test.go：测试 AuthService 结构体定义
- [x] 4.2 🟢 创建 service/auth.go，定义 AuthService 结构体和构造函数

### 4.2 Register 方法（TDD）

- [x] 4.3 🔴 编写 Register 测试：成功注册新用户（表格驱动）
- [x] 4.4 🟢 实现 Register 方法：创建用户并生成 token
- [x] 4.5 🔴 编写 Register 测试：密码使用 bcrypt 哈希（表格驱动）
- [x] 4.6 🟢 实现密码哈希逻辑
- [x] 4.7 🔴 编写 Register 测试：邮箱已存在返回错误（表格驱动）
- [x] 4.8 🟢 实现邮箱唯一性检查
- [x] 4.9 🔴 编写 Register 测试：用户名已存在返回错误（表格驱动）
- [x] 4.10 🟢 实现用户名唯一性检查
- [x] 4.11 🔴 编写 Register 测试：密码强度验证（表格驱动，含弱密码场景）
- [x] 4.12 🟢 实现密码强度验证逻辑
- [x] 4.13 🔴 编写 Register 测试：邮箱格式验证（表格驱动）
- [x] 4.14 🟢 实现邮箱格式验证逻辑
- [x] 4.15 🔴 编写 Register 测试：用户名格式验证（表格驱动）
- [x] 4.16 🟢 实现用户名格式验证逻辑

### 4.3 Login 方法（TDD）

- [x] 4.17 🔴 编写 Login 测试：成功登录返回 token（表格驱动）
- [x] 4.18 🟢 实现 Login 方法：验证密码并生成 token
- [x] 4.19 🔴 编写 Login 测试：邮箱不存在返回错误（表格驱动）
- [x] 4.20 🟢 实现邮箱检查逻辑
- [x] 4.21 🔴 编写 Login 测试：密码错误返回错误（表格驱动）
- [x] 4.22 🟢 实现密码验证逻辑

### 4.4 RefreshToken 方法（TDD）

- [x] 4.23 🔴 编写 RefreshToken 测试：成功刷新返回新 token（表格驱动）
- [x] 4.24 🟢 实现 RefreshToken 方法：验证并生成新 token
- [x] 4.25 🔴 编写 RefreshToken 测试：token 过期返回错误（表格驱动）
- [x] 4.26 🟢 实现 token 过期检查
- [x] 4.27 🔴 编写 RefreshToken 测试：token 在黑名单中返回错误（表格驱动）
- [x] 4.28 🟢 实现 Redis 黑名单检查
- [x] 4.29 🔴 编写 RefreshToken 测试：使用 access token 刷新返回错误（表格驱动）
- [x] 4.30 🟢 实现 token 类型检查
- [x] 4.31 🔴 编写 RefreshToken 测试：刷新后旧 token 失效（表格驱动）
- [x] 4.32 🟢 实现 token 轮换机制（旧 token 加入黑名单）

### 4.5 Logout 方法（TDD）

- [x] 4.33 🔴 编写 Logout 测试：成功将 token 加入黑名单（表格驱动）
- [x] 4.34 🟢 实现 Logout 方法：将 refresh token 加入 Redis 黑名单

---

## 5. 认证中间件

### 5.1 Auth 中间件（TDD）

- [x] 5.1 🔴 编写 middleware/auth_test.go：测试有效 token 通过（表格驱动）
- [x] 5.2 🟢 创建 middleware/auth.go，实现 Auth 中间件：验证 JWT
- [x] 5.3 🔴 编写 Auth 中间件测试：缺少 Authorization 头返回 401（表格驱动）
- [x] 5.4 🟢 实现 Authorization 头检查逻辑
- [x] 5.5 🔴 编写 Auth 中间件测试：Bearer 格式错误返回 401（表格驱动）
- [x] 5.6 🟢 实现 Bearer 格式解析逻辑
- [x] 5.7 🔴 编写 Auth 中间件测试：token 过期返回 401（表格驱动）
- [x] 5.8 🟢 实现 token 过期处理
- [x] 5.9 🔴 编写 Auth 中间件测试：token 无效返回 401（表格驱动）
- [x] 5.10 🟢 实现 token 验证错误处理

### 5.2 用户上下文注入（TDD）

- [x] 5.11 🔴 编写 GetCurrentUser 测试：成功获取用户信息（表格驱动）
- [x] 5.12 🟢 实现 GetCurrentUser 辅助函数
- [x] 5.13 🔴 编写 GetCurrentUser 测试：未认证返回 nil（表格驱动）
- [x] 5.14 🟢 实现未认证场景处理

---

## 6. 权限中间件

### 6.1 RequireRole 中间件（TDD）

- [x] 6.1 🔴 编写 middleware/permission_test.go：测试有权限用户通过（表格驱动）
- [x] 6.2 🟢 创建 middleware/permission.go，实现 RequireRole 中间件
- [x] 6.3 🔴 编写 RequireRole 测试：无权限返回 403（表格驱动）
- [x] 6.4 🟢 实现权限检查和拒绝逻辑

### 6.2 RequireAdmin 中间件（TDD）

- [x] 6.5 🔴 编写 RequireAdmin 测试：admin 角色通过（表格驱动）
- [x] 6.6 🟢 实现 RequireAdmin 便捷中间件
- [x] 6.7 🔴 编写 RequireAdmin 测试：global_admin 角色通过（表格驱动）
- [x] 6.8 🟢 实现 global_admin 包含在 admin 权限中

### 6.3 RequireGlobalAdmin 中间件（TDD）

- [x] 6.9 🔴 编写 RequireGlobalAdmin 测试：仅 global_admin 通过（表格驱动）
- [x] 6.10 🟢 实现 RequireGlobalAdmin 中间件
- [x] 6.11 🔴 编写 RequireGlobalAdmin 测试：admin 角色返回 403（表格驱动）
- [x] 6.12 🟢 实现严格的 global_admin 检查

### 6.4 辅助函数（TDD）

- [x] 6.13 🔴 编写 GetCurrentUserRole 测试：成功获取角色（表格驱动）
- [x] 6.14 🟢 实现 GetCurrentUserRole 辅助函数
- [x] 6.15 🔴 编写 IsAdmin 测试：admin 和 global_admin 返回 true（表格驱动）
- [x] 6.16 🟢 实现 IsAdmin 辅助函数

---

## 7. 认证 HTTP Handler

### 7.1 AuthHandler 结构体

- [x] 7.1 🔴 编写 handler/auth_test.go：测试 AuthHandler 结构体定义
- [x] 7.2 🟢 创建 handler/auth.go，定义 AuthHandler 结构体和构造函数

### 7.2 Register 处理器（TDD）

- [x] 7.3 🔴 编写 Register 处理器测试：成功注册返回 201（表格驱动）
- [x] 7.4 🟢 实现 Register 处理器（POST /api/v1/auth/register）
- [x] 7.5 🔴 编写 Register 处理器测试：邮箱已存在返回 409（表格驱动）
- [x] 7.6 🟢 实现邮箱冲突错误处理
- [x] 7.7 🔴 编写 Register 处理器测试：用户名已存在返回 409（表格驱动）
- [x] 7.8 🟢 实现用户名冲突错误处理
- [x] 7.9 🔴 编写 Register 处理器测试：输入验证失败返回 400（表格驱动）
- [x] 7.10 🟢 实现请求验证逻辑

### 7.3 Login 处理器（TDD）

- [x] 7.11 🔴 编写 Login 处理器测试：成功登录返回 200（表格驱动）
- [x] 7.12 🟢 实现 Login 处理器（POST /api/v1/auth/login）
- [x] 7.13 🔴 编写 Login 处理器测试：认证失败返回 401（表格驱动）
- [x] 7.14 🟢 实现认证失败错误处理
- [x] 7.15 🔴 编写 Login 处理器测试：缺少字段返回 400（表格驱动）
- [x] 7.16 🟢 实现请求验证逻辑

### 7.4 Refresh 处理器（TDD）

- [x] 7.17 🔴 编写 Refresh 处理器测试：成功刷新返回 200（表格驱动）
- [x] 7.18 🟢 实现 Refresh 处理器（POST /api/v1/auth/refresh）
- [x] 7.19 🔴 编写 Refresh 处理器测试：token 无效返回 401（表格驱动）
- [x] 7.20 🟢 实现 token 无效错误处理

### 7.5 Logout 处理器（TDD）

- [x] 7.21 🔴 编写 Logout 处理器测试：成功登出返回 200（表格驱动）
- [x] 7.22 🟢 实现 Logout 处理器（POST /api/v1/auth/logout）

---

## 8. 用户 Profile Handler

### 8.1 UserHandler 结构体

- [x] 8.1 🔴 编写 handler/user_test.go：测试 UserHandler 结构体定义
- [x] 8.2 🟢 创建 handler/user.go，定义 UserHandler 结构体和构造函数

### 8.2 GetMe 处理器（TDD）

- [x] 8.3 🔴 编写 GetMe 处理器测试：成功返回用户信息（表格驱动）
- [x] 8.4 🟢 实现 GetMe 处理器（GET /api/v1/users/me）
- [x] 8.5 🔴 编写 GetMe 处理器测试：未认证返回 401（表格驱动）
- [x] 8.6 🟢 实现认证检查（依赖 Auth 中间件）

### 8.3 UpdateMe 处理器（TDD）

- [x] 8.7 🔴 编写 UpdateMe 处理器测试：成功更新返回 200（表格驱动）
- [x] 8.8 🟢 实现 UpdateMe 处理器（PATCH /api/v1/users/me）
- [x] 8.9 🔴 编写 UpdateMe 处理器测试：邮箱冲突返回 409（表格驱动）
- [x] 8.10 🟢 实现邮箱冲突错误处理
- [x] 8.11 🔴 编写 UpdateMe 处理器测试：用户名冲突返回 409（表格驱动）
- [x] 8.12 🟢 实现用户名冲突错误处理

### 8.4 UploadAvatar 处理器（TDD）

- [x] 8.13 🔴 编写 UploadAvatar 处理器测试：成功上传返回 200（表格驱动）
- [x] 8.14 🟢 实现 UploadAvatar 处理器（POST /api/v1/users/me/avatar）
- [x] 8.15 🔴 编写 UploadAvatar 处理器测试：文件过大返回 400（表格驱动）
- [x] 8.16 🟢 实现文件大小检查
- [x] 8.17 🔴 编写 UploadAvatar 处理器测试：文件类型不支持返回 400（表格驱动）
- [x] 8.18 🟢 实现文件类型检查

---

## 9. 头像上传服务

### 9.1 AvatarService 结构体

- [x] 9.1 🔴 编写 service/avatar_test.go：测试 AvatarService 结构体定义
- [x] 9.2 🟢 创建 service/avatar.go，定义 AvatarService 接口

### 9.2 UploadAvatar 方法（TDD）

- [x] 9.3 🔴 编写 UploadAvatar 测试：成功上传到 MinIO（表格驱动）
- [x] 9.4 🟢 实现 UploadAvatar 方法：上传到 MinIO
- [x] 9.5 🔴 编写 UploadAvatar 测试：验证图片 magic number（表格驱动）
- [x] 9.6 🟢 实现文件类型验证（magic number 检测）
- [x] 9.7 🔴 编写 UploadAvatar 测试：生成正确的存储路径（表格驱动）
- [x] 9.8 🟢 实现存储路径生成逻辑（avatars/{user_id}/{uuid}.{ext}）

---

## 10. 路由集成

### 10.1 服务初始化

- [x] 10.1 更新 main.go，初始化所有服务（JWTService, AuthService, UserService）
- [x] 10.2 初始化所有 Handler（AuthHandler, UserHandler）

### 10.2 路由注册（TDD）

- [x] 10.3 🔴 编写集成测试：测试公开路由（/api/v1/auth/*）无需认证
- [x] 10.4 🟢 注册认证相关路由（公开）
- [x] 10.5 🔴 编写集成测试：测试受保护路由（/api/v1/users/*）需要认证
- [x] 10.6 🟢 注册用户相关路由（需认证）
- [x] 10.7 🔴 编写集成测试：测试完整认证流程（注册→登录→访问→刷新→登出）
- [x] 10.8 🟢 验证并修复集成问题

---

## 11. 前端 - 类型定义

- [x] 11.1 创建 web/src/types/auth.ts，定义认证相关 TypeScript 类型（LoginRequest, RegisterRequest, AuthResponse, TokenPair）
- [x] 11.2 创建 web/src/types/user.ts，定义用户相关 TypeScript 类型（User, UpdateUserRequest）

## 12. 前端 - 认证 API 层

- [x] 12.1 创建 web/src/api/auth.ts，实现登录 API 调用
- [x] 12.2 实现注册 API 调用
- [x] 12.3 实现刷新 Token API 调用
- [x] 12.4 实现登出 API 调用
- [x] 12.5 创建 web/src/api/user.ts，实现获取当前用户 API 调用
- [x] 12.6 实现更新用户信息 API 调用
- [x] 12.7 实现上传头像 API 调用

## 13. 前端 - 认证状态管理

- [x] 13.1 创建 web/src/stores/authStore.ts，定义 AuthState 接口
- [x] 13.2 实现 login action
- [x] 13.3 实现 register action
- [x] 13.4 实现 logout action
- [x] 13.5 实现 refreshTokens action
- [x] 13.6 实现 localStorage 持久化（refresh_token）
- [x] 13.7 创建 web/src/lib/axios.ts，配置 Axios 实例
- [x] 13.8 实现 Axios 请求拦截器（添加 Authorization 头）
- [x] 13.9 实现 Axios 响应拦截器（401 时自动刷新 Token）
- [x] 13.10 实现刷新失败后跳转登录页

## 14. 前端 - 登录/注册页面

- [x] 14.1 创建 web/src/pages/Login.tsx 登录页面组件
- [x] 14.2 实现登录表单（基础验证）
- [x] 14.3 实现表单提交和错误提示
- [x] 14.4 实现登录成功后跳转
- [x] 14.5 创建 web/src/pages/Register.tsx 注册页面组件
- [x] 14.6 实现注册表单（基础验证）
- [x] 14.7 实现表单提交和错误提示
- [x] 14.8 实现注册成功后自动登录并跳转

## 15. 前端 - Profile 设置页

- [x] 15.1 创建 web/src/pages/Profile.tsx Profile 设置页
- [x] 15.2 实现用户信息展示
- [x] 15.3 实现用户信息编辑表单
- [x] 15.4 实现保存功能和成功/失败提示
- [x] 15.5 创建 web/src/components/AvatarUpload.tsx 头像上传组件
- [x] 15.6 实现图片选择和预览
- [x] 15.7 实现上传功能和进度显示

## 16. 前端 - 路由守卫

- [x] 16.1 创建 web/src/components/ProtectedRoute.tsx 路由守卫组件
- [x] 16.2 实现未认证重定向到登录页逻辑
- [x] 16.3 实现登录后跳转回原页面逻辑
- [x] 16.4 更新路由配置，保护需要认证的页面

## 17. 端到端验证

- [x] 17.1 启动完整开发环境（Docker Compose）
- [x] 17.2 验证用户注册流程（前端 + 后端）
- [x] 17.3 验证用户登录流程（前端 + 后端）
- [x] 17.4 验证 Token 刷新流程
- [x] 17.5 验证 Profile 更新流程
- [x] 17.6 验证头像上传流程（需要 MinIO 配置）
- [x] 17.7 验证权限控制（Admin vs Member）
- [x] 17.8 验证登出后 Token 失效

---

**任务统计**：
- 后端 TDD 任务：92 个（测试 + 实现 配对）
- 前端任务：32 个
- 验证任务：8 个
- **总计：132 个任务**
- **预估工时**：约 5-6 天

**TDD 任务格式说明**：
- 🔴 = Red 阶段（编写失败的测试）
- 🟢 = Green 阶段（编写实现让测试通过）
- 每个 🔴 任务后必须紧跟对应的 🟢 任务
