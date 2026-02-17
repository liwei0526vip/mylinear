/**
 * Team 状态管理
 */

import { create } from 'zustand';
import type {
  Team,
  TeamMember,
  CreateTeamRequest,
  UpdateTeamRequest,
  AddMemberRequest,
  TeamRole,
} from '../types/team';
import * as teamApi from '../api/team';

interface TeamState {
  // 状态
  teams: Team[];
  currentTeam: Team | null;
  members: TeamMember[];
  total: number;
  page: number;
  isLoading: boolean;
  error: string | null;

  // Team Actions
  fetchTeams: (workspaceId: string, page?: number, pageSize?: number) => Promise<void>;
  fetchTeam: (teamId: string) => Promise<void>;
  createTeam: (data: CreateTeamRequest) => Promise<Team | null>;
  updateTeam: (teamId: string, data: UpdateTeamRequest) => Promise<void>;
  deleteTeam: (teamId: string) => Promise<void>;

  // Member Actions
  fetchMembers: (teamId: string) => Promise<void>;
  addMember: (teamId: string, data: AddMemberRequest) => Promise<void>;
  removeMember: (teamId: string, userId: string) => Promise<void>;
  updateMemberRole: (teamId: string, userId: string, role: TeamRole) => Promise<void>;

  // Utility
  clearError: () => void;
  reset: () => void;
}

export const useTeamStore = create<TeamState>((set, get) => ({
  // 初始状态
  teams: [],
  currentTeam: null,
  members: [],
  total: 0,
  page: 1,
  isLoading: false,
  error: null,

  // 获取团队列表
  fetchTeams: async (workspaceId: string, page = 1, pageSize = 20) => {
    set({ isLoading: true, error: null });
    const response = await teamApi.listTeams(workspaceId, page, pageSize);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      return;
    }
    set({
      teams: response.data?.teams || [],
      total: response.data?.total || 0,
      page: response.data?.page || 1,
      isLoading: false,
    });
  },

  // 获取团队详情
  fetchTeam: async (teamId: string) => {
    set({ isLoading: true, error: null });
    const response = await teamApi.getTeam(teamId);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      return;
    }
    set({ currentTeam: response.data, isLoading: false });
  },

  // 创建团队
  createTeam: async (data: CreateTeamRequest) => {
    set({ isLoading: true, error: null });
    const response = await teamApi.createTeam(data);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      throw new Error(response.error);
    }
    const newTeam = response.data;
    set((state) => ({
      teams: [newTeam!, ...state.teams],
      total: state.total + 1,
      isLoading: false,
    }));
    return newTeam;
  },

  // 更新团队
  updateTeam: async (teamId: string, data: UpdateTeamRequest) => {
    set({ isLoading: true, error: null });
    const response = await teamApi.updateTeam(teamId, data);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      throw new Error(response.error);
    }
    set((state) => ({
      currentTeam: response.data,
      teams: state.teams.map((t) => (t.id === teamId ? response.data! : t)),
      isLoading: false,
    }));
  },

  // 删除团队
  deleteTeam: async (teamId: string) => {
    set({ isLoading: true, error: null });
    const response = await teamApi.deleteTeam(teamId);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      throw new Error(response.error);
    }
    set((state) => ({
      teams: state.teams.filter((t) => t.id !== teamId),
      total: state.total - 1,
      currentTeam: state.currentTeam?.id === teamId ? null : state.currentTeam,
      isLoading: false,
    }));
  },

  // 获取成员列表
  fetchMembers: async (teamId: string) => {
    set({ isLoading: true, error: null });
    const response = await teamApi.listMembers(teamId);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      return;
    }
    set({ members: response.data?.members || [], isLoading: false });
  },

  // 添加成员
  addMember: async (teamId: string, data: AddMemberRequest) => {
    set({ isLoading: true, error: null });
    const response = await teamApi.addMember(teamId, data);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      throw new Error(response.error);
    }
    // 刷新成员列表
    await get().fetchMembers(teamId);
  },

  // 移除成员
  removeMember: async (teamId: string, userId: string) => {
    set({ isLoading: true, error: null });
    const response = await teamApi.removeMember(teamId, userId);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      throw new Error(response.error);
    }
    set((state) => ({
      members: state.members.filter((m) => m.user_id !== userId),
      isLoading: false,
    }));
  },

  // 更新成员角色
  updateMemberRole: async (teamId: string, userId: string, role: TeamRole) => {
    set({ isLoading: true, error: null });
    const response = await teamApi.updateMemberRole(teamId, userId, role);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      throw new Error(response.error);
    }
    set((state) => ({
      members: state.members.map((m) =>
        m.user_id === userId ? { ...m, role } : m
      ),
      isLoading: false,
    }));
  },

  // 清除错误
  clearError: () => set({ error: null }),

  // 重置状态
  reset: () =>
    set({
      teams: [],
      currentTeam: null,
      members: [],
      total: 0,
      page: 1,
      error: null,
      isLoading: false,
    }),
}));
