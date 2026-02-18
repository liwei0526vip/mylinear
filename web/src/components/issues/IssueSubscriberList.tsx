/**
 * IssueSubscriberList 组件 - 订阅者列表
 */

import { useEffect } from 'react';
import { useIssueStore } from '@/stores/issueStore';
import type { IssueSubscriber } from '@/types/issue';

interface IssueSubscriberListProps {
  issueId: string;
  onSubscribe?: () => void;
  // 注：onUnsubscribe 待实现
}

export function IssueSubscriberList({
  issueId,
  onSubscribe,
}: IssueSubscriberListProps) {
  const { subscribers, fetchSubscribers, subscribe, isLoading } = useIssueStore();

  useEffect(() => {
    if (issueId) {
      fetchSubscribers(issueId);
    }
  }, [issueId, fetchSubscribers]);

  const handleSubscribe = async () => {
    await subscribe(issueId);
    fetchSubscribers(issueId);
    onSubscribe?.();
  };

  // 注：取消订阅功能待实现，unsubscribe 已在 store 中准备好

  return (
    <div className="rounded-lg border border-gray-200 p-3">
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-medium text-gray-700">
          订阅者 ({subscribers.length})
        </h4>
        <button
          onClick={handleSubscribe}
          disabled={isLoading}
          className="text-xs text-indigo-500 hover:text-indigo-600 disabled:opacity-50"
        >
          + 订阅
        </button>
      </div>

      {subscribers.length === 0 ? (
        <p className="mt-2 text-sm text-gray-500">暂无订阅者</p>
      ) : (
        <div className="mt-2 flex flex-wrap gap-2">
          {subscribers.map((subscriber) => (
            <SubscriberAvatar key={subscriber.id} subscriber={subscriber} />
          ))}
        </div>
      )}
    </div>
  );
}

/**
 * 订阅者头像组件
 */
interface SubscriberAvatarProps {
  subscriber: IssueSubscriber;
  size?: 'sm' | 'md';
}

function SubscriberAvatar({ subscriber, size = 'sm' }: SubscriberAvatarProps) {
  const sizeClass = size === 'sm' ? 'h-6 w-6' : 'h-8 w-8';

  return (
    <div
      className="group relative flex items-center gap-1.5 rounded-full bg-gray-100 px-2 py-1"
      title={subscriber.name}
    >
      {subscriber.avatar_url ? (
        <img
          src={subscriber.avatar_url}
          alt={subscriber.name}
          className={`${sizeClass} rounded-full object-cover`}
        />
      ) : (
        <div
          className={`${sizeClass} flex items-center justify-center rounded-full bg-indigo-100 text-xs font-medium text-indigo-600`}
        >
          {subscriber.name.charAt(0).toUpperCase()}
        </div>
      )}
      <span className="text-xs text-gray-600">{subscriber.name}</span>
    </div>
  );
}
