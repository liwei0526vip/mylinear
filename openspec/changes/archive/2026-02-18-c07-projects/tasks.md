# C07 — Projects 管理 任务清单

> **TDD 开发原则**：严格遵循 Red-Green-Refactor 循环
> - 🔴 Red：先写失败的测试
> - 🟢 Green：写最少的代码让测试通过
> - 🔵 Refactor：重构代码（保持测试通过）

---

## 1. Store 层 - Project CRUD

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

- [x] 1.1 🔴 编写 `project_store_test.go` - CreateProject 测试（表格驱动：正常创建、名称为空、描述超长）
- [x] 1.2 🟢 实现 `CreateProject` store 方法
- [x] 1.3 🔴 编写 `GetProject` 测试（表格驱动：正常获取、不存在的ID、已删除的项目）
- [x] 1.4 🟢 实现 `GetProject` store 方法
- [x] 1.5 🔴 编写 `UpdateProject` 测试（表格驱动：更新名称、更新状态、更新负责人、更新日期）
- [x] 1.6 🟢 实现 `UpdateProject` store 方法
- [x] 1.7 🔴 编写 `DeleteProject` 测试（表格驱动：软删除、恢复删除）
- [x] 1.8 🟢 实现 `DeleteProject` store 方法

---

## 2. Store 层 - Project 查询

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

- [x] 2.1 🔴 编写 `ListProjectsByWorkspace` 测试（表格驱动：空列表、多条记录、分页）
- [x] 2.2 🟢 实现 `ListProjectsByWorkspace` store 方法
- [x] 2.3 🔴 编写 `ListProjectsByTeam` 测试（表格驱动：按团队过滤、按状态过滤）
- [x] 2.4 🟢 实现 `ListProjectsByTeam` store 方法
- [x] 2.5 🔴 编写 `GetProjectProgress` 测试（表格驱动：无Issue、部分完成、全部完成、有取消Issue）
- [x] 2.6 🟢 实现 `GetProjectProgress` store 方法
- [x] 2.7 🔴 编写 `ListProjectIssues` 测试（表格驱动：空列表、多条记录、按状态过滤、分页）
- [x] 2.8 🟢 实现 `ListProjectIssues` store 方法

---

## 3. Service 层 - Project 业务逻辑

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

- [x] 3.1 🔴 编写 `project_service_test.go` - CreateProject 测试（表格驱动：正常创建、leadId验证失败、日期逻辑验证失败）
- [x] 3.2 🟢 实现 `CreateProject` service 方法（含负责人验证、日期逻辑验证）
- [x] 3.3 🔴 编写 `UpdateProject` 测试（表格驱动：状态转换到completed设置时间戳、转换到cancelled设置时间戳）
- [x] 3.4 🟢 实现 `UpdateProject` service 方法（含状态时间戳处理）
- [x] 3.5 🔴 编写 `DeleteProject` 测试（表格驱动：权限校验-非Admin拒绝）
- [x] 3.6 🟢 实现 `DeleteProject` service 方法（含权限校验）

---

## 4. Handler 层 - HTTP 接口

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

- [x] 4.1 🔴 编写 `project_handler_test.go` - CreateProject HTTP 测试（表格驱动：成功创建、参数校验失败、未授权）
- [x] 4.2 🟢 实现 `CreateProject` handler（POST /api/v1/workspaces/:workspaceId/projects）
- [x] 4.3 🔴 编写 `ListTeamProjects` HTTP 测试（表格驱动：成功获取、空列表、分页参数）
- [x] 4.4 🟢 实现 `ListTeamProjects` handler（GET /api/v1/teams/:teamId/projects）
- [x] 4.5 🔴 编写 `GetProject` HTTP 测试（表格驱动：成功获取、不存在的ID）
- [x] 4.6 🟢 实现 `GetProject` handler（GET /api/v1/projects/:id）
- [x] 4.7 🔴 编写 `UpdateProject` HTTP 测试（表格驱动：成功更新、部分更新、不存在的ID）
- [x] 4.8 🟢 实现 `UpdateProject` handler（PUT /api/v1/projects/:id）
- [x] 4.9 🔴 编写 `DeleteProject` HTTP 测试（表格驱动：成功删除、权限不足）
- [x] 4.10 🟢 实现 `DeleteProject` handler（DELETE /api/v1/projects/:id）
- [x] 4.11 🔴 编写 `GetProjectProgress` HTTP 测试（表格驱动：成功获取进度）
- [x] 4.12 🟢 实现 `GetProjectProgress` handler（GET /api/v1/projects/:id/progress）
- [x] 4.13 🔴 编写 `ListProjectIssues` HTTP 测试（表格驱动：成功获取、分页、过滤）
- [x] 4.14 🟢 实现 `ListProjectIssues` handler（GET /api/v1/projects/:id/issues）

---

## 5. 路由注册与数据库迁移

