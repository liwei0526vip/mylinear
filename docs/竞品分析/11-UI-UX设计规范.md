## 十一、UI/UX 设计规范

### 11.1 设计令牌系统

#### 11.1.1 颜色系统

```css
:root {
  /* 主色调 */
  --color-primary-50: #EEF2FF;
  --color-primary-100: #E0E7FF;
  --color-primary-200: #C7D2FE;
  --color-primary-300: #A5B4FC;
  --color-primary-400: #818CF8;
  --color-primary-500: #6366F1;  /* 主色 */
  --color-primary-600: #4F46E5;
  --color-primary-700: #4338CA;
  --color-primary-800: #3730A3;
  --color-primary-900: #312E81;

  /* 灰色系 */
  --color-gray-50: #F9FAFB;
  --color-gray-100: #F3F4F6;
  --color-gray-200: #E5E7EB;
  --color-gray-300: #D1D5DB;
  --color-gray-400: #9CA3AF;
  --color-gray-500: #6B7280;
  --color-gray-600: #4B5563;
  --color-gray-700: #374151;
  --color-gray-800: #1F2937;
  --color-gray-900: #111827;

  /* 状态色 */
  --color-status-backlog: #6B7280;
  --color-status-todo: #9CA3AF;
  --color-status-in-progress: #F59E0B;
  --color-status-in-review: #10B981;
  --color-status-done: #3B82F6;
  --color-status-cancelled: #6B7280;
  --color-status-duplicate: #6B7280;

  /* 优先级色 */
  --color-priority-urgent: #EF4444;
  --color-priority-high: #F97316;
  --color-priority-medium: #EAB308;
  --color-priority-low: #6B7280;
  --color-priority-none: #9CA3AF;

  /* 语义色 */
  --color-success: #10B981;
  --color-warning: #F59E0B;
  --color-error: #EF4444;
  --color-info: #3B82F6;
}

/* 深色主题 */
[data-theme="dark"] {
  --color-gray-50: #111827;
  --color-gray-100: #1F2937;
  --color-gray-200: #374151;
  --color-gray-300: #4B5563;
  --color-gray-400: #6B7280;
  --color-gray-500: #9CA3AF;
  --color-gray-600: #D1D5DB;
  --color-gray-700: #E5E7EB;
  --color-gray-800: #F3F4F6;
  --color-gray-900: #F9FAFB;
}
```

#### 11.1.2 间距系统

```css
:root {
  /* 基础间距（4px 基数） */
  --space-0: 0;
  --space-1: 0.25rem;   /* 4px */
  --space-2: 0.5rem;    /* 8px */
  --space-3: 0.75rem;   /* 12px */
  --space-4: 1rem;      /* 16px */
  --space-5: 1.25rem;   /* 20px */
  --space-6: 1.5rem;    /* 24px */
  --space-8: 2rem;      /* 32px */
  --space-10: 2.5rem;   /* 40px */
  --space-12: 3rem;     /* 48px */
  --space-16: 4rem;     /* 64px */
}
```

#### 11.1.3 排版系统

```css
:root {
  /* 字体家族 */
  --font-sans: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', sans-serif;
  --font-mono: 'SF Mono', Monaco, 'Andale Mono', monospace;

  /* 字号 */
  --text-xs: 0.75rem;     /* 12px */
  --text-sm: 0.875rem;    /* 14px */
  --text-base: 1rem;      /* 16px */
  --text-lg: 1.125rem;    /* 18px */
  --text-xl: 1.25rem;     /* 20px */
  --text-2xl: 1.5rem;     /* 24px */
  --text-3xl: 1.875rem;   /* 30px */

  /* 行高 */
  --leading-tight: 1.25;
  --leading-normal: 1.5;
  --leading-relaxed: 1.75;

  /* 字重 */
  --font-normal: 400;
  --font-medium: 500;
  --font-semibold: 600;
  --font-bold: 700;
}
```

#### 11.1.4 阴影系统

```css
:root {
  --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
  --shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1);
  --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
  --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);
  --shadow-xl: 0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1);

  /* 面板阴影 */
  --shadow-panel: 0 0 0 1px rgb(0 0 0 / 0.05), 0 4px 6px -1px rgb(0 0 0 / 0.1);

  /* 弹窗阴影 */
  --shadow-modal: 0 0 0 1px rgb(0 0 0 / 0.1), 0 20px 25px -5px rgb(0 0 0 / 0.1);
}
```

