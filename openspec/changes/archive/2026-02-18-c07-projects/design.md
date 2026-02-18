# C07 — Projects 管理 技术设计

## Context

### 背景
当前系统已完成 Issue 核心 CRUD（C05），Issue 模型已支持 `ProjectID` 字段关联项目。但 Project 相关的 CRUD API 和前端视图尚未实现。本 change 需要补齐项目管理的完整能力。

### 现有状态
- **数据模型**：`Project` 模型已在 C02 定义，包含 `WorkspaceID`、`Status`、`LeadID` 等字段
- **Issue 关联**：Issue 模型已有 `ProjectID` 字段，可关联项目
- **权限体系**：Workspace/Team 级权限已实现

### 约束
- 遵循现有的 handler → service → store 三层架构
- 使用 Gin 框架 + GORM
- API 路径格式：`/api/v1/<resource>`
- 需要通过 TDD 方式实现（先写测试）

---

## Goals / Non-Goals

### Goals
- 实现 Project CRUD API（创建、读取、更新、删除）
- 实现项目进度统计 API（基于关联 Issue 计算完成百分比）
- 实现项目关联 Issue 列表查询
- 实现前端项目列表页和详情页
- 权限控制：团队成员可创建/查看，Admin 可删除

### Non-Goals
- 不实现跨团队项目的高级功能（Phase 2）
- 不实现项目模板功能（Phase 3）
- 不实现项目里程碑（C18）
- 不实现项目导入导出（C32）

---

## Decisions

### D1: 项目归属模型（Project Scope）

**背景**：现有 `Project` 模型使用 `WorkspaceID` 而非 `TeamID`，且包含 `Teams pq.StringArray` 字段。

**决策**：保持现有模型不变，Project 是 **Workspace 级资源**，但支持关联多个 Team。

**理由**：
1. Linear 的 Project 本质上是跨团队的概念，一个项目可包含多个团队的工作
2. 现有模型设计已考虑这一点，`Teams` 字段支持多团队关联
3. API 层面提供 `/api/v1/teams/:teamId/projects` 便捷入口，返回与该团队相关的项目

**替代方案**：
- 方案 A：改为 `TeamID` 单团队归属 → 限制性太强，不符合 Linear 设计
- 方案 B：同时保留 `WorkspaceID` 和 `TeamID` → 数据冗余，查询复杂

### D2: API 路径设计

**决策**：采用混合路径设计

| 路径 | 说明 |
|------|------|
| `GET /api/v1/teams/:teamId/projects` | 获取与团队关联的项目列表（便捷入口） |
| `POST /api/v1/workspaces/:workspaceId/projects` | 在工作区创建项目 |
| `GET /api/v1/projects/:id` | 获取项目详情 |
| `PUT /api/v1/projects/:id` | 更新项目 |
| `DELETE /api/v1/projects/:id` | 删除项目 |
| `GET /api/v1/projects/:id/issues` | 获取项目关联的 Issue |
| `GET /api/v1/projects/:id/progress` | 获取项目进度统计 |

**理由**：
- 团队视角的项目列表是高频操作，提供便捷路径
- 项目创建需要 workspace 上下文（Project 是 workspace 级资源）
- 项目操作使用 `/projects/:id` 统一前缀

### D3: 进度统计算法

**决策**：实时计算，不缓存

```go
type ProjectProgress struct {
    TotalIssues      int     `json:"total_issues"`
    CompletedIssues  int     `json:"completed_issues"`
    CancelledIssues  int     `json:"cancelled_issues"`
    ProgressPercent  float64 `json:"progress_percent"`
}
```

**计算规则**：
```
effectiveTotal = totalIssues - cancelledIssues
progressPercent = (completedIssues / effectiveTotal) * 100
if effectiveTotal == 0 → progressPercent = 0
```

**理由**：
1. 项目 Issue 数通常不会很大（几十到几百），实时计算性能可接受
2. 避免缓存一致性问题
3. 后续如需优化可添加 Redis 缓存

**替代方案**：
- 方案 A：在 Project 表存储 `progress_percent` 字段 → Issue 变更时需同步更新，复杂度高
- 方案 B：定时任务批量计算 → 实时性差

### D4: 前端状态管理

**决策**：使用 Zustand 创建独立的 `projectStore`

