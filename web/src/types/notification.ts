/**
 * 通知相关类型定义
 */

// 通知类型
export type NotificationType =
  | 'issue_assigned'
  | 'issue_mentioned'
  | 'issue_status_changed'
  | 'issue_priority_changed'
  | 'issue_commented';

// 通知渠道
export type NotificationChannel = 'in_app' | 'email';

// 通知信息
export interface Notification {
  id: string;
  user_id: string;
  type: NotificationType;
  title: string;
  body?: string;
  resource_type?: string;
  resource_id?: string;
  read_at?: string;
  created_at: string;
}

// 通知偏好配置
export interface NotificationPreference {
  id?: string;
  user_id?: string;
  channel: NotificationChannel;
  type: NotificationType;
  enabled: boolean;
}

// 通知列表响应
export interface NotificationListResponse {
  notifications: Notification[];
  total: number;
  page: number;
  page_size: number;
}

// 未读数量响应
export interface UnreadCountResponse {
  count: number;
}

// 批量标记已读请求
export interface BatchReadRequest {
  ids: string[];
}

// 批量标记已读响应
export interface BatchReadResponse {
  message: string;
  marked: number;
}

// 更新偏好配置请求
export interface UpdatePreferencesRequest {
  preferences: {
    channel: NotificationChannel;
    type: NotificationType;
    enabled: boolean;
  }[];
}

// 通知类型标签
export const NOTIFICATION_TYPE_LABELS: Record<NotificationType, string> = {
  issue_assigned: 'Issue 指派',
  issue_mentioned: '@提及',
  issue_status_changed: '状态变更',
  issue_priority_changed: '优先级变更',
  issue_commented: '新评论',
};

// 获取通知类型标签
export function getNotificationTypeLabel(type: NotificationType): string {
  return NOTIFICATION_TYPE_LABELS[type] || type;
}