### 11.2 状态图标规范

#### 11.2.1 SVG 图标定义

```svg
<!-- Backlog（虚线圆） -->
<svg viewBox="0 0 16 16" class="status-icon backlog">
  <circle cx="8" cy="8" r="7" fill="none" stroke="currentColor"
          stroke-width="1.5" stroke-dasharray="2 2"/>
</svg>

<!-- Todo（空心圆） -->
<svg viewBox="0 0 16 16" class="status-icon todo">
  <circle cx="8" cy="8" r="7" fill="none" stroke="currentColor"
          stroke-width="1.5"/>
</svg>

<!-- In Progress（半圆） -->
<svg viewBox="0 0 16 16" class="status-icon in-progress">
  <circle cx="8" cy="8" r="7" fill="none" stroke="currentColor"
          stroke-width="1.5"/>
  <path d="M8 1 A7 7 0 0 1 15 8 A7 7 0 0 1 8 15"
        fill="currentColor" opacity="0.5"/>
</svg>

<!-- In Review（加号圆） -->
<svg viewBox="0 0 16 16" class="status-icon in-review">
  <circle cx="8" cy="8" r="7" fill="none" stroke="currentColor"
          stroke-width="1.5"/>
  <line x1="8" y1="4" x2="8" y2="12" stroke="currentColor" stroke-width="1.5"/>
  <line x1="4" y1="8" x2="12" y2="8" stroke="currentColor" stroke-width="1.5"/>
</svg>

<!-- Done（勾选） -->
<svg viewBox="0 0 16 16" class="status-icon done">
  <circle cx="8" cy="8" r="7" fill="currentColor" opacity="0.2"/>
  <path d="M4 8 L7 11 L12 5" fill="none" stroke="currentColor"
        stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
</svg>

<!-- Cancelled（叉号） -->
<svg viewBox="0 0 16 16" class="status-icon cancelled">
  <circle cx="8" cy="8" r="7" fill="none" stroke="currentColor"
          stroke-width="1.5"/>
  <line x1="5" y1="5" x2="11" y2="11" stroke="currentColor" stroke-width="1.5"/>
  <line x1="11" y1="5" x2="5" y2="11" stroke="currentColor" stroke-width="1.5"/>
</svg>
```

#### 11.2.2 图标尺寸规范

| 用途 | 尺寸 | CSS 类 |
|------|------|--------|
| 列表项状态 | 16px | `.status-icon-sm` |
| 标准使用 | 20px | `.status-icon` |
| 大尺寸 | 24px | `.status-icon-lg` |

### 11.3 核心组件规格

#### 11.3.1 Button 组件

```tsx
// Button 变体
type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger';
type ButtonSize = 'sm' | 'md' | 'lg';

const buttonStyles = {
  base: `
    inline-flex items-center justify-center
    font-medium rounded-md
    transition-colors duration-150
    focus:outline-none focus:ring-2 focus:ring-offset-2
    disabled:opacity-50 disabled:cursor-not-allowed
  `,
  variants: {
    primary: 'bg-primary-600 text-white hover:bg-primary-700 focus:ring-primary-500',
    secondary: 'bg-gray-100 text-gray-900 hover:bg-gray-200 focus:ring-gray-500',
    ghost: 'bg-transparent text-gray-600 hover:bg-gray-100 focus:ring-gray-500',
    danger: 'bg-red-600 text-white hover:bg-red-700 focus:ring-red-500',
  },
  sizes: {
    sm: 'text-xs px-2.5 py-1.5',
    md: 'text-sm px-4 py-2',
    lg: 'text-base px-6 py-3',
  },
};
```

#### 11.3.2 Input 组件

```tsx
const inputStyles = `
  block w-full
  px-3 py-2
  text-sm text-gray-900
  bg-white border border-gray-300
  rounded-md
  placeholder:text-gray-400
  focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500
  disabled:bg-gray-100 disabled:cursor-not-allowed
`;
```

#### 11.3.3 Dropdown 组件

