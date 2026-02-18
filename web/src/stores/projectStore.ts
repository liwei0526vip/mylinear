/**
 * Project 状态管理
 */

import { create } from 'zustand';
import type {
  Project,
  ProjectProgress,
  ProjectFilter,
  CreateProjectRequest,
  UpdateProjectRequest,
} from '../types/project';
import type { IssueFilter } from '../types/issue';
import * as projectApi from '../api/project';

interface ProjectState {
  // 状态
  projects: Project[];
  currentProject: Project | null;
  progress: ProjectProgress | null;
  projectIssues: {
    items: {
      id: string;
      team_id: string;
      number: number;
      title: string;
      status_id: string;
      priority: number;
      assignee_id?: string;
      created_at: string;
      completed_at?: string;
    }[];
    total: number;
    page: number;
  };
  total: number;
  page: number;
  isLoading: boolean;
  error: string | null;

  // Project Actions
  fetchTeamProjects: (
    teamId: string,
    filter?: ProjectFilter,
    page?: number,
    pageSize?: number
  ) => Promise<void>;
  fetchProject: (projectId: string) => Promise<void>;
  createProject: (workspaceId: string, data: CreateProjectRequest) => Promise<Project | null>;
  updateProject: (projectId: string, data: UpdateProjectRequest) => Promise<void>;
  deleteProject: (projectId: string) => Promise<void>;

  // Progress Action
  fetchProjectProgress: (projectId: string) => Promise<void>;

  // Issues Action
  fetchProjectIssues: (
    projectId: string,
    filter?: IssueFilter,
    page?: number,
    pageSize?: number
  ) => Promise<void>;

  // Utility
  clearError: () => void;
  reset: () => void;
}

export const useProjectStore = create<ProjectState>((set) => ({
  // 初始状态
  projects: [],
  currentProject: null,
  progress: null,
  projectIssues: {
    items: [],
    total: 0,
    page: 1,
  },
  total: 0,
  page: 1,
  isLoading: false,
  error: null,

  // 获取团队项目列表
  fetchTeamProjects: async (
    teamId: string,
    filter?: ProjectFilter,
    page = 1,
    pageSize = 20
  ) => {
    set({ isLoading: true, error: null });
    const response = await projectApi.fetchTeamProjects(teamId, filter, page, pageSize);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      return;
    }
    set({
      projects: response.data?.items || [],
      total: response.data?.total || 0,
      page: response.data?.page || 1,
      isLoading: false,
    });
  },

  // 获取项目详情
  fetchProject: async (projectId: string) => {
    set({ isLoading: true, error: null });
    const response = await projectApi.fetchProject(projectId);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      return;
    }
    set({ currentProject: response.data, isLoading: false });
  },

  // 创建项目
  createProject: async (workspaceId: string, data: CreateProjectRequest) => {
    set({ isLoading: true, error: null });
    const response = await projectApi.createProject(workspaceId, data);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      throw new Error(response.error);
    }
    const newProject = response.data;
    set((state) => ({
      projects: [newProject!, ...state.projects],
      total: state.total + 1,
      isLoading: false,
    }));
    return newProject;
  },

  // 更新项目
  updateProject: async (projectId: string, data: UpdateProjectRequest) => {
    set({ isLoading: true, error: null });
    const response = await projectApi.updateProject(projectId, data);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      throw new Error(response.error);
    }
    set((state) => ({
      currentProject: response.data,
      projects: state.projects.map((p) => (p.id === projectId ? response.data! : p)),
      isLoading: false,
    }));
  },

  // 删除项目
  deleteProject: async (projectId: string) => {
    set({ isLoading: true, error: null });
    const response = await projectApi.deleteProject(projectId);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      throw new Error(response.error);
    }
    set((state) => ({
      projects: state.projects.filter((p) => p.id !== projectId),
      total: state.total - 1,
      currentProject:
        state.currentProject?.id === projectId ? null : state.currentProject,
      isLoading: false,
    }));
  },

  // 获取项目进度
  fetchProjectProgress: async (projectId: string) => {
    set({ isLoading: true, error: null });
    const response = await projectApi.fetchProjectProgress(projectId);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      return;
    }
    set({ progress: response.data, isLoading: false });
  },

  // 获取项目关联 Issue 列表
  fetchProjectIssues: async (
    projectId: string,
    filter?: IssueFilter,
    page = 1,
    pageSize = 20
  ) => {
    set({ isLoading: true, error: null });
    const response = await projectApi.fetchProjectIssues(projectId, filter, page, pageSize);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      return;
    }
    set({
      projectIssues: {
        items: response.data?.items || [],
        total: response.data?.total || 0,
        page: response.data?.page || 1,
      },
      isLoading: false,
    });
  },

  // 清除错误
  clearError: () => set({ error: null }),

  // 重置状态
  reset: () =>
    set({
      projects: [],
      currentProject: null,
      progress: null,
      projectIssues: {
        items: [],
        total: 0,
        page: 1,
      },
      total: 0,
      page: 1,
      error: null,
      isLoading: false,
    }),
}));
