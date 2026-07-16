import axios from 'axios'

const TOKEN_KEY = 'itcfg_token'

// Token 管理
export const getToken = (): string | null => localStorage.getItem(TOKEN_KEY)
export const setToken = (token: string) => localStorage.setItem(TOKEN_KEY, token)
export const removeToken = () => localStorage.removeItem(TOKEN_KEY)
export const isLoggedIn = (): boolean => !!getToken()
export const getUser = () => {
  try {
    const raw = localStorage.getItem('itcfg_user')
    return raw ? JSON.parse(raw) : null
  } catch {
    return null
  }
}
export const setUser = (user: any) => localStorage.setItem('itcfg_user', JSON.stringify(user))
export const removeUser = () => localStorage.removeItem('itcfg_user')

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
})

// 请求拦截器 - 自动附加 token
api.interceptors.request.use((config) => {
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器
api.interceptors.response.use(
  (response) => response.data,
  (error) => {
    const message = error.response?.data?.error || error.message || '请求失败'
    console.error('API Error:', message)
    // 401 未认证时清除 token
    if (error.response?.status === 401) {
      removeToken()
      removeUser()
    }
    return Promise.reject(error)
  }
)

// ==================== 认证 API ====================

export const login = (username: string, password: string) =>
  api.post('/auth/login', { username, password })

// ==================== 用户管理 API ====================

export const getUsers = () => api.get('/users')

export const createUser = (data: { username: string; password: string; nickname?: string; role?: string }) =>
  api.post('/users', data)

export const updateUser = (id: string, data: { username: string; password?: string; nickname?: string; role?: string }) =>
  api.put(`/users/${id}`, data)

export const deleteUser = (id: string) => api.delete(`/users/${id}`)

// ==================== 仪表盘 API ====================

export const getDashboardStats = () => api.get('/dashboard/stats')

// ==================== 客户 API ====================

export const getCustomers = () => api.get('/customers')

export const createCustomer = (data: { name: string; code: string; contact?: string }) =>
  api.post('/customers', data)

export const updateCustomer = (id: string, data: { name: string; code: string; contact?: string }) =>
  api.put(`/customers/${id}`, data)

export const deleteCustomer = (id: string) => api.delete(`/customers/${id}`)

// ==================== 环境 API ====================

export const getEnvs = (customerId: string) => api.get(`/customers/${customerId}/envs`)

export const createEnv = (customerId: string, data: { env_name: string; env_key: string; description?: string }) =>
  api.post(`/customers/${customerId}/envs`, data)

export const deleteEnv = (customerId: string, envId: string) =>
  api.delete(`/customers/${customerId}/envs/${envId}`)

// ==================== 模板 API ====================

export const getTemplates = () => api.get('/templates')

// ==================== 组件 API ====================

export const getComponents = () => api.get('/components')

export const getComponentVariables = (componentId: string) =>
  api.get(`/components/${componentId}/variables`)

// ==================== 配置 API ====================

export const getEnvConfigs = (envId: string) => api.get(`/envs/${envId}/configs`)

export const updateEnvConfigs = (envId: string, data: { values: Record<string, string>; updated_by?: string }) =>
  api.put(`/envs/${envId}/configs`, data)

export const previewConfigs = (envId: string, componentName: string) =>
  api.post(`/envs/${envId}/configs/preview`, { component_name: componentName })

export const exportPackage = (envId: string) =>
  axios.post(`/api/v1/envs/${envId}/export`, {}, { responseType: 'blob' })

// ==================== 部署记录 API ====================

export const getDeployRecords = (envId: string) => api.get(`/envs/${envId}/deploy-records`)

export const createDeployRecord = (envId: string, data: {
  version_tag: string
  deployed_by?: string
  status?: string
  notes?: string
}) => api.post(`/envs/${envId}/deploy-records`, data)

// ==================== 配置版本 API ====================

export const getVersions = (envId: string) => api.get(`/envs/${envId}/versions`)

export const createSnapshot = (envId: string, data: { created_by?: string; change_summary?: string }) =>
  api.post(`/envs/${envId}/versions/snapshot`, data)

export const diffVersions = (envId: string, from: number, to: number) =>
  api.get(`/envs/${envId}/versions/diff`, { params: { from, to } })

export const rollbackVersion = (envId: string, data: { target_version: number; operator?: string }) =>
  api.post(`/envs/${envId}/versions/rollback`, data)

// ==================== 配置克隆 API ====================

export const cloneConfigs = (envId: string, data: { from_env_id: string; updated_by?: string }) =>
  api.post(`/envs/${envId}/configs/clone`, data)

// ==================== 环境克隆 API ====================

export const cloneEnv = (envId: string, data: { from_env_id: string; operator?: string }) =>
  api.post(`/envs/${envId}/clone-env`, data)

// ==================== 制品版本 API ====================

export const getArtifacts = (envId: string) => api.get(`/envs/${envId}/artifacts`)

export const createArtifact = (envId: string, data: {
  component_id: string
  artifact_type: string
  artifact_name: string
  artifact_version: string
  registry_url?: string
}) => api.post(`/envs/${envId}/artifacts`, data)

export const updateArtifact = (envId: string, id: string, data: {
  component_id: string
  artifact_type: string
  artifact_name: string
  artifact_version: string
  registry_url?: string
}) => api.put(`/envs/${envId}/artifacts/${id}`, data)

export const deleteArtifact = (envId: string, id: string) =>
  api.delete(`/envs/${envId}/artifacts/${id}`)

// ==================== 通知配置 API ====================

export const getNotifyConfigs = () => api.get('/notify-configs')

export const createNotifyConfig = (data: {
  name: string
  type: string
  webhook_url: string
  secret?: string
  events?: string
  is_active?: boolean
}) => api.post('/notify-configs', data)

export const updateNotifyConfig = (id: string, data: {
  name: string
  type: string
  webhook_url: string
  secret?: string
  events?: string
  is_active?: boolean
}) => api.put(`/notify-configs/${id}`, data)

export const deleteNotifyConfig = (id: string) => api.delete(`/notify-configs/${id}`)

export const testNotifyConfig = (id: string) => api.post(`/notify-configs/${id}/test`)