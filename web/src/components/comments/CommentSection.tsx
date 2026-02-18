/**
 * CommentSection 组件 - 评论区容器
 */

import { useEffect, useState } from 'react';
import { useCommentStore } from '@/stores/commentStore';
import { CommentList } from './CommentList';
import { CommentInput } from './CommentInput';
import type { Comment } from '@/types/comment';

interface CommentSectionProps {
  issueId: string;
  currentUserId?: string;
}

export function CommentSection({ issueId, currentUserId }: CommentSectionProps) {
  const {
    currentIssueComments,
    fetchIssueComments,
    createComment,
    updateComment,
    deleteComment,
    isLoading,
    error,
  } = useCommentStore();

  const [replyingTo, setReplyingTo] = useState<Comment | null>(null);
  const [editingComment, setEditingComment] = useState<Comment | null>(null);

  // 获取评论列表
  useEffect(() => {
    if (issueId) {
      fetchIssueComments(issueId);
    }
  }, [issueId, fetchIssueComments]);

  // 创建新评论
  const handleCreateComment = async (body: string) => {
    await createComment(issueId, { body });
    await fetchIssueComments(issueId);
  };

  // 创建回复
  const handleCreateReply = async (body: string) => {
    if (!replyingTo) return;
    await createComment(issueId, { body, parent_id: replyingTo.id });
    setReplyingTo(null);
    await fetchIssueComments(issueId);
  };

  // 更新评论
  const handleUpdateComment = async (body: string) => {
    if (!editingComment) return;
    await updateComment(editingComment.id, issueId, { body });
    setEditingComment(null);
    await fetchIssueComments(issueId);
  };

  // 删除评论
  const handleDeleteComment = async (comment: Comment) => {
    if (!window.confirm('确定要删除这条评论吗？')) return;
    await deleteComment(comment.id, issueId);
    await fetchIssueComments(issueId);
  };

  // 回复评论
  const handleReply = (comment: Comment) => {
    setReplyingTo(comment);
    setEditingComment(null);
  };

  // 编辑评论
  const handleEdit = (comment: Comment) => {
    setEditingComment(comment);
    setReplyingTo(null);
  };

  // 取消回复
  const handleCancelReply = () => {
    setReplyingTo(null);
  };

  // 取消编辑
  const handleCancelEdit = () => {
    setEditingComment(null);
  };

  return (
    <div className="flex flex-col">
      {/* 标题 */}
      <div className="flex items-center justify-between border-b border-gray-100 px-4 py-2">
        <h3 className="text-sm font-medium text-gray-700">
          评论
          {currentIssueComments.length > 0 && (
            <span className="ml-1 text-gray-400">({currentIssueComments.length})</span>
          )}
        </h3>
      </div>

      {/* 评论输入框 */}
      <div className="border-b border-gray-100 p-4">
        <CommentInput
          onSubmit={handleCreateComment}
          placeholder="添加评论...（支持 @mention 提及团队成员）"
          submitText="评论"
        />
      </div>

      {/* 错误提示 */}
      {error && (
        <div className="mx-4 my-2 rounded bg-red-50 px-3 py-2 text-sm text-red-500">
          {error}
        </div>
      )}

      {/* 评论列表 */}
      <div className="flex-1 overflow-y-auto px-4">
        {isLoading && currentIssueComments.length === 0 ? (
          <div className="py-8 text-center text-sm text-gray-400">加载中...</div>
        ) : (
          <CommentList
            comments={currentIssueComments}
            onReply={handleReply}
            onEdit={handleEdit}
            onDelete={handleDeleteComment}
            currentUserId={currentUserId}
            replyingToId={replyingTo?.id}
            editCommentId={editingComment?.id}
            renderReplyInput={(parentId) => (
              <div className="mt-2">
                <div className="mb-2 text-xs text-gray-500">
                  回复给 <span className="font-medium">{replyingTo?.user?.name || '用户'}</span>
                </div>
                <CommentInput
                  onSubmit={handleCreateReply}
                  onCancel={handleCancelReply}
                  placeholder={`回复给 ${replyingTo?.user?.name || '用户'}...`}
                  submitText="回复"
                  isReply
                />
              </div>
            )}
            renderEditInput={(comment) => (
              <div className="mt-2">
                <CommentInput
                  onSubmit={handleUpdateComment}
                  onCancel={handleCancelEdit}
                  initialValue={comment.body}
                  placeholder="编辑评论..."
                  submitText="保存"
                  isReply
                />
              </div>
            )}
          />
        )}
      </div>
    </div>
  );
}
