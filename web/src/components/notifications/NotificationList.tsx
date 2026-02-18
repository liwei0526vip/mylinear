/**
 * 通知列表组件
 */

import { useEffect, useCallback } from 'react';
import { Inbox } from 'lucide-react';
import { useNotificationStore } from '../../stores/notificationStore';
import { NotificationItem } from './NotificationItem';
import { NotificationItemSkeleton } from './NotificationItemSkeleton';
import { Button } from '../ui/button';

interface NotificationListProps {
  showMarkAllRead?: boolean;
  emptyMessage?: string;
  compact?: boolean;
  filterRead?: boolean;
}

export function NotificationList({
  showMarkAllRead = true,
  emptyMessage = '暂无通知',
  compact = false,
  filterRead,
}: NotificationListProps) {
  const {
    notifications,
    unreadCount,
    total,
    page,
    isLoading,
    fetchNotifications,
    fetchUnreadCount,
    markAsRead,
    markAllAsRead,
  } = useNotificationStore();

  useEffect(() => {
    fetchNotifications(1, 20, filterRead);
    fetchUnreadCount();
  }, [fetchNotifications, fetchUnreadCount, filterRead]);

  const handleMarkAsRead = useCallback(
    async (id: string) => {
      await markAsRead(id);
    },
    [markAsRead]
  );

  const handleMarkAllAsRead = useCallback(async () => {
    await markAllAsRead();
    fetchNotifications(page, 20, filterRead);
  }, [markAllAsRead, fetchNotifications, page, filterRead]);

  const handleLoadMore = useCallback(() => {
    fetchNotifications(page + 1, 20, filterRead);
  }, [fetchNotifications, page, filterRead]);

  const hasMore = notifications.length < total;

  if (isLoading && notifications.length === 0) {
    return (
      <div className="space-y-2">
        {Array.from({ length: 5 }).map((_, i) => (
          <NotificationItemSkeleton key={i} compact={compact} />
        ))}
      </div>
    );
  }

  if (notifications.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-gray-500">
        <Inbox className="w-12 h-12 mb-4 text-gray-300" />
        <p className="text-sm">{emptyMessage}</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      {/* 工具栏 */}
      {showMarkAllRead && unreadCount > 0 && (
        <div className="flex items-center justify-between px-4 py-2 border-b border-gray-100 bg-gray-50">
          <span className="text-sm text-gray-600">
            {unreadCount} 条未读
          </span>
          <Button
            variant="ghost"
            size="sm"
            onClick={handleMarkAllAsRead}
            className="text-xs"
          >
            全部标记已读
          </Button>
        </div>
      )}

      {/* 通知列表 */}
      <div className="flex-1 overflow-y-auto">
        {notifications.map((notification) => (
          <NotificationItem
            key={notification.id}
            notification={notification}
            onMarkAsRead={handleMarkAsRead}
            compact={compact}
          />
        ))}
      </div>

      {/* 加载更多 */}
      {hasMore && (
        <div className="p-4 text-center border-t border-gray-100">
          <Button
            variant="outline"
            size="sm"
            onClick={handleLoadMore}
            disabled={isLoading}
          >
            {isLoading ? '加载中...' : '加载更多'}
          </Button>
        </div>
      )}
    </div>
  );
}
