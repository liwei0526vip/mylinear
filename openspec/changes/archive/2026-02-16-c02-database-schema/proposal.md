## Why

C01 已完成项目脚手架搭建，但系统尚无数据模型。本 change 需要建立完整的数据库 schema，定义所有核心业务实体及其关系，为后续 C03~C13 的业务功能开发提供数据持久化基础。

没有数据模型，用户认证（C03）、工作区/团队（C04）、Issue 管理（C05）等所有业务功能都无法实现。

## What Changes

- **18 张表 DDL 设计**（UUID 主键）
  - 核心业务表（15 张）：`workspaces`、`teams`、`users`、`team_members`、`issues`、`issue_relations`、`workflow_states`、`projects`、`milestones`、`cycles`、`labels`、`comments`、`attachments`、`documents`、`notifications`
  - 辅助表（3 张）：`issue_closure`（闭包表，支持子 Issue 层级查询）、`issue_status_history`（状态变更历史）、`workflow_transitions`（状态转换规则）
- **GORM 模型定义**：为所有表定义 Go struct，包含 GORM 标签
- **数据库迁移文件**：使用 golang-migrate 管理 schema 版本
- **`position FLOAT` 字段**：支持拖拽排序（看板视图、列表排序）
- **索引设计**：Issue 表核心查询索引、GIN 索引（标签数组）
- **JSONB 配置字段**：`settings`、`cycle_settings`、`workflow_settings` 等灵活配置

## Capabilities

### New Capabilities

- `database-schema`: 数据库 schema 定义——18 张表的 DDL、GORM 模型、迁移文件、索引设计
- `gorm-models`: GORM 数据模型——所有表的 Go struct 定义，遵循 UUID 主键、时间戳标准

### Modified Capabilities

（无，这是数据层基础 change，不修改已有 capability）

## Impact

- **新增代码**：
  - `server/internal/model/*.go` — 所有数据模型定义
  - `server/migrations/*.sql` — 数据库迁移文件
- **新增依赖**：
  - `github.com/google/uuid` — UUID 生成
  - `github.com/golang-migrate/migrate` — 数据库迁移工具
- **数据库变更**：创建 18 张表及相关索引
- **前置依赖**：C01（项目脚手架）— 已完成
- **后续 change 依赖此 change**：C03（用户认证）依赖 `users` 表，C04（Workspace 与 Teams）依赖 `workspaces`、`teams`、`team_members` 表

**对应路线图功能项：** 数据模型全部 4 项