```tsx
// 下拉菜单动画
const dropdownAnimation = {
  enter: 'transition ease-out duration-100',
  enterFrom: 'transform opacity-0 scale-95',
  enterTo: 'transform opacity-100 scale-100',
  leave: 'transition ease-in duration-75',
  leaveFrom: 'transform opacity-100 scale-100',
  leaveTo: 'transform opacity-0 scale-95',
};
```

### 11.4 响应式断点

```css
/* 断点定义 */
:root {
  --breakpoint-sm: 640px;   /* 手机横屏 */
  --breakpoint-md: 768px;   /* 平板竖屏 */
  --breakpoint-lg: 1024px;  /* 平板横屏 */
  --breakpoint-xl: 1280px;  /* 桌面 */
  --breakpoint-2xl: 1536px; /* 大屏 */
}

/* 响应式布局策略 */
/*
  < 640px:  单栏布局，隐藏侧边栏
  640-1024px: 两栏布局（导航 + 内容）
  > 1024px: 三栏布局（导航 + 内容 + 详情面板）
*/
```

### 11.5 三栏布局实现

```tsx
// Linear 风格的三栏布局
const Layout = () => (
  <div className="flex h-screen bg-gray-50">
    {/* 左侧导航 */}
    <aside className="w-56 flex-shrink-0 border-r border-gray-200 bg-white">
      <Sidebar />
    </aside>

    {/* 中间内容区 */}
    <main className="flex-1 min-w-0 overflow-auto">
      <Content />
    </main>

    {/* 右侧详情面板（条件渲染） */}
    {showDetailPanel && (
      <aside className="w-96 flex-shrink-0 border-l border-gray-200 bg-white">
        <DetailPanel />
      </aside>
    )}
  </div>
);
```

### 11.6 动画时长标准

```css
:root {
  --duration-fast: 150ms;    /* 微交互：hover、focus */
  --duration-normal: 200ms;  /* 标准过渡：展开/收起 */
  --duration-slow: 300ms;    /* 大动作：弹窗、面板 */
  --duration-slower: 500ms;  /* 页面过渡 */

  --easing-default: cubic-bezier(0.4, 0, 0.2, 1);
  --easing-in: cubic-bezier(0.4, 0, 1, 1);
  --easing-out: cubic-bezier(0, 0, 0.2, 1);
  --easing-bounce: cubic-bezier(0.68, -0.55, 0.265, 1.55);
}
```

---

### 11.7 UI/UX 设计核心要点

#### 标志性交互设计

| 设计 | 核心理念 | 实现要点 |
|------|---------|---------|
| 三栏布局 | 左 56px 导航 + 中间自适应 + 右 384px 详情面板 | CSS Flexbox |
| 命令面板 | Cmd+K 全局操作入口 | 模糊搜索 + 分组命令 |
| Peek 预览 | 右侧浮动面板，不离开列表 | 高信息密度展示 |
| 状态图标 | 每种状态独特的 SVG 图标（虚线圆/半圆/勾选…） | 直觉化状态识别 |
| 键盘驱动 | Vim 风格 J/K 导航 + 单键操作（C/S/A/L/P） | 全键盘操作无需鼠标 |

#### 设计系统总结

| 维度 | 关键参数 |
|------|---------|
| 主色调 | Indigo (#6366F1) |
| 状态色 | 7 种（Backlog 灰/Todo 灰/In Progress 黄/In Review 绿/Done 蓝/Cancelled 灰/Duplicate 灰） |
| 优先级色 | Urgent 红/High 橙/Medium 黄/Low 灰/None 浅灰 |
| 间距基数 | 4px（4/8/12/16/20/24/32/40/48/64px） |
| 动画时长 | Fast 150ms (hover) / Normal 200ms (展开) / Slow 300ms (弹窗) |
| 字体 | 系统字体栈（-apple-system 等） |

#### 最佳实践

- **深色/浅色主题**：使用 CSS 变量 + `[data-theme]` 属性切换，一套代码支持双主题
- **响应式布局**：< 640px 单栏 / 640-1024px 双栏 / > 1024px 三栏
- **SVG 状态图标**应从第一个版本就规范化，这是 Linear 视觉识别的核心
- **4px 间距基数**保证整个应用的视觉节奏一致

---
