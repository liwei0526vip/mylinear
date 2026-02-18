# C07 — Projects 管理

## Why

当前系统已支持 Issue 的创建和管理，但缺乏将 Issue 组织到项目（Project）中的能力。团队需要一种方式来按项目聚合相关 Issue，追踪项目整体进度，并指定项目负责人。这是从单Issue管理向项目级管理演进的关键能力，是 MVP 阶段的核心功能之一。

## What Changes

- 新增 **Projects CRUD API**：支持创建、读取、更新、删除项目
- 新增 **项目状态管理**：5 种状态（Planned / In Progress / Paused / Completed / Cancelled）
- 新增 **项目进度自动统计**：基于关联 Issue 的完成比例自动计算
- 新增 **项目负责人（Lead）**：指定用户作为项目负责人
- 新增 **项目描述**：支持 Markdown 格式的项目描述
- 新增 **前端 Project 列表页**：展示团队下所有项目
- 新增 **前端 Project 详情页**：展示项目进度、关联 Issue 列表、描述等

## Capabilities

### New Capabilities

- `projects`: 项目 CRUD API，包括项目状态、进度统计、负责人、描述管理
- `project-views`: 前端项目视图，包括列表页和详情页

### Modified Capabilities

- `issues`: Issue 模型需要支持与 Project 的关联关系（已在 C02 数据模型和 C05 Issue CRUD 中定义，本 change 仅实现 API 集成）

## Impact

### 后端
- 新增 `Project` 相关 handler、service、store 层代码
- 新增 `internal/handler/project_handler.go`
- 新增 `internal/service/project_service.go`
- 新增 `internal/store/project_store.go`
- 新增 `/api/v1/projects/*` 路由组

### 前端
- 新增 `src/pages/ProjectsPage.tsx`（项目列表页）
- 新增 `src/pages/ProjectDetailPage.tsx`（项目详情页）
- 新增 `src/api/projects.ts`（API 调用层）
- 新增 `src/stores/projectStore.ts`（Zustand store）

### 数据库
- `projects` 表已在 C02 中定义，无需新增迁移

### API 端点
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/teams/:teamId/projects` | 获取团队项目列表 |
| POST | `/api/v1/teams/:teamId/projects` | 创建项目 |
| GET | `/api/v1/projects/:id` | 获取项目详情 |
| PUT | `/api/v1/projects/:id` | 更新项目 |
| DELETE | `/api/v1/projects/:id` | 删除项目 |
| GET | `/api/v1/projects/:id/issues` | 获取项目关联的 Issue |
| GET | `/api/v1/projects/:id/progress` | 获取项目进度统计 |
