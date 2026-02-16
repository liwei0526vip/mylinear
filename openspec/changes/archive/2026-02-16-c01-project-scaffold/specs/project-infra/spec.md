## ADDED Requirements

### Requirement: Docker Compose 容器编排

系统 SHALL 提供 `docker-compose.yml` 文件，用于一键部署所有基础设施服务。编排内容包括：
- PostgreSQL 16 数据库服务
- Redis 7 缓存服务
- MinIO 文件存储服务（S3 兼容）
- Caddy 反向代理服务

所有服务 MUST 通过 Docker 内部网络互联，并使用环境变量管理敏感配置（如数据库密码、Redis 密码）。

#### Scenario: 成功启动所有容器服务
- **WHEN** 用户在项目根目录执行 `docker compose up -d`
- **THEN** 所有服务（PostgreSQL、Redis、MinIO、Caddy）MUST 成功启动并处于 running 状态

#### Scenario: 服务间网络连通性
- **WHEN** 所有容器服务启动完成
- **THEN** 后端服务 MUST 能通过容器名称访问 PostgreSQL（端口 5432）、Redis（端口 6379）和 MinIO（端口 9000）

#### Scenario: 容器异常自动重启
- **WHEN** 某个服务容器异常退出
- **THEN** Docker Compose MUST 自动重启该容器（配置 `restart: unless-stopped`）

#### Scenario: 数据持久化
- **WHEN** 容器被停止并重新启动
- **THEN** PostgreSQL 和 Redis 的数据 MUST 通过 Docker volumes 持久化保留，不会丢失

---

### Requirement: Caddy 反向代理路由

Caddy MUST 正确代理前后端请求，实现统一入口访问。在开发环境中 Caddy 为可选组件，Docker Compose 中通过 `profiles: [proxy]` 标记，默认不启动。开发时直接使用 Vite 内置 proxy 代理 API 请求。

#### Scenario: API 请求路由至后端
- **WHEN** 客户端请求路径匹配 `/api/*`
- **THEN** Caddy SHALL 将请求代理至 Go 后端服务

#### Scenario: 前端资源请求路由至前端
- **WHEN** 客户端请求路径不匹配 `/api/*`
- **THEN** Caddy SHALL 将请求代理至 React 前端服务

#### Scenario: 前端 SPA 路由支持
- **WHEN** 客户端请求一个前端路由路径（如 `/issues/123`）
- **THEN** Caddy SHALL 返回前端的 `index.html`，由前端路由处理

#### Scenario: 开发环境无 Caddy 时 API 代理
- **WHEN** 开发环境未启动 Caddy，前端通过 Vite dev server 访问
- **THEN** Vite proxy SHALL 将 `/api/*` 请求代理至后端服务（默认 `http://localhost:8080`）

---

### Requirement: Makefile 标准化构建

项目 MUST 提供 `Makefile`，定义所有常用的构建和开发操作命令。

#### Scenario: 启动开发环境
- **WHEN** 用户执行 `make dev`
- **THEN** 系统 SHALL 启动基础设施容器（PostgreSQL、Redis、MinIO，不含 Caddy）、后端服务和前端开发服务器

#### Scenario: 构建后端
- **WHEN** 用户执行 `make build`
- **THEN** 系统 SHALL 编译 Go 后端为可执行文件

#### Scenario: 运行测试
- **WHEN** 用户执行 `make test`
- **THEN** 系统 SHALL 运行 Go 后端和前端的所有测试

#### Scenario: 停止所有服务
- **WHEN** 用户执行 `make down`
- **THEN** 系统 SHALL 停止并移除所有 Docker Compose 容器

---

### Requirement: 环境配置管理

系统 MUST 支持通过环境变量和 `.env` 文件管理配置。

#### Scenario: 默认配置可用
- **WHEN** 用户首次启动项目且未创建 `.env` 文件
- **THEN** 系统 SHALL 使用 `.env.example` 中定义的默认配置正常运行

#### Scenario: 自定义配置覆盖
- **WHEN** 用户创建 `.env` 文件并修改了配置值（如数据库密码）
- **THEN** 系统 SHALL 使用 `.env` 文件中的配置覆盖默认值
