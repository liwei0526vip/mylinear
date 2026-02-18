/**
 * 单条通知组件
 */

import { useNavigate } from 'react-router-dom';
import { Bell, CheckCircle, AtSign, AlertCircle, MessageSquare, ArrowRight } from 'lucide-react';
import { cn } from '../../lib/utils';
import type { Notification, NotificationType } from '../../types/notification';
import { getNotificationTypeLabel } from '../../types/notification';
import { formatDistanceToNow } from '../../lib/date';

interface NotificationItemProps {
  notification: Notification;
  onMarkAsRead?: (id: string) => void;
  compact?: boolean;
}

// 获取通知图标
function getNotificationIcon(type: NotificationType) {
  switch (type) {
    case 'issue_assigned':
      return <CheckCircle className="w-4 h-4 text-indigo-500" />;
    case 'issue_mentioned':
      return <AtSign className="w-4 h-4 text-blue-500" />;
    case 'issue_status_changed':
      return <ArrowRight className="w-4 h-4 text-yellow-500" />;
    case 'issue_priority_changed':
      return <AlertCircle className="w-4 h-4 text-orange-500" />;
    case 'issue_commented':
      return <MessageSquare className="w-4 h-4 text-green-500" />;
    default:
      return <Bell className="w-4 h-4 text-gray-500" />;
  }
}

export function NotificationItem({ notification, onMarkAsRead, compact = false }: NotificationItemProps) {
  const navigate = useNavigate();
  const isUnread = !notification.read_at;

  const handleClick = () => {
    // 标记为已读
    if (isUnread && onMarkAsRead) {
      onMarkAsRead(notification.id);
    }

    // 跳转到资源详情页
    if (notification.resource_type === 'issue' && notification.resource_id) {
      navigate(`/issues/${notification.resource_id}`);
    }
  };

  const timeAgo = formatDistanceToNow(new Date(notification.created_at));

  if (compact) {
    return (
      <div
        onClick={handleClick}
        className={cn(
          'flex items-start gap-2 p-2 cursor-pointer hover:bg-gray-50 rounded-md transition-colors',
          isUnread && 'bg-indigo-50/50'
        )}
      >
        <div className="flex-shrink-0 mt-0.5">
          {getNotificationIcon(notification.type)}
        </div>
        <div className="flex-1 min-w-0">
          <p className={cn('text-sm truncate', isUnread && 'font-medium')}>
            {notification.title}
          </p>
          <p className="text-xs text-gray-500">{timeAgo}</p>
        </div>
        {isUnread && (
          <div className="w-2 h-2 bg-indigo-500 rounded-full flex-shrink-0 mt-1.5" />
        )}
      </div>
    );
  }

  return (
    <div
      onClick={handleClick}
      className={cn(
        'flex items-start gap-3 p-4 cursor-pointer hover:bg-gray-50 transition-colors border-b border-gray-100',
        isUnread && 'bg-indigo-50/30'
      )}
    >
      <div className="flex-shrink-0 mt-1">
        {getNotificationIcon(notification.type)}
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-1">
          <span className="text-xs text-gray-500 bg-gray-100 px-1.5 py-0.5 rounded">
            {getNotificationTypeLabel(notification.type)}
          </span>
          {isUnread && (
            <span className="text-xs text-indigo-600 font-medium">未读</span>
          )}
        </div>
        <p className={cn('text-sm mb-1', isUnread && 'font-medium')}>
          {notification.title}
        </p>
        {notification.body && (
          <p className="text-sm text-gray-600 line-clamp-2">{notification.body}</p>
        )}
        <p className="text-xs text-gray-400 mt-2">{timeAgo}</p>
      </div>
      {isUnread && (
        <div className="w-2.5 h-2.5 bg-indigo-500 rounded-full flex-shrink-0 mt-2" />
      )}
    </div>
  );
}
