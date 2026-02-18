/**
 * ProjectCard - 项目卡片组件
 */

import { useNavigate } from 'react-router-dom';
import { cn } from '@/lib/utils';
import { Card, CardContent } from '@/components/ui/card';
import { ProjectStatusBadge, ProjectStatusIcon } from './ProjectStatusBadge';
import { ProgressBar } from './ProgressBar';
import type { Project, ProjectProgress } from '@/types/project';

interface ProjectCardProps {
  project: Project;
  progress?: ProjectProgress | null;
  className?: string;
}

export function ProjectCard({ project, progress, className }: ProjectCardProps) {
  const navigate = useNavigate();

  const handleClick = () => {
    navigate(`/projects/${project.id}`);
  };

  const progressPercent = progress?.progress_percent ?? 0;
  const issueCount = progress?.total_issues ?? 0;

  return (
    <Card
      className={cn(
        'cursor-pointer transition-all hover:border-indigo-500/50 hover:shadow-md',
        className
      )}
      onClick={handleClick}
    >
      <CardContent className="p-4">
        {/* 头部：状态图标 + 名称 */}
        <div className="mb-3 flex items-start justify-between">
          <div className="flex items-center gap-2">
            <ProjectStatusIcon status={project.status} size={18} />
            <h3 className="font-medium text-gray-900 dark:text-white">
              {project.name}
            </h3>
          </div>
          <ProjectStatusBadge status={project.status} size="sm" />
        </div>

        {/* 描述 */}
        {project.description && (
          <p className="mb-3 line-clamp-2 text-sm text-gray-600 dark:text-gray-400">
            {project.description}
          </p>
        )}

        {/* 进度条 */}
        <div className="mb-3">
          <ProgressBar percent={progressPercent} size="md" showLabel />
        </div>

        {/* 底部信息 */}
        <div className="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
          <div className="flex items-center gap-3">
            {/* Issue 数量 */}
            <span className="flex items-center gap-1">
              <svg className="h-4 w-4" viewBox="0 0 16 16" fill="currentColor">
                <circle cx="8" cy="8" r="6" stroke="currentColor" strokeWidth="1.5" fill="none" />
                <circle cx="8" cy="8" r="2" fill="currentColor" />
              </svg>
              {issueCount} 个任务
            </span>

            {/* 负责人 */}
            {project.lead && (
              <span className="flex items-center gap-1">
                <svg className="h-4 w-4" viewBox="0 0 16 16" fill="currentColor">
                  <circle cx="8" cy="6" r="3" />
                  <path d="M3 14c0-2.76 2.24-5 5-5s5 2.24 5 5" />
                </svg>
                {project.lead.name}
              </span>
            )}
          </div>

          {/* 目标日期 */}
          {project.target_date && (
            <span className="flex items-center gap-1">
              <svg className="h-4 w-4" viewBox="0 0 16 16" fill="currentColor">
                <rect x="2" y="3" width="12" height="11" rx="2" stroke="currentColor" strokeWidth="1.5" fill="none" />
                <path d="M2 6h12M5 1v2M11 1v2" stroke="currentColor" strokeWidth="1.5" />
              </svg>
              {new Date(project.target_date).toLocaleDateString('zh-CN')}
            </span>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
