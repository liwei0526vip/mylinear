/**
 * 工作流状态相关类型定义
 */

export type StateType = 'backlog' | 'unstarted' | 'started' | 'completed' | 'canceled';

export interface WorkflowState {
  id: string;
  team_id: string;
  name: string;
  type: StateType;
  color: string;
  position: number;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateStateRequest {
  name: string;
  type: StateType;
  color?: string;
  position?: number;
  description?: string;
}

export interface UpdateStateRequest {
  name?: string;
  type?: StateType;
  color?: string;
  position?: number;
  description?: string;
}
