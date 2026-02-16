## Context

MyLinear 项目当前处于零代码状态，仅有产品文档（路线图、竞品分析）和 OpenSpec 配置。本 change 需要搭建完整的项目脚手架，包括 Go 后端、React 前端、Docker Compose 基础设施和构建脚本。

这是项目的第一个 change，所有后续 change（C02~C13）都依赖于本 change 提供的项目结构和基础设施。

## Goals / Non-Goals

**Goals:**
- 建立可运行的 Go 后端项目，包含标准目录结构和健康检查端点
- 建立可运行的 React + TypeScript 前端项目，集成 shadcn/ui 和 Zustand
- 提供 Docker Compose 一键部署所有基础设施服务
- 通过 Caddy 统一代理前后端请求
- 通过 Makefile 标准化常用操作

**Non-Goals:**
- 不实现任何业务逻辑（Issue、Workspace、Team 等）
- 不创建数据库表结构（C02 负责）
- 不实现用户认证（C03 负责）
- 不实现前端三栏布局（C10 负责）
- 不配置生产环境 HTTPS（开发阶段不需要）

## Decisions

### D1：后端框架选择 Gin

**决策**：使用 Gin 作为 HTTP 框架。

**理由**：
- Go 生态中最成熟的 Web 框架，社区活跃
- 性能优异，中间件机制成熟
- 路由组（Group）便于组织 API 版本（`/api/v1`）

**替代方案**：
- `net/http` + `chi`：更轻量，但缺少内置的 JSON 绑定/验证
- `Echo`：功能类似 Gin，但社区规模稍小
- `Fiber`：基于 fasthttp，性能最高但与标准库 `net/http` 不兼容

### D2：ORM 选择 GORM

**决策**：使用 GORM 作为 ORM 框架。

**理由**：
- 路线图和 AGENTS.md 中明确推荐 GORM 或 sqlc
- GORM 提供迁移工具、关联查询等功能，对快速开发更友好
- 支持 PostgreSQL，与项目数据库选型一致
- 本阶段仅配置连接，不创建模型（留给 C02）

**替代方案**：
- `sqlc`：类型安全，但需要手写 SQL 和额外的代码生成步骤
- 混合方案：先用 GORM 快速迭代，复杂查询用 raw SQL

### D3：前端初始化使用 Vite 官方模板

**决策**：使用 `create-vite` 官方模板初始化 React + TypeScript 项目。

**理由**：
- 官方维护，稳定可靠
- 内置 TypeScript 配置、HMR、ESBuild 预编译
- 与 shadcn/ui 兼容

**替代方案**：
- `create-react-app`：已不再推荐，社区迁移至 Vite
- `Next.js`：SSR 对内部工具不必要，增加复杂性

### D4：Caddy 作为反向代理（生产部署用，开发环境可选）

**决策**：使用 Caddy 作为反向代理，但在开发环境中为**可选组件**。开发时直接使用 Vite 内置 proxy 代理 API 请求，不需要通过 Caddy 中转。

**理由**：
- 配置简洁（Caddyfile 语法直观）
- 内置 HTTPS 自动签发（生产环境可用）
- 路线图已明确要求使用 Caddy
- 开发环境中 Vite 自带 proxy 已足够，额外启动 Caddy 增加不必要的复杂性

**实现方式**：Docker Compose 中 Caddy 服务使用 `profiles: [proxy]` 标记，默认 `docker compose up -d` 不启动 Caddy，需要时通过 `docker compose --profile proxy up -d` 显式启用。

**替代方案**：
- Nginx：功能更强大，但配置更复杂
- Traefik：动态配置强，但对 Docker Compose 单机部署过重

### D5：配置管理使用 godotenv + 环境变量

**决策**：通过 `github.com/joho/godotenv` 加载 `.env` 文件，再使用 Go 标准 `os.Getenv` 读取配置。

**理由**：
- 符合 12-Factor App 规范
- Docker Compose 原生支持 `.env` 文件（容器环境）
- godotenv 解决本地开发（非 Docker）时 Go 应用无法自动读取 `.env` 文件的问题
- 加载逻辑：godotenv 将 `.env` 文件中的变量注入 `os.Environ`，不覆盖已存在的系统环境变量
- 在生产环境（Docker 容器）中，`.env` 文件不存在也不影响运行，godotenv 会静默跳过

