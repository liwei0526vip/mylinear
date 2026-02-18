import { useEffect } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { useAppStore } from '@/stores/app-store'
import { checkHealth } from '@/api/client'
import { ProtectedRoute } from '@/components/ProtectedRoute'
import { MainLayout } from '@/components/layout'
import LoginPage from '@/pages/Login'
import RegisterPage from '@/pages/Register'
import ProfilePage from '@/pages/Profile'
import { WorkspaceSettingsPage } from '@/pages/Settings/Workspace'
import { TeamsPage } from '@/pages/Settings/Teams'
import { TeamDetailPage } from '@/pages/Settings/TeamDetail'
import { ProjectsPage } from '@/pages/ProjectsPage'
import { ProjectDetailPage } from '@/pages/ProjectDetailPage'
import { InboxPage } from '@/pages/inbox/InboxPage'

function HomePage() {
  const { backendStatus, setBackendStatus } = useAppStore()

  useEffect(() => {
    // 检查后端健康状态
    const checkBackendHealth = async () => {
      const { data, error } = await checkHealth()
      if (error || !data || data.status !== 'ok') {
        setBackendStatus('error')
      } else {
        setBackendStatus('ok')
      }
    }

    checkBackendHealth()
  }, [setBackendStatus])

  const getStatusColor = () => {
    switch (backendStatus) {
      case 'ok':
        return 'text-green-500'
      case 'error':
        return 'text-red-500'
      default:
        return 'text-gray-500'
    }
  }

  const getStatusText = () => {
    switch (backendStatus) {
      case 'ok':
        return '后端连接正常'
      case 'error':
        return '后端连接失败'
      default:
        return '正在检查后端连接...'
    }
  }

  return (
    <div className="min-h-screen bg-background flex flex-col items-center justify-center">
      <div className="text-center">
        <h1 className="text-4xl font-bold text-foreground mb-2">
          MyLinear
        </h1>
        <p className="text-muted-foreground mb-8">
          项目管理工具
        </p>

        <div className="flex items-center gap-2">
          <span className={`inline-block w-2 h-2 rounded-full ${
            backendStatus === 'ok' ? 'bg-green-500' :
            backendStatus === 'error' ? 'bg-red-500' :
            'bg-gray-500 animate-pulse'
          }`} />
          <span className={`text-sm ${getStatusColor()}`}>
            {getStatusText()}
          </span>
        </div>
      </div>
    </div>
  )
}

function App() {
  return (
    <Routes>
      {/* 公开路由（不使用侧边栏） */}
      <Route
        path="/login"
        element={
          <MainLayout showSidebar={false}>
            <LoginPage />
          </MainLayout>
        }
      />
      <Route
        path="/register"
        element={
          <MainLayout showSidebar={false}>
            <RegisterPage />
          </MainLayout>
        }
      />

      {/* 受保护路由（使用侧边栏） */}
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <MainLayout>
              <HomePage />
            </MainLayout>
          </ProtectedRoute>
        }
      />
      <Route
        path="/settings/profile"
        element={
          <ProtectedRoute>
            <MainLayout>
              <ProfilePage />
            </MainLayout>
          </ProtectedRoute>
        }
      />
      <Route
        path="/settings/workspace"
        element={
          <ProtectedRoute>
            <MainLayout>
              <WorkspaceSettingsPage />
            </MainLayout>
          </ProtectedRoute>
        }
      />
      <Route
        path="/settings/teams"
        element={
          <ProtectedRoute>
            <MainLayout>
              <TeamsPage />
            </MainLayout>
          </ProtectedRoute>
        }
      />
      <Route
        path="/settings/teams/:teamId"
        element={
          <ProtectedRoute>
            <MainLayout>
              <TeamDetailPage />
            </MainLayout>
          </ProtectedRoute>
        }
      />

      {/* 项目路由 */}
      <Route
        path="/teams/:teamId/projects"
        element={
          <ProtectedRoute>
            <MainLayout>
              <ProjectsPage />
            </MainLayout>
          </ProtectedRoute>
        }
      />
      <Route
        path="/projects/:projectId"
        element={
          <ProtectedRoute>
            <MainLayout>
              <ProjectDetailPage />
            </MainLayout>
          </ProtectedRoute>
        }
      />

      {/* 通知路由 */}
      <Route
        path="/inbox"
        element={
          <ProtectedRoute>
            <MainLayout>
              <InboxPage />
            </MainLayout>
          </ProtectedRoute>
        }
      />

      {/* 默认重定向 */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default App
