## ADDED Requirements

### Requirement: NotificationPreference 模型

NotificationPreference 模型 MUST 定义通知偏好的完整结构。

#### Scenario: NotificationPreference 模型字段

- **WHEN** 定义 `NotificationPreference` 结构体
- **THEN** MUST 包含字段：`ID`（UUID）、`UserID`（FK）、`Channel`（string）、`Type`（string）、`Enabled`（boolean）、`CreatedAt`、`UpdatedAt`

#### Scenario: Channel 字段约束

- **WHEN** 定义 `Channel` 字段
- **THEN** MUST 支持：`in_app`、`email`、`slack`（MVP 仅实现 `in_app`）

#### Scenario: Type 字段约束

- **WHEN** 定义 `Type` 字段
- **THEN** MUST 支持与 `NotificationType` 相同的类型：`issue_assigned`、`issue_mentioned`、`issue_status_changed`、`issue_commented`、`issue_priority_changed`

#### Scenario: 唯一约束

- **WHEN** 用户对同一渠道和类型设置偏好
- **THEN** 数据库 SHALL 强制 `user_id + channel + type` 唯一

#### Scenario: NotificationPreference 表名

- **WHEN** 定义 `NotificationPreference` 的 `TableName()` 方法
- **THEN** MUST 返回 `notification_preferences`

---

### Requirement: 通知配置查询 API

系统 SHALL 支持查询用户的通知配置。

#### Scenario: 获取通知配置列表

- **WHEN** 用户通过 `GET /api/v1/notification-preferences` 查询
- **THEN** 系统 SHALL 返回当前用户所有通知配置
- **AND** 默认包含所有渠道和类型的配置

#### Scenario: 按渠道过滤

- **WHEN** 用户指定 `channel=in_app` 参数
- **THEN** 系统 SHALL 仅返回该渠道的配置

#### Scenario: 未配置返回默认值

- **WHEN** 用户从未配置过某个类型的通知
- **THEN** 系统 SHALL 返回默认配置（`enabled: true`）

---

### Requirement: 通知配置更新 API

系统 SHALL 支持更新用户的通知配置。

#### Scenario: 更新单个配置

- **WHEN** 用户通过 `PUT /api/v1/notification-preferences` 更新配置
- **WITH** 请求体 `{"channel": "in_app", "type": "issue_assigned", "enabled": false}`
- **THEN** 系统 SHALL 更新或创建该配置记录
- **AND** 返回更新后的配置对象

#### Scenario: 批量更新配置

- **WHEN** 用户提供 `preferences` 数组
- **WITH** 请求体 `{"preferences": [{"channel": "in_app", "type": "issue_assigned", "enabled": false}, ...]}`
- **THEN** 系统 SHALL 批量更新所有配置
- **AND** 返回更新后的配置列表

#### Scenario: 更新他人配置

- **WHEN** 用户尝试更新其他用户的配置
- **THEN** 系统 SHALL 返回 403 Forbidden

---

### Requirement: 默认配置

系统 SHALL 为新用户自动使用默认配置。

#### Scenario: 默认全部启用

- **WHEN** 新用户首次查询通知配置
- **THEN** 所有通知类型 SHALL 默认启用（`enabled: true`）

#### Scenario: MVP 仅支持 in_app 渠道

- **WHEN** 用户尝试配置 `email` 或 `slack` 渠道
- **THEN** 系统 SHALL 返回 400 Bad Request
- **AND** 提示 "该渠道暂不支持"

---

### Requirement: 通知发送时检查配置

系统 SHALL 在发送通知前检查用户配置。

#### Scenario: 配置启用时发送通知

- **WHEN** 用户对某类型通知配置为 `enabled: true`
- **THEN** 系统 SHALL 正常创建通知

#### Scenario: 配置禁用时不发送通知

- **WHEN** 用户对某类型通知配置为 `enabled: false`
- **THEN** 系统 SHALL NOT 创建该类型通知

#### Scenario: 无配置时使用默认值

- **WHEN** 用户未配置某类型通知
- **THEN** 系统 SHALL 按默认值（`enabled: true`）处理
