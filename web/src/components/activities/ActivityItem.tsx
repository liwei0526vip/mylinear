/**
 * ActivityItem 组件 - 单条活动
 */

import type { Activity, ActivityType } from '@/types/activity';
import { ACTIVITY_TYPE_CONFIG } from '@/types/activity';

interface ActivityItemProps {
  activity: Activity;
}

// 格式化相对时间
function formatRelativeTime(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSec = Math.floor(diffMs / 1000);
  const diffMin = Math.floor(diffSec / 60);
  const diffHour = Math.floor(diffMin / 60);
  const diffDay = Math.floor(diffHour / 24);

  if (diffSec < 60) return '刚刚';
  if (diffMin < 60) return `${diffMin} 分钟前`;
  if (diffHour < 24) return `${diffHour} 小时前`;
  if (diffDay < 7) return `${diffDay} 天前`;

  return date.toLocaleDateString('zh-CN', {
    month: 'short',
    day: 'numeric',
  });
}

// 获取活动图标
function getActivityIcon(type: ActivityType): React.ReactNode {
  const iconClass = 'w-4 h-4';

  switch (type) {
    case 'issue_created':
      return (
        <svg className={iconClass} fill="currentColor" viewBox="0 0 16 16">
          <path d="M8 4a.5.5 0 0 1 .5.5v3h3a.5.5 0 0 1 0 1h-3v3a.5.5 0 0 1-1 0v-3h-3a.5.5 0 0 1 0-1h3v-3A.5.5 0 0 1 8 4z" />
        </svg>
      );
    case 'title_changed':
    case 'description_changed':
      return (
        <svg className={iconClass} fill="currentColor" viewBox="0 0 16 16">
          <path d="M15.502 1.94a.5.5 0 0 1 0 .706L14.459 3.69l-2-2L13.502.646a.5.5 0 0 1 .707 0l1.293 1.293zm-1.75 2.456-2-2L4.939 9.21a.5.5 0 0 0-.121.196l-.805 2.414a.25.25 0 0 0 .316.316l2.414-.805a.5.5 0 0 0 .196-.12l6.813-6.814z" />
          <path fillRule="evenodd" d="M1 13.5A1.5 1.5 0 0 0 2.5 15h11a1.5 1.5 0 0 0 1.5-1.5v-6a.5.5 0 0 0-1 0v6a.5.5 0 0 1-.5.5h-11a.5.5 0 0 1-.5-.5v-11a.5.5 0 0 1 .5-.5H9a.5.5 0 0 0 0-1H2.5A1.5 1.5 0 0 0 1 2.5v11z" />
        </svg>
      );
    case 'status_changed':
      return (
        <svg className={iconClass} fill="currentColor" viewBox="0 0 16 16">
          <path fillRule="evenodd" d="M1 8a7 7 0 1 0 14 0A7 7 0 0 0 1 8zm15 0A8 8 0 1 1 0 8a8 8 0 0 1 16 0zM4.5 7.5a.5.5 0 0 0 0 1h5.793l-2.147 2.146a.5.5 0 0 0 .708.708l3-3a.5.5 0 0 0 0-.708l-3-3a.5.5 0 1 0-.708.708L10.293 7.5H4.5z" />
        </svg>
      );
    case 'priority_changed':
      return (
        <svg className={iconClass} fill="currentColor" viewBox="0 0 16 16">
          <path d="M5 3.5a.5.5 0 0 1 .5-.5h5a.5.5 0 0 1 0 1H6.707l5.147 5.146a.5.5 0 0 1-.708.708L6 4.707V10.5a.5.5 0 0 1-1 0v-7z" />
        </svg>
      );
    case 'assignee_changed':
      return (
        <svg className={iconClass} fill="currentColor" viewBox="0 0 16 16">
          <path d="M3 14s-1 0-1-1 1-4 6-4 6 3 6 4-1 1-1 1H3zm5-6a3 3 0 1 0 0-6 3 3 0 0 0 0 6z" />
        </svg>
      );
    case 'due_date_changed':
      return (
        <svg className={iconClass} fill="currentColor" viewBox="0 0 16 16">
          <path d="M3.5 0a.5.5 0 0 1 .5.5V1h8V.5a.5.5 0 0 1 1 0V1h1a2 2 0 0 1 2 2v11a2 2 0 0 1-2 2H2a2 2 0 0 1-2-2V3a2 2 0 0 1 2-2h1V.5a.5.5 0 0 1 .5-.5zM1 4v10a1 1 0 0 0 1 1h12a1 1 0 0 0 1-1V4H1z" />
        </svg>
      );
    case 'project_changed':
      return (
        <svg className={iconClass} fill="currentColor" viewBox="0 0 16 16">
          <path d="M9.293 0H4a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V4.707A1 1 0 0 0 13.707 4L10 .293A1 1 0 0 0 9.293 0zM9.5 3.5v-2l3 3h-2a1 1 0 0 1-1-1z" />
        </svg>
      );
    case 'labels_changed':
      return (
        <svg className={iconClass} fill="currentColor" viewBox="0 0 16 16">
          <path d="M2 2a2 2 0 0 1 2-2h5.293A1 1 0 0 1 11 .293L13.707 3a1 1 0 0 1 .293.707V14a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V2z" />
        </svg>
      );
    case 'comment_added':
      return (
        <svg className={iconClass} fill="currentColor" viewBox="0 0 16 16">
          <path d="M2.678 11.894a1 1 0 0 1 .287.801 10.97 10.97 0 0 1-.398 2c1.395-.323 2.247-.697 2.634-.893a1 1 0 0 1 .71-.074A8.06 8.06 0 0 0 8 14c3.996 0 7-2.807 7-6 0-3.192-3.004-6-7-6S1 4.808 1 8c0 1.468.617 2.83 1.678 3.894zm-.493 3.905a21.682 21.682 0 0 1-.713.129c-.2.032-.352-.176-.273-.362a9.68 9.68 0 0 0 .244-.637l.003-.01c.248-.72.45-1.548.524-2.319C.743 11.37 0 9.76 0 8c0-3.866 3.582-7 8-7s8 3.134 8 7-3.582 7-8 7a9.06 9.06 0 0 1-2.347-.306c-.52.263-1.639.742-3.468 1.105z" />
        </svg>
      );
    default:
      return (
        <svg className={iconClass} fill="currentColor" viewBox="0 0 16 16">
          <circle cx="8" cy="8" r="4" />
        </svg>
      );
  }
}

