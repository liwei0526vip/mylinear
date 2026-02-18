/**
 * 通知徽章组件
 */

import { useEffect, useRef } from 'react';
import { Bell } from 'lucide-react';
import { useNotificationStore } from '../../stores/notificationStore';
import { cn } from '../../lib/utils';

interface NotificationBadgeProps {
  className?: string;
  iconClassName?: string;
  badgeClassName?: string;
  maxCount?: number;
  showZero?: boolean;
}

export function NotificationBadge({
  className,
  iconClassName,
  badgeClassName,
  maxCount = 99,
  showZero = false,
}: NotificationBadgeProps) {
  const { unreadCount, fetchUnreadCount } = useNotificationStore();
  const intervalRef = useRef<NodeJS.Timeout | null>(null);

  // 初始化获取未读数量
  useEffect(() => {
    fetchUnreadCount();

    // 定时刷新未读数量（每 30 秒）
    intervalRef.current = setInterval(() => {
      fetchUnreadCount();
    }, 30000);

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [fetchUnreadCount]);

  const displayCount = unreadCount > maxCount ? `${maxCount}+` : unreadCount;

  if (!showZero && unreadCount === 0) {
    return (
      <div className={cn('relative', className)}>
        <Bell className={cn('w-5 h-5 text-gray-500', iconClassName)} />
      </div>
    );
  }

  return (
    <div className={cn('relative inline-flex', className)}>
      <Bell className={cn('w-5 h-5 text-gray-500', iconClassName)} />
      {unreadCount > 0 && (
        <span
          className={cn(
            'absolute -top-1 -right-1 min-w-[18px] h-[18px] px-1',
            'flex items-center justify-center',
            'text-xs font-medium text-white bg-red-500 rounded-full',
            badgeClassName
          )}
        >
          {displayCount}
        </span>
      )}
    </div>
  );
}
