/**
 * 标签 API
 */
import apiClient from '../lib/axios';
import type { Label, CreateLabelRequest, UpdateLabelRequest } from '../types/label';

const labelApi = {
    /**
     * 获取团队的所有标签 (包含工作区标签)
     */
    listLabels: (teamId: string) => {
        return apiClient.get<{ data: Label[] }>(`/api/v1/teams/${teamId}/labels`);
    },

    /**
     * 为团队创建专属标签
     */
    createLabel: (teamId: string, data: CreateLabelRequest) => {
        return apiClient.post<{ data: Label }>(`/api/v1/teams/${teamId}/labels`, data);
    },

    /**
     * 更新标签 (前端目前主要用在团队标签)
     */
    updateLabel: (id: string, data: UpdateLabelRequest) => {
        return apiClient.put<{ data: Label }>(`/api/v1/labels/${id}`, data);
    },

    /**
     * 删除标签
     */
    deleteLabel: (id: string) => {
        return apiClient.delete(`/api/v1/labels/${id}`);
    },
};

export default labelApi;
