# C04 — Workspace 与 Teams 技术设计

## Context

本 change 实现 MyLinear 的顶层组织结构：Workspace（工作区）和 Team（团队）。

**当前状态**：
- C01 已完成：项目脚手架搭建
- C02 已完成：数据模型定义（`workspaces`、`teams`、`team_members` 表已创建）
- C03 已完成：用户认证与权限中间件（支持 Admin / Member 角色）

**约束**：
- 数据表已在 C02 定义，本 change 实现业务逻辑和 API
- 权限中间件需扩展以支持团队级角色（Team Owner / Member）
- 团队标识符 Key 用于后续 Issue ID 生成，需保证在工作区内唯一

**利益相关者**：
- 后端开发：实现 API 和业务逻辑
- 前端开发：实现设置页面和团队管理 UI

## Goals / Non-Goals

**Goals：**
1. 实现 Workspace CRUD API（读取 + 更新）
2. 实现 Team CRUD API（创建/读取/更新/删除）
3. 实现团队成员管理 API（添加/移除/角色更新）
4. 扩展权限中间件支持团队级权限检查
5. 实现前端 Workspace 设置页面
6. 实现前端 Teams 管理页面和成员管理 UI

**Non-Goals：**
- Workspace 创建 API（用户注册时自动创建，不在本 change 范围）
- 私有团队完整权限隔离（Phase 5 功能）
- Workspace Owner 角色（Phase 4 功能）
- 团队嵌套（parent_id 字段预留，但不实现业务逻辑）
- 跨团队项目支持（Phase 2 功能）

## Decisions

### D1: API 路由设计

**决策**：采用资源嵌套路由设计

```
/api/v1/workspaces/:id                    # Workspace 资源
/api/v1/teams                             # Team 列表（通过 query 参数过滤 workspace）
/api/v1/teams/:id                         # Team 资源
/api/v1/teams/:id/members                 # TeamMember 嵌套资源
/api/v1/teams/:id/members/:uid            # TeamMember 单个成员
```

**理由**：
- 符合 RESTful 约定，资源层级清晰
- Team 通过 `workspace_id` query 参数过滤，而非 URL 路径嵌套，避免深层嵌套
- TeamMember 作为 Team 的子资源，符合领域模型关系

**替代方案**：
- `/api/v1/workspaces/:wid/teams` — 拒绝，嵌套层级过深
- `/api/v1/team-members` — 拒绝，独立资源会导致路由分散

### D2: 团队标识符 Key 格式

**决策**：Key 格式为 `^[A-Z][A-Z0-9]{1,9}$`（2-10 位大写字母和数字，首字母必须为大写字母）

**理由**：
- 大写字母易识别，与 Issue ID 格式（如 ENG-123）保持一致
- 长度限制 2-10 位，保证可读性和唯一性
- 首字母必须为字母，避免纯数字造成混淆

**实现**：
```go
var teamKeyRegex = regexp.MustCompile(`^[A-Z][A-Z0-9]{1,9}$`)

func ValidateTeamKey(key string) error {
    if !teamKeyRegex.MatchString(key) {
        return errors.New("团队标识符必须为大写字母开头，2-10 位大写字母和数字")
    }
    return nil
}
```

### D3: 团队角色存储

**决策**：在 `team_members` 表中使用 `role` 字段存储团队角色（owner / member）

**理由**：
- 团队角色与团队成员关系紧密，存储在同一表减少 JOIN
- 与 C02 定义的表结构一致
- 支持快速角色查询

