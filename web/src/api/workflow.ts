/**
 * 工作流状态 API
 */
import apiClient from '../lib/axios';
import type { WorkflowState, CreateStateRequest, UpdateStateRequest } from '../types/workflow';

const workflowApi = {
    /**
     * 获取团队的所有工作流状态
     */
    listStates: (teamId: string) => {
        return apiClient.get<{ data: WorkflowState[] }>(`/api/v1/teams/${teamId}/workflow-states`);
    },

    /**
     * 为团队创建工作流状态
     */
    createState: (teamId: string, data: CreateStateRequest) => {
        return apiClient.post<{ data: WorkflowState }>(`/api/v1/teams/${teamId}/workflow-states`, data);
    },

    /**
     * 更新工作流状态
     */
    updateState: (id: string, data: UpdateStateRequest) => {
        return apiClient.put<{ data: WorkflowState }>(`/api/v1/workflow-states/${id}`, data);
    },

    /**
     * 删除工作流状态
     */
    deleteState: (id: string) => {
        return apiClient.delete(`/api/v1/workflow-states/${id}`);
    },
};

export default workflowApi;
