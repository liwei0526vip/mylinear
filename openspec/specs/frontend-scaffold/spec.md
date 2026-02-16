## ADDED Requirements

### Requirement: React 前端项目结构

前端 MUST 使用 Vite + React + TypeScript 初始化，代码组织在 `web/` 目录下。

目录结构：
```
web/
├── src/
│   ├── components/          # UI 组件
│   ├── pages/               # 页面组件
│   ├── stores/              # Zustand 状态管理
│   ├── api/                 # API 调用层
│   ├── lib/                 # 工具函数
│   ├── App.tsx              # 根组件
│   ├── main.tsx             # 入口
│   └── index.css            # 全局样式
├── public/
├── index.html
├── package.json
├── tsconfig.json
└── vite.config.ts
```

#### Scenario: 项目结构完整性
- **WHEN** 开发者查看 `web/` 目录
- **THEN** MUST 包含 `src/`（源码）目录，以及 `components/`、`pages/`、`stores/`、`api/`、`lib/` 子目录

#### Scenario: TypeScript 编译检查
- **WHEN** 开发者在 `web/` 目录执行 `npm run build`
- **THEN** TypeScript 编译 SHALL 成功通过，无类型错误

---

### Requirement: Vite 开发服务器配置

前端开发服务器 MUST 正确配置 API 代理，将 `/api/*` 请求代理至后端。

#### Scenario: 前端开发服务器启动
- **WHEN** 用户执行 `npm run dev`
- **THEN** Vite 开发服务器 SHALL 在配置端口（默认 5173）上启动，并支持 HMR（热模块替换）

#### Scenario: API 请求代理
- **WHEN** 前端代码发起 `/api/v1/health` 请求
- **THEN** Vite 开发服务器 SHALL 将请求代理至后端服务（默认 `http://localhost:8080`）

---

### Requirement: shadcn/ui 组件库集成

前端 MUST 集成 shadcn/ui（基于 Radix UI）组件库，提供基础 UI 组件。

#### Scenario: 组件库配置完成
- **WHEN** 开发者查看项目配置
- **THEN** `components.json` 配置文件 SHALL 存在，包含 shadcn/ui 的基本配置（样式、别名、CSS 变量等）

#### Scenario: 基础组件可用
- **WHEN** 开发者需要使用 shadcn/ui 组件（如 Button）
- **THEN** SHALL 能通过 `npx shadcn@latest add button` 命令添加组件并正常使用

---

### Requirement: Zustand 状态管理

前端 MUST 集成 Zustand 作为全局状态管理方案。

#### Scenario: Zustand 可用
- **WHEN** 开发者在 `web/src/stores/` 目录创建 store
- **THEN** SHALL 能正常导入 `zustand` 并创建状态 store

#### Scenario: 示例 Store 结构
- **WHEN** 开发者查看 stores 目录
- **THEN** SHOULD 包含一个示例 store 文件，展示标准的 store 定义模式

---

### Requirement: API 调用层基础

前端 MUST 提供统一的 API 调用层，封装 HTTP 请求逻辑。

#### Scenario: API 客户端初始化
- **WHEN** 开发者需要调用后端 API
- **THEN** SHALL 能使用 `web/src/api/` 目录中的 API 客户端发起请求

#### Scenario: 基础 URL 配置
- **WHEN** API 客户端发起请求
- **THEN** 请求 SHALL 自动添加 `/api/v1` 前缀

#### Scenario: 错误响应处理
- **WHEN** 后端返回非 2xx 状态码
- **THEN** API 客户端 SHALL 抛出包含状态码和错误信息的异常

---

### Requirement: 前端首页占位

前端 MUST 提供一个可访问的首页，确认前端正常运行。

#### Scenario: 首页正常渲染
- **WHEN** 用户在浏览器访问前端根路径 `/`
- **THEN** 页面 SHALL 显示 MyLinear 项目名称和基本信息，确认前端应用正常运行

#### Scenario: 后端连通性验证
- **WHEN** 首页加载完成
- **THEN** 页面 SHOULD 调用健康检查 API 并显示后端连接状态（已连接/未连接）