// 获取活动颜色
function getActivityColor(type: ActivityType): string {
  switch (type) {
    case 'issue_created':
      return 'text-green-500 bg-green-50';
    case 'status_changed':
      return 'text-blue-500 bg-blue-50';
    case 'priority_changed':
      return 'text-orange-500 bg-orange-50';
    case 'assignee_changed':
      return 'text-purple-500 bg-purple-50';
    case 'comment_added':
      return 'text-gray-500 bg-gray-50';
    default:
      return 'text-indigo-500 bg-indigo-50';
  }
}

// 渲染活动描述
function renderActivityDescription(activity: Activity): React.ReactNode {
  const actorName = activity.actor?.name || '用户';
  const payload = activity.payload as Record<string, unknown> | undefined;

  switch (activity.type) {
    case 'issue_created':
      return (
        <span>
          <strong>{actorName}</strong> 创建了这个 Issue
        </span>
      );
    case 'title_changed':
      return (
        <span>
          <strong>{actorName}</strong> 修改了标题
          {payload && (
            <span className="ml-2 text-gray-500">
              "{payload.old_value as string}" → "{payload.new_value as string}"
            </span>
          )}
        </span>
      );
    case 'description_changed':
      return (
        <span>
          <strong>{actorName}</strong> 修改了描述
        </span>
      );
    case 'status_changed':
      return (
        <span>
          <strong>{actorName}</strong> 修改了状态
          {payload && (payload.new_status as Record<string, unknown>) && (
            <span className="ml-2 text-gray-500">
              → {(payload.new_status as Record<string, unknown>).name as string}
            </span>
          )}
        </span>
      );
    case 'priority_changed':
      return (
        <span>
          <strong>{actorName}</strong> 修改了优先级
        </span>
      );
    case 'assignee_changed':
      return (
        <span>
          <strong>{actorName}</strong> 修改了负责人
        </span>
      );
    case 'due_date_changed':
      return (
        <span>
          <strong>{actorName}</strong> 修改了截止日期
        </span>
      );
    case 'project_changed':
      return (
        <span>
          <strong>{actorName}</strong> 修改了项目
        </span>
      );
    case 'labels_changed':
      return (
        <span>
          <strong>{actorName}</strong> 修改了标签
        </span>
      );
    case 'comment_added':
      return (
        <span>
          <strong>{actorName}</strong> 添加了评论
        </span>
      );
    default:
      return (
        <span>
          <strong>{actorName}</strong> 执行了操作
        </span>
      );
  }
}

export function ActivityItem({ activity }: ActivityItemProps) {
  const config = ACTIVITY_TYPE_CONFIG[activity.type] || { label: '未知操作', icon: 'Circle' };
  const iconColorClass = getActivityColor(activity.type);

  return (
    <div className="flex items-start gap-3 py-3">
      {/* 时间线圆点和图标 */}
      <div className="relative flex flex-col items-center">
        <div className={`flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full ${iconColorClass}`}>
          {getActivityIcon(activity.type)}
        </div>
        {/* 连接线 */}
        <div className="absolute top-8 h-full w-px bg-gray-100" />
      </div>

      {/* 内容区 */}
      <div className="flex-1 min-w-0 pb-3">
        {/* 活动描述 */}
        <div className="text-sm text-gray-700">
          {renderActivityDescription(activity)}
        </div>

        {/* 时间 */}
        <div className="mt-1 text-xs text-gray-400">
          {formatRelativeTime(activity.created_at)}
        </div>
      </div>
    </div>
  );
}
