/**
 * Issue 相关类型定义
 */

// Issue 优先级 (0-4, 无优先级/紧急/高/中/低)
export type IssuePriority = 0 | 1 | 2 | 3 | 4;

// Issue 信息
export interface Issue {
  id: string;
  team_id: string;
  number: number;
  title: string;
  description?: string;
  status_id: string;
  priority: IssuePriority;
  assignee_id?: string;
  project_id?: string;
  position: number;
  created_at: string;
  updated_at: string;
  created_by: string;
  completed_at?: string;
  cancelled_at?: string;
  labels?: string[];

  // 关联数据
  status?: {
    id: string;
    name: string;
    type: string;
    color: string;
  };
  assignee?: {
    id: string;
    name: string;
    email: string;
    avatar_url?: string;
  };
  created_by_user?: {
    id: string;
    name: string;
    email: string;
  };
}

// Issue 订阅者
export interface IssueSubscriber {
  id: string;
  username: string;
  name: string;
  email: string;
  avatar_url?: string;
}

// Issue 过滤条件
export interface IssueFilter {
  status_id?: string;
  priority?: number;
  assignee_id?: string;
  project_id?: string;
  cycle_id?: string;
  created_by_id?: string;
}

// Issue 列表响应
export interface IssueListResponse {
  issues: Issue[];
  total: number;
  page: number;
}

// 创建 Issue 请求
export interface CreateIssueRequest {
  title: string;
  description?: string;
  status_id: string;
  priority?: IssuePriority;
  assignee_id?: string;
  project_id?: string;
  labels?: string[];
  due_date?: string;
}

// 更新 Issue 请求
export interface UpdateIssueRequest {
  title?: string;
  description?: string;
  status_id?: string;
  priority?: IssuePriority;
  assignee_id?: string;
  project_id?: string;
}

// 更新位置请求
export interface UpdatePositionRequest {
  position: number;
  status_id?: string;
}

// 订阅者列表响应
export interface SubscriberListResponse {
  subscribers: IssueSubscriber[];
}

// 优先级选项
export const PRIORITY_OPTIONS: { value: IssuePriority; label: string; color: string }[] = [
  { value: 0, label: '无优先级', color: '#9ca3af' },
  { value: 1, label: '紧急', color: '#ef4444' },
  { value: 2, label: '高', color: '#f97316' },
  { value: 3, label: '中', color: '#eab308' },
  { value: 4, label: '低', color: '#6b7280' },
];

// 获取优先级配置
export function getPriorityConfig(priority: IssuePriority) {
  return PRIORITY_OPTIONS.find((p) => p.value === priority) || PRIORITY_OPTIONS[0];
}
