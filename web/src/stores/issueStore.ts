/**
 * Issue 状态管理
 */

import { create } from 'zustand';
import type {
  Issue,
  IssueFilter,
  CreateIssueRequest,
  UpdateIssueRequest,
  UpdatePositionRequest,
  IssueSubscriber,
  IssuePriority,
} from '../types/issue';
import * as issueApi from '../api/issue';

interface IssueState {
  // 状态
  issues: Map<string, Issue>; // issueId -> Issue
  issuesByTeam: Map<string, string[]>; // teamId -> issueId[]
  currentIssue: Issue | null;
  subscribers: IssueSubscriber[];
  filters: IssueFilter;
  total: number;
  page: number;
  isLoading: boolean;
  error: string | null;

  // Issue CRUD Actions
  fetchIssues: (teamId: string, filter?: IssueFilter, page?: number, pageSize?: number) => Promise<void>;
  fetchIssue: (issueId: string) => Promise<void>;
  createIssue: (teamId: string, data: CreateIssueRequest) => Promise<Issue | null>;
  updateIssue: (issueId: string, data: UpdateIssueRequest) => Promise<void>;
  deleteIssue: (issueId: string) => Promise<void>;
  restoreIssue: (issueId: string) => Promise<void>;

  // Position Actions
  updatePosition: (issueId: string, data: UpdatePositionRequest) => Promise<void>;

  // Subscription Actions
  subscribe: (issueId: string) => Promise<void>;
  unsubscribe: (issueId: string) => Promise<void>;
  fetchSubscribers: (issueId: string) => Promise<void>;

  // Filter Actions
  setFilter: (filter: Partial<IssueFilter>) => void;
  clearFilter: () => void;

  // Utility
  getIssueById: (issueId: string) => Issue | undefined;
  getIssuesByTeam: (teamId: string) => Issue[];
  setCurrentIssue: (issue: Issue | null) => void;
  clearError: () => void;
  reset: () => void;
}

const initialFilter: IssueFilter = {};

