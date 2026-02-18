/**
 * Activity 相关类型定义
 */

// 活动类型
export type ActivityType =
  | 'issue_created'
  | 'title_changed'
  | 'description_changed'
  | 'status_changed'
  | 'priority_changed'
  | 'assignee_changed'
  | 'due_date_changed'
  | 'project_changed'
  | 'labels_changed'
  | 'comment_added';

// 活动状态引用
export interface ActivityStatusRef {
  id: string;
  name?: string;
  color?: string;
}

// 活动用户引用
export interface ActivityUserRef {
  id: string;
  name?: string;
  username?: string;
}

// 活动项目引用
export interface ActivityProjectRef {
  id: string;
  name?: string;
}

// 活动标签引用
export interface ActivityLabelRef {
  id: string;
  name?: string;
  color?: string;
}

// 活动载荷 - 标题变更
export interface ActivityPayloadTitle {
  old_value: string;
  new_value: string;
}

// 活动载荷 - 描述变更
export interface ActivityPayloadDescription {
  old_value?: string;
  new_value?: string;
}

// 活动载荷 - 状态变更
export interface ActivityPayloadStatus {
  old_status?: ActivityStatusRef;
  new_status?: ActivityStatusRef;
}

// 活动载荷 - 优先级变更
export interface ActivityPayloadPriority {
  old_value: number;
  new_value: number;
}

// 活动载荷 - 负责人变更
export interface ActivityPayloadAssignee {
  old_assignee?: ActivityUserRef;
  new_assignee?: ActivityUserRef;
}

// 活动载荷 - 截止日期变更
export interface ActivityPayloadDueDate {
  old_value?: string;
  new_value?: string;
}

// 活动载荷 - 项目变更
export interface ActivityPayloadProject {
  old_project?: ActivityProjectRef;
  new_project?: ActivityProjectRef;
}

// 活动载荷 - 标签变更
export interface ActivityPayloadLabels {
  added: ActivityLabelRef[];
  removed: ActivityLabelRef[];
}

// 活动载荷 - 评论添加
export interface ActivityPayloadComment {
  comment_id: string;
  comment_preview: string;
}

// 活动信息
export interface Activity {
  id: string;
  issue_id: string;
  type: ActivityType;
  actor_id: string;
  payload?: Record<string, unknown>;
  created_at: string;

  // 关联数据
  actor?: {
    id: string;
    name: string;
    username: string;
    email: string;
    avatar_url?: string;
  };
}

// 活动列表响应
export interface ActivityListResponse {
  activities: Activity[];
  total: number;
  page: number;
  page_size: number;
}

// 活动列表查询参数
export interface ActivityListParams {
  page?: number;
  page_size?: number;
  types?: ActivityType[];
}

// 活动类型配置
export const ACTIVITY_TYPE_CONFIG: Record<ActivityType, { label: string; icon: string }> = {
  issue_created: { label: '创建 Issue', icon: 'Plus' },
  title_changed: { label: '修改标题', icon: 'Edit' },
  description_changed: { label: '修改描述', icon: 'FileText' },
  status_changed: { label: '修改状态', icon: 'ArrowRight' },
  priority_changed: { label: '修改优先级', icon: 'Flag' },
  assignee_changed: { label: '修改负责人', icon: 'User' },
  due_date_changed: { label: '修改截止日期', icon: 'Calendar' },
  project_changed: { label: '修改项目', icon: 'Folder' },
  labels_changed: { label: '修改标签', icon: 'Tag' },
  comment_added: { label: '添加评论', icon: 'MessageSquare' },
};
