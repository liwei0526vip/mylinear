/**
 * 标签相关类型定义
 */

export interface Label {
    id: string;
    workspace_id: string;
    team_id?: string;
    name: string;
    color: string;
    created_at: string;
    updated_at: string;
}

export interface CreateLabelRequest {
    name: string;
    color?: string;
}

export interface UpdateLabelRequest {
    name?: string;
    color?: string;
}
