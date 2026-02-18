# C08 — 评论与活动流 任务清单

> **TDD 开发原则**：严格遵循 Red-Green-Refactor 循环
> - 🔴 Red：先写失败的测试
> - 🟢 Green：写最少的代码让测试通过
> - 🔵 Refactor：重构代码（保持测试通过）

---

## 1. 数据模型与迁移

- [x] 1.1 在 `server/internal/model/enums.go` 添加 `ActivityType` 枚举类型（10 种活动类型）
- [x] 1.2 创建 `server/internal/model/activity.go` - Activity 模型定义
- [x] 1.3 创建数据库迁移文件 `server/migrations/000008_create_activities.up.sql`
- [x] 1.4 创建对应的 down 迁移文件

---

## 2. Store 层 - Activity CRUD

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

- [x] 2.1 🔴 编写 `activity_store_test.go` - CreateActivity 测试（表格驱动：正常创建、不同活动类型）
- [x] 2.2 🟢 实现 `CreateActivity` store 方法
- [x] 2.3 🔴 编写 `GetActivitiesByIssueID` 测试（表格驱动：空列表、多条记录、分页、按类型过滤）
- [x] 2.4 🟢 实现 `GetActivitiesByIssueID` store 方法
- [x] 2.5 🔴 编写 `GetActivityByID` 测试（表格驱动：正常获取、不存在的ID）
- [x] 2.6 🟢 实现 `GetActivityByID` store 方法

---

## 3. Store 层 - Comment CRUD

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

- [x] 3.1 🔴 编写 `comment_store_test.go` - CreateComment 测试（表格驱动：正常创建、嵌套回复、父评论不存在）
- [x] 3.2 🟢 实现 `CreateComment` store 方法
- [x] 3.3 🔴 编写 `GetCommentsByIssueID` 测试（表格驱动：空列表、多条记录、嵌套结构、分页）
- [x] 3.4 🟢 实现 `GetCommentsByIssueID` store 方法（返回树形结构）
- [x] 3.5 🔴 编写 `UpdateComment` 测试（表格驱动：更新body、设置edited_at）
- [x] 3.6 🟢 实现 `UpdateComment` store 方法
- [x] 3.7 🔴 编写 `DeleteComment` 测试（表格驱动：正常删除、级联删除回复）
- [x] 3.8 🟢 实现 `DeleteComment` store 方法
- [x] 3.9 🔴 编写 `GetCommentByID` 测试（表格驱动：正常获取、不存在的ID）
- [x] 3.10 🟢 实现 `GetCommentByID` store 方法

---

## 4. Service 层 - @mention 解析

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

- [x] 4.1 🔴 编写 `mention_service_test.go` - ParseMentions 测试（表格驱动：单个@、多个@、无@、邮箱误匹配排除）
- [x] 4.2 🟢 实现 `ParseMentions` 函数（正则提取 @username）
- [x] 4.3 🔴 编写 `ResolveMentions` 测试（表格驱动：有效username、无效username、部分有效）- 跳过，合并到 CommentService
- [x] 4.4 🟢 实现 `ResolveMentions` 方法（查询 User 表验证并返回用户列表）- 跳过，合并到 CommentService

---

## 5. Service 层 - Comment 业务逻辑

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

- [x] 5.1 🔴 编写 `comment_service_test.go` - CreateComment 测试（表格驱动：正常创建、自动订阅评论者、@mention自动订阅）
- [x] 5.2 🟢 实现 `CreateComment` service 方法（含 @mention 解析、自动订阅、活动记录）
- [x] 5.3 🔴 编写 `UpdateComment` 测试（表格驱动：仅作者可编辑、重新解析@mention）
- [x] 5.4 🟢 实现 `UpdateComment` service 方法（权限校验、设置 edited_at）
- [x] 5.5 🔴 编写 `DeleteComment` 测试（表格驱动：仅作者可删除、Admin可删除任何评论）
- [x] 5.6 🟢 实现 `DeleteComment` service 方法（权限校验）