**数据模型**：
```go
type TeamMember struct {
    TeamID    uuid.UUID `gorm:"type:uuid;primaryKey"`
    UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`
    Role      string    `gorm:"type:varchar(20);not null;default:'member'"` // owner, member
    JoinedAt  time.Time `gorm:"not null"`
}
```

### D4: 权限中间件扩展

**决策**：新增 `RequireTeamOwner()` 和 `RequireTeamMember()` 中间件

**实现方式**：
1. 从路由参数获取 `team_id`（如 `/api/v1/teams/:id/members`）
2. 查询 `team_members` 表获取用户角色
3. Admin/GlobalAdmin 角色绕过团队级权限检查

```go
// RequireTeamOwner 要求用户是团队 Owner 或 Admin
func RequireTeamOwner() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Admin 绕过检查
        if IsAdmin(c) {
            c.Next()
            return
        }

        teamID := getTeamIDFromContext(c)
        role, err := store.GetTeamRole(c, GetCurrentUserID(c), teamID)
        if err != nil || role != "owner" {
            c.JSON(403, gin.H{"error": "需要 Team Owner 权限"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

**替代方案**：
- 在 User 模型中存储团队角色 — 拒绝，不支持一个用户在多个团队有不同角色
- 使用 Redis 缓存团队角色 — 本 change 暂不引入，后续优化时考虑

### D5: 团队删除策略

**决策**：软删除 + 存在 Issue 时禁止删除

**理由**：
- 软删除保证历史数据可追溯
- Issue 依赖 Team 存在，删除前需清空关联 Issue
- 符合数据安全原则

**实现**：
```go
func (s *TeamService) DeleteTeam(ctx context.Context, teamID uuid.UUID) error {
    // 检查是否存在 Issue
    issueCount, err := s.store.CountIssuesByTeam(ctx, teamID)
    if err != nil {
        return err
    }
    if issueCount > 0 {
        return errors.New("团队下存在 Issue，无法删除")
    }

    // 软删除团队
    return s.store.SoftDeleteTeam(ctx, teamID)
}
```

### D6: 前端状态管理

**决策**：使用独立 Zustand store 管理 Workspace 和 Team 状态

```
web/src/stores/
├── workspaceStore.ts   # Workspace 状态
└── teamStore.ts        # Team 状态 + TeamMember 状态
```

**理由**：
- Workspace 和 Team 是独立领域，分离 store 职责清晰
- teamStore 包含团队成员状态，避免过度拆分

### D7: 前端路由设计

**决策**：Settings 页面作为独立路由模块

```
/settings/workspace     # Workspace 设置
/settings/teams         # Teams 列表
/settings/teams/:id     # Team 详情 + 成员管理
```

**理由**：
- Settings 是用户配置入口，独立路由便于权限控制
- 团队详情页包含成员管理，减少路由跳转

## Risks / Trade-offs

### R1: 团队标识符 Key 修改风险

**风险**：修改 Team Key 会影响后续创建的 Issue ID 格式

**缓解措施**：
- 前端显示警告提示："修改团队标识符后，新创建的 Issue 将使用新标识符"
- 后端记录 Key 变更历史（Phase 2 功能）

### R2: 权限中间件性能

**风险**：每次请求查询团队角色会影响性能

**缓解措施**：
- 短期：接受数据库查询开销（用户量不大时可接受）
- 长期：引入 Redis 缓存团队角色（需在角色变更时失效缓存）

### R3: 最后一个 Team Owner 删除

**风险**：删除最后一个 Owner 会导致团队无管理员

**缓解措施**：
- 后端强制校验：团队必须至少有一个 Owner
- 前端显示警告并禁用删除按钮

## Migration Plan

本 change 无需数据库迁移（表已在 C02 创建）。

**部署步骤**：
1. 部署后端 API
2. 部署前端页面
3. 验证功能可用性

**回滚策略**：
- 后端回滚：恢复上一版本 API
- 前端回滚：恢复上一版本页面
- 无数据迁移，回滚无风险

## Open Questions

1. **团队成员上限**：是否需要限制团队成员数量？（建议：暂不限制，后续根据实际使用情况调整）
2. **团队图标上传**：本 change 是否实现图标上传功能？（建议：暂不实现，使用默认图标，图标上传在 C19 附件上传时统一实现）
3. **工作区 Slug 用途**：Slug 字段预留用于自定义域名，本 change 是否需要暴露 API？（建议：暂不暴露，保持内部字段）
