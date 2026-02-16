import { create } from 'zustand'

interface AppState {
  // 后端连接状态
  backendStatus: 'unknown' | 'ok' | 'error'
  setBackendStatus: (status: 'unknown' | 'ok' | 'error') => void

  // 主题模式
  theme: 'light' | 'dark'
  toggleTheme: () => void

  // 侧边栏状态
  sidebarCollapsed: boolean
  toggleSidebar: () => void
}

export const useAppStore = create<AppState>((set) => ({
  // 后端连接状态
  backendStatus: 'unknown',
  setBackendStatus: (status) => set({ backendStatus: status }),

  // 主题模式
  theme: 'light',
  toggleTheme: () => set((state) => ({ theme: state.theme === 'light' ? 'dark' : 'light' })),

  // 侧边栏状态
  sidebarCollapsed: false,
  toggleSidebar: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
}))
