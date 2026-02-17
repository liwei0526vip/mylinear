/**
 * Workspace 状态管理
 */

import { create } from 'zustand';
import type { Workspace, WorkspaceStats, UpdateWorkspaceRequest } from '../types/workspace';
import * as workspaceApi from '../api/workspace';

interface WorkspaceState {
  // 状态
  workspace: Workspace | null;
  stats: WorkspaceStats | null;
  isLoading: boolean;
  error: string | null;

  // Actions
  fetchWorkspace: (id: string) => Promise<void>;
  updateWorkspace: (id: string, data: UpdateWorkspaceRequest) => Promise<void>;
  clearError: () => void;
  reset: () => void;
}

export const useWorkspaceStore = create<WorkspaceState>((set) => ({
  // 初始状态
  workspace: null,
  stats: null,
  isLoading: false,
  error: null,

  // 获取工作区
  fetchWorkspace: async (id: string) => {
    set({ isLoading: true, error: null });
    const response = await workspaceApi.getWorkspace(id);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      return;
    }
    set({ workspace: response.data, isLoading: false });
  },

  // 更新工作区
  updateWorkspace: async (id: string, data: UpdateWorkspaceRequest) => {
    set({ isLoading: true, error: null });
    const response = await workspaceApi.updateWorkspace(id, data);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      throw new Error(response.error);
    }
    set({ workspace: response.data, isLoading: false });
  },

  // 清除错误
  clearError: () => set({ error: null }),

  // 重置状态
  reset: () => set({ workspace: null, stats: null, error: null, isLoading: false }),
}));
