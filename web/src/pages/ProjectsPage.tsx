/**
 * ProjectsPage - 项目列表页
 */

import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { useProjectStore } from '@/stores/projectStore';
import { useTeamStore } from '@/stores/teamStore';
import { useAuthStore } from '@/stores/authStore';
import { ProjectCard, ProjectForm } from '@/components/projects';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import type { ProjectStatus, CreateProjectRequest, UpdateProjectRequest } from '@/types/project';
import { STATUS_OPTIONS, getStatusConfig } from '@/types/project';

export function ProjectsPage() {
  const { teamId } = useParams<{ teamId: string }>();
  const { user } = useAuthStore();
  const { currentTeam, fetchTeam } = useTeamStore();
  const {
    projects,
    isLoading,
    error,
    fetchTeamProjects,
    createProject,
    clearError,
  } = useProjectStore();

  const [statusFilter, setStatusFilter] = useState<ProjectStatus | 'all'>('all');
  const [showCreateModal, setShowCreateModal] = useState(false);

  // 加载团队信息和项目列表
  useEffect(() => {
    if (teamId) {
      fetchTeam(teamId);
      fetchTeamProjects(teamId);
    }
  }, [teamId, fetchTeam, fetchTeamProjects]);

  // 过滤后的项目列表
  const filteredProjects =
    statusFilter === 'all'
      ? projects
      : projects.filter((p) => p.status === statusFilter);

  // 处理创建项目
  const handleCreateProject = async (data: CreateProjectRequest | UpdateProjectRequest) => {
    if (!user?.workspace_id) {
      throw new Error('无法获取工作区信息');
    }

    const createData = data as CreateProjectRequest;
    await createProject(user.workspace_id, {
      ...createData,
      teams: teamId ? [teamId] : undefined,
    });
    setShowCreateModal(false);

    // 刷新项目列表
    if (teamId) {
      fetchTeamProjects(teamId);
    }
  };

  return (
    <div className="p-6">
      {/* 页面标题 */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
            {currentTeam?.name ?? '项目'}列表
          </h1>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            管理团队的所有项目
          </p>
        </div>
        <Button onClick={() => setShowCreateModal(true)}>
          <svg className="mr-2 h-4 w-4" viewBox="0 0 16 16" fill="currentColor">
            <path d="M8 1v14M1 8h14" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
          </svg>
          创建项目
        </Button>
      </div>

      {/* 状态过滤器 */}
      <div className="mb-6 flex flex-wrap gap-2">
        <button
          onClick={() => setStatusFilter('all')}
          className={`rounded-full px-4 py-1.5 text-sm font-medium transition-all ${
            statusFilter === 'all'
              ? 'bg-indigo-500 text-white'
              : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:hover:bg-gray-700'
          }`}
        >
          全部
        </button>
        {STATUS_OPTIONS.map((option) => (
          <button
            key={option.value}
            onClick={() => setStatusFilter(option.value)}
            className={`rounded-full px-4 py-1.5 text-sm font-medium transition-all ${
              statusFilter === option.value
                ? 'ring-2 ring-indigo-500 ring-offset-2'
                : 'hover:opacity-80'
            }`}
            style={{
              backgroundColor:
                statusFilter === option.value ? `${option.color}30` : `${option.color}15`,
              color: option.color,
            }}
          >
            {option.label}
          </button>
        ))}
      </div>

      {/* 错误提示 */}
      {error && (
        <div className="mb-4 rounded-md bg-red-50 p-4 text-sm text-red-600 dark:bg-red-900/20 dark:text-red-400">
          <div className="flex items-center justify-between">
            <span>{error}</span>
            <button onClick={clearError} className="text-red-800 hover:text-red-600">
              <svg className="h-4 w-4" viewBox="0 0 16 16" fill="currentColor">
                <path
                  d="M4 4l8 8M12 4l-8 8"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                />
              </svg>
            </button>
          </div>
        </div>
      )}

      {/* 加载状态 */}
      {isLoading && projects.length === 0 && (
        <div className="flex items-center justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-indigo-500 border-t-transparent" />
        </div>
      )}

      {/* 空状态 */}
      {!isLoading && filteredProjects.length === 0 && (
        <Card className="border-dashed">
          <CardContent className="flex flex-col items-center justify-center py-12">
            <svg
              className="mb-4 h-12 w-12 text-gray-400"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
            >
              <rect x="3" y="3" width="18" height="18" rx="2" strokeWidth="1.5" />
              <path d="M3 9h18M9 21V9" strokeWidth="1.5" />
            </svg>
            <p className="mb-2 text-lg font-medium text-gray-900 dark:text-white">
              暂无项目
            </p>
            <p className="mb-4 text-sm text-gray-500 dark:text-gray-400">
              {statusFilter === 'all'
                ? '创建第一个项目开始管理工作'
                : `没有${getStatusConfig(statusFilter as ProjectStatus).label}状态的项目`}
            </p>
            {statusFilter === 'all' && (
              <Button onClick={() => setShowCreateModal(true)}>创建项目</Button>
            )}
          </CardContent>
        </Card>
      )}

      {/* 项目网格 */}
      {filteredProjects.length > 0 && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {filteredProjects.map((project) => (
            <ProjectCard key={project.id} project={project} />
          ))}
        </div>
      )}

      {/* 创建项目模态框 */}
      {showCreateModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <Card className="w-full max-w-lg mx-4">
            <CardContent className="p-6">
              <h2 className="mb-4 text-xl font-bold text-gray-900 dark:text-white">
                创建新项目
              </h2>
              <ProjectForm
                workspaceId={user?.workspace_id ?? ''}
                onSubmit={handleCreateProject}
                onCancel={() => setShowCreateModal(false)}
                isLoading={isLoading}
              />
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}
