## ADDED Requirements

### Requirement: Project 列表页

前端 SHALL 提供项目列表页面，展示团队下所有项目。

#### Scenario: 列表页路由

- **WHEN** 用户访问 `/teams/:teamSlug/projects` 路由
- **THEN** 系统 SHALL 展示项目列表页面

#### Scenario: 项目卡片展示

- **WHEN** 项目列表加载完成
- **THEN** 每个 项目 SHALL 以卡片形式展示，包含：
  - 项目名称
  - 项目状态（带颜色标识）
  - 进度条（可视化完成百分比）
  - 负责人头像和名称
  - 关联 Issue 数量
  - 目标日期（若有）

#### Scenario: 状态过滤器

- **WHEN** 用户点击状态过滤器
- **THEN** 系统 SHALL 提供按状态筛选的选项：全部 / Planned / In Progress / Paused / Completed / Cancelled

#### Scenario: 项目排序

- **WHEN** 用户查看项目列表
- **THEN** 系统 SHALL 默认按 `updatedAt` 倒序排列
- **AND** 系统 SHALL 支持按名称、状态、目标日期排序

#### Scenario: 空状态展示

- **WHEN** 团队暂无项目
- **THEN** 系统 SHALL 展示空状态提示，包含"创建第一个项目"的引导按钮

---

### Requirement: Project 创建入口

前端 SHALL 提供便捷的项目创建入口。

#### Scenario: 列表页创建按钮

- **WHEN** 用户在项目列表页点击"新建项目"按钮
- **THEN** 系统 SHALL 弹出项目创建模态框

#### Scenario: 创建模态框字段

- **WHEN** 项目创建模态框打开
- **THEN** 模态框 SHALL 包含以下字段：
  - 项目名称（必填）
  - 项目描述（可选，Markdown 编辑器）
  - 负责人选择器（可选）
  - 开始日期（可选）
  - 目标日期（可选）

#### Scenario: 创建成功反馈

- **WHEN** 项目创建成功
- **THEN** 系统 SHALL 关闭模态框
- **AND** 系统 SHALL 在列表中新增项目卡片
- **AND** 系统 SHALL 展示成功提示

---

### Requirement: Project 详情页

前端 SHALL 提供项目详情页面，展示项目完整信息。

#### Scenario: 详情页路由

- **WHEN** 用户访问 `/projects/:projectId` 路由
- **THEN** 系统 SHALL 展示项目详情页面

#### Scenario: 详情页头部信息

- **WHEN** 项目详情页加载完成
- **THEN** 页面头部 SHALL 展示：
  - 项目名称
  - 项目状态（可点击切换）
  - 负责人头像和名称
  - 进度统计（完成/总数 + 百分比）
  - 开始日期和目标日期

#### Scenario: 项目描述区域

- **WHEN** 项目有描述内容
- **THEN** 系统 SHALL 渲染 Markdown 描述
- **AND** 系统 SHALL 支持点击切换到编辑模式

#### Scenario: 状态切换

- **WHEN** 用户点击项目状态选择器
- **THEN** 系统 SHALL 展示状态选项下拉菜单
- **AND** 选择新状态后 SHALL 立即保存并更新界面

---

### Requirement: Project 关联 Issue 列表

项目详情页 SHALL 展示关联的 Issue 列表。

#### Scenario: Issue 列表展示

- **WHEN** 项目详情页加载完成
- **THEN** 页面 SHALL 展示该项目关联的所有 Issue 列表
- **AND** 每个 Issue SHALL 展示：标识符、标题、状态图标、优先级、负责人头像

#### Scenario: Issue 分组

- **WHEN** 用户选择按状态分组
- **THEN** 系统 SHALL 按 Issue 工作流状态分组展示

#### Scenario: Issue 排序

- **WHEN** 用户查看 Issue 列表
- **THEN** 系统 SHALL 默认按 `position` 排序
- **AND** 系统 SHALL 支持按优先级、创建时间排序

#### Scenario: 点击 Issue 跳转

- **WHEN** 用户点击某个 Issue
- **THEN** 系统 SHALL 导航到 Issue 详情页（或打开 Peek 预览面板）

---

### Requirement: Project 编辑功能

前端 SHALL 支持内联编辑项目信息。

#### Scenario: 编辑项目名称

- **WHEN** 用户点击项目名称旁的编辑图标
- **THEN** 系统 SHALL 切换名称为可编辑输入框
- **AND** 用户按 Enter 或失焦后 SHALL 保存

#### Scenario: 编辑项目描述

- **WHEN** 用户点击项目描述区域的编辑按钮
- **THEN** 系统 SHALL 切换为 Markdown 编辑器
- **AND** 编辑器 SHALL 支持预览切换
- **AND** 用户点击保存后 SHALL 提交更新

#### Scenario: 编辑日期

- **WHEN** 用户点击开始日期或目标日期
- **THEN** 系统 SHALL 展示日期选择器
- **AND** 选择日期后 SHALL 自动保存

---

### Requirement: Project 删除功能

前端 SHALL 支持删除项目。

#### Scenario: 删除确认

- **WHEN** 用户点击"删除项目"按钮
- **THEN** 系统 SHALL 弹出确认对话框
- **AND** 对话框 SHALL 提示"删除后关联 Issue 将保留，但不再显示项目关联"

#### Scenario: 删除成功反馈

- **WHEN** 项目删除成功
- **THEN** 系统 SHALL 导航回项目列表页
- **AND** 系统 SHALL 展示删除成功提示

---

### Requirement: 侧边栏 Project 入口

前端侧边栏 SHALL 提供项目快速入口。

#### Scenario: 侧边栏展示项目

- **WHEN** 用户查看侧边栏
- **THEN** 在团队节点下 SHALL 展示"Projects"入口
- **AND** 点击后 SHALL 导航到项目列表页

#### Scenario: 最近项目快捷入口

- **WHEN** 用户展开侧边栏的 Projects 节点
- **THEN** 系统 SHALL 展示最近访问的 3-5 个项目
- **AND** 点击项目名称 SHALL 直接导航到项目详情页

---

### Requirement: 响应式适配

Project 视图 SHALL 适配不同屏幕尺寸。

#### Scenario: 桌面端布局

- **WHEN** 屏幕宽度 >= 1024px
- **THEN** 项目列表 SHALL 使用多列卡片网格布局

#### Scenario: 移动端布局

- **WHEN** 屏幕宽度 < 640px
- **THEN** 项目列表 SHALL 使用单列列表布局
- **AND** 项目详情 SHALL 使用全屏模式
