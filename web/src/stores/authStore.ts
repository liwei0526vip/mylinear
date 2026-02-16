/**
 * 认证状态管理
 */

import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User } from '../types/user';
import type { LoginRequest, RegisterRequest } from '../types/auth';
import * as authApi from '../api/auth';
import * as userApi from '../api/user';
import { getRefreshToken, setTokens, clearTokens, getAccessToken } from '../lib/axios';

interface AuthState {
  // 状态
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  error: string | null;

  // Actions
  login: (data: LoginRequest) => Promise<void>;
  register: (data: RegisterRequest) => Promise<void>;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
  updateUser: (data: Partial<User>) => Promise<void>;
  clearError: () => void;
  checkAuth: () => Promise<void>;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
  // 初始状态
  user: null,
  isLoading: false,
  isAuthenticated: false,
  error: null,

  // 登录
  login: async (data: LoginRequest) => {
    set({ isLoading: true, error: null });
    try {
      const response = await authApi.login(data);
      setTokens(response.access_token, response.refresh_token);
      set({
        user: response.user,
        isAuthenticated: true,
        isLoading: false,
      });
    } catch (err) {
      const message = err instanceof Error ? err.message : '登录失败';
      set({ error: message, isLoading: false });
      throw err;
    }
  },

  // 注册
  register: async (data: RegisterRequest) => {
    set({ isLoading: true, error: null });
    try {
      const response = await authApi.register(data);
      setTokens(response.access_token, response.refresh_token);
      set({
        user: response.user,
        isAuthenticated: true,
        isLoading: false,
      });
    } catch (err) {
      const message = err instanceof Error ? err.message : '注册失败';
      set({ error: message, isLoading: false });
      throw err;
    }
  },

  // 登出
  logout: async () => {
    const refreshToken = getRefreshToken();
    if (refreshToken) {
      try {
        await authApi.logout({ refresh_token: refreshToken });
      } catch {
        // 忽略登出错误
      }
    }
    clearTokens();
    set({
      user: null,
      isAuthenticated: false,
      error: null,
    });
  },

  // 刷新用户信息
  refreshUser: async () => {
    try {
      const user = await userApi.getCurrentUser();
      set({ user });
    } catch (err) {
      // 如果获取用户失败，清除认证状态
      set({ user: null, isAuthenticated: false });
    }
  },

  // 更新用户信息
  updateUser: async (data: Partial<User>) => {
    set({ isLoading: true, error: null });
    try {
      const user = await userApi.updateCurrentUser(data);
      set({ user, isLoading: false });
    } catch (err) {
      const message = err instanceof Error ? err.message : '更新失败';
      set({ error: message, isLoading: false });
      throw err;
    }
  },

  // 清除错误
  clearError: () => set({ error: null }),

  // 检查认证状态
  checkAuth: async () => {
    const accessToken = getAccessToken();
    const refreshToken = getRefreshToken();

    // 如果没有 token，直接设置为未认证
    if (!accessToken && !refreshToken) {
      set({ isAuthenticated: false, user: null, isLoading: false });
      return;
    }

    // 如果有 token 但还没有用户信息，尝试获取
    set({ isLoading: true });
    try {
      const user = await userApi.getCurrentUser();
      set({ user, isAuthenticated: true, isLoading: false });
    } catch {
      // API 调用失败，可能是 token 过期或无效
      // 检查是否还有 refresh token（可能刚刷新过）
      const currentRefreshToken = getRefreshToken();
      if (!currentRefreshToken) {
        set({ isAuthenticated: false, user: null, isLoading: false });
      } else {
        // 还有 refresh token，保持认证状态，下次请求会自动刷新
        set({ isLoading: false });
      }
    }
  },
}),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);
