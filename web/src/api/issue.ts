/**
 * Issue API
 */

import { api } from './client';
import type {
  Issue,
  IssueListResponse,
  CreateIssueRequest,
  UpdateIssueRequest,
  UpdatePositionRequest,
  SubscriberListResponse,
  IssueFilter,
} from '../types/issue';

// =============================================================================
// Issue CRUD API
// =============================================================================

/**
 * 获取 Issue 列表
 */
export async function listIssues(
  teamId: string,
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
    if (filter.project_id) params.append('project_id', filter.project_id);
    if (filter.cycle_id) params.append('cycle_id', filter.cycle_id);
    if (filter.created_by_id) params.append('created_by_id', filter.created_by_id);
  }

  return api<IssueListResponse>(`/teams/${teamId}/issues?${params.toString()}`);
}

/**
 * 创建 Issue
 */
export async function createIssue(teamId: string, data: CreateIssueRequest) {
  return api<Issue>(`/teams/${teamId}/issues`, {
    method: 'POST',
    body: data,
  });
}

/**
 * 获取 Issue 详情
 */
export async function getIssue(issueId: string) {
  return api<Issue>(`/issues/${issueId}`);
}

/**
 * 更新 Issue
 */
export async function updateIssue(issueId: string, data: UpdateIssueRequest) {
  return api<Issue>(`/issues/${issueId}`, {
    method: 'PUT',
    body: data,
  });
}

/**
 * 删除 Issue
 */
export async function deleteIssue(issueId: string) {
  return api<{ message: string }>(`/issues/${issueId}`, {
    method: 'DELETE',
  });
}

/**
 * 恢复已删除的 Issue
 */
export async function restoreIssue(issueId: string) {
  return api<{ message: string }>(`/issues/${issueId}/restore`, {
    method: 'POST',
  });
}

// =============================================================================
// Issue 位置 API
// =============================================================================

/**
 * 更新 Issue 位置
 */
export async function updatePosition(issueId: string, data: UpdatePositionRequest) {
  return api<{ message: string }>(`/issues/${issueId}/position`, {
    method: 'PUT',
    body: data,
  });
}

// =============================================================================
// Issue 订阅 API
// =============================================================================

/**
 * 订阅 Issue
 */
export async function subscribeIssue(issueId: string) {
  return api<{ message: string }>(`/issues/${issueId}/subscribe`, {
    method: 'POST',
  });
}

/**
 * 取消订阅 Issue
 */
export async function unsubscribeIssue(issueId: string) {
  return api<{ message: string }>(`/issues/${issueId}/subscribe`, {
    method: 'DELETE',
  });
}

/**
 * 获取 Issue 订阅者列表
 */
export async function listSubscribers(issueId: string) {
  return api<SubscriberListResponse>(`/issues/${issueId}/subscribers`);
}
