## ADDED Requirements

### Requirement: UUID 主键规范

所有表 MUST 使用 UUID 作为主键，支持分布式 ID 生成，兼容 Local-First 架构。

#### Scenario: 主键类型为 UUID

- **WHEN** 创建任意数据表
- **THEN** 主键字段 MUST 为 `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`

#### Scenario: 支持 PostgreSQL gen_random_uuid()

- **WHEN** 插入新记录且未指定 ID
- **THEN** 数据库 MUST 自动生成 UUID（使用 `gen_random_uuid()` 函数）

---

### Requirement: 核心业务表结构

系统 MUST 定义 15 张核心业务表，覆盖 Workspace、Team、User、Issue、Project 等领域实体。

#### Scenario: Workspace 工作区表

- **WHEN** 执行数据库迁移
- **THEN** MUST 创建 `workspaces` 表，包含字段：`id`（UUID 主键）、`name`（VARCHAR 255）、`slug`（VARCHAR 50 UNIQUE）、`logo_url`（TEXT）、`settings`（JSONB）、`created_at`/`updated_at`（TIMESTAMPTZ）

#### Scenario: Team 团队表

- **WHEN** 执行数据库迁移
- **THEN** MUST 创建 `teams` 表，包含字段：`id`、`workspace_id`（FK）、`parent_id`（FK nullable）、`name`、`key`（VARCHAR 10 UNIQUE）、`icon_url`、`timezone`、`is_private`、`cycle_settings`（JSONB）、`workflow_settings`（JSONB）、`created_at`/`updated_at`

#### Scenario: User 用户表

- **WHEN** 执行数据库迁移
- **THEN** MUST 创建 `users` 表，包含字段：`id`、`workspace_id`（FK）、`email`（UNIQUE）、`name`、`username`（UNIQUE）、`avatar_url`、`password_hash`、`role`（ENUM）、`settings`（JSONB）、`created_at`/`updated_at`

#### Scenario: TeamMember 团队成员表

- **WHEN** 执行数据库迁移
- **THEN** MUST 创建 `team_members` 表，复合主键（`team_id` + `user_id`），包含 `role`（ENUM）、`joined_at`

#### Scenario: Issue 工单表

- **WHEN** 执行数据库迁移
- **THEN** MUST 创建 `issues` 表，包含字段：`id`、`team_id`（FK）、`number`（团队内序号）、`title`、`description`（TEXT）、`status_id`（FK）、`priority`（INTEGER）、`assignee_id`（FK nullable）、`project_id`（FK nullable）、`milestone_id`（FK nullable）、`cycle_id`（FK nullable）、`parent_id`（FK nullable）、`estimate`（nullable）、`due_date`（DATE nullable）、`sla_due_at`（TIMESTAMPTZ nullable）、`labels`（UUID[]）、`created_by_id`（FK）、`created_at`/`updated_at`、`completed_at`/`cancelled_at`（nullable）

#### Scenario: IssueRelation Issue 关系表

- **WHEN** 执行数据库迁移
- **THEN** MUST 创建 `issue_relations` 表，包含字段：`id`、`issue_id`（FK）、`related_issue_id`（FK）、`type`（ENUM: blocked_by/blocking/related/duplicate）、`created_at`，UNIQUE 约束（`issue_id` + `related_issue_id` + `type`）

#### Scenario: WorkflowState 工作流状态表

- **WHEN** 执行数据库迁移
- **THEN** MUST 创建 `workflow_states` 表，包含字段：`id`、`team_id`（FK）、`name`、`type`（ENUM: backlog/unstarted/started/completed/cancelled）、`color`、`position`（FLOAT）、`is_default`、`created_at`

#### Scenario: Project 项目表

- **WHEN** 执行数据库迁移
- **THEN** MUST 创建 `projects` 表，包含字段：`id`、`workspace_id`（FK）、`name`、`description`（TEXT）、`status`（ENUM: planned/in_progress/paused/completed/cancelled）、`priority`、`lead_id`（FK nullable）、`start_date`/`target_date`（DATE nullable）、`teams`/`labels`（UUID[]）、`created_at`/`updated_at`、`completed_at`（nullable）

#### Scenario: Milestone 里程碑表

- **WHEN** 执行数据库迁移
- **THEN** MUST 创建 `milestones` 表，包含字段：`id`、`project_id`（FK）、`name`、`description`（TEXT）、`target_date`（DATE nullable）、`position`（FLOAT）、`created_at`

#### Scenario: Cycle 迭代表

- **WHEN** 执行数据库迁移
- **THEN** MUST 创建 `cycles` 表，包含字段：`id`、`team_id`（FK）、`number`、`name`、`description`（TEXT）、`start_date`/`end_date`（DATE）、`cooldown_end_date`（DATE nullable）、`status`（ENUM: upcoming/active/completed）、`created_at`，UNIQUE 约束（`team_id` + `number`）

#### Scenario: Label 标签表

- **WHEN** 执行数据库迁移
- **THEN** MUST 创建 `labels` 表，包含字段：`id`、`workspace_id`（FK）、`team_id`（FK nullable）、`name`、`description`（TEXT）、`color`、`parent_id`（FK nullable）、`is_archived`、`created_at`

