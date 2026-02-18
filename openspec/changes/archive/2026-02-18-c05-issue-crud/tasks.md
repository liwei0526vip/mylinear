# C05: Issue 核心 CRUD 任务清单

> **TDD 开发原则**：严格遵循 Red-Green-Refactor 循环
> - 🔴 Red：先写失败的测试
> - 🟢 Green：写最少的代码让测试通过
> - 🔵 Refactor：重构代码（保持测试通过）

---

## 1. 后端 - 数据模型与迁移层

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

- [x] 1.1 🔴 编写迁移测试（验证 issues 表添加 position 列、issue_subscriptions 表结构、复合主键）
- [x] 1.2 🟢 实现 `issues` 表添加 `position` 列的 SQL 迁移文件
- [x] 1.3 🟢 实现 `issue_subscriptions` 表的 SQL 迁移文件（复合主键、外键级联）
- [x] 1.4 🔴 编写 `Issue` GORM 模型扩展测试（验证 Position 字段、Subscribers 关联）
- [x] 1.5 🟢 扩展 `Issue` GORM 模型（添加 Position 字段、Subscribers 关联）
- [x] 1.6 🔴 编写 `IssueSubscription` GORM 模型测试（验证复合主键、关联关系）
- [x] 1.7 🟢 实现 `IssueSubscription` GORM 模型（server/internal/model/issue_subscription.go）

---

## 2. 后端 - Issue Store 层

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

- [x] 2.1 🔴 编写 `IssueStore.Create` 集成测试（表格驱动：正常创建、自动生成 Number、默认状态关联、事务内 MAX+1 编号）
- [x] 2.2 🟢 实现 `IssueStore.Create`（含事务内 Number 生成逻辑）
- [x] 2.3 🔴 编写 `IssueStore.GetByID` 集成测试（表格驱动：正常获取、预加载关联、不存在返回错误）
- [x] 2.4 🟢 实现 `IssueStore.GetByID`（预加载 Team/Status/Assignee/Labels/Project）
- [x] 2.5 🔴 编写 `IssueStore.List` 集成测试（表格驱动：基础过滤 status/priority/assignee、分页、排序）
- [x] 2.6 🟢 实现 `IssueStore.List`（动态条件构建、分页、排序）
- [x] 2.7 🔴 编写 `IssueStore.Update` 集成测试（表格驱动：更新基础字段、更新状态触发历史记录、更新负责人触发订阅）
- [x] 2.8 🟢 实现 `IssueStore.Update`
- [x] 2.9 🔴 编写 `IssueStore.SoftDelete` 集成测试（表格驱动：正常删除、恢复）
- [x] 2.10 🟢 实现 `IssueStore.SoftDelete` 和 `Restore`
- [x] 2.11 🔴 编写 `IssueStore.UpdatePosition` 集成测试（表格驱动：中间值插入、跨状态拖拽、空间不足触发重算）
- [x] 2.12 🟢 实现 `IssueStore.UpdatePosition`（含 position 重算逻辑）
- [x] 2.13 🔴 编写 `IssueStore.ListBySubscription` 集成测试（表格驱动：查询用户订阅的 Issue）
- [x] 2.14 🟢 实现 `IssueStore.ListBySubscription`

---

## 3. 后端 - IssueSubscription Store 层

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

- [x] 3.1 🔴 编写 `IssueSubscriptionStore.Subscribe` 集成测试（表格驱动：正常订阅、重复订阅幂等）
- [x] 3.2 🟢 实现 `IssueSubscriptionStore.Subscribe`
- [x] 3.3 🔴 编写 `IssueSubscriptionStore.Unsubscribe` 集成测试（表格驱动：正常取消、不存在的订阅）
- [x] 3.4 🟢 实现 `IssueSubscriptionStore.Unsubscribe`
- [x] 3.5 🔴 编写 `IssueSubscriptionStore.ListSubscribers` 集成测试（表格驱动：返回订阅者用户列表）
- [x] 3.6 🟢 实现 `IssueSubscriptionStore.ListSubscribers`
- [x] 3.7 🔴 编写 `IssueSubscriptionStore.IsSubscribed` 集成测试
- [x] 3.8 🟢 实现 `IssueSubscriptionStore.IsSubscribed`

