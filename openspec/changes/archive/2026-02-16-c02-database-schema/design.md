## Context

C01 已完成项目脚手架搭建，包括 Docker Compose 编排（PostgreSQL 16）、Go 后端框架（Gin + GORM）和基础目录结构。本 change 需要在现有基础上建立完整的数据库 schema，定义 18 张表及其关联关系，为后续业务功能开发提供数据持久化基础。

**当前状态：**
- PostgreSQL 16 已通过 Docker Compose 部署
- GORM 已在 `server/go.mod` 中引入
- `server/migrations/` 目录已创建（空）
- `server/internal/model/` 目录已创建（空）

**约束：**
- 必须使用 UUID 主键（兼容 Local-First 架构）
- 必须通过迁移文件管理 schema 变更
- 遵循 Go 标准项目布局

## Goals / Non-Goals

**Goals:**
- 创建 18 张表的 DDL（15 张核心业务表 + 3 张辅助表）
- 为所有表定义 GORM 模型（Go struct）
- 集成 golang-migrate 迁移工具
- 设计高效的索引策略
- 支持 `position FLOAT` 字段实现拖拽排序

**Non-Goals:**
- 不实现任何业务逻辑（CRUD API 在 C03~C05）
- 不创建初始数据（默认工作流状态在 C06）
- 不实现数据库连接池的高级配置（使用 GORM 默认值）

## Decisions

### D1：使用 golang-migrate 管理迁移

**决策：** 使用 `github.com/golang-migrate/migrate/v4` 管理数据库迁移。

**理由：**
- 业界标准，社区活跃
- 支持版本化迁移，确保团队开发一致性
- 提供 CLI 工具和 Go 库两种使用方式
- 支持 PostgreSQL 驱动

**替代方案：**
- GORM AutoMigrate：简单但不适合团队协作，无法追踪历史变更
- Atlas：功能强大但学习曲线陡峭
- 手写 SQL：缺乏版本管理和回滚支持

**实施方式：**
- 迁移文件存放在 `server/migrations/`
- 命名格式：`{version}_{description}.up.sql` / `{version}_{description}.down.sql`
- 在 `main.go` 启动时自动执行迁移

---

### D2：GORM 模型使用 `gorm.io/datatypes` 处理 JSONB

**决策：** 使用 `gorm.io/datatypes.JSON` 类型处理 JSONB 字段。

**理由：**
- GORM 官方提供，兼容性好
- 支持 JSON 查询语法
- 类型安全（相对于 `interface{}`）

**替代方案：**
- `map[string]interface{}`：灵活但缺乏类型安全
- 自定义类型：需要额外开发成本
- `github.com/lib/pq` 的 JSON 类型：功能类似，但 datatypes 更符合 GORM 生态

**适用字段：**
- `workspaces.settings`
- `teams.cycle_settings` / `teams.workflow_settings`
- `users.settings`

---

### D3：使用 `github.com/lib/pq` 处理数组字段

**决策：** 使用 `github.com/lib/pq` 的 `pq.StringArray` 处理数组字段。

**理由：**
- PostgreSQL 官方驱动支持
- `pq.StringArray` 实现了 `sql.Scanner` 和 `driver.Valuer` 接口
- 支持 GIN 索引查询

**替代方案：**
- `pq.Array()` 包装器：每次查询需要手动包装，不够优雅
- JSONB 存储数组：查询性能较差，无法使用数组索引

**适用字段：**
- `issues.labels`（UUID 数组，存储为字符串数组）
- `projects.teams` / `projects.labels`

---

### D4：闭包表存储 Issue 层级关系

**决策：** 使用闭包表（Closure Table）存储 Sub-Issue 层级关系。

**理由：**
- 支持任意深度层级查询
- 查询性能稳定，不随深度增加而下降
- 易于获取所有子孙节点或祖先节点

**替代方案：**
- 邻接表（Adjacency List）：仅存储 `parent_id`，深度查询需要递归 CTE
- 路径枚举（Path Enumeration）：存储完整路径，更新复杂
- 嵌套集（Nested Set）：插入/移动操作复杂

**表结构：**
```sql
CREATE TABLE issue_closure (
    ancestor_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    descendant_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    depth INTEGER NOT NULL,
    PRIMARY KEY (ancestor_id, descendant_id)
);
```

---

### D5：`position FLOAT` 支持拖拽排序

**决策：** 使用 `FLOAT` 类型存储排序位置，支持拖拽排序。

**理由：**
- 插入新元素时无需更新其他记录
- 使用小数实现"插入到两个元素之间"的语义
- 简单直观，易于实现

**替代方案：**
- `INTEGER`：插入时需要更新后续所有记录
- `DECIMAL`：精度更高但实现复杂
- Linked List：需要额外维护 prev/next 指针

**实现方式：**
- 新元素 `position = (prev.position + next.position) / 2`
- 拖拽到末尾时 `position = last.position + 65536`
- 位置重排：当 `position` 差值过小时触发批量重排

**适用字段：**
- `workflow_states.position`
- `milestones.position`

---

### D6：枚举类型使用自定义 Go 类型

**决策：** 为枚举字段定义自定义 Go 类型，实现 `Scanner`/`Valuer` 接口。

**理由：**
- 类型安全，避免字符串硬编码
- IDE 自动补全支持
- 便于添加验证逻辑

**替代方案：**
- 纯字符串：缺乏类型安全
- `int` 常量：可读性差，数据库存储不直观

