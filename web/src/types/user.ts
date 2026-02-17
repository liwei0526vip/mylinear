/**
 * 用户相关类型定义
 */

// 用户信息（从 auth.ts 重导出）
export interface User {
  id: string;
  email: string;
  username: string;
  name: string;
  role: string;
  workspace_id?: string;
  avatar_url?: string;
}

// 更新用户请求
export interface UpdateUserRequest {
  email?: string;
  username?: string;
  name?: string;
}

// 上传头像响应
export interface UploadAvatarResponse {
  avatar_url: string;
}
