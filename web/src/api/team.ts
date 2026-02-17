/**
 * Team 和 TeamMember API
 */

import { api } from './client';
import type {
  Team,
  TeamMember,
  TeamListResponse,
  CreateTeamRequest,
  UpdateTeamRequest,
  AddMemberRequest,
  MemberListResponse,
} from '../types/team';

// =============================================================================
// Team API
// =============================================================================

/**
 * 获取团队列表
 */
export async function listTeams(workspaceId: string, page = 1, pageSize = 20) {
  return api<TeamListResponse>(
    `/teams?workspace_id=${workspaceId}&page=${page}&page_size=${pageSize}`
  );
}

/**
 * 创建团队
 */
export async function createTeam(data: CreateTeamRequest) {
  return api<Team>('/teams', {
    method: 'POST',
    body: data,
  });
}

/**
 * 获取团队详情
 */
export async function getTeam(teamId: string) {
  return api<Team>(`/teams/${teamId}`);
}

/**
 * 更新团队
 */
export async function updateTeam(teamId: string, data: UpdateTeamRequest) {
  return api<Team>(`/teams/${teamId}`, {
    method: 'PUT',
    body: data,
  });
}

/**
 * 删除团队
 */
export async function deleteTeam(teamId: string) {
  return api<{ message: string }>(`/teams/${teamId}`, {
    method: 'DELETE',
  });
}

// =============================================================================
// TeamMember API
// =============================================================================

/**
 * 获取团队成员列表
 */
export async function listMembers(teamId: string) {
  return api<MemberListResponse>(`/teams/${teamId}/members`);
}

/**
 * 添加团队成员
 */
export async function addMember(teamId: string, data: AddMemberRequest) {
  return api<{ message: string }>(`/teams/${teamId}/members`, {
    method: 'POST',
    body: data,
  });
}

/**
 * 移除团队成员
 */
export async function removeMember(teamId: string, userId: string) {
  return api<{ message: string }>(`/teams/${teamId}/members/${userId}`, {
    method: 'DELETE',
  });
}

/**
 * 更新成员角色
 */
export async function updateMemberRole(
  teamId: string,
  userId: string,
  role: string
) {
  return api<{ message: string }>(`/teams/${teamId}/members/${userId}`, {
    method: 'PUT',
    body: { role },
  });
}
