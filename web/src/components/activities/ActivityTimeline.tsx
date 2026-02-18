/**
 * ActivityTimeline 组件 - 活动时间线容器
 */

import { useEffect } from 'react';
import { useActivityStore } from '@/stores/activityStore';
import { ActivityItem } from './ActivityItem';
import type { ActivityType } from '@/types/activity';

interface ActivityTimelineProps {
  issueId: string;
  filterTypes?: ActivityType[];
  showFilter?: boolean;
}

// 活动类型过滤选项
const ACTIVITY_FILTER_OPTIONS: { value: ActivityType | 'all'; label: string }[] = [
  { value: 'all', label: '全部' },
  { value: 'issue_created', label: '创建' },
  { value: 'status_changed', label: '状态变更' },
  { value: 'priority_changed', label: '优先级变更' },
  { value: 'assignee_changed', label: '负责人变更' },
  { value: 'comment_added', label: '评论' },
];

export function ActivityTimeline({
  issueId,
  filterTypes,
  showFilter = false,
}: ActivityTimelineProps) {
  const {
    currentIssueActivities,
    filterTypes: storeFilterTypes,
    fetchIssueActivities,
    setFilterTypes,
    isLoading,
    error,
  } = useActivityStore();

  // 获取活动列表
  useEffect(() => {
    if (issueId) {
      fetchIssueActivities(issueId, { types: filterTypes });
    }
  }, [issueId, filterTypes, fetchIssueActivities]);

  // 处理过滤变更
  const handleFilterChange = (types: ActivityType[]) => {
    setFilterTypes(types);
    fetchIssueActivities(issueId, { types });
  };

  // 切换过滤类型
  const toggleFilter = (type: ActivityType | 'all') => {
    if (type === 'all') {
      handleFilterChange([]);
    } else {
      const currentTypes = storeFilterTypes;
      if (currentTypes.includes(type)) {
        handleFilterChange(currentTypes.filter((t) => t !== type));
      } else {
        handleFilterChange([...currentTypes, type]);
      }
    }
  };

  return (
    <div className="flex flex-col">
      {/* 标题和过滤器 */}
      <div className="border-b border-gray-100 px-4 py-2">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-medium text-gray-700">
            活动
            {currentIssueActivities.length > 0 && (
              <span className="ml-1 text-gray-400">({currentIssueActivities.length})</span>
            )}
          </h3>
        </div>

        {/* 过滤器 */}
        {showFilter && (
          <div className="mt-2 flex flex-wrap gap-1">
            {ACTIVITY_FILTER_OPTIONS.map((option) => {
              const isActive =
                option.value === 'all'
                  ? storeFilterTypes.length === 0
                  : storeFilterTypes.includes(option.value);

              return (
                <button
                  key={option.value}
                  onClick={() => toggleFilter(option.value)}
                  className={`rounded-full px-2 py-0.5 text-xs transition-colors ${
                    isActive
                      ? 'bg-indigo-100 text-indigo-700'
                      : 'bg-gray-100 text-gray-500 hover:bg-gray-200'
                  }`}
                >
                  {option.label}
                </button>
              );
            })}
          </div>
        )}
      </div>

      {/* 错误提示 */}
      {error && (
        <div className="mx-4 my-2 rounded bg-red-50 px-3 py-2 text-sm text-red-500">
          {error}
        </div>
      )}

      {/* 活动列表 */}
      <div className="flex-1 overflow-y-auto px-4">
        {isLoading && currentIssueActivities.length === 0 ? (
          <div className="py-8 text-center text-sm text-gray-400">加载中...</div>
        ) : currentIssueActivities.length === 0 ? (
          <div className="py-8 text-center text-sm text-gray-400">
            暂无活动记录
          </div>
        ) : (
          <div className="relative">
            {/* 移除最后一条活动的连接线 */}
            {currentIssueActivities.map((activity, index) => (
              <div
                key={activity.id}
                className={index === currentIssueActivities.length - 1 ? '[&>div>div:first-child>div:last-child]:hidden' : ''}
              >
                <ActivityItem activity={activity} />
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
