/**
 * ProjectForm - 项目创建/编辑表单
 */

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { cn } from '@/lib/utils';
import type { Project, CreateProjectRequest, UpdateProjectRequest, ProjectStatus } from '@/types/project';
import { STATUS_OPTIONS } from '@/types/project';

interface ProjectFormProps {
  workspaceId: string;
  project?: Project | null;
  onSubmit: (data: CreateProjectRequest | UpdateProjectRequest) => Promise<void>;
  onCancel: () => void;
  isLoading?: boolean;
  className?: string;
}

export function ProjectForm({
  workspaceId: _workspaceId,
  project,
  onSubmit,
  onCancel,
  isLoading = false,
  className,
}: ProjectFormProps) {
  // workspaceId is used for future workspace-scoped operations
  void _workspaceId;
  const [name, setName] = useState(project?.name ?? '');
  const [description, setDescription] = useState(project?.description ?? '');
  const [leadId, setLeadId] = useState(project?.lead_id ?? '');
  const [startDate, setStartDate] = useState(
    project?.start_date ? project.start_date.split('T')[0] : ''
  );
  const [targetDate, setTargetDate] = useState(
    project?.target_date ? project.target_date.split('T')[0] : ''
  );
  const [status, setStatus] = useState<ProjectStatus>(project?.status ?? 'planned');
  const [error, setError] = useState<string | null>(null);

  // 重置表单当 project 变化
  useEffect(() => {
    if (project) {
      setName(project.name);
      setDescription(project.description ?? '');
      setLeadId(project.lead_id ?? '');
      setStartDate(project.start_date ? project.start_date.split('T')[0] : '');
      setTargetDate(project.target_date ? project.target_date.split('T')[0] : '');
      setStatus(project.status);
    }
  }, [project]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!name.trim()) {
      setError('项目名称不能为空');
      return;
    }

    try {
      const data: CreateProjectRequest | UpdateProjectRequest = {
        name: name.trim(),
        description: description.trim() || undefined,
        lead_id: leadId.trim() || undefined,
        start_date: startDate || undefined,
        target_date: targetDate || undefined,
      };

      // 编辑模式包含状态
      if (project) {
        (data as UpdateProjectRequest).status = status;
      }

      await onSubmit(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : '保存失败');
    }
  };

  return (
    <form onSubmit={handleSubmit} className={cn('space-y-4', className)}>
      {/* 项目名称 */}
      <div className="space-y-2">
        <Label htmlFor="name">项目名称 *</Label>
        <Input
          id="name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="输入项目名称"
          disabled={isLoading}
        />
      </div>

      {/* 项目描述 */}
      <div className="space-y-2">
        <Label htmlFor="description">描述</Label>
        <textarea
          id="description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="输入项目描述（可选）"
          rows={3}
          className="flex w-full rounded-md border border-gray-300 bg-transparent px-3 py-2 text-sm placeholder:text-gray-400 focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500 disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-600"
          disabled={isLoading}
        />
      </div>

      {/* 状态（仅编辑模式） */}
      {project && (
        <div className="space-y-2">
          <Label>状态</Label>
          <div className="flex flex-wrap gap-2">
            {STATUS_OPTIONS.map((option) => (
              <button
                key={option.value}
                type="button"
                onClick={() => setStatus(option.value)}
                disabled={isLoading}
                className={cn(
                  'rounded-full px-3 py-1 text-sm font-medium transition-all',
                  status === option.value
                    ? 'ring-2 ring-indigo-500 ring-offset-2'
                    : 'hover:opacity-80'
                )}
                style={{
                  backgroundColor: `${option.color}20`,
                  color: option.color,
                }}
              >
                {option.label}
              </button>
            ))}
          </div>
        </div>
      )}

      {/* 负责人（预留） */}
      <div className="space-y-2">
        <Label htmlFor="lead">负责人</Label>
        <Input
          id="lead"
          value={leadId}
          onChange={(e) => setLeadId(e.target.value)}
          placeholder="输入负责人 ID（暂不支持选择）"
          disabled={isLoading}
        />
      </div>

      {/* 日期 */}
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="start-date">开始日期</Label>
          <Input
            id="start-date"
            type="date"
            value={startDate}
            onChange={(e) => setStartDate(e.target.value)}
            disabled={isLoading}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="target-date">目标日期</Label>
          <Input
            id="target-date"
            type="date"
            value={targetDate}
            onChange={(e) => setTargetDate(e.target.value)}
            disabled={isLoading}
          />
        </div>
      </div>

      {/* 错误提示 */}
      {error && (
        <div className="rounded-md bg-red-50 p-3 text-sm text-red-600 dark:bg-red-900/20 dark:text-red-400">
          {error}
        </div>
      )}

      {/* 操作按钮 */}
      <div className="flex justify-end gap-3 pt-4">
        <Button type="button" variant="outline" onClick={onCancel} disabled={isLoading}>
          取消
        </Button>
        <Button type="submit" disabled={isLoading}>
          {isLoading ? '保存中...' : project ? '更新项目' : '创建项目'}
        </Button>
      </div>
    </form>
  );
}