**实现示例：**
```go
type Role string

const (
    RoleGlobalAdmin Role = "global_admin"
    RoleAdmin       Role = "admin"
    RoleMember      Role = "member"
    RoleGuest       Role = "guest"
)

func (r *Role) Scan(value interface{}) error {
    *r = Role(value.(string))
    return nil
}

func (r Role) Value() (driver.Value, error) {
    return string(r), nil
}
```

---

### D7：迁移拆分为多个文件

**决策：** 将 18 张表拆分为 3 个迁移文件，按依赖顺序执行。

**理由：**
- 单文件过大，难以维护
- 便于定位问题
- 支持部分回滚

**拆分策略：**
| 版本 | 内容 | 表数量 |
|------|------|--------|
| `000001` | 基础实体（Workspace, Team, User, TeamMember） | 4 |
| `000002` | 核心业务（Issue, IssueRelation, WorkflowState, Project, Milestone, Cycle, Label, Comment, Attachment, Document, Notification） | 11 |
| `000003` | 辅助表（IssueClosure, IssueStatusHistory, WorkflowTransition） | 3 |

## 表结构设计

### 核心业务表（15 张）

| 表名 | 说明 | 主要字段 |
|------|------|----------|
| `workspaces` | 工作区 | id, name, slug, settings |
| `teams` | 团队 | id, workspace_id, name, key, cycle_settings |
| `users` | 用户 | id, workspace_id, email, name, role |
| `team_members` | 团队成员 | team_id, user_id, role |
| `issues` | 工单 | id, team_id, number, title, status_id, priority, labels |
| `issue_relations` | Issue 关系 | id, issue_id, related_issue_id, type |
| `workflow_states` | 工作流状态 | id, team_id, name, type, position |
| `projects` | 项目 | id, workspace_id, name, status, lead_id |
| `milestones` | 里程碑 | id, project_id, name, target_date, position |
| `cycles` | 迭代 | id, team_id, number, start_date, end_date, status |
| `labels` | 标签 | id, workspace_id, team_id, name, color |
| `comments` | 评论 | id, issue_id, user_id, body |
| `attachments` | 附件 | id, issue_id, user_id, filename, url |
| `documents` | 文档 | id, workspace_id, title, content |
| `notifications` | 通知 | id, user_id, type, title, read_at |

### 辅助表（3 张）

| 表名 | 说明 | 主要字段 |
|------|------|----------|
| `issue_closure` | Issue 层级闭包表 | ancestor_id, descendant_id, depth |
| `issue_status_history` | 状态变更历史 | id, issue_id, from_status_id, to_status_id, changed_at |
| `workflow_transitions` | 工作流转换规则 | id, team_id, from_state_id, to_state_id, is_allowed |

## 索引设计

### Issue 表核心索引

```sql
-- Issue 标识符唯一性
CREATE UNIQUE INDEX idx_issue_team_number ON issues(team_id, number);

-- 按负责人查询
CREATE INDEX idx_issue_assignee_id ON issues(assignee_id);

-- 按项目查询
CREATE INDEX idx_issue_project_id ON issues(project_id);

-- 按迭代查询
CREATE INDEX idx_issue_cycle_id ON issues(cycle_id);

-- 按状态查询
CREATE INDEX idx_issue_status_id ON issues(status_id);

-- 标签数组查询（GIN）
CREATE INDEX idx_issue_labels ON issues USING GIN (labels);
```

### 闭包表索引

```sql
CREATE INDEX idx_issue_closure_descendant ON issue_closure(descendant_id);
CREATE INDEX idx_issue_closure_depth ON issue_closure(depth);
```

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|----------|
| 迁移执行失败导致 schema 不一致 | 每个迁移包裹在事务中，失败自动回滚；提供 down 迁移支持回滚 |
| GORM 模型与 DDL 不同步 | 在测试中添加 schema 一致性检查，对比 GORM AutoMigrate 输出与实际表结构 |
| JSONB 字段查询性能 | 为常用 JSONB 路径创建表达式索引（如需要） |
| `position FLOAT` 精度问题 | 当差值小于 0.0001 时触发位置重排 |
| 闭包表维护成本 | 使用数据库触发器或应用层钩子自动维护（C14 实现） |
| 数组字段 GIN 索引写入性能 | 监控写入性能，必要时考虑使用中间表替代 |

## Migration Plan

### 部署步骤

1. 确保 PostgreSQL 容器运行中（`docker compose up -d postgres`）
2. 执行迁移：`make migrate-up` 或启动后端服务自动执行
3. 验证表结构：`\dt` 查看所有表

### 回滚策略

```bash
# 回滚最近一次迁移
make migrate-down

# 回滚所有迁移
make migrate-down-all

# 强制回滚到指定版本
migrate -path server/migrations -database $DATABASE_URL goto 1
```

## Open Questions

- **Q1**: `issue_closure` 表的触发器在哪个 change 实现？
  - **A**: 在 C14（Sub-Issues 与 Issue Relations）实现，本 change 仅创建表结构

- **Q2**: 默认工作流状态如何初始化？
  - **A**: 在 C06（工作流状态与标签）实现，本 change 仅创建 `workflow_states` 表

- **Q3**: 是否需要为 `documents.content` 创建全文搜索索引？
  - **A**: Phase 3 C22 实现全局搜索时再添加，本 change 不涉及
