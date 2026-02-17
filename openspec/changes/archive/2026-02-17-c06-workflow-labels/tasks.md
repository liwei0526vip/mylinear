# C06: Workflow States & Labels 任务清单

> **TDD 开发原则**：严格遵循 Red-Green-Refactor 循环
> - 🔴 Red：先写失败的测试
> - 🟢 Green：写最少的代码让测试通过
> - 🔵 Refactor：重构代码（保持测试通过）

---

## 1. 后端 - 数据模型与迁移层

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

- [x] 1.1 🔴 编写 `db/migrations` 测试（验证表结构、Type枚举、(team_id, name)唯一性、Labels partial索引）
- [x] 1.2 🟢 实现 `workflow_states` 和 `labels` 表的 SQL 迁移文件
- [x] 1.3 🔴 编写 `WorkflowState` GORM 模型测试（验证 CRUD 基础映射）
- [x] 1.4 🟢 实现 `WorkflowState` GORM 模型结构体（server/internal/model/workflow_state.go）
- [x] 1.5 🔴 编写 `Label` GORM 模型测试（验证 CRUD 基础映射）
- [x] 1.6 🟢 实现 `Label` GORM 模型结构体（server/internal/model/label.go）

---

## 2. 后端 - Workflow Store 层

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

- [x] 2.1 🔴 编写 `WorkflowStore.Create` 集成测试（表格驱动：正常创建、同名冲突、非法Type）
- [x] 2.2 🟢 实现 `WorkflowStore.Create`
- [x] 2.3 🔴 编写 `WorkflowStore.List` 集成测试（表格驱动：按Type分组排序、Position升序）
- [x] 2.4 🟢 实现 `WorkflowStore.List`
- [x] 2.5 🔴 编写 `WorkflowStore.Update` 集成测试（表格驱动：更新名称、更新Position、重名检测）
- [x] 2.6 🟢 实现 `WorkflowStore.Update`
- [x] 2.7 🔴 编写 `WorkflowStore.Delete` 集成测试（表格驱动：正常删除、ID不存在）
- [x] 2.8 🟢 实现 `WorkflowStore.Delete`
- [x] 2.9 🔴 编写 `WorkflowStore.CountByType` 集成测试（表格驱动：统计各Type数量，用于删除校验）
- [x] 2.10 🟢 实现 `WorkflowStore.CountByType`

---

## 3. 后端 - Label Store 层

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

- [x] 3.1 🔴 编写 `LabelStore.Create` 集成测试（表格驱动：Team级标签、Workspace级标签、重名检测）
- [x] 3.2 🟢 实现 `LabelStore.Create`
- [x] 3.3 🔴 编写 `LabelStore.List` 集成测试（表格驱动：混合Team和Workspace标签、仅Workspace标签）
- [x] 3.4 🟢 实现 `LabelStore.List`
- [x] 3.5 🔴 编写 `LabelStore.Update` 集成测试（表格驱动：更新名称颜色）
- [x] 3.6 🟢 实现 `LabelStore.Update`
- [x] 3.7 🔴 编写 `LabelStore.Delete` 集成测试（表格驱动：正常删除）
- [x] 3.8 🟢 实现 `LabelStore.Delete`

---

## 4. 后端 - Service 层

> 使用真实 PostgreSQL 数据库进行集成测试

- [x] 4.1 🔴 编写 `WorkflowService.Create` 集成测试（表格驱动：自动计算Position、默认颜色）
- [x] 4.2 🟢 实现 `WorkflowService.Create`
- [x] 4.3 🔴 编写 `WorkflowService.Delete` 校验测试（表格驱动：删除最后一个状态拒绝、正常删除）
- [x] 4.4 🟢 实现 `WorkflowService.Delete`（含前置校验逻辑）
- [x] 4.5 🔴 编写 `TeamService` 扩展测试：创建Team时自动初始化默认工作流
- [x] 4.6 🟢 实现 `TeamService` 默认工作流初始化逻辑 (扩展 `CreateTeam`)
- [x] 4.7 🔴 编写 `LabelService` 通用测试（业务逻辑较少，主要是透传和权限校验准备）
- [x] 4.8 🟢 实现 `LabelService` CRUD 方法

---

## 5. 后端 - Handler 层

> 使用 `httptest` 进行端到端 HTTP 测试

- [x] 5.1 🔴 编写 `GET /api/v1/teams/:id/workflow-states` 测试（表格驱动：JSON 结构验证）
- [x] 5.2 🟢 实现 `WorkflowHandler.List`
- [x] 5.3 🔴 编写 `POST /api/v1/teams/:id/workflow-states` 测试（表格驱动：创建状态、参数校验）
- [x] 5.4 🟢 实现 `WorkflowHandler.Create`
- [x] 5.5 🔴 编写 `PUT /api/v1/workflow-states/:id` 测试
- [x] 5.6 🟢 实现 `WorkflowHandler.Update`
- [x] 5.7 🔴 编写 `DELETE /api/v1/workflow-states/:id` 测试（表格驱动：正常删除、非法删除返回400）
- [x] 5.8 🟢 实现 `WorkflowHandler.Delete`
- [x] 5.9 🔴 编写 `GET /api/v1/teams/:id/labels` 测试（表格驱动：验证返回混合标签）
- [x] 5.10 🟢 实现 `LabelHandler.List`
- [x] 5.11 🔴 编写 `POST /api/v1/teams/:id/labels` 测试（表格驱动：创建Team标签）
- [x] 5.12 🟢 实现 `LabelHandler.Create`
- [x] 5.13 🔴 编写路由注册测试
- [x] 5.14 🟢 注册所有 Workflow 和 Label 相关路由

---

## 6. 前端 - API 与 Store 层

- [x] 6.1 创建 `web/src/types/workflow.ts` 定义 WorkflowState, Label 类型
- [x] 6.2 创建 `web/src/api/workflow.ts` 实现 Workflow API 请求
- [x] 6.3 创建 `web/src/api/label.ts` 实现 Label API 请求
- [x] 6.4 创建 `web/src/stores/workflowStore.ts` (Zustand: states map, actions)
- [x] 6.5 创建 `web/src/stores/labelStore.ts` (Zustand: labels map, actions)

---

## 7. 前端 - 组件与页面开发

- [x] 7.1 创建 `WorkflowIcon` 组件（支持 5 种 Type 的 SVG）
- [x] 7.2 创建 `StateBadge` 和 `LabelBadge` 组件
- [x] 7.3 创建 `web/src/components/settings/WorkflowList.tsx`（状态列表、分组展示）
- [x] 7.4 创建 `web/src/components/settings/CreateStateDialog.tsx`（添加状态表单）
- [x] 7.5 创建 `web/src/components/settings/LabelList.tsx`（标签列表，区分 Team/Workspace）
- [x] 7.6 创建 `web/src/components/settings/CreateLabelDialog.tsx`
- [x] 7.7 集成到 `web/src/pages/Settings/TeamDetail.tsx`（增加 Workflow 和 Label Tab 页）

---

## 8. 端到端验证

- [x] 8.1 🔵 运行完整后端测试套件 `make test`
- [x] 8.2 🔵 启动环境：创建新 Team，验证默认 5 个状态是否自动生成
- [x] 8.3 🔵 验证状态管理：添加/编辑/删除操作，尝试删除最后一个状态确认报错
- [x] 8.4 🔵 验证标签管理：创建 Team 标签和 Workspace 标签，确认列表合并显示

---

**任务统计**：
- 后端 TDD 任务：44 个
- 前端任务：11 个
- 验证任务：4 个
- **总计：59 个任务**
