/**
 * Comment 相关类型定义
 */

// 评论信息
export interface Comment {
  id: string;
  issue_id: string;
  parent_id?: string;
  user_id: string;
  body: string;
  created_at: string;
  updated_at: string;
  edited_at?: string;

  // 关联数据
  user?: {
    id: string;
    name: string;
    username: string;
    email: string;
    avatar_url?: string;
  };
  replies?: Comment[];
}

// 评论列表响应
export interface CommentListResponse {
  comments: Comment[];
  total: number;
  page: number;
  page_size: number;
}

// 创建评论请求
export interface CreateCommentRequest {
  body: string;
  parent_id?: string;
}

// 更新评论请求
export interface UpdateCommentRequest {
  body: string;
}

// 评论列表查询参数
export interface CommentListParams {
  page?: number;
  page_size?: number;
}
