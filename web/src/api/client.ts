import { getAccessToken } from '../lib/axios';

const API_BASE = '/api/v1'

interface ApiResponse<T> {
  data: T | null
  error: string | null
}

interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  body?: unknown
  headers?: Record<string, string>
}

/**
 * 发送 API 请求
 */
export async function api<T>(
  endpoint: string,
  options: RequestOptions = {}
): Promise<ApiResponse<T>> {
  const { method = 'GET', body, headers = {} } = options

  const url = `${API_BASE}${endpoint}`
  const token = getAccessToken();

  const config: RequestInit = {
    method,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { 'Authorization': `Bearer ${token}` } : {}),
      ...headers,
    },
  }

  if (body && method !== 'GET') {
    config.body = JSON.stringify(body)
  }

  try {
    const response = await fetch(url, config)

    // 尝试解析 JSON 响应
    let data: T | null = null
    const contentType = response.headers.get('content-type')
    if (contentType && contentType.includes('application/json')) {
      data = await response.json()
    }

    if (!response.ok) {
      const errorMessage = (data as { message?: string })?.message || `HTTP ${response.status}`
      return { data: null, error: errorMessage }
    }

    return { data, error: null }
  } catch (err) {
    const message = err instanceof Error ? err.message : '网络请求失败'
    return { data: null, error: message }
  }
}

/**
 * 健康检查 API
 */
export interface HealthResponse {
  status: 'ok' | 'error'
  message?: string
}

export async function checkHealth(): Promise<ApiResponse<HealthResponse>> {
  return api<HealthResponse>('/health')
}