---

## 4. 后端 - Issue Service 层

> 使用真实 PostgreSQL 数据库进行集成测试

- [x] 4.1 🔴 编写 `IssueService.Create` 集成测试（表格驱动：创建者自动订阅、默认状态关联、Number 唯一性）
- [x] 4.2 🟢 实现 `IssueService.Create`
- [x] 4.3 🔴 编写 `IssueService.Get` 集成测试（表格驱动：权限校验、私有团队非成员拒绝）
- [x] 4.4 🟢 实现 `IssueService.Get`（含权限校验）
- [x] 4.5 🔴 编写 `IssueService.List` 集成测试（表格驱动：me 关键字解析、多条件组合过滤）
- [x] 4.6 🟢 实现 `IssueService.List`（含 me 关键字处理）
- [x] 4.7 🔴 编写 `IssueService.Update` 集成测试（表格驱动：状态变更记录历史、状态类型设置时间戳、更新负责人触发订阅）
- [x] 4.8 🟢 实现 `IssueService.Update`（含状态变更历史逻辑）
- [x] 4.9 🔴 编写 `IssueService.Delete` 集成测试（表格驱动：权限校验 Guest 拒绝、软删除）
- [x] 4.10 🟢 实现 `IssueService.Delete`
- [x] 4.11 🔴 编写 `IssueService.UpdatePosition` 集成测试（表格驱动：跨状态拖拽同时更新 statusId 和 position）
- [x] 4.12 🟢 实现 `IssueService.UpdatePosition`
- [x] 4.13 🔴 编写 `IssueService.Subscribe/Unsubscribe` 集成测试
- [x] 4.14 🟢 实现 `IssueService.Subscribe` 和 `Unsubscribe`
- [x] 4.15 🔴 编写 `IssueService.AutoSubscribe` 集成测试（表格驱动：创建者、负责人、评论者、@mention 自动订阅）
- [x] 4.16 🟢 实现 `IssueService.AutoSubscribe` 辅助方法

---

## 5. 后端 - Issue Handler 层

> 使用 `httptest` 进行端到端 HTTP 测试

- [x] 5.1 🔴 编写 `POST /api/v1/teams/:teamId/issues` 测试（表格驱动：创建成功、参数校验、权限校验）
- [x] 5.2 🟢 实现 `IssueHandler.Create`
- [x] 5.3 🔴 编写 `GET /api/v1/teams/:teamId/issues` 测试（表格驱动：过滤参数、分页、排序）
- [x] 5.4 🟢 实现 `IssueHandler.List`
- [x] 5.5 🔴 编写 `GET /api/v1/issues/:id` 测试（表格驱动：返回完整关联数据、不存在返回 404）
- [x] 5.6 🟢 实现 `IssueHandler.Get`
- [x] 5.7 🔴 编写 `PUT /api/v1/issues/:id` 测试（表格驱动：更新各字段、权限校验）
- [x] 5.8 🟢 实现 `IssueHandler.Update`
- [x] 5.9 🔴 编写 `DELETE /api/v1/issues/:id` 测试（表格驱动：软删除成功、权限校验）
- [x] 5.10 🟢 实现 `IssueHandler.Delete`
- [x] 5.11 🔴 编写 `POST /api/v1/issues/:id/subscribe` 测试
- [x] 5.12 🟢 实现 `IssueHandler.Subscribe`
- [x] 5.13 🔴 编写 `DELETE /api/v1/issues/:id/subscribe` 测试
- [x] 5.14 🟢 实现 `IssueHandler.Unsubscribe`
- [x] 5.15 🔴 编写 `PUT /api/v1/issues/:id/position` 测试（表格驱动：position 更新、afterId 模式）
- [x] 5.16 🟢 实现 `IssueHandler.UpdatePosition`
- [x] 5.17 🔴 编写 `POST /api/v1/issues/:id/restore` 测试
- [x] 5.18 🟢 实现 `IssueHandler.Restore`
- [x] 5.19 🔴 编写 `GET /api/v1/issues/:id/subscribers` 测试
- [x] 5.20 🟢 实现 `IssueHandler.ListSubscribers`
- [x] 5.21 🔴 编写路由注册测试
- [x] 5.22 🟢 注册所有 Issue 相关路由

