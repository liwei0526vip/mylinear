## Context

Issue 是项目管理系统的核心实体。当前系统已具备：
- Issue 数据模型（`server/internal/model/issue.go`），但缺少 `position` 字段
- 工作流状态（C06 已完成）
- 团队和用户体系（C04、C03 已完成）
- 标准的 Handler → Service → Store 三层架构

本 change 在现有基础上实现 Issue 的完整 CRUD 能力，包括标识符生成、拖拽排序、订阅机制。

## Goals / Non-Goals

**Goals:**
- 实现 Issue CRUD API（创建、读取、更新、删除/归档）
- 实现 Issue 标识符自动生成（`{Team.Key}-{Number}`，如 `ENG-123`）
- 实现拖拽排序（基于 `position` 字段）
- 实现 Issue 订阅/取消订阅机制
- 实现基础过滤和排序查询
- 记录 Issue 状态变更历史

**Non-Goals:**
- Issue 列表视图、看板视图的 UI 实现（属于 C11）
- Issue 评论功能（属于 C08）
- Issue 通知推送（属于 C09）
- Issue 子任务和关系（属于 C14）
- 批量操作（属于 C20）

## Decisions

### D1: Issue 标识符生成策略

**决策**: 使用数据库事务 + `MAX(number) + 1` 确保团队内 Issue 编号唯一递增。

**理由**:
- 简单可靠，不需要引入额外的序列表
- PostgreSQL 的 `UNIQUE INDEX (team_id, number)` 约束保证唯一性
- 事务内查询和插入避免并发冲突

**实现**:
```go
// 在事务中执行
var maxNumber int
db.Model(&model.Issue{}).Where("team_id = ?", teamID).Select("COALESCE(MAX(number), 0)").Scan(&maxNumber)
issue.Number = maxNumber + 1
```

**备选方案**:
- PostgreSQL 序列：需要为每个团队创建单独序列，管理复杂
- UUID 替代编号：不符合 Linear 习惯，不利于用户记忆和沟通

---

### D2: Position 字段设计

**决策**: 使用 `FLOAT` 类型存储 position，支持任意位置插入。

**理由**:
- 拖拽到两个 Issue 之间时，计算中间值（如 A=1000, B=2000 → 新位置=1500）
- 无需更新其他记录，性能好
- 与 WorkflowState 的 position 设计保持一致

**实现**:
```go
// 拖拽到 afterId 之后
newPosition := afterIssue.Position + 1000
issue.Position = newPosition

// 如果空间不足（afterIssue.Position + 1000 > nextIssue.Position），需要重算
```

**重算策略**: 当 position 差值小于 1 时，触发该状态列下所有 Issue 的 position 重算（以 1000 为基数递增）。

---

### D3: 订阅表设计

**决策**: 使用独立的 `issue_subscriptions` 关联表。

**表结构**:
```sql
CREATE TABLE issue_subscriptions (
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (issue_id, user_id)
);
```

**理由**:
- 符合关系型设计范式
- 支持高效的订阅者查询（`WHERE issue_id = ?`）
- 支持用户订阅列表查询（`WHERE user_id = ?`）
- 级联删除简化清理逻辑

---

### D4: 状态变更历史记录

**决策**: 状态变更时自动写入 `issue_status_history` 表。

**触发时机**: 更新 Issue 的 `statusId` 时。

**记录内容**:
- `issue_id`: Issue ID
- `from_status_id`: 原状态 ID（可为空，创建时）
- `to_status_id`: 新状态 ID
- `changed_by_id`: 操作用户 ID
- `changed_at`: 变更时间

**理由**:
- 支持后续 Time in Status 分析（C16）
- 活动流展示（C08）
- 燃尽图计算（C17）

---

### D5: API 路由设计

**决策**: 遵循 RESTful 风格，资源嵌套在团队下。

