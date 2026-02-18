/**
 * CommentList 组件 - 评论列表（递归渲染嵌套回复）
 */

import type { Comment } from '@/types/comment';
import { CommentItem } from './CommentItem';

interface CommentListProps {
  comments: Comment[];
  onReply?: (comment: Comment) => void;
  onEdit?: (comment: Comment) => void;
  onDelete?: (comment: Comment) => void;
  currentUserId?: string;
  replyingToId?: string | null;
  editCommentId?: string | null;
  renderReplyInput?: (parentId: string) => React.ReactNode;
  renderEditInput?: (comment: Comment) => React.ReactNode;
}

// 递归渲染评论树
function CommentTree({
  comments,
  onReply,
  onEdit,
  onDelete,
  currentUserId,
  replyingToId,
  editCommentId,
  renderReplyInput,
  renderEditInput,
  depth = 0,
}: CommentListProps & { depth?: number }) {
  if (!comments || comments.length === 0) {
    return null;
  }

  return (
    <div className={depth > 0 ? 'ml-10' : ''}>
      {comments.map((comment) => (
        <div key={comment.id}>
          {/* 编辑模式 */}
          {editCommentId === comment.id && renderEditInput ? (
            <div className={depth > 0 ? 'ml-10 mt-2' : 'mt-4'}>
              {renderEditInput(comment)}
            </div>
          ) : (
            <CommentItem
              comment={comment}
              onReply={onReply}
              onEdit={onEdit}
              onDelete={onDelete}
              isReply={depth > 0}
              currentUserId={currentUserId}
            />
          )}

          {/* 回复输入框 */}
          {replyingToId === comment.id && renderReplyInput && (
            <div className="ml-10 mt-2">
              {renderReplyInput(comment.id)}
            </div>
          )}

          {/* 递归渲染子评论 */}
          {comment.replies && comment.replies.length > 0 && (
            <CommentTree
              comments={comment.replies}
              onReply={onReply}
              onEdit={onEdit}
              onDelete={onDelete}
              currentUserId={currentUserId}
              replyingToId={replyingToId}
              editCommentId={editCommentId}
              renderReplyInput={renderReplyInput}
              renderEditInput={renderEditInput}
              depth={depth + 1}
            />
          )}
        </div>
      ))}
    </div>
  );
}

export function CommentList({
  comments,
  onReply,
  onEdit,
  onDelete,
  currentUserId,
  replyingToId,
  editCommentId,
  renderReplyInput,
  renderEditInput,
}: CommentListProps) {
  if (!comments || comments.length === 0) {
    return (
      <div className="py-8 text-center text-sm text-gray-400">
        暂无评论，来添加第一条评论吧
      </div>
    );
  }

  return (
    <div className="py-2">
      <CommentTree
        comments={comments}
        onReply={onReply}
        onEdit={onEdit}
        onDelete={onDelete}
        currentUserId={currentUserId}
        replyingToId={replyingToId}
        editCommentId={editCommentId}
        renderReplyInput={renderReplyInput}
        renderEditInput={renderEditInput}
      />
    </div>
  );
}
