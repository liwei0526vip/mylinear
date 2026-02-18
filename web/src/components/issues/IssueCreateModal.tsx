/**
 * IssueCreateModal 组件 - 创建 Issue 的模态框
 */

import { useState, useEffect } from 'react';
import { useIssueStore } from '@/stores/issueStore';
import { useWorkflowStore } from '@/stores/workflowStore';
import { PrioritySelector } from './PriorityIcon';
import type { CreateIssueRequest, IssuePriority } from '@/types/issue';

interface IssueCreateModalProps {
  teamId: string;
  isOpen: boolean;
  onClose: () => void;
  onSuccess?: (issueId: string) => void;
}

export function IssueCreateModal({ teamId, isOpen, onClose, onSuccess }: IssueCreateModalProps) {
  const { createIssue, isLoading, error, clearError } = useIssueStore();
  const { states, fetchStates } = useWorkflowStore();

  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [statusId, setStatusId] = useState<string>('');
  const [priority, setPriority] = useState<IssuePriority>(0);

  // 获取默认状态
  useEffect(() => {
    if (isOpen && teamId) {
      fetchStates(teamId);
    }
  }, [isOpen, teamId, fetchStates]);

  // 设置默认状态
  useEffect(() => {
    if (states.length > 0 && !statusId) {
      const backlogStatus = states.find((s) => s.type === 'backlog');
      setStatusId(backlogStatus?.id || states[0]?.id);
    }
  }, [states, statusId]);

  // 重置表单
  const resetForm = () => {
    setTitle('');
    setDescription('');
    setPriority(0);
    clearError();
  };

  // 关闭模态框
  const handleClose = () => {
    resetForm();
    onClose();
  };

  // 提交表单
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!title.trim()) {
      return;
    }

    const data: CreateIssueRequest = {
      title: title.trim(),
      description: description.trim() || undefined,
      status_id: statusId,
      priority,
    };

    try {
      const issue = await createIssue(teamId, data);
      if (issue) {
        resetForm();
        onClose();
        onSuccess?.(issue.id);
      }
    } catch {
      // 错误已经在 store 中处理
    }
  };

  // 快捷键支持
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        handleClose();
      }
    };

    if (isOpen) {
      window.addEventListener('keydown', handleKeyDown);
    }

    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-[15vh]">
      {/* 背景遮罩 */}
      <div
        className="absolute inset-0 bg-black/50"
        onClick={handleClose}
      />

      {/* 模态框 */}
      <div className="relative w-full max-w-lg rounded-lg bg-white shadow-xl dark:bg-gray-800">
        <form onSubmit={handleSubmit}>
          {/* 标题输入 */}
          <div className="p-4">
            <input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Issue 标题"
              className="w-full border-0 bg-transparent text-lg font-medium placeholder:text-gray-400 focus:outline-none focus:ring-0"
              autoFocus
            />
          </div>

          {/* 描述输入 */}
          <div className="px-4 pb-2">
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="添加描述..."
              rows={3}
              className="w-full resize-none border-0 bg-transparent text-sm placeholder:text-gray-400 focus:outline-none focus:ring-0"
            />
          </div>

          {/* 选项区域 */}
          <div className="flex items-center gap-4 border-t border-gray-100 px-4 py-3">
            {/* 状态选择 */}
            <div className="flex items-center gap-2">
              <label className="text-xs text-gray-500">状态</label>
              <select
                value={statusId}
                onChange={(e) => setStatusId(e.target.value)}
                className="h-7 rounded border border-gray-200 bg-white px-2 text-xs focus:outline-none focus:ring-1 focus:ring-indigo-500"
              >
                {states.map((state) => (
                  <option key={state.id} value={state.id}>
                    {state.name}
                  </option>
                ))}
              </select>
            </div>

            {/* 优先级选择 */}
            <div className="flex items-center gap-2">
              <label className="text-xs text-gray-500">优先级</label>
              <PrioritySelector value={priority} onChange={setPriority} />
            </div>
          </div>

          {/* 错误提示 */}
          {error && (
            <div className="px-4 py-2 text-sm text-red-500">
              {error}
            </div>
          )}

          {/* 操作按钮 */}
          <div className="flex items-center justify-end gap-2 border-t border-gray-100 px-4 py-3">
            <button
              type="button"
              onClick={handleClose}
              className="rounded-md px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100"
            >
              取消
            </button>
            <button
              type="submit"
              disabled={!title.trim() || isLoading}
              className="rounded-md bg-indigo-500 px-3 py-1.5 text-sm font-medium text-white hover:bg-indigo-600 disabled:opacity-50"
            >
              {isLoading ? '创建中...' : '创建 Issue'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
