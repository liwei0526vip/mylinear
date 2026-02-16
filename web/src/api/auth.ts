/**
 * 认证相关 API 调用
 */

import apiClient from '../lib/axios';
import type {
  LoginRequest,
  RegisterRequest,
  AuthResponse,
  RefreshTokenRequest,
  LogoutRequest,
  TokenPair,
} from '../types/auth';

/**
 * 用户登录
 */
export async function login(data: LoginRequest): Promise<AuthResponse> {
  const response = await apiClient.post<{ data: AuthResponse }>('/api/v1/auth/login', data);
  return response.data.data;
}

/**
 * 用户注册
 */
export async function register(data: RegisterRequest): Promise<AuthResponse> {
  const response = await apiClient.post<{ data: AuthResponse }>('/api/v1/auth/register', data);
  return response.data.data;
}

/**
 * 刷新令牌
 */
export async function refreshToken(data: RefreshTokenRequest): Promise<TokenPair> {
  const response = await apiClient.post<{ data: TokenPair }>('/api/v1/auth/refresh', data);
  return response.data.data;
}

/**
 * 用户登出
 */
export async function logout(data: LogoutRequest): Promise<void> {
  await apiClient.post('/api/v1/auth/logout', data);
}
