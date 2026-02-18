# C08 — 评论与活动流

## Why

Issue 协作需要两个核心能力：**评论讨论**和**变更追踪**。

当前 Issue 详情面板缺少评论功能，团队成员无法直接在 Issue 下讨论问题；同时也缺少活动时间线，无法追溯 Issue 的变更历史（谁在什么时候改了什么）。这两个功能是团队协作的基础设施，需要在 C09 通知系统之前完成。

## What Changes

### 新增功能

- **评论 CRUD**：用户可在 Issue 下发表、编辑、删除评论
- **嵌套回复**：支持评论的层级回复（parent_id 机制）
- **@mention 解析**：评论中 @用户名 自动解析并关联，触发订阅
- **Markdown 支持**：评论内容支持 Markdown 格式化
- **编辑标记**：评论被编辑后显示"已编辑"标识和编辑时间
- **活动时间线**：记录 Issue 的所有变更事件，包括：
  - 状态变更（从 X 状态变为 Y 状态）
  - 字段修改（标题、描述、优先级、负责人、截止日期等）
  - 新增评论
  - 关联/取消关联 Project、Cycle、Label 等
- **活动查询 API**：支持按 Issue 查询活动历史，分页返回

### 关联功能项

| 编号 | 功能 | 说明 |
|------|------|------|
| #40 | 评论 | @mention、Markdown、嵌套回复 |
| #41 | 活动流 | 显示所有操作历史和时间线 |

## Scope

### In Scope（范围内）

1. **评论系统**
   - Comment 模型已存在于 C02，需实现 CRUD API
   - 嵌套回复（通过 `parent_id` 实现）
   - @mention 解析（从评论 body 中提取 `@username` 模式）
   - 评论编辑与删除（权限控制）

2. **活动流系统**
   - 新增 `activities` 表记录 Issue 变更事件
   - 活动类型定义：`status_changed`、`field_updated`、`comment_added`、`relation_changed` 等
   - 活动 API：按 Issue 查询活动列表

3. **状态历史写入**
   - Issue 状态变更时写入 `issue_status_history` 表（表已存在，逻辑未实现）

4. **前端组件**
   - Issue 详情面板的评论区（评论列表 + 输入框）
   - Issue 详情面板的活动时间线

### Out of Scope（范围外）

- **通知推送**：C09 实现
- **实时 WebSocket 推送**：C41 实现
- **富文本编辑器增强**（图片上传、代码块语法高亮）：C19 实现
- **评论表情反应**：Phase 4 Pulse 功能

## Capabilities

### New Capabilities

- `comments`：评论 CRUD、嵌套回复、@mention 解析、Markdown 支持
- `activity-stream`：活动记录、活动查询、状态历史写入

### Modified Capabilities

- `issues`：需要在 Issue 更新时触发活动记录写入、状态变更时写入 `issue_status_history`

## Approach

### 技术方案概述

1. **评论系统**
   - 利用已有的 `comments` 表和 GORM 模型
   - 实现 CommentStore、CommentService、CommentHandler 三层架构
   - @mention 解析使用正则提取 `@[\w]+`，查询 User 表匹配 username

2. **活动流系统**
   - 新增 `activities` 表，字段包括：`id`、`issue_id`、`type`、`actor_id`、`payload`（JSONB）、`created_at`
   - 实现统一的 ActivityService，在 Issue 变更时调用记录活动
   - 活动类型使用 ENUM 或常量定义

3. **状态历史**
   - 在 Issue 更新状态时，调用 IssueStatusHistoryService 写入 `issue_status_history`

4. **前端实现**
   - 使用现有的 shadcn/ui 组件
   - 评论列表使用虚拟滚动优化性能（大量评论场景）
   - 活动时间线按时间倒序展示

### 与前置依赖的关系

- **依赖 C05（Issue CRUD）**：评论和活动都关联到 Issue
- **依赖 C02（数据库模型）**：`comments` 表和 `issue_status_history` 表已定义
- **为 C09（通知系统）准备**：活动记录是通知推送的数据源

## Impact

### 后端

- 新增 `internal/store/comment_store.go`
- 新增 `internal/store/activity_store.go`
- 新增 `internal/service/comment_service.go`
- 新增 `internal/service/activity_service.go`
- 新增 `internal/handler/comment_handler.go`
- 新增 `internal/handler/activity_handler.go`
- 修改 `internal/service/issue_service.go`：集成活动记录
- 新增数据库迁移：`activities` 表

### 前端

- 新增 `src/api/comments.ts`
- 新增 `src/api/activities.ts`
- 新增 `src/stores/commentStore.ts`
- 新增 `src/stores/activityStore.ts`
- 新增 `src/components/comments/CommentList.tsx`
- 新增 `src/components/comments/CommentInput.tsx`
- 新增 `src/components/activities/ActivityTimeline.tsx`
- 修改 Issue 详情面板：集成评论区和活动时间线

### API 端点

- `POST /api/v1/issues/:issueId/comments` - 创建评论
- `GET /api/v1/issues/:issueId/comments` - 获取评论列表
- `PUT /api/v1/comments/:id` - 更新评论
- `DELETE /api/v1/comments/:id` - 删除评论
- `GET /api/v1/issues/:issueId/activities` - 获取活动时间线
