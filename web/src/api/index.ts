import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
})

// 响应拦截器
api.interceptors.response.use(
  (response) => response.data,
  (error) => {
    const message = error.response?.data?.error || error.message || '请求失败'
    console.error('API Error:', message)
    return Promise.reject(error)
  }
)

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
  api.post(`/envs/${envId}/export`, {}, { responseType: 'blob' })

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