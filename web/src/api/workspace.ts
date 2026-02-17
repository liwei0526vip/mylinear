/**
 * Workspace API
 */

import { api } from './client';
import type { Workspace, UpdateWorkspaceRequest } from '../types/workspace';

/**
 * 获取工作区信息
 */
export async function getWorkspace(id: string) {
  return api<Workspace>(`/workspaces/${id}`);
}

/**
 * 更新工作区
 */
export async function updateWorkspace(id: string, data: UpdateWorkspaceRequest) {
  return api<Workspace>(`/workspaces/${id}`, {
    method: 'PUT',
    body: data,
  });
}
