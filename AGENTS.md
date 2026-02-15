@./constitution.md

你是一个资深的全栈工程师，正在协助我开发一个名为 **MyLinear** 的项目管理工具。你的所有行动都必须严格遵守上面导入的项目宪法。

---

## 1. 项目概述

MyLinear 是一个对标 [Linear](https://linear.app) 的项目管理工具，面向公司内部软件团队，支持私有部署。

| 属性 | 说明 |
|------|------|
| 竞品对标 | Linear |
| 部署方式 | Docker Compose 私有部署 |
| 许可证 | MIT |
| 国内适配 | 企微/钉钉/飞书替代 Slack |

---

## 2. 技术栈

### 后端
- **语言**: Go (版本 >= 1.24)
- **Web 框架**: Gin
- **API 风格**: REST
- **ORM**: GORM 或 sqlc
- **实时通信**: gorilla/websocket

### 前端
- **框架**: React + Vite + TypeScript
- **UI 组件库**: shadcn/ui (Radix UI)
- **状态管理**: Zustand

### 数据层
- **数据库**: PostgreSQL 16
- **缓存**: Redis 7
- **文件存储**: MinIO (S3 兼容)

### 基础设施
- **反向代理**: Caddy
- **部署**: Docker Compose

### 构建与测试
- 使用 `Makefile` 进行标准化操作
- 运行所有测试: `make test`
- 构建后端服务: `make build`
- 启动开发环境: `make dev`

---

## 3. 项目结构

```
mylinear/
├── AGENTS.md                # AI Agent 协作指南（本文件）
├── constitution.md          # 项目开发宪法（最高优先级）
├── Makefile                 # 标准化构建脚本
├── docker-compose.yml       # 容器编排
├── docs/                    # 产品文档
│   ├── 路线图.md
│   ├── 竞品分析.md
│   └── openspec-plan.md     # OpenSpec SDD 执行计划
├── openspec/                # OpenSpec SDD 工作目录
│   ├── config.yaml          # 项目级配置
│   ├── specs/               # 系统规格（Source of Truth）
│   └── changes/             # 活跃的变更
├── server/                  # Go 后端
│   ├── cmd/                 # 入口
│   ├── internal/            # 内部包
│   │   ├── handler/         # HTTP 处理器
│   │   ├── service/         # 业务逻辑
│   │   ├── store/           # 数据存储层
│   │   └── model/           # 数据模型
│   └── migrations/          # 数据库迁移
└── web/                     # React 前端
    ├── src/
    │   ├── components/      # UI 组件
    │   ├── pages/           # 页面
    │   ├── stores/          # Zustand 状态
    │   ├── api/             # API 调用层
    │   └── lib/             # 工具函数
    └── public/
```

---

## 4. 开发流程 (OpenSpec SDD)

本项目采用 **OpenSpec SDD（Spec-Driven Development）** 工作流。开发计划详见 `docs/openspec-plan.md`。

### 核心流程

每个 change 按以下顺序产出：

```
proposal → specs (Delta Specs) → design → tasks → implement → verify → archive
```

| 产出物 | 说明 |
|--------|------|
| `proposal.md` | 变更的意图、范围和方案概述 |
| `specs/` | Delta Specs：需求规格（GIVEN/WHEN/THEN 场景） |
| `design.md` | 技术方案、架构决策 |
| `tasks.md` | 分组实现清单（带 checkbox） |
| 代码实现 | 按 tasks.md 逐项实现 |

### 重要规则
- **不要一键执行所有操作**：每个阶段完成后等待用户审阅确认
- **每个 change 聚焦单一领域**：避免交叉耦合
- **specs 是 Source of Truth**：归档时 Delta Specs 合并至 `openspec/specs/`

---

## 5. Git 与版本控制

- **Commit Message 规范**: 严格遵循 Conventional Commits 规范
  - 格式: `<type>(<scope>): <subject>`
  - 提交信息内容必须使用**中文**
  - 示例: `feat(issues): 实现 Issue CRUD API`
- **代码提交**：不要擅自执行 git 操作，除非用户明确指令

---

## 6. 编码规范

### Go 后端
- 遵循 Go 标准项目布局（`cmd/`, `internal/`）
- 错误必须显式处理，使用 `fmt.Errorf("...: %w", err)` 包装
- 禁止全局变量传递状态，依赖通过参数或结构体注入
- 测试优先采用**表格驱动测试**风格
- 优先编写集成测试，使用真实依赖

### React 前端
- 使用函数式组件 + Hooks
- 状态管理统一使用 Zustand store
- 组件命名使用 PascalCase，文件名使用 kebab-case
- 样式使用 shadcn/ui 组件 + CSS 变量

### 通用
- 数据库主键使用 UUID
- API 路径格式: `/api/v1/<resource>`
- 日期时间使用 UTC，前端展示时转换为本地时区

---

## 7. AI 协作指令

- **当被要求编写测试时**: 优先编写**表格驱动测试（Table-Driven Tests）**
- **当被要求构建项目时**: 优先使用 `Makefile` 中定义好的命令
- **当被要求开发新功能时**: 遵循 OpenSpec SDD 流程，先 proposal 再 specs 再 design 再 tasks
- **当涉及数据库变更时**: 必须通过迁移文件（migrations），禁止手动修改 schema
- **当不确定需求范围时**: 主动询问，而不是猜测实现
