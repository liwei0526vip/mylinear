## 十二、Webhook 与集成

### 12.1 Webhook 事件类型

#### 12.1.1 Issue 事件

| 事件类型 | 说明 | 触发时机 |
|---------|------|---------|
| `Issue.create` | Issue 创建 | 新 Issue 被创建 |
| `Issue.update` | Issue 更新 | 任意字段变更 |
| `Issue.delete` | Issue 删除 | Issue 被删除 |
| `Issue.archive` | Issue 归档 | Issue 被归档 |
| `Issue.move` | Issue 移动 | 跨团队/跨项目移动 |

#### 12.1.2 Issue 相关事件

| 事件类型 | 说明 |
|---------|------|
| `IssueLabel.create` | 标签添加 |
| `IssueLabel.delete` | 标签移除 |
| `IssueAttachment.create` | 附件上传 |
| `IssueAttachment.delete` | 附件删除 |
| `IssueComment.create` | 评论创建 |
| `IssueComment.update` | 评论更新 |
| `IssueComment.delete` | 评论删除 |
| `IssueCommentReaction.create` | 评论表情反应 |

#### 12.1.3 Project 事件

| 事件类型 | 说明 |
|---------|------|
| `Project.create` | 项目创建 |
| `Project.update` | 项目更新 |
| `Project.delete` | 项目删除 |
| `Project.update.create` | 项目更新通报 |

#### 12.1.4 其他事件

| 事件类型 | 说明 |
|---------|------|
| `Cycle.create` | 迭代创建 |
| `Cycle.update` | 迭代更新 |
| `Cycle.delete` | 迭代删除 |
| `Document.create` | 文档创建 |
| `Document.update` | 文档更新 |
| `Document.delete` | 文档删除 |
| `Customer.create` | 客户创建 |
| `User.create` | 用户加入 |
| `User.delete` | 用户移除 |
| `IssueSLA.breach` | SLA 即将违约 |
| `OAuthApp.revoked` | OAuth 授权撤销 |

### 12.2 Webhook Payload 格式

```json
{
  "id": "evt_1234567890",
  "type": "Issue.update",
  "createdAt": "2026-02-15T10:30:00.000Z",
  "webhookId": "wh_abcdefgh",
  "data": {
    "id": "issue-uuid",
    "identifier": "ENG-123",
    "title": "Fix authentication bug",
    "description": "...",
    "status": {
      "id": "status-uuid",
      "name": "In Progress",
      "type": "started"
    },
    "priority": 2,
    "assignee": {
      "id": "user-uuid",
      "name": "John Doe",
      "email": "john@example.com"
    },
    "team": {
      "id": "team-uuid",
      "key": "ENG",
      "name": "Engineering"
    },
    "updatedAt": "2026-02-15T10:30:00.000Z"
  },
  "changes": {
    "status": {
      "from": { "name": "Todo", "type": "unstarted" },
      "to": { "name": "In Progress", "type": "started" }
    }
  },
  "actor": {
    "id": "user-uuid",
    "name": "John Doe",
    "type": "user"
  }
}
```

### 12.3 Webhook 签名验证

```go
// HMAC-SHA256 签名验证
func VerifyWebhookSignature(payload []byte, signature string, secret string) bool {
    // 签名格式：sha256=<hex-digest>
    if !strings.HasPrefix(signature, "sha256=") {
        return false
    }

    expectedMAC, err := hex.DecodeString(signature[7:])
    if err != nil {
        return false
    }

    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    actualMAC := mac.Sum(nil)

    return hmac.Equal(expectedMAC, actualMAC)
}

// 服务端处理
func HandleWebhook(w http.ResponseWriter, r *http.Request) {
    payload, _ := io.ReadAll(r.Body)
    signature := r.Header.Get("X-Linear-Signature")

    if !VerifyWebhookSignature(payload, signature, webhookSecret) {
        http.Error(w, "Invalid signature", 401)
        return
    }

    // 处理 webhook...
}
```

### 12.4 GitHub 集成配置