#### Scenario: Comment 评论表

- **WHEN** 执行数据库迁移
- **THEN** MUST 创建 `comments` 表，包含字段：`id`、`issue_id`（FK）、`parent_id`（FK nullable）、`user_id`（FK）、`body`（TEXT）、`created_at`/`updated_at`、`edited_at`（nullable）

#### Scenario: Attachment 附件表

- **WHEN** 执行数据库迁移
- **THEN** MUST 创建 `attachments` 表，包含字段：`id`、`issue_id`（FK）、`user_id`（FK）、`filename`、`url`（TEXT）、`size`（BIGINT）、`mime_type`、`created_at`

#### Scenario: Document 文档表

- **WHEN** 执行数据库迁移
- **THEN** MUST 创建 `documents` 表，包含字段：`id`、`workspace_id`（FK）、`project_id`/`issue_id`（FK nullable）、`title`、`content`（TEXT）、`icon`、`created_by_id`（FK）、`created_at`/`updated_at`

#### Scenario: Notification 通知表

- **WHEN** 执行数据库迁移
- **THEN** MUST 创建 `notifications` 表，包含字段：`id`、`user_id`（FK）、`type`（ENUM）、`title`、`body`（TEXT）、`resource_type`、`resource_id`（UUID nullable）、`read_at`（nullable）、`created_at`

---

### Requirement: 辅助表结构

系统 MUST 定义 3 张辅助表，支持子 Issue 层级查询、状态变更历史、工作流转换规则。

#### Scenario: Issue 闭包表

- **WHEN** 执行数据库迁移
- **THEN** MUST 创建 `issue_closure` 表，复合主键（`ancestor_id` + `descendant_id`），包含 `depth`（INTEGER），支持任意深度子 Issue 查询

#### Scenario: Issue 状态历史表

- **WHEN** 执行数据库迁移
- **THEN** MUST 创建 `issue_status_history` 表，包含字段：`id`、`issue_id`（FK）、`from_status_id`/`to_status_id`（FK nullable）、`changed_by_id`（FK）、`changed_at`（TIMESTAMPTZ）

#### Scenario: 工作流转换规则表

- **WHEN** 执行数据库迁移
- **THEN** MUST 创建 `workflow_transitions` 表，包含字段：`id`、`team_id`（FK）、`from_state_id`/`to_state_id`（FK）、`is_allowed`（BOOLEAN）、`created_at`

---

### Requirement: 索引设计

Issue 表是系统核心，MUST 设计高效索引支持常见查询场景。

#### Scenario: Issue 标识符唯一性

- **WHEN** 创建 `issues` 表
- **THEN** MUST 创建 UNIQUE 索引（`team_id` + `number`），确保团队内 Issue 序号唯一

#### Scenario: Issue 按负责人查询

- **WHEN** 创建 `issues` 表
- **THEN** MUST 创建索引 `idx_issue_assignee_id` ON (`assignee_id`)

#### Scenario: Issue 按项目查询

- **WHEN** 创建 `issues` 表
- **THEN** MUST 创建索引 `idx_issue_project_id` ON (`project_id`)

#### Scenario: Issue 按迭代查询

- **WHEN** 创建 `issues` 表
- **THEN** MUST 创建索引 `idx_issue_cycle_id` ON (`cycle_id`)

#### Scenario: Issue 按状态查询

- **WHEN** 创建 `issues` 表
- **THEN** MUST 创建索引 `idx_issue_status_id` ON (`status_id`)

#### Scenario: Issue 标签数组查询

- **WHEN** 创建 `issues` 表
- **THEN** MUST 创建 GIN 索引 ON (`labels`)，支持数组包含查询

#### Scenario: 闭包表索引

- **WHEN** 创建 `issue_closure` 表
- **THEN** MUST 创建索引 `idx_issue_closure_descendant` ON (`descendant_id`) 和 `idx_issue_closure_depth` ON (`depth`)

---

### Requirement: 外键约束

所有关联关系 MUST 通过外键约束保证数据完整性。

#### Scenario: 级联删除

- **WHEN** 删除父记录（如 Team）
- **THEN** 外键 MUST 配置 `ON DELETE CASCADE` 或 `ON DELETE SET NULL`，根据业务语义选择

#### Scenario: 外键引用完整性

- **WHEN** 插入或更新记录
- **THEN** 数据库 MUST 验证外键引用的目标记录存在

---

### Requirement: 数据库迁移管理

系统 MUST 使用 golang-migrate 管理数据库 schema 版本。

#### Scenario: 迁移文件命名规范

- **WHEN** 创建迁移文件
- **THEN** 文件名 MUST 遵循 `{version}_{description}.up.sql` 和 `{version}_{description}.down.sql` 格式

#### Scenario: 迁移版本递增

- **WHEN** 添加新迁移
- **THEN** 版本号 MUST 递增，确保迁移按顺序执行

#### Scenario: 迁移可回滚

- **WHEN** 执行迁移
- **THEN** MUST 提供对应的 down 迁移，支持回滚到之前版本

#### Scenario: 迁移幂等性

- **WHEN** 重复执行迁移
- **THEN** 系统 MUST 记录已执行的迁移版本，避免重复执行