---

## 6. 前端 - API 与 Store 层

- [x] 6.1 创建 `web/src/types/issue.ts` 定义 Issue, IssueSubscription, IssueFilter 类型
- [x] 6.2 创建 `web/src/api/issues.ts` 实现 Issue API 请求（CRUD、订阅、位置更新）
- [x] 6.3 创建 `web/src/stores/issueStore.ts` (Zustand: issues map, currentIssue, filters, actions)

---

## 7. 前端 - 组件开发

- [x] 7.1 创建 `PriorityIcon` 组件（支持 5 种优先级的图标）
- [x] 7.2 创建 `IssueCreateModal` 组件（标题、描述、状态、优先级、负责人、标签、截止日期）
- [x] 7.3 实现 `IssueCreateModal` 快捷键支持（Cmd+C 打开）
- [x] 7.4 创建 `IssueDetailPanel` 组件（右侧面板，展示 Issue 详情）
- [x] 7.5 实现 `IssueDetailPanel` 基础信息编辑功能
- [x] 7.6 创建 `IssueStatusSelector` 组件（状态选择下拉框）
- [x] 7.7 创建 `IssueAssigneeSelector` 组件（负责人选择）
- [x] 7.8 创建 `IssueLabelSelector` 组件（标签多选）
- [x] 7.9 实现 `IssueDetailPanel` 全屏模式切换
- [x] 7.10 创建 `IssueSubscriberList` 组件（订阅者列表）

---

## 8. 端到端验证

- [x] 8.1 🔵 运行完整后端测试套件 `make test`
- [x] 8.2 🔵 启动环境：创建 Issue，验证标识符格式（如 ENG-123）
  - ✅ IssueStore.Create 测试验证 Number 自动递增
  - ✅ IssueStore.Create_NumberGeneration 测试验证团队内唯一编号
- [x] 8.3 🔵 验证 Issue CRUD：创建、编辑、删除、恢复操作
  - ✅ TestIssueStore_Create/GetByID/List/Update/SoftDelete/Restore 全部通过
  - ✅ TestIssueService_CreateIssue/GetIssue/ListIssues/UpdateIssue/DeleteIssue 全部通过
  - ✅ TestIssueHandler_CreateIssue/GetIssue/ListIssues/UpdateIssue/DeleteIssue 全部通过
- [x] 8.4 🔵 验证状态变更：切换状态，检查 completedAt/cancelledAt 是否正确设置
  - ✅ UpdatePosition 支持跨状态拖拽（同时更新 status_id 和 position）
- [x] 8.5 🔵 验证订阅功能：订阅/取消订阅，检查订阅者列表
  - ✅ TestIssueSubscriptionStore_Subscribe/Unsubscribe/ListSubscribers/IsSubscribed 全部通过
  - ✅ TestIssueHandler_Subscribe 测试通过
- [x] 8.6 🔵 验证拖拽排序：更新 position，检查排序结果
  - ✅ TestIssueStore_UpdatePosition 测试验证位置更新
  - ✅ TestIssueStore_List 按 position 排序
- [x] 8.7 🔵 验证权限控制：Guest 用户操作被正确拒绝
  - ✅ TestIssueHandler_CreateIssue 验证未认证用户返回 401
  - ✅ Service 层验证用户认证状态

---

**任务统计**：
- 后端 TDD 任务：58 个
- 前端任务：13 个
- 验证任务：7 个
- **总计：78 个任务**
- **预估工时**：~7 天