**路由**:
```
POST   /api/v1/teams/:teamId/issues          # 创建 Issue
GET    /api/v1/teams/:teamId/issues          # 列表查询
GET    /api/v1/issues/:id                    # 获取详情
PUT    /api/v1/issues/:id                    # 更新 Issue
DELETE /api/v1/issues/:id                    # 删除/归档
POST   /api/v1/issues/:id/subscribe          # 订阅
DELETE /api/v1/issues/:id/subscribe          # 取消订阅
PUT    /api/v1/issues/:id/position           # 更新位置
POST   /api/v1/issues/:id/restore            # 恢复已删除
GET    /api/v1/issues/:id/subscribers        # 获取订阅者列表
```

**理由**:
- 创建和列表需要团队上下文，嵌套在 `/teams/:teamId` 下
- 单个 Issue 操作使用 `/issues/:id`，路径更简洁
- 符合现有 API 设计风格

---

### D6: 查询过滤实现

**决策**: 使用动态条件构建，支持多字段组合过滤。

**支持的过滤参数**:
- `status_id`: 单个状态
- `priority`: 单个优先级
- `assignee_id`: 单个负责人（支持 `me` 关键字）
- `label_ids`: 多个标签（逗号分隔，AND 逻辑）
- `project_id`: 单个项目
- `cycle_id`: 单个迭代
- `created_by_id`: 创建者（支持 `me` 关键字）
- `subscribed`: 是否订阅（`me` 关键字）

**实现**:
```go
query := db.Model(&model.Issue{}).Where("team_id = ?", teamID)

if statusID := c.Query("status_id"); statusID != "" {
    query = query.Where("status_id = ?", statusID)
}

if assigneeID := c.Query("assignee_id"); assigneeID == "me" {
    query = query.Where("assignee_id = ?", currentUserID)
} else if assigneeID != "" {
    query = query.Where("assignee_id = ?", assigneeID)
}
// ...
```

---

### D7: 权限控制策略

**决策**: 基于团队成员角色控制 Issue 操作权限。

| 操作 | Guest | Member | Admin |
|------|-------|--------|-------|
| 查看 Issue | ✅ 非私有团队 | ✅ | ✅ |
| 创建 Issue | ❌ | ✅ | ✅ |
| 更新 Issue | ❌ | ✅ 自己负责/创建的 | ✅ 全部 |
| 删除 Issue | ❌ | ❌ | ✅ |
| 订阅/取消订阅 | ✅ | ✅ | ✅ |

**实现**: 在 Service 层检查权限，使用 `store.IsTeamMember()` 辅助函数。

## Risks / Trade-offs

### R1: Issue 编号并发冲突
**风险**: 高并发创建时可能产生编号冲突。
**缓解**:
- 使用数据库事务
- `UNIQUE INDEX` 约束作为最后防线
- 冲突时重试（重新获取 MAX + 1）

### R2: Position 重算性能
**风险**: 频繁拖拽可能导致 position 空间耗尽，触发全量重算。
**缓解**:
- 只在单个状态列内重算，影响范围小
- 使用批量更新减少数据库压力
- 重算后 position 差值为 1000，足够容纳大量操作

### R3: 订阅表增长
**风险**: 高频操作产生大量订阅记录。
**缓解**:
- 使用复合主键避免重复
- 级联删除自动清理
- 后续可考虑订阅过期策略（不在本期范围）

## Migration Plan

### 数据库迁移
1. 新增 `issue_subscriptions` 表
2. 为 `issues` 表添加 `position` 列（默认值 0）
3. 为现有 Issue 设置 position（按 created_at 排序，递增 1000）

### 回滚策略
1. 删除 `issue_subscriptions` 表
2. 删除 `issues.position` 列

## Open Questions

1. **Issue 描述的 Markdown 渲染**: 后端是否需要预渲染？还是仅存储，由前端渲染？
   - **决策**: 仅存储，前端渲染。后端不做 Markdown 处理。

2. **Labels 字段类型**: 当前使用 `pq.StringArray` 存储 UUID 字符串数组，是否需要改用关联表？
   - **决策**: 保持当前设计。Issue-Label 是多对多关系，但 Label 数量有限，数组足够。后续如有性能问题再优化。
