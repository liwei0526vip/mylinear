import { create } from 'zustand';
import type { Label, CreateLabelRequest, UpdateLabelRequest } from '../types/label';
import labelApi from '../api/label';

interface LabelStore {
    labels: Label[];
    loading: boolean;
    error: string | null;
    fetchLabels: (teamId: string) => Promise<void>;
    addLabel: (teamId: string, data: CreateLabelRequest) => Promise<void>;
    updateLabel: (id: string, data: UpdateLabelRequest) => Promise<void>;
    deleteLabel: (id: string) => Promise<void>;
}

export const useLabelStore = create<LabelStore>((set) => ({
    labels: [],
    loading: false,
    error: null,

    fetchLabels: async (teamId: string) => {
        set({ loading: true, error: null });
        try {
            const response = await labelApi.listLabels(teamId);
            const labels = response?.data?.data || [];
            set({ labels, loading: false });
        } catch (error: any) {
            set({
                error: error.response?.data?.error || error.message || '加载标签失败',
                loading: false
            });
        }
    },

    addLabel: async (teamId: string, data: CreateLabelRequest) => {
        try {
            const response = await labelApi.createLabel(teamId, data);
            const newLabel = response.data.data;
            if (newLabel) {
                set((state) => ({
                    labels: [...state.labels, newLabel],
                }));
            }
        } catch (error: any) {
            throw new Error(error.response?.data?.error || '创建标签失败');
        }
    },

    updateLabel: async (id: string, data: UpdateLabelRequest) => {
        try {
            const response = await labelApi.updateLabel(id, data);
            const updatedLabel = response.data.data;
            if (updatedLabel) {
                set((state) => ({
                    labels: state.labels.map((l) => (l.id === id ? updatedLabel : l)),
                }));
            }
        } catch (error: any) {
            throw new Error(error.response?.data?.error || '更新标签失败');
        }
    },

    deleteLabel: async (id: string) => {
        try {
            await labelApi.deleteLabel(id);
            set((state) => ({
                labels: state.labels.filter((l) => l.id !== id),
            }));
        } catch (error: any) {
            throw new Error(error.response?.data?.error || '删除标签失败');
        }
    },
}));
