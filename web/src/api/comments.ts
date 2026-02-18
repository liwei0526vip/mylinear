/**
 * Comment API
 */

import { api } from './client';
import type {
  Comment,
  CommentListResponse,
  CommentListParams,
  CreateCommentRequest,
  UpdateCommentRequest,
} from '../types/comment';

// =============================================================================
// Comment CRUD API
// =============================================================================

/**
 * 获取 Issue 的评论列表
 */
export async function fetchIssueComments(
  issueId: string,
  params?: CommentListParams
) {
  const searchParams = new URLSearchParams();
  if (params?.page) searchParams.append('page', String(params.page));
  if (params?.page_size) searchParams.append('page_size', String(params.page_size));

  const query = searchParams.toString();
  const endpoint = query
    ? `/issues/${issueId}/comments?${query}`
    : `/issues/${issueId}/comments`;

  return api<CommentListResponse>(endpoint);
}

/**
 * 创建评论
 */
export async function createComment(
  issueId: string,
  data: CreateCommentRequest
) {
  return api<Comment>(`/issues/${issueId}/comments`, {
    method: 'POST',
    body: data,
  });
}

/**
 * 更新评论
 */
export async function updateComment(
  commentId: string,
  data: UpdateCommentRequest
) {
  return api<Comment>(`/comments/${commentId}`, {
    method: 'PUT',
    body: data,
  });
}

/**
 * 删除评论
 */
export async function deleteComment(commentId: string) {
  return api<{ message: string }>(`/comments/${commentId}`, {
    method: 'DELETE',
  });
}