#### 12.4.1 PR 自动化规则

| 规则 | 触发条件 | 动作 |
|------|---------|------|
| PR 创建 | PR 标题/描述包含 Issue ID | 关联 Issue，状态 → In Review |
| PR 合并 | PR 合并到主分支 | 状态 → Done，添加评论 |
| PR 关闭 | PR 关闭（未合并） | 状态 → 原状态 |
| Commit | Commit message 包含 Issue ID | 状态 → In Progress |

#### 12.4.2 Commit Message 解析

```
# 支持的格式
ENG-123 fix: resolve authentication issue
fixes ENG-123
closes ENG-123
resolves ENG-123
ENG-123 #close
```

```go
// 解析 Commit Message 中的 Issue ID
var issuePattern = regexp.MustCompile(`(?i)([A-Z]+-\d+)|(?:fixes|closes|resolves)\s+([A-Z]+-\d+)`)

func ParseIssueIDs(message string) []string {
    matches := issuePattern.FindAllStringSubmatch(message, -1)
    var issueIDs []string
    for _, match := range matches {
        if match[1] != "" {
            issueIDs = append(issueIDs, match[1])
        } else if match[2] != "" {
            issueIDs = append(issueIDs, match[2])
        }
    }
    return issueIDs
}
```

### 12.5 Slack 集成配置

#### 12.5.1 消息格式

```json
{
  "blocks": [
    {
      "type": "header",
      "text": {
        "type": "plain_text",
        "text": "ENG-123: Fix authentication bug"
      }
    },
    {
      "type": "section",
      "fields": [
        {
          "type": "mrkdwn",
          "text": "*Status:*\n🟡 In Progress"
        },
        {
          "type": "mrkdwn",
          "text": "*Priority:*\n🔴 High"
        },
        {
          "type": "mrkdwn",
          "text": "*Assignee:*\nJohn Doe"
        },
        {
          "type": "mrkdwn",
          "text": "*Project:*\nQ1 Security Update"
        }
      ]
    },
    {
      "type": "actions",
      "elements": [
        {
          "type": "button",
          "text": { "type": "plain_text", "text": "View Issue" },
          "url": "https://linear.app/issue/ENG-123"
        },
        {
          "type": "button",
          "text": { "type": "plain_text", "text": "Change Status" },
          "action_id": "change_status"
        }
      ]
    }
  ]
}
```

#### 12.5.2 Interactive Actions

```go
// Slack 交互回调处理
func HandleSlackInteraction(payload SlackInteractionPayload) {
    switch payload.Actions[0].ActionID {
    case "change_status":
        // 弹出状态选择菜单
        showStatusMenu(payload.ResponseURL, payload.IssueID)
    case "assign":
        // 弹出成员选择菜单
        showAssigneeMenu(payload.ResponseURL, payload.IssueID)
    }
}
```

---

### 12.6 集成生态分析

#### Webhook 设计要点

- **30+ 事件类型**覆盖 Issue/Project/Cycle/Document/User/SLA 全生命周期
- **Payload 包含 `changes` 字段**：精确记录"从什么变为什么"，便于下游系统精确处理
- **HMAC-SHA256 签名验证**：安全防篡改

#### GitHub/GitLab 集成设计

- **Commit Message 解析**：自动识别 `ENG-123`、`fixes ENG-123`、`closes ENG-123` 等格式
- **PR 状态联动**：PR → In Review → 合并 → Done，实现代码到任务的闭环
- 代码管理集成是开发团队使用率最高的集成，应优先实现

#### AI Agent 生态

Linear 的 Agent 策略是"**平台化**"——不自建 AI 编码能力，而是接入 Codex/Copilot/Cursor/Factory 等外部 Agent，通过 MCP 协议让 AI 模型直接访问 Linear 数据。

#### 最佳实践

- Webhook Payload 中 `changes` 字段的设计是下游自动化的关键——仅推送变化的字段，减少下游解析和判断开销
- 预留 API Keys + Webhooks 接口为后续 AI 集成做准备

---
