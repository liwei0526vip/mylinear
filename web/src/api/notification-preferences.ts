/**
 * 通知偏好配置 API
 */

import { api } from './client';
import type {
  NotificationPreference,
  UpdatePreferencesRequest,
} from '../types/notification';

interface PreferencesResponse {
  preferences: NotificationPreference[];
}

/**
 * 获取通知偏好配置
 */
export async function getPreferences(
  channel?: string
): Promise<{ data: PreferencesResponse | null; error: string | null }> {
  let endpoint = '/notification-preferences';
  if (channel) {
    endpoint += `?channel=${channel}`;
  }
  return api<PreferencesResponse>(endpoint);
}

/**
 * 更新通知偏好配置
 */
export async function updatePreferences(
  preferences: UpdatePreferencesRequest
): Promise<{ data: { message: string } | null; error: string | null }> {
  return api<{ message: string }>('/notification-preferences', {
    method: 'PUT',
    body: preferences,
  });
}
