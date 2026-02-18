/**
 * MainLayout - 主布局组件
 * 遵循 Linear 设计规范：左 224px 导航 + 中间自适应内容
 */

import type { ReactNode } from 'react';
import { Sidebar } from './Sidebar';

interface MainLayoutProps {
  children: ReactNode;
  /** 是否显示侧边栏 */
  showSidebar?: boolean;
}

export function MainLayout({ children, showSidebar = true }: MainLayoutProps) {
  if (!showSidebar) {
    return <>{children}</>;
  }

  return (
    <div className="flex h-screen bg-gray-50 dark:bg-gray-950">
      {/* 左侧导航 */}
      <Sidebar />

      {/* 中间内容区 */}
      <main className="flex-1 min-w-0 overflow-auto">
        {children}
      </main>
    </div>
  );
}
