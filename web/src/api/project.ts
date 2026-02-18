/**
 * Project API
 */

import { api } from './client';
import type {
  Project,
  ProjectProgress,
  ProjectListResponse,
  ProjectFilter,
  CreateProjectRequest,
  UpdateProjectRequest,
  ProjectIssueListResponse,
} from '../types/project';
import type { IssueFilter } from '../types/issue';

// =============================================================================
// Project CRUD API
// =============================================================================

/**
 * 创建项目（在工作区内）
 */
export async function createProject(workspaceId: string, data: CreateProjectRequest) {
  return api<Project>(`/workspaces/${workspaceId}/projects`, {
    method: 'POST',
    body: data,
  });
}

/**
 * 获取团队项目列表
 */
export async function fetchTeamProjects(
  teamId: string,
  filter?: ProjectFilter,
  page = 1,
  pageSize = 20
) {
  const params = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });

  if (filter?.status) {
    params.append('status', filter.status);
  }

  return api<ProjectListResponse>(`/teams/${teamId}/projects?${params.toString()}`);
}

/**
 * 获取项目详情
 */
export async function fetchProject(projectId: string) {
  return api<Project>(`/projects/${projectId}`);
}

/**
 * 更新项目
 */
export async function updateProject(projectId: string, data: UpdateProjectRequest) {
  return api<Project>(`/projects/${projectId}`, {
    method: 'PUT',
    body: data,
  });
}

/**
 * 删除项目
 */
export async function deleteProject(projectId: string) {
  return api<{ message: string }>(`/projects/${projectId}`, {
    method: 'DELETE',
  });
}

// =============================================================================
// Project 进度 API
// =============================================================================

/**
 * 获取项目进度
 */
export async function fetchProjectProgress(projectId: string) {
  return api<ProjectProgress>(`/projects/${projectId}/progress`);
}

// =============================================================================
// Project Issue 列表 API
// =============================================================================

/**
 * 获取项目关联的 Issue 列表
 */
export async function fetchProjectIssues(
  projectId: string,
  filter?: IssueFilter,
  page = 1,
  pageSize = 20
) {
  const params = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });

  if (filter) {
    if (filter.status_id) params.append('status_id', filter.status_id);
    if (filter.priority !== undefined) params.append('priority', String(filter.priority));
    if (filter.assignee_id) params.append('assignee_id', filter.assignee_id);
  }

  return api<ProjectIssueListResponse>(`/projects/${projectId}/issues?${params.toString()}`);
}
