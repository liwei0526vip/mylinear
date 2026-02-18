/**
 * 通知状态管理
 */

import { create } from 'zustand';
import type { Notification, NotificationPreference } from '../types/notification';
import * as notificationApi from '../api/notifications';
import * as preferenceApi from '../api/notification-preferences';

interface NotificationState {
  // 状态
  notifications: Notification[];
  unreadCount: number;
  preferences: NotificationPreference[];
  total: number;
  page: number;
  pageSize: number;
  isLoading: boolean;
  error: string | null;

  // 通知 Actions
  fetchNotifications: (page?: number, pageSize?: number, read?: boolean) => Promise<void>;
  fetchUnreadCount: () => Promise<void>;
  markAsRead: (notificationId: string) => Promise<boolean>;
  markAllAsRead: () => Promise<number>;
  markBatchAsRead: (ids: string[]) => Promise<number>;

  // 偏好配置 Actions
  fetchPreferences: (channel?: string) => Promise<void>;
  updatePreferences: (preferences: NotificationPreference[]) => Promise<boolean>;

  // Utility
  getUnreadNotifications: () => Notification[];
  clearError: () => void;
  reset: () => void;
}

export const useNotificationStore = create<NotificationState>((set, get) => ({
  // 初始状态
  notifications: [],
  unreadCount: 0,
  preferences: [],
  total: 0,
  page: 1,
  pageSize: 20,
  isLoading: false,
  error: null,

  // 获取通知列表
  fetchNotifications: async (page = 1, pageSize = 20, read?: boolean) => {
    set({ isLoading: true, error: null });
    const response = await notificationApi.listNotifications(page, pageSize, read);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      return;
    }

    set({
      notifications: response.data?.notifications || [],
      total: response.data?.total || 0,
      page: response.data?.page || 1,
      pageSize: response.data?.page_size || 20,
      isLoading: false,
    });
  },

  // 获取未读数量
  fetchUnreadCount: async () => {
    const response = await notificationApi.getUnreadCount();
    if (response.error) {
      // 静默失败，不影响用户体验
      console.error('获取未读数量失败:', response.error);
      return;
    }
    set({ unreadCount: response.data?.count || 0 });
  },

  // 标记单条已读
  markAsRead: async (notificationId: string) => {
    const response = await notificationApi.markAsRead(notificationId);
    if (response.error) {
      set({ error: response.error });
      return false;
    }

    // 更新本地状态
    const notifications = get().notifications.map((n) =>
      n.id === notificationId ? { ...n, read_at: new Date().toISOString() } : n
    );
    const unreadCount = Math.max(0, get().unreadCount - 1);
    set({ notifications, unreadCount });
    return true;
  },

  // 标记全部已读
  markAllAsRead: async () => {
    const response = await notificationApi.markAllAsRead();
    if (response.error) {
      set({ error: response.error });
      return 0;
    }

    const marked = response.data?.marked || 0;
    const now = new Date().toISOString();
    const notifications = get().notifications.map((n) => ({
      ...n,
      read_at: n.read_at || now,
    }));
    set({ notifications, unreadCount: 0 });
    return marked;
  },

  // 批量标记已读
  markBatchAsRead: async (ids: string[]) => {
    const response = await notificationApi.markBatchAsRead(ids);
    if (response.error) {
      set({ error: response.error });
      return 0;
    }

    const marked = response.data?.marked || 0;
    const now = new Date().toISOString();
    const notifications = get().notifications.map((n) =>
      ids.includes(n.id) ? { ...n, read_at: n.read_at || now } : n
    );
    const unreadCount = Math.max(0, get().unreadCount - marked);
    set({ notifications, unreadCount });
    return marked;
  },

  // 获取偏好配置
  fetchPreferences: async (channel?: string) => {
    set({ isLoading: true, error: null });
    const response = await preferenceApi.getPreferences(channel);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      return;
    }
    set({
      preferences: response.data?.preferences || [],
      isLoading: false,
    });
  },

  // 更新偏好配置
  updatePreferences: async (preferences: NotificationPreference[]) => {
    set({ isLoading: true, error: null });
    const response = await preferenceApi.updatePreferences({
      preferences: preferences.map((p) => ({
        channel: p.channel,
        type: p.type,
        enabled: p.enabled,
      })),
    });
    if (response.error) {
      set({ error: response.error, isLoading: false });
      return false;
    }
    set({ preferences, isLoading: false });
    return true;
  },

  // 获取未读通知
  getUnreadNotifications: () => {
    return get().notifications.filter((n) => !n.read_at);
  },

  // 清除错误
  clearError: () => set({ error: null }),

  // 重置状态
  reset: () =>
    set({
      notifications: [],
      unreadCount: 0,
      preferences: [],
      total: 0,
      page: 1,
      pageSize: 20,
      isLoading: false,
      error: null,
    }),
}));
