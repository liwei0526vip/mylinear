/**
 * 用户相关 API 调用
 */

import apiClient from '../lib/axios';
import type { User, UpdateUserRequest } from '../types/user';

/**
 * 获取当前用户信息
 */
export async function getCurrentUser(): Promise<User> {
  const response = await apiClient.get<{ data: User }>('/api/v1/users/me');
  return response.data.data;
}

/**
 * 更新当前用户信息
 */
export async function updateCurrentUser(data: UpdateUserRequest): Promise<User> {
  const response = await apiClient.patch<{ data: User }>('/api/v1/users/me', data);
  return response.data.data;
}

/**
 * 上传头像
 */
export async function uploadAvatar(file: File): Promise<{ avatar_url: string }> {
  const formData = new FormData();
  formData.append('avatar', file);

  const response = await apiClient.post<{ data: { avatar_url: string } }>(
    '/api/v1/users/me/avatar',
    formData,
    {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    }
  );

  return response.data.data;
}
