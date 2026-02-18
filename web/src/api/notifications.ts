/**
 * 通知 API
 */

import { api } from './client';
import type {
  Notification,
  NotificationListResponse,
  UnreadCountResponse,
  BatchReadRequest,
  BatchReadResponse,
} from '../types/notification';

/**
 * 获取通知列表
 */
export async function listNotifications(
  page = 1,
  pageSize = 20,
  read?: boolean
): Promise<{ data: NotificationListResponse | null; error: string | null }> {
  let endpoint = `/notifications?page=${page}&page_size=${pageSize}`;
  if (read !== undefined) {
    endpoint += `&read=${read}`;
  }
  return api<NotificationListResponse>(endpoint);
}

/**
 * 获取未读通知数量
 */
export async function getUnreadCount(): Promise<{
  data: UnreadCountResponse | null;
  error: string | null;
}> {
  return api<UnreadCountResponse>('/notifications/unread-count');
}

/**
 * 标记单条通知为已读
 */
export async function markAsRead(
  notificationId: string
): Promise<{ data: { message: string } | null; error: string | null }> {
  return api<{ message: string }>(`/notifications/${notificationId}/read`, {
    method: 'POST',
  });
}

/**
 * 标记所有通知为已读
 */
export async function markAllAsRead(): Promise<{
  data: BatchReadResponse | null;
  error: string | null;
}> {
  return api<BatchReadResponse>('/notifications/read-all', {
    method: 'POST',
  });
}

/**
 * 批量标记通知为已读
 */
export async function markBatchAsRead(
  ids: string[]
): Promise<{ data: BatchReadResponse | null; error: string | null }> {
  const body: BatchReadRequest = { ids };
  return api<BatchReadResponse>('/notifications/batch-read', {
    method: 'POST',
    body,
  });
}
