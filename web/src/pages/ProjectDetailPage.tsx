/**
 * ProjectDetailPage - 项目详情页
 */

import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useProjectStore } from '@/stores/projectStore';
import { ProjectStatusIcon, ProgressBar, ProjectForm } from '@/components/projects';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import type { UpdateProjectRequest, ProjectStatus } from '@/types/project';
import { STATUS_OPTIONS } from '@/types/project';

export function ProjectDetailPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const navigate = useNavigate();

  const {
    currentProject,
    progress,
    projectIssues,
    isLoading,
    error,
    fetchProject,
    fetchProjectProgress,
    fetchProjectIssues,
    updateProject,
    deleteProject,
    clearError,
  } = useProjectStore();

  const [showEditModal, setShowEditModal] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  // 加载项目详情
  useEffect(() => {
    if (projectId) {
      fetchProject(projectId);
      fetchProjectProgress(projectId);
      fetchProjectIssues(projectId);
    }
  }, [projectId, fetchProject, fetchProjectProgress, fetchProjectIssues]);

  // 更新项目
  const handleUpdateProject = async (data: UpdateProjectRequest) => {
    if (!projectId) return;
    await updateProject(projectId, data);
    setShowEditModal(false);
  };

  // 删除项目
  const handleDeleteProject = async () => {
    if (!projectId) return;
    try {
      await deleteProject(projectId);
      navigate(-1);
    } catch (err) {
      // 错误已在 store 中处理
    }
  };

  // 状态切换
  const handleStatusChange = async (status: ProjectStatus) => {
    if (!projectId || !currentProject) return;
    await updateProject(projectId, { status });
  };

  if (isLoading && !currentProject) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-indigo-500 border-t-transparent" />
      </div>
    );
  }

  if (!currentProject) {
    return (
      <div className="p-6">
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <p className="text-lg text-gray-500">项目不存在或已被删除</p>
            <Button className="mt-4" onClick={() => navigate(-1)}>
              返回
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="p-6">
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

      {/* 项目头部 */}
      <Card className="mb-6">
        <CardContent className="p-6">
          <div className="flex items-start justify-between">
            <div className="flex items-center gap-3">
              <ProjectStatusIcon status={currentProject.status} size={24} />
              <div>
                <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
                  {currentProject.name}
                </h1>
                {currentProject.description && (
                  <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                    {currentProject.description}
                  </p>
                )}
              </div>
            </div>
            <div className="flex gap-2">
              <Button variant="outline" onClick={() => setShowEditModal(true)}>
                编辑
              </Button>
              <Button variant="destructive" onClick={() => setShowDeleteConfirm(true)}>
                删除
              </Button>
            </div>
          </div>

          {/* 状态切换 */}
          <div className="mt-4">
            <p className="mb-2 text-sm font-medium text-gray-700 dark:text-gray-300">状态</p>
            <div className="flex flex-wrap gap-2">
              {STATUS_OPTIONS.map((option) => (
                <button
                  key={option.value}
                  onClick={() => handleStatusChange(option.value)}
                  disabled={isLoading}
                  className={`rounded-full px-4 py-1.5 text-sm font-medium transition-all ${
                    currentProject.status === option.value
                      ? 'ring-2 ring-indigo-500 ring-offset-2'
                      : 'hover:opacity-80'
                  }`}
                  style={{
                    backgroundColor:
                      currentProject.status === option.value
                        ? `${option.color}30`
                        : `${option.color}15`,
                    color: option.color,
                  }}
                >
                  {option.label}
                </button>
              ))}
            </div>
          </div>

          {/* 进度 */}
          {progress && (
            <div className="mt-6">
              <div className="mb-2 flex items-center justify-between">
                <p className="text-sm font-medium text-gray-700 dark:text-gray-300">进度</p>
                <span className="text-sm text-gray-500">
                  {progress.completed_issues} / {progress.total_issues} 任务完成
                </span>
              </div>
              <ProgressBar percent={progress.progress_percent} size="lg" showLabel />
            </div>
          )}

          {/* 元信息 */}
          <div className="mt-6 grid grid-cols-2 gap-4 text-sm sm:grid-cols-4">
            {currentProject.lead && (
              <div>
                <p className="text-gray-500 dark:text-gray-400">负责人</p>
                <p className="font-medium text-gray-900 dark:text-white">
                  {currentProject.lead.name}
                </p>
              </div>
            )}
            {currentProject.start_date && (
              <div>
                <p className="text-gray-500 dark:text-gray-400">开始日期</p>
                <p className="font-medium text-gray-900 dark:text-white">
                  {new Date(currentProject.start_date).toLocaleDateString('zh-CN')}
                </p>
              </div>
            )}
            {currentProject.target_date && (
              <div>
                <p className="text-gray-500 dark:text-gray-400">目标日期</p>
                <p className="font-medium text-gray-900 dark:text-white">
                  {new Date(currentProject.target_date).toLocaleDateString('zh-CN')}
                </p>
              </div>
            )}
            <div>
              <p className="text-gray-500 dark:text-gray-400">创建时间</p>
              <p className="font-medium text-gray-900 dark:text-white">
                {new Date(currentProject.created_at).toLocaleDateString('zh-CN')}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 关联 Issue 列表 */}
      <Card>
        <CardHeader>
          <CardTitle>关联任务</CardTitle>
        </CardHeader>
        <CardContent>
          {projectIssues.items.length === 0 ? (
            <div className="py-8 text-center text-gray-500">
              <p>暂无关联任务</p>
              <p className="mt-1 text-sm">将 Issue 分配到此项目后将在此显示</p>
            </div>
          ) : (
            <div className="divide-y divide-gray-200 dark:divide-gray-700">
              {projectIssues.items.map((issue) => (
                <div
                  key={issue.id}
                  className="flex cursor-pointer items-center justify-between py-3 hover:bg-gray-50 dark:hover:bg-gray-800"
                  onClick={() => navigate(`/issues/${issue.id}`)}
                >
                  <div className="flex items-center gap-3">
                    <span className="text-sm text-gray-500">#{issue.number}</span>
                    <span className="font-medium text-gray-900 dark:text-white">
                      {issue.title}
                    </span>
                  </div>
                  <div className="flex items-center gap-2">
                    {issue.completed_at && (
                      <span className="text-xs text-green-600">已完成</span>
                    )}
                    <svg
                      className="h-4 w-4 text-gray-400"
                      viewBox="0 0 16 16"
                      fill="currentColor"
                    >
                      <path d="M6 4l4 4-4 4" stroke="currentColor" strokeWidth="2" fill="none" />
                    </svg>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* 编辑模态框 */}
      {showEditModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <Card className="w-full max-w-lg mx-4">
            <CardContent className="p-6">
              <h2 className="mb-4 text-xl font-bold text-gray-900 dark:text-white">
                编辑项目
              </h2>
              <ProjectForm
                workspaceId={currentProject.workspace_id}
                project={currentProject}
                onSubmit={handleUpdateProject}
                onCancel={() => setShowEditModal(false)}
                isLoading={isLoading}
              />
            </CardContent>
          </Card>
        </div>
      )}

      {/* 删除确认对话框 */}
      {showDeleteConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <Card className="w-full max-w-md mx-4">
            <CardContent className="p-6">
              <h2 className="mb-4 text-xl font-bold text-gray-900 dark:text-white">
                确认删除
              </h2>
              <p className="mb-6 text-gray-600 dark:text-gray-400">
                确定要删除项目 "{currentProject.name}" 吗？此操作不可撤销。
              </p>
              <div className="flex justify-end gap-3">
                <Button variant="outline" onClick={() => setShowDeleteConfirm(false)}>
                  取消
                </Button>
                <Button variant="destructive" onClick={handleDeleteProject} disabled={isLoading}>
                  {isLoading ? '删除中...' : '确认删除'}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}
