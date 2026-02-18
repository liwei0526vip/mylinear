/**
 * Project 相关类型定义
 */

// Project 状态
export type ProjectStatus = 'planned' | 'in_progress' | 'paused' | 'completed' | 'cancelled';

// Project 信息
export interface Project {
  id: string;
  workspace_id: string;
  name: string;
  description?: string;
  status: ProjectStatus;
  priority: number;
  lead_id?: string;
  start_date?: string;
  target_date?: string;
  teams: string[];
  labels: string[];
  completed_at?: string;
  created_at: string;
  updated_at: string;
  deleted_at?: string;

  // 关联数据
  lead?: {
    id: string;
    name: string;
    email: string;
    avatar_url?: string;
  };
}

// Project 进度
export interface ProjectProgress {
  project_id: string;
  total_issues: number;
  completed_issues: number;
  cancelled_issues: number;
  progress_percent: number;
}

// Project 过滤条件
export interface ProjectFilter {
  status?: ProjectStatus;
}

// Project 列表响应
export interface ProjectListResponse {
  items: Project[];
  total: number;
  page: number;
}

// 创建 Project 请求
export interface CreateProjectRequest {
  name: string;
  description?: string;
  lead_id?: string;
  start_date?: string;
  target_date?: string;
  teams?: string[];
  labels?: string[];
}

// 更新 Project 请求
export interface UpdateProjectRequest {
  name?: string;
  description?: string;
  status?: ProjectStatus;
  priority?: number;
  lead_id?: string;
  start_date?: string;
  target_date?: string;
  teams?: string[];
  labels?: string[];
}

// Project Issue 列表响应
export interface ProjectIssueListResponse {
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
}

// 状态选项
export const STATUS_OPTIONS: { value: ProjectStatus; label: string; color: string }[] = [
  { value: 'planned', label: '计划中', color: '#6b7280' },
  { value: 'in_progress', label: '进行中', color: '#3b82f6' },
  { value: 'paused', label: '已暂停', color: '#f59e0b' },
  { value: 'completed', label: '已完成', color: '#22c55e' },
  { value: 'cancelled', label: '已取消', color: '#ef4444' },
];

// 获取状态配置
export function getStatusConfig(status: ProjectStatus) {
  return STATUS_OPTIONS.find((s) => s.value === status) || STATUS_OPTIONS[0];
}