```typescript
interface ProjectStore {
  projects: Project[];
  currentProject: Project | null;
  loading: boolean;

  fetchTeamProjects: (teamId: string) => Promise<void>;
  fetchProject: (id: string) => Promise<void>;
  createProject: (data: CreateProjectDTO) => Promise<Project>;
  updateProject: (id: string, data: UpdateProjectDTO) => Promise<void>;
  deleteProject: (id: string) => Promise<void>;
}
```

**理由**：
1. 与现有 `issueStore`、`teamStore` 模式一致
2. 项目数据相对独立，不需要复杂的跨 store 交互

### D5: 项目状态转换规则

**决策**：状态转换无严格限制，允许任意状态间切换

```
planned ⇄ in_progress ⇄ paused ⇄ completed
    ↓         ↓           ↓          ↓
    └─────────→ cancelled ←──────────┘
```

**状态时间戳**：
- 切换到 `completed` → 设置 `completed_at`
- 切换到 `cancelled` → 设置 `cancelled_at`（需新增字段）
- 从 `completed`/`cancelled` 切换到其他状态 → 清除对应时间戳

**理由**：
1. Linear 的项目状态是手动控制的，不强制线性流转
2. 提供灵活性，适应不同团队的工作方式

---

## Architecture

### 后端分层

```
┌─────────────────────────────────────────────────────────┐
│                    Handler Layer                         │
│  project_handler.go                                      │
│  - CreateProject, GetProject, UpdateProject, DeleteProject │
│  - ListTeamProjects, ListProjectIssues, GetProjectProgress │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                    Service Layer                         │
│  project_service.go                                      │
│  - 业务逻辑：权限校验、状态转换、进度计算               │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                    Store Layer                           │
│  project_store.go                                        │
│  - 数据访问：CRUD、关联查询、进度统计                    │
└─────────────────────────────────────────────────────────┘
```

### 数据模型关系

```
Workspace 1 ──── * Project * ──── * Team
                     │
                     │ 1
                     │
                     * n
                   Issue
```

### 前端组件结构

```
src/
├── pages/
│   ├── ProjectsPage.tsx          # 项目列表页
│   └── ProjectDetailPage.tsx     # 项目详情页
├── components/
│   └── projects/
│       ├── ProjectCard.tsx       # 项目卡片组件
│       ├── ProjectForm.tsx       # 项目表单（创建/编辑）
│       ├── ProjectStatusBadge.tsx # 状态徽章
│       └── ProgressBar.tsx       # 进度条
├── api/
│   └── projects.ts               # API 调用
└── stores/
    └── projectStore.ts           # Zustand store
```

---

## Risks / Trade-offs

### R1: 跨团队项目可见性
- **风险**：用户可能看到不属于其团队的项目
- **缓解**：在 `ListTeamProjects` 中过滤 `Teams` 数组包含当前团队的项目

### R2: 进度统计性能
- **风险**：大型项目（1000+ Issue）实时计算可能较慢
- **缓解**：
  1. 初期不处理，观察实际性能
  2. 如需优化，添加 Redis 缓存（TTL 5分钟）
  3. Issue 变更时异步刷新缓存

### R3: 项目删除与 Issue 关联
- **风险**：删除项目后 Issue 的 `project_id` 成为悬空引用
- **缓解**：
  1. 使用软删除，项目记录保留
  2. Issue 查询时 LEFT JOIN，正确处理已删除项目
  3. 前端展示"已删除项目"标记

---

## Migration Plan

### 数据库变更
无需数据库迁移（`projects` 表已在 C02 创建）。

**新增字段**（可选）：
- `cancelled_at TIMESTAMPTZ` - 记录取消时间

```sql
ALTER TABLE projects ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ;
```

### 部署步骤
1. 部署后端 API 变更
2. 部署前端页面变更
3. 验证功能正常

### 回滚策略
- 后端：回滚到上一版本镜像
- 前端：回滚静态资源版本
- 数据库：`cancelled_at` 字段可保留（不影响旧版本）

---

## Open Questions

1. **Q: 项目归档功能是否需要在 MVP 阶段实现？**
   - 当前决策：暂不实现，使用软删除替代
   - Phase 2 可考虑增加归档状态

2. **Q: 是否需要项目排序/置顶功能？**
   - 当前决策：暂不实现，按 `updated_at` 倒序
   - 可通过 `position` 字段扩展（类似 Issue）

3. **Q: 项目描述的 Markdown 编辑器使用哪个组件？**
   - 建议：复用 Issue 描述的编辑器组件（待确认 C05 实现）
