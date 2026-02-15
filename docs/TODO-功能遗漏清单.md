# 功能遗漏清单（待处理）

> **创建时间**：2026-02-15
> **来源**：Linear 截图分析 vs 现有文档对比
> **状态**：待逐项处理

---

## 🔴 完全遗漏（文档和路线图均未提及）

| # | 功能 | 重要性 | 建议阶段 | 说明 | 处理状态 |
|---|------|:------:|:------:|------|:------:|
| 1 | ~~用户个人资料（Profile）~~ | ★★★★★ | Phase 1 | 头像、邮箱、全名、用户名、工作区访问管理 | ✅ 已补充 |
| 2 | 用户偏好设置（Preferences） | ★★★★☆ | Phase 2 | 默认主页、名称显示方式、一周起始日、表情转换等 | ⬜ |
| 3 | 关联账号（Connected accounts） | ★★☆☆☆ | Phase 3 | 第三方账号绑定（GitHub、Google 等） | ⬜ |
| 4 | App sidebar 自定义 | ★★★☆☆ | Phase 2 | 侧边栏项目可见性、排序、徽章样式自定义 | ⬜ |
| 5 | 字号设置（Font size） | ★★☆☆☆ | Phase 4 | 全局字体大小调整（无障碍访问相关） | ⬜ |
| 6 | 指针样式设置（Pointer cursors） | ★☆☆☆☆ | Phase 4 | 细节交互偏好 | ⬜ |
| 7 | Emojis 功能开关 | ★★☆☆☆ | Phase 3 | 自定义表情和表情反应 | ⬜ |
| 8 | Applications（OAuth 应用管理） | ★★★☆☆ | Phase 3 | 管理第三方 OAuth 应用注册和授权 | ⬜ |
| 9 | 桌面应用设置 | ★★☆☆☆ | Phase 5 | Open in desktop app、通知徽章样式 | ⬜ |
| 10 | Cycle "Days to start" 指标 | ★★☆☆☆ | Phase 2 | Cycle 距离开始的倒计时天数 | ⬜ |

---

## 🟡 部分遗漏（文档有提及但不够详细）

| # | 功能 | 当前覆盖 | 遗漏点 | 处理状态 |
|---|------|---------|-------|:------:|
| 1 | 安全与访问（Security & access） | 仅提到 SSO/LDAP | 缺少：活跃会话管理、密码修改、2FA、API Token 管理页面、登录设备管理 | ⬜ |
| 2 | 深色/浅色主题 | 已规划 #66 | 缺少 "System（跟随系统）" 选项和自定义主色调 | ⬜ |
| 3 | 成员邀请流程 | 提及成员管理 | 缺少具体的邀请流程描述（邀请链接、邮件邀请、批量邀请） | ⬜ |
| 4 | Projects 设置独立分区 | 项目功能已覆盖 | Settings 中 Projects 作为独立设置分区的设计：Labels / Templates / Statuses / Updates | ⬜ |
| 5 | Features 开关管理 | 各功能已覆盖 | 功能模块可以独立开/关（Initiatives、Documents、Customer requests、Pulse、AI、Agents、Asks、Emojis），需要功能开关机制 | ⬜ |

---

## 📝 建议新增文档

- [ ] 新增 `02-功能模块.md` 章节："Settings/Preferences（设置与偏好）"，统一描述 Linear 设置页面结构
- [ ] 更新 `04-功能索引.md`，将设置相关功能纳入索引
- [ ] 更新 `路线图.md`，将遗漏功能补充到对应阶段

---