export const useIssueStore = create<IssueState>((set, get) => ({
  // 初始状态
  issues: new Map(),
  issuesByTeam: new Map(),
  currentIssue: null,
  subscribers: [],
  filters: initialFilter,
  total: 0,
  page: 1,
  isLoading: false,
  error: null,

  // 获取 Issue 列表
  fetchIssues: async (teamId: string, filter?: IssueFilter, page = 1, pageSize = 20) => {
    set({ isLoading: true, error: null });
    const response = await issueApi.listIssues(teamId, filter, page, pageSize);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      return;
    }

    const issues = response.data?.issues || [];
    const issueMap = new Map(get().issues);
    const teamIssues: string[] = [];

    issues.forEach((issue) => {
      issueMap.set(issue.id, issue);
      teamIssues.push(issue.id);
    });

    const issuesByTeam = new Map(get().issuesByTeam);
    issuesByTeam.set(teamId, teamIssues);

    set({
      issues: issueMap,
      issuesByTeam,
      total: response.data?.total || 0,
      page: response.data?.page || 1,
      filters: filter || initialFilter,
      isLoading: false,
    });
  },

  // 获取 Issue 详情
  fetchIssue: async (issueId: string) => {
    set({ isLoading: true, error: null });
    const response = await issueApi.getIssue(issueId);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      return;
    }
    const issue = response.data;
    if (issue) {
      const issueMap = new Map(get().issues);
      issueMap.set(issue.id, issue);
      set({ currentIssue: issue, issues: issueMap, isLoading: false });
    } else {
      set({ isLoading: false });
    }
  },

  // 创建 Issue
  createIssue: async (teamId: string, data: CreateIssueRequest) => {
    set({ isLoading: true, error: null });
    const response = await issueApi.createIssue(teamId, data);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      throw new Error(response.error);
    }
    const newIssue = response.data;
    if (newIssue) {
      const issueMap = new Map(get().issues);
      issueMap.set(newIssue.id, newIssue);

      const issuesByTeam = new Map(get().issuesByTeam);
      const teamIssues = issuesByTeam.get(teamId) || [];
      issuesByTeam.set(teamId, [newIssue.id, ...teamIssues]);

      set((state) => ({
        issues: issueMap,
        issuesByTeam,
        total: state.total + 1,
        isLoading: false,
      }));
    }
    return newIssue;
  },

  // 更新 Issue
  updateIssue: async (issueId: string, data: UpdateIssueRequest) => {
    set({ isLoading: true, error: null });
    const response = await issueApi.updateIssue(issueId, data);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      throw new Error(response.error);
    }
    const updatedIssue = response.data;
    if (updatedIssue) {
      const issueMap = new Map(get().issues);
      issueMap.set(issueId, updatedIssue);
      set((state) => ({
        issues: issueMap,
        currentIssue: state.currentIssue?.id === issueId ? updatedIssue : state.currentIssue,
        isLoading: false,
      }));
    }
  },

  // 删除 Issue
  deleteIssue: async (issueId: string) => {
    set({ isLoading: true, error: null });
    const response = await issueApi.deleteIssue(issueId);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      throw new Error(response.error);
    }

    const issueMap = new Map(get().issues);
    const issue = issueMap.get(issueId);
    issueMap.delete(issueId);

    const issuesByTeam = new Map(get().issuesByTeam);
    if (issue) {
      const teamIssues = issuesByTeam.get(issue.team_id) || [];
      issuesByTeam.set(
        issue.team_id,
        teamIssues.filter((id) => id !== issueId)
      );
    }

    set((state) => ({
      issues: issueMap,
      issuesByTeam,
      total: state.total - 1,
      currentIssue: state.currentIssue?.id === issueId ? null : state.currentIssue,
      isLoading: false,
    }));
  },

  // 恢复 Issue
  restoreIssue: async (issueId: string) => {
    set({ isLoading: true, error: null });
    const response = await issueApi.restoreIssue(issueId);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      throw new Error(response.error);
    }
    // 恢复后重新获取 Issue
    await get().fetchIssue(issueId);
  },

  // 更新位置
  updatePosition: async (issueId: string, data: UpdatePositionRequest) => {
    set({ isLoading: true, error: null });
    const response = await issueApi.updatePosition(issueId, data);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      throw new Error(response.error);
    }

    // 更新本地状态
    const issueMap = new Map(get().issues);
    const issue = issueMap.get(issueId);
    if (issue) {
      issueMap.set(issueId, {
        ...issue,
        position: data.position,
        status_id: data.status_id || issue.status_id,
      });
      set({ issues: issueMap, isLoading: false });
    } else {
      set({ isLoading: false });
    }
  },

  // 订阅 Issue
  subscribe: async (issueId: string) => {
    set({ isLoading: true, error: null });
    const response = await issueApi.subscribeIssue(issueId);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      throw new Error(response.error);
    }
    set({ isLoading: false });
  },

  // 取消订阅
  unsubscribe: async (issueId: string) => {
    set({ isLoading: true, error: null });
    const response = await issueApi.unsubscribeIssue(issueId);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      throw new Error(response.error);
    }
    set({ isLoading: false });
  },

  // 获取订阅者列表
  fetchSubscribers: async (issueId: string) => {
    set({ isLoading: true, error: null });
    const response = await issueApi.listSubscribers(issueId);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      return;
    }
    set({ subscribers: response.data?.subscribers || [], isLoading: false });
  },

  // 设置过滤条件
  setFilter: (filter: Partial<IssueFilter>) => {
    set((state) => ({
      filters: { ...state.filters, ...filter },
    }));
  },

  // 清除过滤条件
  clearFilter: () => {
    set({ filters: initialFilter });
  },

  // 通过 ID 获取 Issue
  getIssueById: (issueId: string) => {
    return get().issues.get(issueId);
  },

  // 获取团队的所有 Issue
  getIssuesByTeam: (teamId: string) => {
    const issueIds = get().issuesByTeam.get(teamId) || [];
    return issueIds
      .map((id) => get().issues.get(id))
      .filter((issue): issue is Issue => issue !== undefined);
  },

  // 设置当前 Issue
  setCurrentIssue: (issue: Issue | null) => {
    set({ currentIssue: issue });
  },

  // 清除错误
  clearError: () => set({ error: null }),

  // 重置状态
  reset: () =>
    set({
      issues: new Map(),
      issuesByTeam: new Map(),
      currentIssue: null,
      subscribers: [],
      filters: initialFilter,
      total: 0,
      page: 1,
      error: null,
      isLoading: false,
    }),
}));
