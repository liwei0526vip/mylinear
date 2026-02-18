## ADDED Requirements

### Requirement: Issue 订阅机制

系统 SHALL 支持 Issue 订阅功能，用户可追踪关注的 Issue 变更。

#### Scenario: 订阅 Issue

- **WHEN** 用户通过 `POST /api/v1/issues/:id/subscribe` 订阅 Issue
- **THEN** 系统 SHALL 在 `issue_subscriptions` 表创建订阅记录
- **AND** 用户后续 SHALL 收到该 Issue 的变更通知

#### Scenario: 取消订阅 Issue

- **WHEN** 用户通过 `DELETE /api/v1/issues/:id/subscribe` 取消订阅
- **THEN** 系统 SHALL 删除对应的订阅记录
- **AND** 用户不再收到该 Issue 的变更通知（除非重新订阅）

#### Scenario: 重复订阅幂等

- **WHEN** 用户重复订阅同一 Issue
- **THEN** 系统 SHALL 返回成功（幂等操作，不创建重复记录）

#### Scenario: 查询订阅者列表

- **WHEN** 用户通过 `GET /api/v1/issues/:id/subscribers` 查询订阅者
- **THEN** 系统 SHALL 返回该 Issue 的所有订阅者列表（包含用户基本信息）

---

### Requirement: 自动订阅规则

系统 SHALL 根据用户行为自动管理订阅关系。

#### Scenario: 创建者自动订阅

- **WHEN** 用户创建 Issue
- **THEN** 系统 SHALL 自动将创建者添加为订阅者

#### Scenario: 负责人自动订阅

- **WHEN** Issue 被分配给用户
- **THEN** 系统 SHALL 自动将该用户添加为订阅者（若未订阅）

#### Scenario: 评论者自动订阅

- **WHEN** 用户在 Issue 下发表评论
- **THEN** 系统 SHALL 自动将该用户添加为订阅者（若未订阅）

#### Scenario: @mention 自动订阅

- **WHEN** Issue 描述或评论中 @mention 某用户
- **THEN** 系统 SHALL 自动将被提及的用户添加为订阅者

---

### Requirement: 订阅与通知联动

订阅状态 SHALL 影响通知推送行为。

#### Scenario: 订阅者接收通知

- **WHEN** Issue 发生变更（状态、负责人、评论等）
- **THEN** 系统 SHALL 向所有订阅者推送通知

#### Scenario: 取消订阅后停止通知

- **WHEN** 用户取消订阅 Issue
- **THEN** 系统 SHALL 停止向该用户推送该 Issue 的后续通知

#### Scenario: 负责人变更时通知

- **WHEN** Issue 负责人变更
- **THEN** 系统 SHALL 通知原负责人（若仍订阅）和新负责人

---

### Requirement: 订阅数据模型

订阅关系 SHALL 通过独立数据表管理。

#### Scenario: 订阅表结构

- **WHEN** 系统初始化
- **THEN** `issue_subscriptions` 表 SHALL 包含字段：`issue_id`（FK）、`user_id`（FK）、`created_at`
- **AND** 复合主键 SHALL 为 (`issue_id`, `user_id`)

#### Scenario: 级联删除

- **WHEN** Issue 被删除
- **THEN** 相关订阅记录 SHALL 自动删除（级联）

#### Scenario: 用户删除时清理

- **WHEN** 用户被删除
- **THEN** 该用户的所有订阅记录 SHALL 自动删除（级联）