**配置项和默认值**：
| 配置项 | 环境变量 | 开发默认值 |
|--------|---------|----------|
| 数据库连接 | `DATABASE_URL` | `postgres://mylinear:mylinear@localhost:5432/mylinear?sslmode=disable` |
| Redis 地址 | `REDIS_URL` | `redis://localhost:6379/0` |
| 服务端口 | `PORT` | `8080` |
| MinIO 端点 | `MINIO_ENDPOINT` | `localhost:9000` |
| MinIO 用户 | `MINIO_ACCESS_KEY` | `minioadmin` |
| MinIO 密码 | `MINIO_SECRET_KEY` | `minioadmin` |

**替代方案**：
- 纯 `os.Getenv`：最简单，但本地开发需要手动 `export` 或在 shell 中 `source .env`
- Viper 配置库：功能丰富（支持 YAML、TOML、远程配置等），但对 MVP 过重
- YAML 配置文件：需要额外的文件管理

### D6：Redis 客户端选择 go-redis

**决策**：使用 `github.com/redis/go-redis/v9` 作为 Redis 客户端。

**理由**：
- Go 生态中最主流的 Redis 客户端
- 支持连接池、Pipeline、Pub/Sub
- 类型安全的 API

**替代方案**：
- `redigo`：更底层，API 不够现代

## API 端点设计

本阶段仅实现一个健康检查端点：

```
GET /api/v1/health
```

**响应格式：**
```json
// 正常
{
  "status": "ok"
}

// 异常
{
  "status": "error",
  "message": "database unavailable"
}
```

## 项目目录结构设计

```
mylinear/
├── server/                      # Go 后端
│   ├── cmd/
│   │   └── server/
│   │       └── main.go          # 入口：初始化配置、DB、Redis、路由，启动服务器
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go        # 配置结构体和加载逻辑
│   │   ├── handler/
│   │   │   └── health.go        # 健康检查处理器
│   │   ├── middleware/          # 预留中间件目录
│   │   ├── model/               # 预留数据模型目录
│   │   ├── service/             # 预留业务逻辑目录
│   │   └── store/               # 预留数据存储目录
│   ├── migrations/              # 预留数据库迁移目录
│   ├── go.mod
│   └── go.sum
├── web/                         # React 前端
│   ├── src/
│   │   ├── components/          # 预留 UI 组件目录
│   │   ├── pages/               # 预留页面目录
│   │   ├── stores/              # Zustand 状态管理
│   │   ├── api/
│   │   │   └── client.ts        # API 客户端封装
│   │   ├── lib/                 # 预留工具函数目录
│   │   ├── App.tsx              # 根组件
│   │   ├── main.tsx             # 入口
│   │   └── index.css            # 全局样式
│   ├── public/
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   ├── components.json          # shadcn/ui 配置
│   └── vite.config.ts           # Vite 配置（含 API 代理）
├── docker-compose.yml           # 容器编排
├── Caddyfile                    # Caddy 反向代理配置
├── Makefile                     # 构建脚本
├── .env.example                 # 环境变量示例
└── .gitignore                   # Git 忽略规则
```

## Docker Compose 服务编排

| 服务 | 镜像 | 端口映射 | 持久化 | Profile |
|------|------|---------|--------|--------|
| `postgres` | `postgres:16-alpine` | `5432:5432` | `postgres_data:/var/lib/postgresql/data` | 默认 |
| `redis` | `redis:7-alpine` | `6379:6379` | `redis_data:/data` | 默认 |
| `minio` | `minio/minio:latest` | `9000:9000`, `9001:9001`（控制台） | `minio_data:/data` | 默认 |
| `caddy` | `caddy:2-alpine` | `80:80` | 无 | `proxy`（需 `--profile proxy` 显式启用） |

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|---------|
| GORM 在复杂查询场景下性能不及 raw SQL | 本阶段仅配置连接，C02 阶段评估是否需要 sqlc 补充 |
| Docker Compose 在大规模部署中不够灵活 | 当前为私有部署场景，Docker Compose 足够；未来可迁移至 Kubernetes |
| Caddy 开发模式下 HTTP 无加密 | 开发阶段可接受，生产部署时 Caddy 自动启用 HTTPS |
| MinIO 在开发阶段可能未被使用 | 仅启动服务但不配置连接，C19（附件上传）时再集成 |
