/**
 * Sidebar - 侧边栏组件
 * 遵循 Linear 设计规范：左 56px 导航
 */

import { useState, useEffect } from 'react';
import { Link, useLocation, useParams } from 'react-router-dom';
import { useAuthStore } from '@/stores/authStore';
import { useTeamStore } from '@/stores/teamStore';
import { useProjectStore } from '@/stores/projectStore';
import { useNotificationStore } from '@/stores/notificationStore';

// 图标组件
const ProjectsIcon = () => (
  <svg viewBox="0 0 16 16" className="h-4 w-4" fill="none" stroke="currentColor">
    <rect x="2" y="2" width="5" height="5" rx="1" strokeWidth="1.5" />
    <rect x="9" y="2" width="5" height="5" rx="1" strokeWidth="1.5" />
    <rect x="2" y="9" width="5" height="5" rx="1" strokeWidth="1.5" />
    <rect x="9" y="9" width="5" height="5" rx="1" strokeWidth="1.5" />
  </svg>
);

const InboxIcon = () => (
  <svg viewBox="0 0 16 16" className="h-4 w-4" fill="none" stroke="currentColor">
    <path d="M2 4h12v8a2 2 0 01-2 2H4a2 2 0 01-2-2V4z" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
    <path d="M2 4l3 3h6l3-3" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
);

const ChevronIcon = ({ expanded }: { expanded: boolean }) => (
  <svg
    viewBox="0 0 16 16"
    className={`h-3 w-3 transition-transform duration-150 ${expanded ? 'rotate-90' : ''}`}
    fill="currentColor"
  >
    <path d="M6 4l4 4-4 4" stroke="currentColor" strokeWidth="1.5" fill="none" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
);

const SettingsIcon = () => (
  <svg viewBox="0 0 16 16" className="h-4 w-4" fill="none" stroke="currentColor">
    <circle cx="8" cy="8" r="2" strokeWidth="1.5" />
    <path
      d="M8 1v2M8 13v2M14 8h-2M4 8H2M12.95 3.05l-1.414 1.414M4.464 12.536l-1.414 1.414M12.95 12.95l-1.414-1.414M4.464 3.464L3.05 2.05"
      strokeWidth="1.5"
      strokeLinecap="round"
    />
  </svg>
);

const IssuesIcon = () => (
  <svg viewBox="0 0 16 16" className="h-4 w-4" fill="none" stroke="currentColor">
    <circle cx="8" cy="8" r="6" strokeWidth="1.5" />
    <circle cx="8" cy="8" r="2" fill="currentColor" />
  </svg>
);

