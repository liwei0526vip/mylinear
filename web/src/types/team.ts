/**
 * Team 和 TeamMember 相关类型定义
 */

// 团队角色
export type TeamRole = 'admin' | 'member';

// 团队信息
export interface Team {
  id: string;
  workspace_id: string;
  name: string;
  key: string;
  description: string;
  created_at: string;
  updated_at: string;
}

// 团队成员
export interface TeamMember {
  id: string; // user_id
  user_id: string;
  role: TeamRole;
  joined_at: string;
  user?: {
    id: string;
    name: string;
    email: string;
    username: string;
    avatar_url?: string;
  };
}

// 团队列表响应
export interface TeamListResponse {
  teams: Team[];
  total: number;
  page: number;
}

// 创建团队请求
export interface CreateTeamRequest {
  name: string;
  key: string;
  description: string;
  workspace_id: string;
}

// 更新团队请求
export interface UpdateTeamRequest {
  name?: string;
  key?: string;
  description?: string;
}

// 添加成员请求
export interface AddMemberRequest {
  user_id: string;
  role: TeamRole;
}

// 成员列表响应
export interface MemberListResponse {
  members: TeamMember[];
}
