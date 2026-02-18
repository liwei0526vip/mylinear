## Why

Issue 是项目管理系统的核心实体。当前系统已具备 Issue 数据模型（C02）和工作流状态（C06），但缺少 Issue 的业务 API 和前端交互界面。本 change 实现完整的 Issue CRUD 能力，使团队能够创建、管理、追踪日常工作项。

这是 MVP 的核心功能，是后续 Projects、Cycles、Views 等功能的基础。

## What Changes

### 后端
- 新增 Issue CRUD API（创建、读取、更新、删除/归档）
- 实现 Issue 标识符自动生成（Team key + 序号，如 `ENG-123`）
- 支持拖拽排序（position 字段）
- 实现 Issue 订阅/取消订阅机制
- 支持按状态/优先级/负责人/标签/项目等条件查询和过滤

### 前端
- Issue 创建模态框（支持 `Cmd+C` 快捷键）
- Issue 详情面板（右侧面板，支持全屏模式）
- Issue 基础信息编辑（标题、描述、状态、优先级、负责人、标签、截止日期）

### 数据层
- 新增 `position` 字段到 Issue 模型（支持拖拽排序）
- 新增 `issue_subscriptions` 表（用户订阅关系）

## Capabilities

### New Capabilities

- `issue-crud`: Issue 核心 CRUD 操作，包括创建、读取、更新、删除/归档，以及标识符生成（`ENG-123` 格式）、拖拽排序、条件过滤查询
- `issue-subscription`: Issue 订阅机制，支持用户手动订阅/取消订阅，自动订阅规则（创建者、负责人自动订阅）

### Modified Capabilities

- `gorm-models`: 为 Issue 模型新增 `position` 字段（FLOAT 类型），新增 `IssueSubscription` 模型

## Impact

### 后端
- 新增 `server/internal/handler/issue.go` - Issue HTTP 处理器
- 新增 `server/internal/service/issue.go` - Issue 业务逻辑
- 新增 `server/internal/store/issue.go` - Issue 数据访问层
- 修改 `server/internal/model/issue.go` - 添加 position 字段和订阅关联
- 新增数据库迁移 - 添加 `position` 列和 `issue_subscriptions` 表

### 前端
- 新增 `web/src/components/issues/IssueCreateModal.tsx` - Issue 创建模态框
- 新增 `web/src/components/issues/IssueDetailPanel.tsx` - Issue 详情面板
- 新增 `web/src/stores/issueStore.ts` - Issue 状态管理
- 新增 `web/src/api/issues.ts` - Issue API 调用层

### API 端点
- `POST /api/v1/teams/:teamId/issues` - 创建 Issue
- `GET /api/v1/teams/:teamId/issues` - 列表查询（支持过滤、排序）
- `GET /api/v1/issues/:id` - 获取单个 Issue
- `PUT /api/v1/issues/:id` - 更新 Issue
- `DELETE /api/v1/issues/:id` - 删除/归档 Issue
- `POST /api/v1/issues/:id/subscribe` - 订阅 Issue
- `DELETE /api/v1/issues/:id/subscribe` - 取消订阅
- `PUT /api/v1/issues/:id/position` - 更新排序位置

### 依赖
- 依赖 C04（Workspace 与 Teams）- Issue 属于 Team
- 依赖 C06（工作流状态与标签）- Issue 关联 WorkflowState 和 Label

### 路线图功能项
- #27 Issue CRUD
- #28 Issue 标题和描述
- #29 Issue 状态
- #30 Issue 优先级
- #31 Issue 负责人
- #32 Issue 标签
- #33 Issue 截止日期
- #42 订阅/取消订阅
