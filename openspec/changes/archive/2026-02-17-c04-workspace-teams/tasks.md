# C04 — Workspace 与 Teams 实现任务清单

> **TDD 开发原则**：严格遵循 Red-Green-Refactor 循环
> - 🔴 Red：先写失败的测试
> - 🟢 Green：写最少的代码让测试通过
> - 🔵 Refactor：重构代码（保持测试通过）

---

## 1. 后端 - Workspace Store 层

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

- [x] 1.1 🔴 编写 WorkspaceStore.GetByID 集成测试（表格驱动：正常获取、工作区不存在）
- [x] 1.2 🟢 实现 WorkspaceStore.GetByID
- [x] 1.3 🔴 编写 WorkspaceStore.Update 集成测试（表格驱动：更新名称、更新 Logo、Slug 重复）
- [x] 1.4 🟢 实现 WorkspaceStore.Update
- [x] 1.5 🔴 编写 WorkspaceStore.GetStats 集成测试（表格驱动：正常统计、空工作区）
- [x] 1.6 🟢 实现 WorkspaceStore.GetStats

---

## 2. 后端 - Team Store 层

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

- [x] 2.1 🔴 编写 TeamStore.List 集成测试（表格驱动：按 workspace 过滤、分页）
- [x] 2.2 🟢 实现 TeamStore.List
- [x] 2.3 🔴 编写 TeamStore.GetByID 集成测试（表格驱动：正常获取、团队不存在）
- [x] 2.4 🟢 实现 TeamStore.GetByID
- [x] 2.5 🔴 编写 TeamStore.Create 集成测试（表格驱动：正常创建、Key 重复、Key 格式错误）
- [x] 2.6 🟢 实现 TeamStore.Create
- [x] 2.7 🔴 编写 TeamStore.Update 集成测试（表格驱动：更新名称、更新 Key、Key 重复）
- [x] 2.8 🟢 实现 TeamStore.Update
- [x] 2.9 🔴 编写 TeamStore.SoftDelete 集成测试（表格驱动：正常删除、团队不存在）
- [x] 2.10 🟢 实现 TeamStore.SoftDelete
- [x] 2.11 🔴 编写 TeamStore.CountIssuesByTeam 集成测试（表格驱动：有 Issue、无 Issue）
- [x] 2.12 🟢 实现 TeamStore.CountIssuesByTeam

---

## 3. 后端 - TeamMember Store 层

> 使用真实 PostgreSQL 数据库进行集成测试，拒绝 Mocks

- [x] 3.1 🔴 编写 TeamMemberStore.List 集成测试（表格驱动：正常列表、空团队）
- [x] 3.2 🟢 实现 TeamMemberStore.List
- [x] 3.3 🔴 编写 TeamMemberStore.Add 集成测试（表格驱动：正常添加、重复添加、用户不存在）
- [x] 3.4 🟢 实现 TeamMemberStore.Add
- [x] 3.5 🔴 编写 TeamMemberStore.Remove 集成测试（表格驱动：正常移除、成员不存在、最后一个 Owner）
- [x] 3.6 🟢 实现 TeamMemberStore.Remove
- [x] 3.7 🔴 编写 TeamMemberStore.UpdateRole 集成测试（表格驱动：提升为 Owner、降级为 Member、最后一个 Owner 降级）
- [x] 3.8 🟢 实现 TeamMemberStore.UpdateRole
- [x] 3.9 🔴 编写 TeamMemberStore.GetRole 集成测试（表格驱动：Owner、Member、非成员）
- [x] 3.10 🟢 实现 TeamMemberStore.GetRole

---

## 4. 后端 - 权限中间件扩展

> 使用真实数据库进行集成测试

- [x] 4.1 🔴 编写 GetTeamRole 集成测试（表格驱动：Owner、Member、非成员、Admin 绕过）
- [x] 4.2 🟢 实现 GetTeamRole 函数
- [x] 4.3 🔴 编写 IsTeamOwner 集成测试（表格驱动：是 Owner、不是 Owner、Admin 绕过）
- [x] 4.4 🟢 实现 IsTeamOwner 函数
- [x] 4.5 🔴 编写 IsTeamMember 集成测试（表格驱动：是成员、不是成员）
- [x] 4.6 🟢 实现 IsTeamMember 函数
- [x] 4.7 🔴 编写 RequireTeamOwner 中间件集成测试（表格驱动：Owner 通过、Member 拒绝、Admin 绕过）
- [x] 4.8 🟢 实现 RequireTeamOwner 中间件
- [x] 4.9 🔴 编写 RequireTeamMember 中间件集成测试（表格驱动：成员通过、非成员拒绝）
- [x] 4.10 🟢 实现 RequireTeamMember 中间件

---

## 5. 后端 - Workspace Service 层

> 使用真实数据库进行集成测试