- [x] 5.1 在 `router/router.go` 添加 `RegisterProjectRoutes` 函数
- [x] 5.2 在 `cmd/server/main.go` 注册 Project 路由
- [x] 5.3 检查 `projects` 表是否需要添加 `cancelled_at` 字段，如需则创建迁移文件
- [x] 5.4 运行 `make test` 确保所有后端测试通过

---

## 6. 前端 - API 层

- [x] 6.1 创建 `src/api/projects.ts` - 定义 Project、ProjectProgress、CreateProjectDTO 等类型
- [x] 6.2 实现 `fetchTeamProjects(teamId, params)` API 函数
- [x] 6.3 实现 `fetchProject(id)` API 函数
- [x] 6.4 实现 `createProject(workspaceId, data)` API 函数
- [x] 6.5 实现 `updateProject(id, data)` API 函数
- [x] 6.6 实现 `deleteProject(id)` API 函数
- [x] 6.7 实现 `fetchProjectProgress(id)` API 函数
- [x] 6.8 实现 `fetchProjectIssues(id, params)` API 函数

---

## 7. 前端 - 状态管理

- [x] 7.1 创建 `src/stores/projectStore.ts` - Zustand store 定义
- [x] 7.2 实现 `projects` 状态和 `fetchTeamProjects` action
- [x] 7.3 实现 `currentProject` 状态和 `fetchProject` action
- [x] 7.4 实现 `createProject` / `updateProject` / `deleteProject` actions
- [x] 7.5 实现 `progress` 状态和 `fetchProjectProgress` action
- [x] 7.6 实现 `loading` / `error` 状态管理

---

## 8. 前端 - 组件开发

- [x] 8.1 创建 `src/components/projects/` 目录
- [x] 8.2 实现 `ProjectCard.tsx` - 项目卡片组件（展示名称、状态、进度条、负责人、Issue数量）
- [x] 8.3 实现 `ProjectStatusBadge.tsx` - 状态徽章组件（5种状态对应颜色）
- [x] 8.4 实现 `ProgressBar.tsx` - 进度条组件（百分比可视化）
- [x] 8.5 实现 `ProjectForm.tsx` - 项目创建/编辑表单（名称、描述、负责人、日期选择）

---

## 9. 前端 - 页面开发

- [x] 9.1 创建 `src/pages/ProjectsPage.tsx` - 项目列表页骨架
- [x] 9.2 实现项目卡片网格布局展示
- [x] 9.3 实现状态过滤器（全部/Planned/In Progress/Paused/Completed/Cancelled）
- [x] 9.4 实现创建项目模态框（使用 ProjectForm 组件）
- [x] 9.5 实现空状态展示与引导
- [x] 9.6 创建 `src/pages/ProjectDetailPage.tsx` - 项目详情页骨架
- [x] 9.7 实现项目头部信息展示（名称、状态切换器、负责人、日期、进度统计）
- [x] 9.8 实现项目描述区域（Markdown 渲染 + 点击编辑）
- [x] 9.9 实现关联 Issue 列表展示（表格/列表形式）
- [x] 9.10 实现删除项目确认对话框

---

## 10. 前端 - 路由与集成

- [x] 10.1 在 `src/App.tsx` 添加项目相关路由（`/teams/:teamSlug/projects`、`/projects/:projectId`）
- [x] 10.2 在侧边栏添加 Projects 入口（导航到项目列表页）
- [x] 10.3 实现侧边栏最近项目快捷入口（展开显示最近 3-5 个项目）
- [x] 10.4 实现响应式布局适配（桌面端网格、移动端列表）
- [x] 10.5 验证前端 UI 与 Linear 设计规范一致性（颜色、间距、动画）

---

## 11. 集成验证

- [x] 11.1 运行 `make test` 确保所有后端测试通过
- [x] 11.2 手动测试：创建项目 → 查看列表 → 查看详情 → 编辑 → 删除
- [x] 11.3 手动测试：项目进度统计（创建多个 Issue 并完成，验证百分比计算）
- [x] 11.4 手动测试：权限控制（非成员无法访问私有团队项目）
- [x] 11.5 验证 API 响应格式符合设计文档

---

## 任务统计

| 类别                     | 任务数 | 预估工时 |
| ------------------------ | :----: | :------: |
| 后端 TDD 任务（Store）   |   16   |    8h    |
| 后端 TDD 任务（Service） |   6    |    3h    |
| 后端 TDD 任务（Handler） |   14   |    7h    |
| 后端路由与迁移           |   4    |    2h    |
| 前端 API 层              |   8    |    3h    |
| 前端状态管理             |   6    |    2h    |
| 前端组件                 |   5    |    3h    |
| 前端页面                 |   10   |    6h    |
| 前端路由集成             |   5    |    2h    |
| 集成验证                 |   5    |    2h    |
| **总计**                 | **79** | **~38h** |

> **预估工时说明**：基于每个任务约 30 分钟计算，实际可能因复杂度有所浮动
