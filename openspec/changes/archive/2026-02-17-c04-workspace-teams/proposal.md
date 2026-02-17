# C04 — Workspace 与 Teams

## Why

MyLinear 需要工作区（Workspace）和团队（Team）作为顶层组织单元，支撑后续 Issue、Project、Cycle 等核心功能的开发。工作区是最高层级组织，团队归属于工作区，Issue 归属于团队。团队标识符（如 `ENG`）将用于生成 Issue ID（如 `ENG-123`），这是 Issue 管理的基础设施。

此 change 是 **C05 Issue 核心 CRUD** 和 **C06 工作流状态与标签** 的前置依赖。

## What Changes

### 后端 API

- **Workspace 管理**
  - 创建/读取/更新工作区 API
  - 工作区基本设置（名称、Logo、Slug）
- **Team 管理**
  - 创建/读取/更新/删除团队 API
  - 团队基本设置（名称、标识符 Key、图标、时区）
  - 团队标识符 Key 在工作区内唯一，用于生成 Issue ID
- **团队成员管理**
  - 添加/移除团队成员 API
  - 团队成员角色（Team Owner / Member）
  - 查询团队成员列表 API

### 前端 UI

- **Workspace 设置页面**
  - 工作区名称、Logo 编辑
  - 工作区基本配置
- **Teams 管理页面**
  - 团队列表展示
  - 创建/编辑/删除团队
  - 团队标识符 Key 编辑（用于 Issue ID 生成）
- **团队成员管理 UI**
  - 成员列表展示
  - 添加/移除成员
  - 成员角色分配

### 数据模型

- `workspaces` 表：已在 C02 定义，本 change 实现完整 CRUD
- `teams` 表：已在 C02 定义，本 change 实现完整 CRUD
- `team_members` 表：已在 C02 定义，本 change 实现成员管理

## Capabilities

### New Capabilities

- `workspace-management`：工作区管理 API（CRUD + 基本设置）
- `team-management`：团队管理 API（CRUD + 成员管理 + 标识符）

### Modified Capabilities

- `permission-middleware`：扩展权限中间件以支持团队级权限（Team Owner / Member）

## Impact

### 代码影响

| 模块 | 变更 |
|------|------|
| `server/internal/handler/` | 新增 `workspace.go`、`team.go`、`team_member.go` |
| `server/internal/service/` | 新增 `workspace.go`、`team.go`、`team_member.go` |
| `server/internal/store/` | 新增 `workspace.go`、`team.go`、`team_member.go` |
| `web/src/pages/` | 新增 `Settings/Workspace.tsx`、`Settings/Teams.tsx` |
| `web/src/components/` | 新增团队管理相关组件 |
| `web/src/stores/` | 新增 `workspaceStore.ts`、`teamStore.ts` |
| `web/src/api/` | 新增 `workspace.ts`、`team.ts` |

### API 端点

```
# Workspace
GET    /api/v1/workspaces/:id          # 获取工作区详情
PUT    /api/v1/workspaces/:id          # 更新工作区

# Teams
GET    /api/v1/teams                   # 获取团队列表
POST   /api/v1/teams                   # 创建团队
GET    /api/v1/teams/:id               # 获取团队详情
PUT    /api/v1/teams/:id               # 更新团队
DELETE /api/v1/teams/:id               # 删除团队

# Team Members
GET    /api/v1/teams/:id/members       # 获取团队成员列表
POST   /api/v1/teams/:id/members       # 添加团队成员
DELETE /api/v1/teams/:id/members/:uid  # 移除团队成员
PUT    /api/v1/teams/:id/members/:uid  # 更新成员角色
```

### 依赖关系

- **前置依赖**：C03（用户认证与权限）✅ 已完成
- **后续依赖此 change**：C05（Issue 核心 CRUD）、C06（工作流状态与标签）

### 风险

1. **团队标识符唯一性**：需确保 Key 在工作区内唯一，Issue ID 生成时需依赖此约束
2. **权限边界**：需明确 Workspace 级权限和 Team 级权限的边界，为后续权限扩展留空间

## 对应路线图功能项

- #88 Workspace 管理：名称/基本配置
- #89 Teams 管理：创建/成员/权限
- #90 团队标识符：用于生成 Issue ID（如 ENG-123）