- [x] 5.1 🔴 编写 WorkspaceService.GetWorkspace 集成测试（表格驱动：正常获取、无权限访问）
- [x] 5.2 🟢 实现 WorkspaceService.GetWorkspace
- [x] 5.3 🔴 编写 WorkspaceService.UpdateWorkspace 集成测试（表格驱动：Admin 更新名称、Admin 更新 Logo、Member 无权限）
- [x] 5.4 🟢 实现 WorkspaceService.UpdateWorkspace
- [x] 5.5 🔴 编写 WorkspaceService.GetWorkspaceStats 集成测试（表格驱动：正常统计、Admin 权限）
- [x] 5.6 🟢 实现 WorkspaceService.GetWorkspaceStats

---

## 6. 后端 - Team Service 层

> 使用真实数据库进行集成测试

- [x] 6.1 🔴 编写 TeamKey 格式校验单元测试（表格驱动：有效 Key、无效 Key）
- [x] 6.2 🟢 实现 ValidateTeamKey 函数
- [x] 6.3 🔴 编写 TeamService.CreateTeam 集成测试（表格驱动：Admin 创建、创建者成为 Owner、Key 重复、Member 无权限）
- [x] 6.4 🟢 实现 TeamService.CreateTeam
- [x] 6.5 🔴 编写 TeamService.ListTeams 集成测试（表格驱动：按 workspace 过滤、分页）
- [x] 6.6 🟢 实现 TeamService.ListTeams
- [x] 6.7 🔴 编写 TeamService.GetTeam 集成测试（表格驱动：公开团队、私有团队成员访问、私有团队非成员拒绝）
- [x] 6.8 🟢 实现 TeamService.GetTeam
- [x] 6.9 🔴 编写 TeamService.UpdateTeam 集成测试（表格驱动：Team Owner 更新、Admin 更新、普通成员无权限）
- [x] 6.10 🟢 实现 TeamService.UpdateTeam
- [x] 6.11 🔴 编写 TeamService.DeleteTeam 集成测试（表格驱动：Team Owner 删除、存在 Issue 时拒绝、普通成员无权限）
- [x] 6.12 🟢 实现 TeamService.DeleteTeam

---

## 7. 后端 - TeamMember Service 层

> 使用真实数据库进行集成测试

- [x] 7.1 🔴 编写 TeamMemberService.ListMembers 集成测试（表格驱动：成员列表、非成员拒绝访问私有团队）
- [x] 7.2 🟢 实现 TeamMemberService.ListMembers
- [x] 7.3 🔴 编写 TeamMemberService.AddMember 集成测试（表格驱动：Owner 添加、重复添加、Admin 添加）
- [x] 7.4 🟢 实现 TeamMemberService.AddMember
- [x] 7.5 🔴 编写 TeamMemberService.RemoveMember 集成测试（表格驱动：Owner 移除、移除最后一个 Owner 拒绝）
- [x] 7.6 🟢 实现 TeamMemberService.RemoveMember
- [x] 7.7 🔴 编写 TeamMemberService.UpdateRole 集成测试（表格驱动：Owner 提升成员、自己降级拒绝）
- [x] 7.8 🟢 实现 TeamMemberService.UpdateRole

---

## 8. 后端 - Workspace Handler 层

> 使用真实数据库进行 HTTP 集成测试

- [x] 8.1 🔴 编写 GET /api/v1/workspaces/:id 集成测试（表格驱动：正常响应、无权限）
- [x] 8.2 🟢 实现 WorkspaceHandler.GetWorkspace
- [x] 8.3 🔴 编写 PUT /api/v1/workspaces/:id 集成测试（表格驱动：更新名称、更新 Logo、Member 无权限）
- [x] 8.4 🟢 实现 WorkspaceHandler.UpdateWorkspace

---

## 9. 后端 - Team Handler 层

> 使用真实数据库进行 HTTP 集成测试

- [x] 9.1 🔴 编写 GET /api/v1/teams 集成测试（表格驱动：按 workspace 过滤、分页）
- [x] 9.2 🟢 实现 TeamHandler.ListTeams
- [x] 9.3 🔴 编写 POST /api/v1/teams 集成测试（表格驱动：Admin 创建、Key 格式错误、Member 无权限）
- [x] 9.4 🟢 实现 TeamHandler.CreateTeam
- [x] 9.5 🔴 编写 GET /api/v1/teams/:id 集成测试（表格驱动：公开团队、私有团队成员访问、私有团队非成员拒绝）
- [x] 9.6 🟢 实现 TeamHandler.GetTeam
- [x] 9.7 🔴 编写 PUT /api/v1/teams/:id 集成测试（表格驱动：Team Owner 更新、Admin 更新、普通成员无权限）
- [x] 9.8 🟢 实现 TeamHandler.UpdateTeam
- [x] 9.9 🔴 编写 DELETE /api/v1/teams/:id 集成测试（表格驱动：Team Owner 删除、存在 Issue 时拒绝）
- [x] 9.10 🟢 实现 TeamHandler.DeleteTeam

