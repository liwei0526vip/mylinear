/**
 * CommentItem 组件 - 单条评论
 */

import { useState } from 'react';
import type { Comment } from '@/types/comment';

interface CommentItemProps {
  comment: Comment;
  onReply?: (comment: Comment) => void;
  onEdit?: (comment: Comment) => void;
  onDelete?: (comment: Comment) => void;
  isReply?: boolean;
  currentUserId?: string;
}

// 格式化相对时间
function formatRelativeTime(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSec = Math.floor(diffMs / 1000);
  const diffMin = Math.floor(diffSec / 60);
  const diffHour = Math.floor(diffMin / 60);
  const diffDay = Math.floor(diffHour / 24);

  if (diffSec < 60) return '刚刚';
  if (diffMin < 60) return `${diffMin} 分钟前`;
  if (diffHour < 24) return `${diffHour} 小时前`;
  if (diffDay < 7) return `${diffDay} 天前`;

  return date.toLocaleDateString('zh-CN', {
    month: 'short',
    day: 'numeric',
  });
}

// 获取用户头像颜色
function getAvatarColor(name: string): string {
  const colors = [
    'bg-red-500',
    'bg-orange-500',
    'bg-amber-500',
    'bg-yellow-500',
    'bg-lime-500',
    'bg-green-500',
    'bg-emerald-500',
    'bg-teal-500',
    'bg-cyan-500',
    'bg-sky-500',
    'bg-blue-500',
    'bg-indigo-500',
    'bg-violet-500',
    'bg-purple-500',
    'bg-fuchsia-500',
    'bg-pink-500',
  ];
  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash);
  }
  return colors[Math.abs(hash) % colors.length];
}

export function CommentItem({
  comment,
  onReply,
  onEdit,
  onDelete,
  isReply = false,
  currentUserId,
}: CommentItemProps) {
  const [showActions, setShowActions] = useState(false);
  const isOwner = currentUserId === comment.user_id;
  const userName = comment.user?.name || '未知用户';
  const avatarColor = getAvatarColor(userName);

  return (
    <div
      className={`group relative ${isReply ? 'ml-10 mt-2' : 'mt-4'}`}
      onMouseEnter={() => setShowActions(true)}
      onMouseLeave={() => setShowActions(false)}
    >
      <div className="flex gap-3">
        {/* 头像 */}
        <div
          className={`flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full text-xs font-medium text-white ${avatarColor}`}
        >
          {userName.charAt(0).toUpperCase()}
        </div>

        {/* 内容区 */}
        <div className="flex-1 min-w-0">
          {/* 用户名和时间 */}
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium text-gray-900">{userName}</span>
            <span className="text-xs text-gray-400">
              {formatRelativeTime(comment.created_at)}
            </span>
            {comment.edited_at && (
              <span className="text-xs text-gray-400">(已编辑)</span>
            )}
          </div>

          {/* 评论内容 */}
          <div className="mt-1 text-sm text-gray-700 whitespace-pre-wrap break-words">
            {comment.body}
          </div>

          {/* 操作按钮 */}
          {showActions && (
            <div className="mt-1 flex items-center gap-3">
              {!isReply && onReply && (
                <button
                  onClick={() => onReply(comment)}
                  className="text-xs text-gray-400 hover:text-gray-600"
                >
                  回复
                </button>
              )}
              {isOwner && onEdit && (
                <button
                  onClick={() => onEdit(comment)}
                  className="text-xs text-gray-400 hover:text-gray-600"
                >
                  编辑
                </button>
              )}
              {(isOwner || onDelete) && onDelete && (
                <button
                  onClick={() => onDelete(comment)}
                  className="text-xs text-gray-400 hover:text-red-500"
                >
                  删除
                </button>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
