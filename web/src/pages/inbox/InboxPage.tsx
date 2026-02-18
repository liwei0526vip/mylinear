/**
 * 通知收件箱页面
 */

import { useState, useEffect } from 'react';
import { Inbox, CheckCheck, Filter } from 'lucide-react';
import { useNotificationStore } from '../../stores/notificationStore';
import { NotificationList } from '../../components/notifications/NotificationList';
import { Button } from '../../components/ui/button';

type FilterType = 'all' | 'unread' | 'read';

export function InboxPage() {
  const [filter, setFilter] = useState<FilterType>('all');
  const { unreadCount, fetchUnreadCount, markAllAsRead, fetchNotifications } = useNotificationStore();

  useEffect(() => {
    fetchUnreadCount();
  }, [fetchUnreadCount]);

  const handleMarkAllAsRead = async () => {
    await markAllAsRead();
    fetchNotifications(1, 20, filter === 'unread' ? false : undefined);
  };

  const getFilterRead = (): boolean | undefined => {
    switch (filter) {
      case 'unread':
        return false;
      case 'read':
        return true;
      default:
        return undefined;
    }
  };

  return (
    <div className="flex flex-col h-full bg-white">
      {/* 页面标题 */}
      <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
        <div className="flex items-center gap-3">
          <Inbox className="w-6 h-6 text-gray-600" />
          <h1 className="text-xl font-semibold text-gray-900">收件箱</h1>
          {unreadCount > 0 && (
            <span className="px-2 py-0.5 text-xs font-medium text-indigo-600 bg-indigo-50 rounded-full">
              {unreadCount} 条未读
            </span>
          )}
        </div>
        {unreadCount > 0 && (
          <Button
            variant="outline"
            size="sm"
            onClick={handleMarkAllAsRead}
            className="text-sm"
          >
            <CheckCheck className="w-4 h-4 mr-1" />
            全部标记已读
          </Button>
        )}
      </div>

      {/* 过滤器 */}
      <div className="flex items-center gap-2 px-6 py-3 border-b border-gray-100 bg-gray-50">
        <Filter className="w-4 h-4 text-gray-400" />
        <div className="flex gap-1">
          {[
            { value: 'all', label: '全部' },
            { value: 'unread', label: '未读' },
            { value: 'read', label: '已读' },
          ].map((option) => (
            <button
              key={option.value}
              onClick={() => setFilter(option.value as FilterType)}
              className={`px-3 py-1 text-sm rounded-md transition-colors ${
                filter === option.value
                  ? 'bg-white text-gray-900 shadow-sm'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              {option.label}
            </button>
          ))}
        </div>
      </div>

      {/* 通知列表 */}
      <div className="flex-1 overflow-hidden">
        <NotificationList
          showMarkAllRead={false}
          emptyMessage={
            filter === 'unread'
              ? '没有未读通知'
              : filter === 'read'
              ? '没有已读通知'
              : '暂无通知'
          }
          filterRead={getFilterRead()}
        />
      </div>
    </div>
  );
}
