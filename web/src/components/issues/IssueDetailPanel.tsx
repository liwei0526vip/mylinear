/**
 * IssueDetailPanel 组件 - 右侧 Issue 详情面板
 * 支持全屏模式切换和评论/活动 Tab
 */

import { useEffect, useState, useCallback } from 'react';
import { useIssueStore } from '@/stores/issueStore';
import { useWorkflowStore } from '@/stores/workflowStore';
import { useAuthStore } from '@/stores/authStore';
import { PrioritySelector } from './PriorityIcon';
import { CommentSection } from '@/components/comments';
import { ActivityTimeline } from '@/components/activities';
import type { Issue, UpdateIssueRequest, IssuePriority } from '@/types/issue';

// Tab 类型
type DetailTab = 'comments' | 'activity';

interface IssueDetailPanelProps {
  issueId: string;
  onClose: () => void;
  defaultFullscreen?: boolean;
  defaultTab?: DetailTab;
}

export function IssueDetailPanel({ issueId, onClose, defaultFullscreen = false, defaultTab = 'comments' }: IssueDetailPanelProps) {
  const {
    currentIssue,
    fetchIssue,
    updateIssue,
    deleteIssue,
    subscribe,
    fetchSubscribers,
    subscribers,
    isLoading,
    error,
  } = useIssueStore();

  const { states, fetchStates } = useWorkflowStore();
  const { user } = useAuthStore();

  const [isEditing, setIsEditing] = useState(false);
  const [editTitle, setEditTitle] = useState('');
  const [editDescription, setEditDescription] = useState('');
  const [isFullscreen, setIsFullscreen] = useState(defaultFullscreen);
  const [activeTab, setActiveTab] = useState<DetailTab>(defaultTab);

  // 获取 Issue 详情
  useEffect(() => {
    if (issueId) {
      fetchIssue(issueId);
      fetchSubscribers(issueId);
    }
  }, [issueId, fetchIssue, fetchSubscribers]);

  // 获取工作流状态
  useEffect(() => {
    if (currentIssue?.team_id) {
      fetchStates(currentIssue.team_id);
    }
  }, [currentIssue?.team_id, fetchStates]);

  // 同步编辑状态
  useEffect(() => {
    if (currentIssue) {
      setEditTitle(currentIssue.title);
      setEditDescription(currentIssue.description || '');
    }
  }, [currentIssue]);

  // 全屏模式快捷键
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // F 切换全屏
      if (e.key === 'f' && !isEditing) {
        e.preventDefault();
        setIsFullscreen((prev) => !prev);
      }
      // Escape 退出全屏或关闭面板
      if (e.key === 'Escape') {
        if (isFullscreen) {
          setIsFullscreen(false);
        } else {
          onClose();
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isFullscreen, isEditing, onClose]);

  // 切换全屏
  const toggleFullscreen = useCallback(() => {
    setIsFullscreen((prev) => !prev);
  }, []);

  // 更新 Issue
  const handleUpdate = async (data: Partial<UpdateIssueRequest>) => {
    if (!currentIssue) return;
    try {
      await updateIssue(currentIssue.id, data);
    } catch {
      // 错误已在 store 中处理
    }
  };

  // 保存标题和描述
  const handleSaveEdit = async () => {
    if (!currentIssue) return;
    if (editTitle !== currentIssue.title || editDescription !== (currentIssue.description || '')) {
      await handleUpdate({
        title: editTitle,
        description: editDescription || undefined,
      });
    }
    setIsEditing(false);
  };

  // 删除 Issue
  const handleDelete = async () => {
    if (!currentIssue) return;
    if (window.confirm('确定要删除这个 Issue 吗？')) {
      try {
        await deleteIssue(currentIssue.id);
        onClose();
      } catch {
        // 错误已在 store 中处理
      }
    }
  };

  // 订阅/取消订阅
  const handleToggleSubscribe = async () => {
    if (!currentIssue) return;
    // 简化处理：假设当前用户未订阅
    try {
      await subscribe(currentIssue.id);
      fetchSubscribers(currentIssue.id);
    } catch {
      // 错误已在 store 中处理
    }
  };

  if (!currentIssue && !isLoading) {
    return (
      <div className={`flex items-center justify-center bg-white ${isFullscreen ? 'fixed inset-0 z-50' : 'h-full w-96 border-l border-gray-200'}`}>
        <p className="text-gray-500">Issue 不存在</p>
      </div>
    );
  }

  if (isLoading && !currentIssue) {
    return (
      <div className={`flex items-center justify-center bg-white ${isFullscreen ? 'fixed inset-0 z-50' : 'h-full w-96 border-l border-gray-200'}`}>
        <p className="text-gray-500">加载中...</p>
      </div>
    );
  }

  const issue = currentIssue as Issue;

  // 全屏模式和侧边栏模式的样式
  const containerClass = isFullscreen
    ? 'fixed inset-0 z-50 flex flex-col bg-white'
    : 'flex h-full w-96 flex-col border-l border-gray-200 bg-white';

  return (
    <div className={containerClass}>
      {/* 头部 */}
      <div className="flex items-center justify-between border-b border-gray-100 px-4 py-3">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium text-gray-500">
            {issue.team_id}-{issue.number}
          </span>
        </div>
        <div className="flex items-center gap-1">
          {/* 全屏切换按钮 */}
          <button
            onClick={toggleFullscreen}
            className="rounded p-1 hover:bg-gray-100"
            title={isFullscreen ? '退出全屏 (F)' : '全屏 (F)'}
          >
            {isFullscreen ? (
              <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                <path d="M5.5 2a.5.5 0 0 1 0 1H3.707l2.147 2.146a.5.5 0 0 1-.708.708L3 3.707V5.5a.5.5 0 0 1-1 0v-3a.5.5 0 0 1 .5-.5h3zm6.5 0a.5.5 0 0 1 .5.5v3a.5.5 0 0 1-1 0V3.707L9.354 5.854a.5.5 0 1 1-.708-.708L10.793 3H9.5a.5.5 0 0 1 0-1h3zM5.5 14a.5.5 0 0 0 0-1H3.707l2.147-2.146a.5.5 0 0 0-.708-.708L3 12.293V10.5a.5.5 0 0 0-1 0v3a.5.5 0 0 0 .5.5h3zm6.5 0a.5.5 0 0 0 .5-.5v-3a.5.5 0 0 0-1 0v1.793l-2.146-2.147a.5.5 0 0 0-.708.708L10.793 13H9.5a.5.5 0 0 0 0 1h3z" />
              </svg>
            ) : (
              <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                <path d="M1.5 1a.5.5 0 0 0-.5.5v3a.5.5 0 0 0 1 0V2.707l2.146 2.147a.5.5 0 1 0 .708-.708L2.707 2H4.5a.5.5 0 0 0 0-1h-3zm11 0a.5.5 0 0 1 .5.5v3a.5.5 0 0 1-1 0V2.707l-2.146 2.147a.5.5 0 1 1-.708-.708L11.293 2H9.5a.5.5 0 0 1 0-1h3zM1 12.5a.5.5 0 0 1 .5-.5h1.793l-2.146-2.146a.5.5 0 1 1 .708-.708L4 11.293V9.5a.5.5 0 0 1 1 0v3a.5.5 0 0 1-.5.5h-3a.5.5 0 0 1-.5-.5zm14 0a.5.5 0 0 1-.5.5h-3a.5.5 0 0 1 0-1h1.793l-2.146-2.146a.5.5 0 1 1 .708-.708L14 11.293V9.5a.5.5 0 0 1 1 0v3z" />
              </svg>
            )}
          </button>
          {/* 关闭按钮 */}
          <button
            onClick={onClose}
            className="rounded p-1 hover:bg-gray-100"
            title="关闭 (Esc)"
          >
            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
              <path d="M4.646 4.646a.5.5 0 0 1 .708 0L8 7.293l2.646-2.647a.5.5 0 0 1 .708.708L8.707 8l2.647 2.646a.5.5 0 0 1-.708.708L8 8.707l-2.646 2.647a.5.5 0 0 1-.708-.708L7.293 8 4.646 5.354a.5.5 0 0 1 0-.708z" />
            </svg>
          </button>
        </div>
      </div>

      {/* 内容区 */}
      <div className="flex-1 overflow-y-auto">
        {/* 标题和属性区 */}
        <div className={`${isFullscreen ? 'mx-auto max-w-3xl px-6' : 'px-4'} py-4`}>
          {/* 标题 */}
          {isEditing ? (
            <div className="mb-4">
              <input
                type="text"
                value={editTitle}
                onChange={(e) => setEditTitle(e.target.value)}
                className="w-full rounded border border-gray-300 px-2 py-1 text-lg font-medium focus:outline-none focus:ring-2 focus:ring-indigo-500"
                autoFocus
              />
              <textarea
                value={editDescription}
                onChange={(e) => setEditDescription(e.target.value)}
                placeholder="添加描述..."
                rows={4}
                className="mt-2 w-full rounded border border-gray-300 px-2 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
              />
              <div className="mt-2 flex gap-2">
                <button
                  onClick={handleSaveEdit}
                  className="rounded bg-indigo-500 px-3 py-1 text-sm text-white hover:bg-indigo-600"
                >
                  保存
                </button>
                <button
                  onClick={() => setIsEditing(false)}
                  className="rounded px-3 py-1 text-sm text-gray-600 hover:bg-gray-100"
                >
                  取消
                </button>
              </div>
            </div>
          ) : (
            <div className="mb-4">
              <h2
                className="cursor-pointer text-lg font-medium hover:bg-gray-50"
                onClick={() => setIsEditing(true)}
              >
                {issue.title}
              </h2>
              {issue.description && (
                <p className="mt-2 text-sm text-gray-600">{issue.description}</p>
              )}
            </div>
          )}

          {/* 属性区 */}
          <div className="space-y-3 border-t border-gray-100 pt-4">
            {/* 状态 */}
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-500">状态</span>
              <select
                value={issue.status_id}
                onChange={(e) => handleUpdate({ status_id: e.target.value })}
                className="h-7 rounded border border-gray-200 bg-white px-2 text-xs focus:outline-none focus:ring-1 focus:ring-indigo-500"
              >
                {states.map((s) => (
                  <option key={s.id} value={s.id}>
                    {s.name}
                  </option>
                ))}
              </select>
            </div>

            {/* 优先级 */}
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-500">优先级</span>
              <PrioritySelector
                value={issue.priority}
                onChange={(p: IssuePriority) => handleUpdate({ priority: p })}
              />
            </div>

            {/* 创建者 */}
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-500">创建者</span>
              <span className="text-sm">{issue.created_by_user?.name || '未知'}</span>
            </div>

            {/* 创建时间 */}
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-500">创建时间</span>
              <span className="text-sm text-gray-600">
                {new Date(issue.created_at).toLocaleDateString('zh-CN')}
              </span>
            </div>
          </div>

          {/* 订阅者 */}
          <div className="mt-4 border-t border-gray-100 pt-4">
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-500">订阅者 ({subscribers.length})</span>
              <button
                onClick={handleToggleSubscribe}
                className="text-xs text-indigo-500 hover:text-indigo-600"
              >
                订阅
              </button>
            </div>
            <div className="mt-2 flex flex-wrap gap-2">
              {subscribers.map((sub) => (
                <div
                  key={sub.id}
                  className="flex items-center gap-1 rounded-full bg-gray-100 px-2 py-1"
                >
                  <div className="h-5 w-5 rounded-full bg-gray-300" />
                  <span className="text-xs">{sub.name}</span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Tab 切换 */}
        <div className="border-t border-gray-100">
          <div className="flex">
            <button
              onClick={() => setActiveTab('comments')}
              className={`flex-1 px-4 py-2 text-sm font-medium transition-colors ${
                activeTab === 'comments'
                  ? 'border-b-2 border-indigo-500 text-indigo-600'
                  : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              评论
            </button>
            <button
              onClick={() => setActiveTab('activity')}
              className={`flex-1 px-4 py-2 text-sm font-medium transition-colors ${
                activeTab === 'activity'
                  ? 'border-b-2 border-indigo-500 text-indigo-600'
                  : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              活动
            </button>
          </div>
        </div>

        {/* Tab 内容 */}
        <div className="flex-1">
          {activeTab === 'comments' ? (
            <CommentSection
              issueId={issueId}
              currentUserId={user?.id}
            />
          ) : (
            <ActivityTimeline
              issueId={issueId}
              showFilter={true}
            />
          )}
        </div>
      </div>

      {/* 错误提示 */}
      {error && (
        <div className="border-t border-gray-100 px-4 py-2 text-sm text-red-500">
          {error}
        </div>
      )}

      {/* 底部操作 */}
      <div className="border-t border-gray-100 px-4 py-3">
        <button
          onClick={handleDelete}
          className="text-sm text-red-500 hover:text-red-600"
        >
          删除 Issue
        </button>
      </div>
    </div>
  );
}