export function Sidebar() {
  const location = useLocation();
  const { teamId } = useParams<{ teamId: string }>();
  const { user } = useAuthStore();
  const { teams, fetchTeams } = useTeamStore();
  const { projects, fetchTeamProjects } = useProjectStore();
  const { unreadCount, fetchUnreadCount } = useNotificationStore();

  const [projectsExpanded, setProjectsExpanded] = useState(true);

  // 加载团队列表
  useEffect(() => {
    if (user?.workspace_id) {
      fetchTeams(user.workspace_id);
    }
  }, [user?.workspace_id, fetchTeams]);

  // 加载未读通知数量
  useEffect(() => {
    fetchUnreadCount();
    // 每 60 秒刷新一次
    const interval = setInterval(fetchUnreadCount, 60000);
    return () => clearInterval(interval);
  }, [fetchUnreadCount]);

  // 加载当前团队的项目（用于最近项目）
  useEffect(() => {
    if (teamId) {
      fetchTeamProjects(teamId);
    }
  }, [teamId, fetchTeamProjects]);

  // 获取最近 5 个项目（按更新时间排序）
  const recentProjects = [...projects]
    .sort((a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime())
    .slice(0, 5);

  // 当前选中的团队
  const currentTeam = teams.find((t) => t.id === teamId);

  // 检查路由是否激活
  const isProjectsActive = location.pathname.includes('/projects');
  const isSettingsActive = location.pathname.startsWith('/settings');
  const isIssuesActive = location.pathname.includes('/issues') && !location.pathname.includes('/projects');
  const isInboxActive = location.pathname === '/inbox';

  // 获取项目列表的路由
  const getProjectsRoute = () => {
    if (teamId) {
      return `/teams/${teamId}/projects`;
    }
    // 如果没有选中团队，使用第一个团队
    if (teams.length > 0) {
      return `/teams/${teams[0].id}/projects`;
    }
    return '/';
  };

  return (
    <aside className="flex h-screen w-56 flex-col border-r border-gray-200 bg-white dark:border-gray-800 dark:bg-gray-900">
      {/* Logo 区域 */}
      <div className="flex h-14 items-center border-b border-gray-200 px-4 dark:border-gray-800">
        <Link to="/" className="flex items-center gap-2">
          <div className="flex h-7 w-7 items-center justify-center rounded bg-indigo-500 text-white text-sm font-bold">
            M
          </div>
          <span className="text-sm font-semibold text-gray-900 dark:text-white">MyLinear</span>
        </Link>
      </div>

      {/* 团队信息 */}
      {currentTeam && (
        <div className="border-b border-gray-200 px-4 py-3 dark:border-gray-800">
          <div className="flex items-center gap-2">
            <div className="flex h-6 w-6 items-center justify-center rounded bg-gray-100 text-xs font-medium text-gray-600 dark:bg-gray-800 dark:text-gray-400">
              {currentTeam.key?.slice(0, 2) || currentTeam.name.slice(0, 2)}
            </div>
            <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
              {currentTeam.name}
            </span>
          </div>
        </div>
      )}

      {/* 导航菜单 */}
      <nav className="flex-1 overflow-y-auto py-2">
        {/* Issues 入口 */}
        <div className="px-2">
          <Link
            to={teamId ? `/teams/${teamId}/issues` : '/'}
            className={`flex items-center gap-2 rounded-md px-3 py-1.5 text-sm transition-colors duration-150 ${
              isIssuesActive
                ? 'bg-indigo-50 text-indigo-600 dark:bg-indigo-900/20 dark:text-indigo-400'
                : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800'
            }`}
          >
            <IssuesIcon />
            <span>Issues</span>
          </Link>
        </div>

        {/* Inbox 入口 */}
        <div className="px-2 mt-1">
          <Link
            to="/inbox"
            className={`flex items-center gap-2 rounded-md px-3 py-1.5 text-sm transition-colors duration-150 ${
              isInboxActive
                ? 'bg-indigo-50 text-indigo-600 dark:bg-indigo-900/20 dark:text-indigo-400'
                : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800'
            }`}
          >
            <InboxIcon />
            <span>收件箱</span>
            {unreadCount > 0 && (
              <span className="ml-auto flex h-5 min-w-[20px] items-center justify-center rounded-full bg-red-500 px-1.5 text-xs font-medium text-white">
                {unreadCount > 99 ? '99+' : unreadCount}
              </span>
            )}
          </Link>
        </div>

        {/* Projects 入口 */}
        <div className="mt-1 px-2">
          {/* Projects 主入口 */}
          <Link
            to={getProjectsRoute()}
            className={`flex items-center gap-2 rounded-md px-3 py-1.5 text-sm transition-colors duration-150 ${
              isProjectsActive
                ? 'bg-indigo-50 text-indigo-600 dark:bg-indigo-900/20 dark:text-indigo-400'
                : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800'
            }`}
          >
            <ProjectsIcon />
            <span>项目</span>
          </Link>

          {/* 最近项目快捷入口 */}
          {recentProjects.length > 0 && (
            <div className="mt-1">
              {/* 展开/收起按钮 */}
              <button
                onClick={() => setProjectsExpanded(!projectsExpanded)}
                className="flex w-full items-center gap-1.5 rounded-md px-3 py-1 text-xs text-gray-500 hover:bg-gray-100 dark:text-gray-500 dark:hover:bg-gray-800"
              >
                <ChevronIcon expanded={projectsExpanded} />
                <span>最近项目</span>
              </button>

              {/* 项目列表 */}
              {projectsExpanded && (
                <div className="ml-4 mt-1 space-y-0.5">
                  {recentProjects.map((project) => (
                    <Link
                      key={project.id}
                      to={`/projects/${project.id}`}
                      className={`flex items-center gap-2 rounded-md px-2 py-1 text-xs transition-colors duration-150 ${
                        location.pathname === `/projects/${project.id}`
                          ? 'bg-indigo-50 text-indigo-600 dark:bg-indigo-900/20 dark:text-indigo-400'
                          : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800'
                      }`}
                    >
                      {/* 项目状态指示器 */}
                      <span
                        className={`h-1.5 w-1.5 rounded-full ${
                          project.status === 'completed'
                            ? 'bg-green-500'
                            : project.status === 'in_progress'
                            ? 'bg-blue-500'
                            : project.status === 'paused'
                            ? 'bg-yellow-500'
                            : project.status === 'cancelled'
                            ? 'bg-red-500'
                            : 'bg-gray-400'
                        }`}
                      />
                      <span className="truncate">{project.name}</span>
                    </Link>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      </nav>

      {/* 底部区域 */}
      <div className="border-t border-gray-200 p-2 dark:border-gray-800">
        {/* 设置入口 */}
        <Link
          to="/settings/workspace"
          className={`flex items-center gap-2 rounded-md px-3 py-1.5 text-sm transition-colors duration-150 ${
            isSettingsActive
              ? 'bg-indigo-50 text-indigo-600 dark:bg-indigo-900/20 dark:text-indigo-400'
              : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800'
          }`}
        >
          <SettingsIcon />
          <span>设置</span>
        </Link>

        {/* 用户信息 */}
        {user && (
          <div className="mt-2 flex items-center gap-2 rounded-md px-3 py-2">
            <div className="flex h-7 w-7 items-center justify-center rounded-full bg-indigo-100 text-xs font-medium text-indigo-600 dark:bg-indigo-900/30 dark:text-indigo-400">
              {user.name?.slice(0, 1).toUpperCase() || user.email.slice(0, 1).toUpperCase()}
            </div>
            <div className="flex-1 overflow-hidden">
              <p className="truncate text-xs font-medium text-gray-900 dark:text-white">
                {user.name || user.email}
              </p>
              <p className="truncate text-xs text-gray-500 dark:text-gray-400">
                {user.email}
              </p>
            </div>
          </div>
        )}
      </div>
    </aside>
  );
}
