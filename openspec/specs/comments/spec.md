## ADDED Requirements

### Requirement: Comment 模型

Comment 模型 MUST 定义评论实体的完整结构，支持嵌套回复。

#### Scenario: Comment 模型字段

- **WHEN** 定义 `Comment` 结构体
- **THEN** MUST 包含字段：`ID`（UUID）、`IssueID`（FK）、`ParentID`（*UUID, FK nullable）、`UserID`（FK）、`Body`（TEXT）、`CreatedAt`、`UpdatedAt`、`EditedAt`（*time.Time）

#### Scenario: Comment 关联关系

- **WHEN** 定义 `Comment` 关联
- **THEN** MUST 支持 `Issue`、`Parent`、`User` 的 belongs-to 关系
- **AND** MUST 支持 `Replies []Comment` 的 has-many 关系（自引用）

#### Scenario: Comment 表名

- **WHEN** 定义 `Comment` 的 `TableName()` 方法
- **THEN** MUST 返回 `comments`

---

### Requirement: 评论创建

系统 SHALL 支持在 Issue 下创建评论。

#### Scenario: 创建评论基础字段

- **WHEN** 用户通过 `POST /api/v1/issues/:issueId/comments` 创建评论
- **THEN** 系统 SHALL 创建 Comment 记录，包含 `body`（必填，Markdown 格式）
- **AND** 系统 SHALL 自动设置 `user_id` 为当前用户
- **AND** 系统 SHALL 返回完整的 Comment 对象

#### Scenario: 创建嵌套回复

- **WHEN** 用户创建评论时指定 `parent_id`
- **THEN** 系统 SHALL 验证父评论属于同一 Issue
- **AND** 系统 SHALL 创建嵌套回复记录

#### Scenario: 评论者自动订阅

- **WHEN** 用户在 Issue 下发表评论
- **THEN** 系统 SHALL 自动将该用户添加为 Issue 订阅者（若未订阅）

#### Scenario: 非成员无法评论

- **WHEN** 非团队成员尝试评论私有团队的 Issue
- **THEN** 系统 SHALL 返回 403 Forbidden

---

### Requirement: @mention 解析

系统 SHALL 支持在评论中 @mention 用户。

#### Scenario: 解析 @mention

- **WHEN** 用户创建或更新评论，body 包含 `@username` 格式的提及
- **THEN** 系统 SHALL 解析所有有效的 @mentions
- **AND** 系统 SHALL 验证被提及的 username 存在于工作区

#### Scenario: @mention 触发自动订阅

- **WHEN** 评论中 @mention 某用户
- **THEN** 系统 SHALL 自动将被提及的用户添加为 Issue 订阅者

#### Scenario: 返回已解析的 mentions

- **WHEN** 返回评论对象
- **THEN** 系统 SHALL 包含 `mentions` 字段，列出被提及用户的 `id`、`username`、`name`

#### Scenario: 无效 username 忽略

- **WHEN** 评论中 @mention 不存在的 username
- **THEN** 系统 SHALL 忽略该提及（不报错，不触发订阅）

---

### Requirement: 评论查询

系统 SHALL 支持查询 Issue 的评论列表。

#### Scenario: 获取评论列表

- **WHEN** 用户通过 `GET /api/v1/issues/:issueId/comments` 查询评论
- **THEN** 系统 SHALL 返回该 Issue 的所有评论
- **AND** 默认按 `created_at` 正序排列

#### Scenario: 评论列表包含嵌套回复

- **WHEN** 返回评论列表
- **THEN** 每条评论 SHALL 包含 `replies` 数组（嵌套回复列表）

#### Scenario: 分页查询

- **WHEN** 用户指定 `page` 和 `page_size` 参数
- **THEN** 系统 SHALL 返回分页结果，包含 `total`（总数）、`items`（当前页）
- **AND** 默认 `page_size` SHALL 为 50，最大 100

#### Scenario: 包含用户信息

- **WHEN** 返回评论列表
- **THEN** 每条评论 SHALL 包含 `user` 对象（`id`、`name`、`username`、`avatar_url`）

---

### Requirement: 评论更新

系统 SHALL 支持更新评论。

#### Scenario: 更新评论内容

- **WHEN** 用户通过 `PUT /api/v1/comments/:id` 更新评论
- **THEN** 系统 SHALL 更新 `body` 字段
- **AND** 系统 SHALL 设置 `edited_at` 为当前时间

#### Scenario: 仅作者可编辑

- **WHEN** 非评论作者尝试编辑评论
- **THEN** 系统 SHALL 返回 403 Forbidden

#### Scenario: 更新时重新解析 @mention

- **WHEN** 用户更新评论内容
- **THEN** 系统 SHALL 重新解析 @mentions
- **AND** 新提及的用户 SHALL 自动订阅 Issue

---

### Requirement: 评论删除

系统 SHALL 支持删除评论。

#### Scenario: 删除评论

- **WHEN** 用户通过 `DELETE /api/v1/comments/:id` 删除评论
- **THEN** 系统 SHALL 删除评论记录
- **AND** 嵌套回复 SHALL 同时删除（级联）

#### Scenario: 仅作者可删除

- **WHEN** 非评论作者尝试删除评论
- **THEN** 系统 SHALL 返回 403 Forbidden

#### Scenario: 管理员可删除任何评论

- **WHEN** 团队 Admin 用户删除其他成员的评论
- **THEN** 系统 SHALL 允许操作

---

### Requirement: Markdown 支持

评论内容 SHALL 支持 Markdown 格式。

#### Scenario: 存储 Markdown 原文

- **WHEN** 用户提交评论
- **THEN** 系统 SHALL 存储 Markdown 原文到 `body` 字段
- **AND** 前端负责渲染

#### Scenario: 安全渲染

- **WHEN** 前端渲染评论 Markdown
- **THEN** 系统（前端）MUST 对用户输入进行 XSS 过滤
- **AND** 禁止执行 `<script>` 等危险标签
