/**
 * Activity 状态管理
 */

import { create } from 'zustand';
import type {
  Activity,
  ActivityListParams,
  ActivityType,
} from '../types/activity';
import * as activityApi from '../api/activities';

interface ActivityState {
  // 状态
  activitiesByIssue: Map<string, Activity[]>; // issueId -> activities
  currentIssueActivities: Activity[];
  total: number;
  page: number;
  pageSize: number;
  filterTypes: ActivityType[];
  isLoading: boolean;
  error: string | null;

  // Actions
  fetchIssueActivities: (issueId: string, params?: ActivityListParams) => Promise<void>;
  setFilterTypes: (types: ActivityType[]) => void;
  clearFilterTypes: () => void;

  // Utility
  getActivitiesByIssue: (issueId: string) => Activity[];
  clearError: () => void;
  reset: () => void;
}

export const useActivityStore = create<ActivityState>((set, get) => ({
  // 初始状态
  activitiesByIssue: new Map(),
  currentIssueActivities: [],
  total: 0,
  page: 1,
  pageSize: 50,
  filterTypes: [],
  isLoading: false,
  error: null,

  // 获取 Issue 的活动列表
  fetchIssueActivities: async (issueId: string, params?: ActivityListParams) => {
    set({ isLoading: true, error: null });

    // 合并过滤条件
    const requestParams: ActivityListParams = {
      ...params,
      types: params?.types || get().filterTypes,
    };

    const response = await activityApi.fetchIssueActivities(issueId, requestParams);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      return;
    }

    const activities = response.data?.activities || [];
    const activitiesByIssue = new Map(get().activitiesByIssue);
    activitiesByIssue.set(issueId, activities);

    set({
      activitiesByIssue,
      currentIssueActivities: activities,
      total: response.data?.total || 0,
      page: response.data?.page || 1,
      pageSize: response.data?.page_size || 50,
      isLoading: false,
    });
  },

  // 设置活动类型过滤
  setFilterTypes: (types: ActivityType[]) => {
    set({ filterTypes: types });
  },

  // 清除活动类型过滤
  clearFilterTypes: () => {
    set({ filterTypes: [] });
  },

  // 获取指定 Issue 的活动
  getActivitiesByIssue: (issueId: string) => {
    return get().activitiesByIssue.get(issueId) || [];
  },

  // 清除错误
  clearError: () => set({ error: null }),

  // 重置状态
  reset: () =>
    set({
      activitiesByIssue: new Map(),
      currentIssueActivities: [],
      total: 0,
      page: 1,
      pageSize: 50,
      filterTypes: [],
      error: null,
      isLoading: false,
    }),
}));
