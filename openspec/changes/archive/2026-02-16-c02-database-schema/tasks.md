## 1. 依赖安装与配置

- [x] 1.1 添加 `github.com/google/uuid` 依赖到 `server/go.mod`
- [x] 1.2 添加 `gorm.io/datatypes` 依赖（JSONB 支持）
- [x] 1.3 添加 `github.com/lib/pq` 依赖（数组支持）
- [x] 1.4 添加 `github.com/golang-migrate/migrate/v4` 依赖
- [x] 1.5 更新 `Makefile`，添加 `migrate-up`、`migrate-down`、`migrate-create` 命令

## 2. 基础类型与工具

- [x] 2.1 创建 `server/internal/model/base.go`，定义 `Model` 基础结构体（ID、CreatedAt、UpdatedAt）
- [x] 2.2 创建 `server/internal/model/enums.go`，定义所有枚举类型（Role、StateType、IssueRelationType、ProjectStatus、CycleStatus、NotificationType）及 Scanner/Valuer 实现
- [x] 2.3 编写枚举类型的表格驱动测试 `server/internal/model/enums_test.go`

## 3. 迁移文件 - 基础实体

- [x] 3.1 创建迁移文件 `server/migrations/000001_init.up.sql`，包含 workspaces、teams、users、team_members 四张表
- [x] 3.2 创建对应的 down 迁移 `server/migrations/000001_init.down.sql`
- [x] 3.3 验证迁移文件语法：`migrate -path server/migrations -database $DATABASE_URL version`

## 4. 迁移文件 - 核心业务

- [x] 4.1 创建迁移文件 `server/migrations/000002_core.up.sql`，包含 11 张核心业务表（issues、issue_relations、workflow_states、projects、milestones、cycles、labels、comments、attachments、documents、notifications）
- [x] 4.2 创建对应的 down 迁移 `server/migrations/000002_core.down.sql`
- [x] 4.3 在 issues 表上创建所有必需索引（UNIQUE team_id+number、assignee_id、project_id、cycle_id、status_id、GIN labels）

## 5. 迁移文件 - 辅助表

- [x] 5.1 创建迁移文件 `server/migrations/000003_auxiliary.up.sql`，包含 3 张辅助表（issue_closure、issue_status_history、workflow_transitions）
- [x] 5.2 创建对应的 down 迁移 `server/migrations/000003_auxiliary.down.sql`
- [x] 5.3 在 issue_closure 表上创建索引（descendant_id、depth）

## 6. GORM 模型 - 基础实体

- [x] 6.1 创建 `server/internal/model/workspace.go`，定义 Workspace 模型及关联关系
- [x] 6.2 创建 `server/internal/model/team.go`，定义 Team 和 TeamMember 模型及关联关系
- [x] 6.3 创建 `server/internal/model/user.go`，定义 User 模型及关联关系

## 7. GORM 模型 - Issue 相关

- [x] 7.1 创建 `server/internal/model/issue.go`，定义 Issue 模型及所有关联关系（Team、Status、Assignee、Project、Milestone、Cycle、Parent、Labels）
- [x] 7.2 创建 `server/internal/model/issue_relation.go`，定义 IssueRelation 模型
- [x] 7.3 创建 `server/internal/model/issue_closure.go`，定义 IssueClosure 模型
- [x] 7.4 创建 `server/internal/model/issue_status_history.go`，定义 IssueStatusHistory 模型

## 8. GORM 模型 - 工作流与项目

- [x] 8.1 创建 `server/internal/model/workflow_state.go`，定义 WorkflowState 模型
- [x] 8.2 创建 `server/internal/model/workflow_transition.go`，定义 WorkflowTransition 模型
- [x] 8.3 创建 `server/internal/model/project.go`，定义 Project 模型及关联关系
- [x] 8.4 创建 `server/internal/model/milestone.go`，定义 Milestone 模型
- [x] 8.5 创建 `server/internal/model/cycle.go`，定义 Cycle 模型

## 9. GORM 模型 - 其他业务实体

- [x] 9.1 创建 `server/internal/model/label.go`，定义 Label 模型
- [x] 9.2 创建 `server/internal/model/comment.go`，定义 Comment 模型
- [x] 9.3 创建 `server/internal/model/attachment.go`，定义 Attachment 模型
- [x] 9.4 创建 `server/internal/model/document.go`，定义 Document 模型
- [x] 9.5 创建 `server/internal/model/notification.go`，定义 Notification 模型

## 10. 迁移集成与验证

- [x] 10.1 编写迁移执行测试 `server/migrations/migrate_test.go`，验证迁移文件格式正确
- [x] 10.2 更新 `server/cmd/server/main.go`，在启动时自动执行迁移
- [x] 10.3 验证 `docker compose up -d postgres` 后执行 `make migrate-up` 成功创建所有 18 张表
- [x] 10.4 验证 `make migrate-down` 能成功回滚所有表

## 11. GORM 模型验证

- [x] 11.1 编写 GORM 模型表名测试 `server/internal/model/table_name_test.go`，验证所有模型的 `TableName()` 方法返回正确
- [x] 11.2 编写 GORM 模型字段标签测试，验证所有模型字段与数据库列正确映射（可使用反射检查 gorm tag）
- [x] 11.3 启动后端服务，验证 GORM 能正确连接数据库并无 schema 错误