---

## 10. 后端 - TeamMember Handler 层

> 使用真实数据库进行 HTTP 集成测试

- [x] 10.1 🔴 编写 GET /api/v1/teams/:id/members 集成测试（表格驱动：成员列表、非成员拒绝）
- [x] 10.2 🟢 实现 TeamMemberHandler.ListMembers
- [x] 10.3 🔴 编写 POST /api/v1/teams/:id/members 集成测试（表格驱动：Owner 添加、重复添加 409）
- [x] 10.4 🟢 实现 TeamMemberHandler.AddMember
- [x] 10.5 🔴 编写 DELETE /api/v1/teams/:id/members/:uid 集成测试（表格驱动：Owner 移除、移除最后一个 Owner 拒绝）
- [x] 10.6 🟢 实现 TeamMemberHandler.RemoveMember
- [x] 10.7 🔴 编写 PUT /api/v1/teams/:id/members/:uid 集成测试（表格驱动：Owner 更新角色、自己降级拒绝）
- [x] 10.8 🟢 实现 TeamMemberHandler.UpdateMemberRole

---

## 11. 后端 - 路由注册

- [x] 11.1 🔴 编写路由集成测试（表格驱动：Workspace 路由、Team 路由、TeamMember 路由、中间件链）
- [x] 11.2 🟢 注册所有路由并配置中间件

---

## 12. 前端 - API 层

- [x] 12.1 创建 web/src/types/workspace.ts，定义 Workspace 相关 TypeScript 类型
- [x] 12.2 创建 web/src/types/team.ts，定义 Team、TeamMember 相关 TypeScript 类型
- [x] 12.3 创建 web/src/api/workspace.ts，实现 Workspace API（getWorkspace、updateWorkspace）
- [x] 12.4 创建 web/src/api/team.ts，实现 Team API（listTeams、createTeam、getTeam、updateTeam、deleteTeam）
- [x] 12.5 实现 TeamMember API（listMembers、addMember、removeMember、updateRole）

---

## 13. 前端 - 状态管理（Zustand Store）

- [x] 13.1 创建 web/src/stores/workspaceStore.ts（workspace 状态、loading、error、fetchWorkspace、updateWorkspace）
- [x] 13.2 创建 web/src/stores/teamStore.ts（teams 列表、当前团队、成员列表、CRUD actions）

---

## 14. 前端 - 组件开发与 API 对接

- [x] 14.1 创建 web/src/components/settings/WorkspaceSettings.tsx（名称、Logo 编辑表单）+ API 对接
- [x] 14.2 创建 web/src/components/settings/TeamList.tsx（团队列表、创建按钮）+ API 对接
- [x] 14.3 创建 web/src/components/settings/CreateTeamDialog.tsx（团队名称、Key 输入、校验）+ API 对接
- [x] 14.4 创建 web/src/components/settings/TeamDetail.tsx（团队信息展示、编辑）+ API 对接
- [x] 14.5 创建 web/src/components/settings/TeamMemberList.tsx（成员列表、角色显示、操作按钮）+ API 对接
- [x] 14.6 创建 web/src/components/settings/AddMemberDialog.tsx（用户选择、角色选择）+ API 对接
- [x] 14.7 创建 web/src/components/settings/TeamKeyInput.tsx（Key 格式校验、唯一性异步校验）

---

## 15. 前端 - 页面与路由集成

- [x] 15.1 创建 web/src/pages/Settings/Workspace.tsx 页面并集成 WorkspaceSettings 组件
- [x] 15.2 创建 web/src/pages/Settings/Teams.tsx 页面并集成 TeamList、CreateTeamDialog 组件
- [x] 15.3 创建 web/src/pages/Settings/TeamDetail.tsx 页面并集成 TeamDetail、TeamMemberList、AddMemberDialog 组件
- [x] 15.4 配置前端路由（/settings/workspace、/settings/teams、/settings/teams/:id）

---

## 16. 端到端验证

- [x] 16.1 运行完整后端测试套件（make test）确保通过
- [x] 16.2 启动完整开发环境（Docker Compose）
- [x] 16.3 验证 Workspace 设置页面功能（查看、更新名称）
- [x] 16.4 验证 Teams 管理页面功能（创建、编辑、删除团队）
- [x] 16.5 验证团队成员管理功能（添加、移除、角色更新）
- [x] 16.6 验证权限控制（Admin、Team Owner、Member）

---

**任务统计**：
- 后端 TDD 任务：84 个（🔴 测试 + 🟢 实现 配对）
- 前端任务：16 个
- 验证任务：6 个
- **总计：106 个任务**
- **预估工时**：约 5 天

**TDD 任务格式说明**：
- 🔴 = Red 阶段（编写失败的测试）
- 🟢 = Green 阶段（编写实现让测试通过）
- 每个 🔴 任务后必须紧跟对应的 🟢 任务
