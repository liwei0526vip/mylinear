# C06: Workflow States & Labels (Proposal)

## 1. Intent / 意图

构建 MyLinear 的 Issue 状态管理和标签分类系统。这是 Issue 追踪的核心基础，直接影响到 Issue 的生命周期管理。
目标是提供一套可定制但有标准约束（5 种核心类型）的工作流系统，以及轻量级的标签分类机制。这直接支持了 OpenSpec Plan 中的 C06 规划。

## 2. Scope / 范围

### In Scope

#### Backend (Golang)
-   **WorkflowState CRUD** (路线图 #29, #38)
    -   Schema: `id`, `team_id`, `name`, `type` (backlog, unstarted, started, completed, canceled), `color`, `position`, `description`.
    -   API: `GET`, `POST`, `PUT`, `DELETE` (带有关联检查).
    -   Validation: 每种类型至少保留一个状态；删除状态时需指定迁移目标或确认无 Issue 使用。
-   **Default Workflow Initialization**
    -   Team 创建时自动生成默认状态集 (Backlog -> Todo -> In Progress -> Done -> Canceled).
-   **Labels CRUD** (路线图 #32)
    -   Schema: `id`, `team_id` (null for workspace), `name`, `color`, `description`.
    -   API: `GET` (支持按 Team 过滤), `POST` (支持 Workspace/Team 级), `PUT`, `DELETE`.

#### Frontend (React)
-   **Status Icons**: 5 种状态类型的标准 SVG 图标组件（虚线圆、半圆、实心圆、勾选圆、叉号圆）。
-   **Workflow Settings UI**: 团队设置页面的工作流管理列表（增删改）。
-   **Labels Settings UI**: 团队设置页面的标签管理列表（增删改）。

### Out of Scope
-   **Issue CRUD**: 具体 Issue 的创建和状态分配在 C05 实现。
-   **SLA / Time in Status**: 状态停留时间追踪在 C16 实现。
-   **Automation**: 状态自动流转（如合并 PR 自动完成）在 C26 实现。
-   **Kanban Board**: 看板视图和拖拽排序 UI 在 C11 实现（本阶段仅支持列表管理和基本的 Position 字段）。

## 3. Approach / 方案

-   **Backend Architecture**: 
    -   遵循现有的 `internal/handler`, `internal/service`, `internal/store` 分层架构。
    -   数据模型设计遵循 `docs/openspec-plan.md` 中的 C02 定义。
    -   使用 GORM Hooks 处理默认状态初始化。
-   **Database Strategy**: 
    -   `workflow_state.type` 存储为字符串，应用层做枚举校验。
    -   `position` 字段使用 `float64` 实现，支持中间插入排序。
-   **Frontend Architecture**:
    -   使用 Zustand 存储 `WorkflowStore` 和 `LabelStore`。
    -   UI 组件复用 shadcn/ui 的 Table, Dialog, Form。
    -   图标系统构建在 `lucide-react` 之上或作为独立 SVG 组件。

## 4. Dependencies / 依赖

-   [x] **C04 (Workspace & Teams)**: `teams` 表已存在，作为外键依赖。
-   [ ] **C05 (Issue CRUD)**: C06 是 C05 的前置依赖，Issue 的创建依赖于 WorkflowState 和 Labels 的存在。
