import { create } from 'zustand';
import type { WorkflowState, CreateStateRequest, UpdateStateRequest } from '../types/workflow';
import workflowApi from '../api/workflow';

interface WorkflowStateStore {
    states: WorkflowState[];
    loading: boolean;
    error: string | null;
    fetchStates: (teamId: string) => Promise<void>;
    addState: (teamId: string, data: CreateStateRequest) => Promise<void>;
    updateState: (id: string, data: UpdateStateRequest) => Promise<void>;
    deleteState: (id: string) => Promise<void>;
}

export const useWorkflowStore = create<WorkflowStateStore>((set) => ({
    states: [],
    loading: false,
    error: null,

    fetchStates: async (teamId: string) => {
        set({ loading: true, error: null });
        try {
            const response = await workflowApi.listStates(teamId);
            // 增加容错性处理，确保 states 始终为数组
            const states = response?.data?.data || [];
            set({ states, loading: false });
        } catch (error: any) {
            // 即使出错也保留原有数组，不设为 undefined
            set({
                error: error.response?.data?.error || error.message || '加载工作流状态失败',
                loading: false
            });
        }
    },

    addState: async (teamId: string, data: CreateStateRequest) => {
        try {
            const response = await workflowApi.createState(teamId, data);
            const newState = response.data.data;
            if (newState) {
                set((state) => ({
                    states: [...state.states, newState].sort((a, b) => a.position - b.position),
                }));
            }
        } catch (error: any) {
            throw new Error(error.response?.data?.error || '创建状态失败');
        }
    },

    updateState: async (id: string, data: UpdateStateRequest) => {
        try {
            const response = await workflowApi.updateState(id, data);
            const updatedState = response.data.data;
            if (updatedState) {
                set((state) => ({
                    states: state.states
                        .map((s) => (s.id === id ? updatedState : s))
                        .sort((a, b) => a.position - b.position),
                }));
            }
        } catch (error: any) {
            throw new Error(error.response?.data?.error || '更新状态失败');
        }
    },

    deleteState: async (id: string) => {
        try {
            await workflowApi.deleteState(id);
            set((state) => ({
                states: state.states.filter((s) => s.id !== id),
            }));
        } catch (error: any) {
            throw new Error(error.response?.data?.error || '删除状态失败');
        }
    },
}));
