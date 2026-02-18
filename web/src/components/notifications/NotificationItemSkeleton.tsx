/**
 * 通知项骨架屏
 */

import { cn } from '../../lib/utils';

interface NotificationItemSkeletonProps {
  compact?: boolean;
}

export function NotificationItemSkeleton({ compact = false }: NotificationItemSkeletonProps) {
  if (compact) {
    return (
      <div className="flex items-start gap-2 p-2 animate-pulse">
        <div className="w-4 h-4 bg-gray-200 rounded flex-shrink-0" />
        <div className="flex-1 space-y-1">
          <div className="h-4 bg-gray-200 rounded w-3/4" />
          <div className="h-3 bg-gray-100 rounded w-1/2" />
        </div>
      </div>
    );
  }

  return (
    <div className="flex items-start gap-3 p-4 border-b border-gray-100 animate-pulse">
      <div className="w-5 h-5 bg-gray-200 rounded flex-shrink-0" />
      <div className="flex-1 space-y-2">
        <div className="flex items-center gap-2">
          <div className="h-4 bg-gray-200 rounded w-16" />
          <div className="h-4 bg-gray-100 rounded w-8" />
        </div>
        <div className="h-4 bg-gray-200 rounded w-3/4" />
        <div className="h-3 bg-gray-100 rounded w-1/2" />
      </div>
    </div>
  );
}
