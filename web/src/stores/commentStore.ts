/**
 * Comment 状态管理
 */

import { create } from 'zustand';
import type {
  Comment,
  CommentListParams,
  CreateCommentRequest,
  UpdateCommentRequest,
} from '../types/comment';
import * as commentApi from '../api/comments';

interface CommentState {
  // 状态
  commentsByIssue: Map<string, Comment[]>; // issueId -> comments
  currentIssueComments: Comment[];
  total: number;
  page: number;
  pageSize: number;
  isLoading: boolean;
  error: string | null;

  // Actions
  fetchIssueComments: (issueId: string, params?: CommentListParams) => Promise<void>;
  createComment: (issueId: string, data: CreateCommentRequest) => Promise<Comment | null>;
  updateComment: (commentId: string, issueId: string, data: UpdateCommentRequest) => Promise<void>;
  deleteComment: (commentId: string, issueId: string) => Promise<void>;

  // Utility
  getCommentsByIssue: (issueId: string) => Comment[];
  clearError: () => void;
  reset: () => void;
}

export const useCommentStore = create<CommentState>((set, get) => ({
  // 初始状态
  commentsByIssue: new Map(),
  currentIssueComments: [],
  total: 0,
  page: 1,
  pageSize: 50,
  isLoading: false,
  error: null,

  // 获取 Issue 的评论列表
  fetchIssueComments: async (issueId: string, params?: CommentListParams) => {
    set({ isLoading: true, error: null });
    const response = await commentApi.fetchIssueComments(issueId, params);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      return;
    }

    const comments = response.data?.comments || [];
    const commentsByIssue = new Map(get().commentsByIssue);
    commentsByIssue.set(issueId, comments);

    set({
      commentsByIssue,
      currentIssueComments: comments,
      total: response.data?.total || 0,
      page: response.data?.page || 1,
      pageSize: response.data?.page_size || 50,
      isLoading: false,
    });
  },

  // 创建评论
  createComment: async (issueId: string, data: CreateCommentRequest) => {
    set({ isLoading: true, error: null });
    const response = await commentApi.createComment(issueId, data);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      throw new Error(response.error);
    }

    const newComment = response.data;
    if (newComment) {
      const commentsByIssue = new Map(get().commentsByIssue);
      const existingComments = commentsByIssue.get(issueId) || [];

      // 如果是回复评论，添加到父评论的 replies 中
      if (newComment.parent_id) {
        const addToReplies = (comments: Comment[]): Comment[] => {
          return comments.map((c) => {
            if (c.id === newComment.parent_id) {
              return {
                ...c,
                replies: [...(c.replies || []), newComment],
              };
            }
            if (c.replies) {
              return {
                ...c,
                replies: addToReplies(c.replies),
              };
            }
            return c;
          });
        };
        commentsByIssue.set(issueId, addToReplies(existingComments));
      } else {
        // 顶级评论，添加到列表开头
        commentsByIssue.set(issueId, [newComment, ...existingComments]);
      }

      set((state) => ({
        commentsByIssue,
        currentIssueComments: commentsByIssue.get(issueId) || [],
        total: state.total + 1,
        isLoading: false,
      }));
    }
    return newComment;
  },

  // 更新评论
  updateComment: async (commentId: string, issueId: string, data: UpdateCommentRequest) => {
    set({ isLoading: true, error: null });
    const response = await commentApi.updateComment(commentId, data);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      throw new Error(response.error);
    }

    const updatedComment = response.data;
    if (updatedComment) {
      const commentsByIssue = new Map(get().commentsByIssue);
      const existingComments = commentsByIssue.get(issueId) || [];

      // 递归更新评论
      const updateInTree = (comments: Comment[]): Comment[] => {
        return comments.map((c) => {
          if (c.id === commentId) {
            return updatedComment;
          }
          if (c.replies) {
            return {
              ...c,
              replies: updateInTree(c.replies),
            };
          }
          return c;
        });
      };

      commentsByIssue.set(issueId, updateInTree(existingComments));
      set({
        commentsByIssue,
        currentIssueComments: commentsByIssue.get(issueId) || [],
        isLoading: false,
      });
    }
  },

  // 删除评论
  deleteComment: async (commentId: string, issueId: string) => {
    set({ isLoading: true, error: null });
    const response = await commentApi.deleteComment(commentId);
    if (response.error) {
      set({ error: response.error, isLoading: false });
      throw new Error(response.error);
    }

    const commentsByIssue = new Map(get().commentsByIssue);
    const existingComments = commentsByIssue.get(issueId) || [];

    // 递归删除评论
    const removeFromTree = (comments: Comment[]): Comment[] => {
      return comments
        .filter((c) => c.id !== commentId)
        .map((c) => {
          if (c.replies) {
            return {
              ...c,
              replies: removeFromTree(c.replies),
            };
          }
          return c;
        });
    };

    commentsByIssue.set(issueId, removeFromTree(existingComments));
    set((state) => ({
      commentsByIssue,
      currentIssueComments: commentsByIssue.get(issueId) || [],
      total: state.total - 1,
      isLoading: false,
    }));
  },

  // 获取指定 Issue 的评论
  getCommentsByIssue: (issueId: string) => {
    return get().commentsByIssue.get(issueId) || [];
  },

  // 清除错误
  clearError: () => set({ error: null }),

  // 重置状态
  reset: () =>
    set({
      commentsByIssue: new Map(),
      currentIssueComments: [],
      total: 0,
      page: 1,
      pageSize: 50,
      error: null,
      isLoading: false,
    }),
}));