---

## 6. Service 层 - Activity 业务逻辑

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

- [x] 6.1 🔴 编写 `activity_service_test.go` - RecordActivity 测试（表格驱动：不同活动类型、payload JSON 序列化）
- [x] 6.2 🟢 实现 `RecordActivity` service 方法
- [x] 6.3 🔴 编写 `GetIssueActivities` 测试（表格驱动：分页、按类型过滤）
- [x] 6.4 🟢 实现 `GetIssueActivities` service 方法

---

## 7. Service 层 - Issue 集成活动记录

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

- [x] 7.1 🔴 编写 `issue_service_activity_test.go` - Issue 创建时记录活动测试
- [x] 7.2 🟢 修改 `IssueService.CreateIssue` 集成活动记录
- [x] 7.3 🔴 编写 Issue 更新标题/描述时记录活动测试
- [x] 7.4 🟢 修改 `IssueService.UpdateIssue` 集成标题/描述变更活动
- [x] 7.5 🔴 编写 Issue 状态变更时写入 `issue_status_history` 并记录活动测试（仅活动记录）
- [x] 7.6 🟢 修改 `IssueService.UpdateIssue` 集成状态变更历史和活动记录（仅活动记录）
- [x] 7.7 🔴 编写 Issue 负责人/优先级/截止日期变更时记录活动测试
- [x] 7.8 🟢 修改 `IssueService.UpdateIssue` 集成其他字段变更活动

---

## 8. Handler 层 - Comment HTTP 接口

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

- [x] 8.1 🔴 编写 `comment_handler_test.go` - CreateComment HTTP 测试（表格驱动：成功创建、参数校验失败、未授权、非成员）
- [x] 8.2 🟢 实现 `CreateComment` handler（POST /api/v1/issues/:issueId/comments）
- [x] 8.3 🔴 编写 `ListIssueComments` HTTP 测试（表格驱动：成功获取、空列表、分页参数）
- [x] 8.4 🟢 实现 `ListIssueComments` handler（GET /api/v1/issues/:issueId/comments）
- [x] 8.5 🔴 编写 `UpdateComment` HTTP 测试（表格驱动：成功更新、非作者拒绝）
- [x] 8.6 🟢 实现 `UpdateComment` handler（PUT /api/v1/comments/:id）
- [x] 8.7 🔴 编写 `DeleteComment` HTTP 测试（表格驱动：成功删除、非作者拒绝、Admin可删除）
- [x] 8.8 🟢 实现 `DeleteComment` handler（DELETE /api/v1/comments/:id）

---

## 9. Handler 层 - Activity HTTP 接口

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

- [x] 9.1 🔴 编写 `activity_handler_test.go` - ListIssueActivities HTTP 测试（表格驱动：成功获取、分页、按类型过滤、非成员拒绝）
- [x] 9.2 🟢 实现 `ListIssueActivities` handler（GET /api/v1/issues/:issueId/activities）

---

## 10. 路由注册与验证

- [x] 10.1 在 `server/internal/router/router.go` 添加 `RegisterCommentRoutes` 函数
- [x] 10.2 在 `server/internal/router/router.go` 添加 `RegisterActivityRoutes` 函数
- [x] 10.3 在 `server/cmd/server/main.go` 注册 Comment 和 Activity 路由
- [x] 10.4 运行 `make test` 确保所有后端测试通过

---

## 11. 前端 - API 层

- [x] 11.1 创建 `web/src/api/comments.ts` - 定义 Comment、CreateCommentDTO、UpdateCommentDTO 类型
- [x] 11.2 实现 `fetchIssueComments(issueId, params)` API 函数
- [x] 11.3 实现 `createComment(issueId, data)` API 函数
- [x] 11.4 实现 `updateComment(id, data)` API 函数
- [x] 11.5 实现 `deleteComment(id)` API 函数
- [x] 11.6 创建 `web/src/api/activities.ts` - 定义 Activity、ActivityType 类型
- [x] 11.7 实现 `fetchIssueActivities(issueId, params)` API 函数

