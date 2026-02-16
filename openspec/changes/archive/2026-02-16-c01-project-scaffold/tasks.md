## 1. 基础设施配置

- [x] 1.1 创建 `.env.example` 文件，定义所有环境变量及默认值（DATABASE_URL、REDIS_URL、PORT、MINIO 配置等）
- [x] 1.2 创建 `docker-compose.yml`，编排 PostgreSQL 16、Redis 7、MinIO、Caddy 四个服务，配置 volumes 持久化和 `restart: unless-stopped`（Caddy 使用独立 profile `proxy`，开发时默认不启动）
- [x] 1.3 创建 `Caddyfile`，配置 `/api/*` 代理至后端（localhost:8080）、`/*` 代理至前端（localhost:5173），用于生产部署或 `docker compose --profile proxy up`
- [x] 1.4 创建 `Makefile`，定义 `dev`（启动 DB/Redis/MinIO + 后端 + 前端，不含 Caddy）、`build`、`test`、`down` 等常用命令
- [x] 1.5 更新 `.gitignore`，补充 Go（`/server/bin/`）、Node.js（`node_modules/`、`dist/`）、IDE、`.env` 等忽略规则

## 2. Go 后端初始化

- [x] 2.1 初始化 Go module（`server/go.mod`），添加 Gin、GORM（postgres driver）、go-redis、godotenv 依赖
- [x] 2.2 编写 config 的表格驱动测试 `server/internal/config/config_test.go`（先写失败测试 — Red）
- [x] 2.3 创建 `server/internal/config/config.go`，实现配置加载逻辑：使用 godotenv 加载 `.env` 文件 → 从环境变量读取 → 支持开发默认值 → 关键配置缺失时报错（使测试通过 — Green）
- [x] 2.4 编写健康检查端点的表格驱动测试 `server/internal/handler/health_test.go`（先写失败测试 — Red）
- [x] 2.5 创建 `server/internal/handler/health.go`，实现健康检查端点 `GET /api/v1/health`（检查数据库连接状态）（使测试通过 — Green）
- [x] 2.6 创建 `server/cmd/server/main.go`，实现应用入口：加载 .env → 初始化配置 → 连接数据库 → 连接 Redis → 注册路由 → 启动 Gin 服务器 → 优雅关闭
- [x] 2.7 创建预留目录结构：`server/internal/middleware/`、`server/internal/model/`、`server/internal/service/`、`server/internal/store/`、`server/migrations/`

## 3. React 前端初始化

- [x] 3.1 使用 `create-vite` 初始化 React + TypeScript 项目到 `web/` 目录
- [x] 3.2 配置 `vite.config.ts`，添加 `/api` 代理至 `http://localhost:8080`
- [x] 3.3 安装并配置 shadcn/ui（创建 `components.json`，配置路径别名和 CSS 变量）
- [x] 3.4 安装 Zustand，创建示例 store `web/src/stores/app-store.ts`
- [x] 3.5 创建 API 客户端 `web/src/api/client.ts`，封装 fetch 逻辑（自动添加 `/api/v1` 前缀、JSON 解析、错误处理）
- [x] 3.6 创建预留目录结构：`web/src/components/`、`web/src/pages/`、`web/src/lib/`
- [x] 3.7 创建首页占位组件 `web/src/App.tsx`，显示 MyLinear 项目名称并调用健康检查 API 展示后端连接状态

## 4. 端到端验证

- [x] 4.1 验证 `docker compose up -d` 能成功启动数据库和缓存容器（PostgreSQL、Redis、MinIO）
- [x] 4.2 验证后端 `go run ./cmd/server` 能正常启动（自动加载 `.env`）并响应 `/api/v1/health`
- [x] 4.3 验证前端 `npm run dev` 能正常启动且 Vite proxy 代理 API 请求至后端
- [x] 4.4 验证 `make dev` 能一键启动完整开发环境（容器 + 后端 + 前端，不含 Caddy）
- [x] 4.5 运行 `make test` 确认所有测试通过
- [x] 4.6 验证 `docker compose --profile proxy up -d` 能启动含 Caddy 的完整编排
