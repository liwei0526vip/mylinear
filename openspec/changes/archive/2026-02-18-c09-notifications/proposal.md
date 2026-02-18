## Why

用户需要一个集中的地方查看所有与己相关的通知，包括被指派任务、被 @mention、订阅的 Issue 发生变更等。通知系统是团队协作的基础功能，确保用户不会错过重要的工作更新。作为 MVP 的核心组成部分，通知系统需要支持应用内收件箱和通知配置功能。

## What Changes

- **新增通知模型与 API**：创建 `notifications` 表，支持 CRUD、已读/未读标记、批量操作
- **新增通知触发逻辑**：在 Issue 被指派、用户被 @mention、订阅的 Issue 变更时自动创建通知
- **新增通知配置 API**：支持按类型配置通知偏好（后续 Phase 扩展邮件/IM 通知时使用）
- **新增前端通知收件箱页面**：展示通知列表、已读/未读状态、批量操作

## Capabilities

### New Capabilities

- `notifications`: 通知模型、CRUD API、通知触发逻辑、应用内收件箱
- `notification-preferences`: 通知配置模型与 API（预留扩展能力）

### Modified Capabilities

- `activity-stream`: 需要在活动创建时触发通知（扩展集成点，非需求变更）

## Impact

### 依赖关系

- **C08 评论与活动流**：通知系统依赖 Activity 模型和 Comment 模型来触发通知
- **C05 Issue 核心 CRUD**：Issue 指派、状态变更等触发通知
- **C03 用户认证与权限**：通知按用户隔离

### 受影响的代码

- **后端**：
  - `internal/model/notification.go`：新增 Notification 模型
  - `internal/handler/notification.go`：新增通知 API 处理器
  - `internal/service/notification.go`：新增通知业务逻辑
  - `internal/store/notification.go`：新增通知数据访问层
  - `internal/service/activity.go`：扩展活动创建时触发通知
  - `internal/service/comment.go`：扩展评论创建时解析 @mention 触发通知
  - `internal/service/issue.go`：扩展 Issue 指派时触发通知
  - `migrations/`：新增 notifications 表迁移

- **前端**：
  - `src/stores/notification-store.ts`：新增通知状态管理
  - `src/api/notifications.ts`：新增通知 API 调用
  - `src/pages/inbox/`：新增通知收件箱页面
  - `src/components/notification/`：新增通知相关组件

### 对应路线图功能项

- #68 Inbox 通知收件箱
- #72 通知配置
