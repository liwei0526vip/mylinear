## ADDED Requirements

### Requirement: Go 后端项目结构

Go 后端 MUST 遵循标准项目布局，代码组织在 `server/` 目录下。

目录结构：
```
server/
├── cmd/
│   └── server/
│       └── main.go          # 应用入口
├── internal/
│   ├── config/              # 配置管理
│   ├── handler/             # HTTP 处理器
│   ├── middleware/          # 中间件
│   ├── model/               # 数据模型
│   ├── service/             # 业务逻辑
│   └── store/               # 数据存储层
├── migrations/              # 数据库迁移
├── go.mod
└── go.sum
```

#### Scenario: 项目结构完整性
- **WHEN** 开发者查看 `server/` 目录
- **THEN** MUST 包含 `cmd/server/main.go`（入口）、`internal/`（业务代码）、`migrations/`（迁移文件）目录

#### Scenario: Go Module 初始化
- **WHEN** 开发者在 `server/` 目录执行 `go build ./...`
- **THEN** 项目 SHALL 编译成功，无编译错误

---

### Requirement: Gin 框架集成

后端 MUST 使用 Gin 作为 HTTP 框架，支持路由注册、中间件和 JSON 响应。

#### Scenario: HTTP 服务器正常启动
- **WHEN** 后端应用启动
- **THEN** Gin HTTP 服务器 SHALL 在配置的端口（默认 8080）上监听请求

#### Scenario: 服务器优雅关闭
- **WHEN** 后端应用接收到 SIGINT 或 SIGTERM 信号
- **THEN** 服务器 SHALL 优雅关闭，等待处理中的请求完成（最长 10 秒超时）

---

### Requirement: 健康检查端点

后端 MUST 提供健康检查端点，用于监控服务状态。

#### Scenario: 服务正常时的健康检查
- **WHEN** 客户端发送 `GET /api/v1/health` 请求
- **THEN** 服务器 SHALL 返回 HTTP 200 和 JSON 响应 `{"status": "ok"}`

#### Scenario: 数据库连接异常时的健康检查
- **WHEN** 客户端发送 `GET /api/v1/health` 请求，且数据库连接不可用
- **THEN** 服务器 SHALL 返回 HTTP 503 和 JSON 响应 `{"status": "error", "message": "database unavailable"}`

---

### Requirement: 后端配置管理

后端应用 MUST 支持通过 godotenv 加载 `.env` 文件，然后从环境变量读取配置，包括数据库连接、Redis 地址、服务端口等。所有配置项 MUST 提供开发环境默认值，确保仅使用 `.env.example` 即可正常运行。

#### Scenario: 从 .env 文件加载配置
- **WHEN** 后端应用启动且工作目录存在 `.env` 文件
- **THEN** 应用 SHALL 使用 godotenv 加载 `.env` 中的变量，不覆盖已存在的系统环境变量

#### Scenario: .env 文件不存在时静默跳过
- **WHEN** 后端应用启动且工作目录不存在 `.env` 文件（如生产环境 Docker 容器）
- **THEN** 应用 SHALL 静默跳过，继续从系统环境变量读取配置

#### Scenario: 从环境变量读取配置
- **WHEN** 后端应用启动且环境变量 `DATABASE_URL`、`REDIS_URL`、`PORT` 已设置
- **THEN** 应用 SHALL 使用这些环境变量中的值进行配置

#### Scenario: 配置缺失时使用开发默认值
- **WHEN** 后端应用启动且部分环境变量（如 `PORT`、`DATABASE_URL`）未设置
- **THEN** 应用 SHALL 使用预定义的开发默认值（如端口 8080、本地 PostgreSQL 连接串）

---

### Requirement: 数据库连接管理

后端 MUST 管理 PostgreSQL 数据库连接池，支持连接健康检查和自动重连。

#### Scenario: 数据库连接池初始化
- **WHEN** 后端应用启动并连接到 PostgreSQL
- **THEN** 应用 SHALL 建立数据库连接池，包含合理的最大连接数和空闲连接数配置

#### Scenario: 数据库连接池关闭
- **WHEN** 后端应用关闭
- **THEN** 应用 SHALL 正确关闭数据库连接池，释放所有连接资源

---

### Requirement: Redis 连接管理

后端 MUST 管理 Redis 客户端连接，用于缓存和会话管理。

#### Scenario: Redis 客户端初始化
- **WHEN** 后端应用启动并连接到 Redis
- **THEN** 应用 SHALL 建立 Redis 客户端连接

#### Scenario: Redis 不可用时的降级处理
- **WHEN** Redis 服务不可用
- **THEN** 应用 SHOULD 记录警告日志但不阻止启动，非关键功能可降级运行