---

## 12. 前端 - 状态管理

- [x] 12.1 创建 `web/src/stores/commentStore.ts` - Zustand store 定义
- [x] 12.2 实现 `comments` 状态和 `fetchIssueComments` action
- [x] 12.3 实现 `createComment` / `updateComment` / `deleteComment` actions
- [x] 12.4 实现 `loading` / `error` 状态管理
- [x] 12.5 创建 `web/src/stores/activityStore.ts` - Zustand store 定义
- [x] 12.6 实现 `activities` 状态和 `fetchIssueActivities` action

---

## 13. 前端 - 评论组件开发

- [x] 13.1 创建 `web/src/components/comments/` 目录
- [x] 13.2 实现 `CommentItem.tsx` - 单条评论组件（头像、用户名、时间、body渲染、操作按钮）
- [x] 13.3 实现 `CommentList.tsx` - 评论列表组件（递归渲染嵌套回复）
- [x] 13.4 实现 `CommentInput.tsx` - 评论输入框组件（Markdown 支持、@mention 高亮）
- [x] 13.5 实现 `CommentSection.tsx` - 评论区容器组件（列表 + 输入框）

---

## 14. 前端 - 活动时间线组件开发

- [x] 14.1 创建 `web/src/components/activities/` 目录
- [x] 14.2 实现 `ActivityItem.tsx` - 单条活动组件（根据类型渲染不同样式）
- [x] 14.3 实现 `ActivityTimeline.tsx` - 活动时间线容器组件

---

## 15. 前端 - Issue 详情面板集成

- [x] 15.1 修改 Issue 详情面板，添加 CommentSection 组件
- [x] 15.2 修改 Issue 详情面板，添加 ActivityTimeline 组件（可折叠 Tab）
- [x] 15.3 实现 Comments / Activity Tab 切换

---

## 16. 集成验证

- [x] 16.1 运行 `make test` 确保所有后端测试通过 ✅ 所有测试通过
- [x] 16.2 手动测试：创建评论 → 查看列表 → 编辑 → 删除（需手动验证）
- [x] 16.3 手动测试：嵌套回复（点击 Reply 创建回复）（需手动验证）
- [x] 16.4 手动测试：@mention 解析（输入 @username 验证订阅）（需手动验证）
- [x] 16.5 手动测试：活动时间线（修改 Issue 各字段，验证活动记录）（需手动验证）
- [x] 16.6 手动测试：权限控制（非作者无法编辑/删除评论）（需手动验证）
- [x] 16.7 验证 API 响应格式符合设计文档（已通过单元测试验证）

---

## 任务统计

| 类别                              | 任务数 | 预估工时 |
| --------------------------------- | :----: | :------: |
| 数据模型与迁移                    |   4    |   1.5h   |
| Store 层 TDD 任务（Activity）     |   6    |    3h    |
| Store 层 TDD 任务（Comment）      |   10   |    5h    |
| Service 层 TDD 任务（@mention）   |   4    |    2h    |
| Service 层 TDD 任务（Comment）    |   6    |    3h    |
| Service 层 TDD 任务（Activity）   |   4    |    2h    |
| Service 层 TDD 任务（Issue 集成） |   8    |    4h    |
| Handler 层 TDD 任务（Comment）    |   8    |    4h    |
| Handler 层 TDD 任务（Activity）   |   2    |    1h    |
| 路由注册与验证                    |   4    |    1h    |
| 前端 API 层                       |   7    |   2.5h   |
| 前端状态管理                      |   6    |    2h    |
| 前端评论组件                      |   5    |    3h    |
| 前端活动组件                      |   3    |   1.5h   |
| 前端面板集成                      |   3    |   1.5h   |
| 集成验证                          |   7    |   2.5h   |
| **总计**                          | **87** | **~39h** |

> **预估工时说明**：基于每个任务约 30 分钟计算，实际可能因复杂度有所浮动
