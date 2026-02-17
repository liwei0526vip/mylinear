/**
 * Workspace 相关类型定义
 */

// 工作区信息
export interface Workspace {
  id: string;
  name: string;
  slug: string;
  logo_url?: string;
  created_at: string;
  updated_at: string;
}

// 工作区统计
export interface WorkspaceStats {
  teams_count: number;
  members_count: number;
  issues_count: number;
}

// 更新工作区请求
export interface UpdateWorkspaceRequest {
  name?: string;
  logo_url?: string;
}
