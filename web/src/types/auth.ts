/**
 * 认证相关类型定义
 */

// 登录请求
export interface LoginRequest {
  email: string;
  password: string;
}

// 注册请求
export interface RegisterRequest {
  email: string;
  username: string;
  password: string;
  name: string;
  workspace_id: string;
}

// 认证响应
export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

// 令牌对
export interface TokenPair {
  access_token: string;
  refresh_token: string;
}

// 刷新令牌请求
export interface RefreshTokenRequest {
  refresh_token: string;
}

// 登出请求
export interface LogoutRequest {
  refresh_token: string;
}

// 用户信息
export interface User {
  id: string;
  workspace_id: string;
  email: string;
  username: string;
  name: string;
  role: string;
  avatar_url?: string;
}

// 更新用户请求
export interface UpdateUserRequest {
  email?: string;
  username?: string;
  name?: string;
}
