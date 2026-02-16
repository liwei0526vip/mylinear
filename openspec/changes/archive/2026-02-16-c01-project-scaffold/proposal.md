## Why

MyLinear 项目当前仅有产品文档和 OpenSpec 配置，尚未初始化任何代码或基础设施。作为整个 Phase 1 的起点（42 个 change 中的第 1 个），本 change 需要搭建完整的项目脚手架——包括 Go 后端、React 前端、Docker Compose 容器编排及构建脚本，为后续所有功能开发提供可运行的基础框架。

没有这个基础设施，C02~C13 的所有 change 都无法开展。

## What Changes

- **Go 后端项目初始化**：创建 Go module，搭建标准目录结构（`cmd/`、`internal/handler,service,store,model`、`migrations/`），集成 Gin 框架，实现健康检查端点 `/api/v1/health`
- **React 前端项目初始化**：使用 Vite + TypeScript 初始化项目，配置 shadcn/ui（Radix UI）组件库，集成 Zustand 状态管理，配置 API 代理
- **Docker Compose 容器编排**：编排 PostgreSQL 16、Redis 7、MinIO（S3 兼容）、Caddy 反向代理等服务容器
- **Caddy 反向代理配置**：`/api/*` 路由至后端，`/*` 路由至前端
- **Makefile 构建脚本**：标准化 `make dev`、`make build`、`make test` 等常用命令
- **`.gitignore` 更新**：补充 Go、Node.js、IDE 相关的忽略规则

## Capabilities

### New Capabilities

- `project-infra`: 项目基础设施——Docker Compose 编排、Caddy 反向代理配置、Makefile 构建脚本、开发环境启动流程
- `backend-scaffold`: Go 后端脚手架——项目结构、Gin 框架集成、健康检查端点、配置管理基础
- `frontend-scaffold`: React 前端脚手架——Vite + TypeScript 构建、shadcn/ui 组件库、Zustand 状态管理、API 层基础

### Modified Capabilities

（无，这是项目的第一个 change，不存在需要修改的已有 capability）

## Impact

- **新增代码**：`server/` 目录（Go 后端全部代码）、`web/` 目录（React 前端全部代码）
- **新增配置**：`docker-compose.yml`、`Caddyfile`、`Makefile`
- **依赖引入**：
  - 后端：`github.com/gin-gonic/gin`（HTTP 框架）
  - 前端：`react`、`vite`、`typescript`、`@radix-ui/*`、`zustand`
- **基础设施**：PostgreSQL 16、Redis 7、MinIO、Caddy 容器
- **前置依赖**：无（这是第一个 change）
- **后续 change 依赖此 change**：C02（数据模型与数据库）依赖本 change 提供的项目结构和数据库连接基础

**对应路线图功能项：** #85（Web App）+ 基础设施 4 项
