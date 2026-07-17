<template>
  <div class="page envs-page">
    <PageHeader title="环境管理" :subtitle="`客户 ID：${customerId}`" back="/customers">
      <template #actions>
        <el-button type="primary" @click="openCreateDialog">
          <el-icon><Plus /></el-icon> 新建环境
        </el-button>
      </template>
    </PageHeader>

    <PageCard title="环境列表" :subtitle="`共 ${envs.length} 个环境，配置、制品和导出都从这里进入`">
      <el-table :data="envs" stripe v-loading="loading" empty-text="暂无环境，请先新建环境">
        <el-table-column label="环境" width="150">
          <template #default="{ row }">
            <el-tag effect="light" type="primary">{{ row.env_name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="环境密钥" min-width="260">
          <template #default="{ row }">
            <div class="env-key-wrap">
              <span class="env-key code-text">{{ row.env_key }}</span>
              <el-button size="small" text @click="copyEnvKey(row.env_key)">
                <el-icon><CopyDocument /></el-icon>
              </el-button>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="220">
          <template #default="{ row }">
            <span v-if="row.description">{{ row.description }}</span>
            <span v-else class="muted">未填写</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="150">
          <template #default="{ row }">
            {{ row.created_at ? new Date(row.created_at).toLocaleDateString('zh-CN') : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="500" fixed="right">
          <template #default="{ row }">
            <el-space wrap>
              <el-button size="small" type="primary" @click="viewConfigs(row)">
                <el-icon><Edit /></el-icon> 配置
              </el-button>
              <el-button size="small" type="success" plain @click="exportPkg(row)">
                <el-icon><Download /></el-icon> 导出
              </el-button>
              <el-button size="small" plain @click="openComponentDialog(row)">组件</el-button>
              <el-button size="small" plain @click="openEditDialog(row)">编辑</el-button>
              <el-button size="small" type="warning" plain @click="openCloneDialog(row)">克隆</el-button>
              <el-popconfirm title="确认删除该环境?" @confirm="handleDelete(row.id)">
                <template #reference>
                  <el-button size="small" type="danger" plain>删除</el-button>
                </template>
              </el-popconfirm>
            </el-space>
          </template>
        </el-table-column>
      </el-table>
    </PageCard>

    <el-dialog :title="isEditing ? '编辑环境' : '新建环境'" v-model="showCreateDialog" width="500px" @close="resetForm">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="环境名称" prop="env_name">
          <el-select v-model="form.env_name" placeholder="请选择环境" style="width: 100%">
            <el-option label="生产环境" value="生产环境" />
            <el-option label="测试环境" value="测试环境" />
            <el-option label="灾备环境" value="灾备环境" />
          </el-select>
        </el-form-item>
        <el-form-item label="环境密钥" prop="env_key">
          <el-input v-model="form.env_key" placeholder="请输入唯一密钥">
            <template #append><el-button @click="generateKey">生成</el-button></template>
          </el-input>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="可选描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确认</el-button>
      </template>
    </el-dialog>

    <el-dialog title="环境组件" v-model="showComponentDialog" width="640px">
      <el-alert type="info" :closable="false" class="dialog-alert">
        请选择当前环境需要部署的组件。配置页和制品页只会展示已启用组件。
      </el-alert>
      <el-table :data="componentOptions" v-loading="componentLoading" size="small" empty-text="暂无可选组件">
        <el-table-column label="启用" width="80">
          <template #default="{ row }"><el-switch v-model="selectedComponents[row.id]" /></template>
        </el-table-column>
        <el-table-column label="组件" min-width="180">
          <template #default="{ row }">
            <div class="component-name">{{ row.display_name }}</div>
            <div class="muted code-text">{{ row.template_dir || row.name }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="category" label="分类" width="130" />
        <el-table-column label="部署顺序" width="110">
          <template #default="{ $index }">{{ $index + 1 }}</template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="showComponentDialog = false">取消</el-button>
        <el-button type="primary" @click="saveEnvComponents" :loading="componentSaving">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog title="克隆环境" v-model="showCloneDialog" width="500px">
      <el-alert type="info" :closable="false" class="dialog-alert">将源环境的配置和制品版本完整复制到目标环境</el-alert>
      <el-form label-width="100px">
        <el-form-item label="目标环境"><el-tag type="primary">{{ cloneTarget?.env_name }}</el-tag></el-form-item>
        <el-form-item label="源环境" required>
          <el-select v-model="cloneSource" placeholder="请选择要克隆的源环境" style="width: 100%">
            <el-option v-for="env in envs" :key="env.id" :label="`${env.env_name} (${env.env_key})`" :value="env.id" :disabled="env.id === cloneTarget?.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCloneDialog = false">取消</el-button>
        <el-button type="primary" @click="handleClone" :loading="cloneLoading">开始克隆</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import PageHeader from '../components/PageHeader.vue'
import PageCard from '../components/PageCard.vue'
import { getEnvs, createEnv, updateEnv, deleteEnv, cloneEnv, exportPackage, getComponents, getEnvComponents, replaceEnvComponents } from '../api'

interface Env { id: string; env_name: string; env_key: string; description?: string; created_at?: string }

const route = useRoute()
const router = useRouter()
const customerId = route.params.id as string
const envs = ref<Env[]>([])
const loading = ref(false)
const showCreateDialog = ref(false)
const isEditing = ref(false)
const editingId = ref('')
const submitting = ref(false)
const formRef = ref()
const form = ref({ env_name: '生产环境', env_key: '', description: '' })
const rules = {
  env_name: [{ required: true, message: '请选择环境名称', trigger: 'change' }],
  env_key: [{ required: true, message: '请输入环境密钥', trigger: 'blur' }],
}

const generateKey = () => {
  const chars = 'abcdefghijklmnopqrstuvwxyz0123456789'
  let key = 'env-'
  for (let i = 0; i < 16; i++) key += chars.charAt(Math.floor(Math.random() * chars.length))
  form.value.env_key = key
}

const fetchEnvs = async () => {
  loading.value = true
  try {
    const res: any = await getEnvs(customerId)
    envs.value = res.data || []
  } catch { ElMessage.error('获取环境列表失败') } finally { loading.value = false }
}

const openCreateDialog = () => { isEditing.value = false; editingId.value = ''; form.value = { env_name: '生产环境', env_key: '', description: '' }; showCreateDialog.value = true }
const openEditDialog = (env: Env) => { isEditing.value = true; editingId.value = env.id; form.value = { env_name: env.env_name, env_key: env.env_key, description: env.description || '' }; showCreateDialog.value = true }
const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    if (isEditing.value) { await updateEnv(customerId, editingId.value, form.value); ElMessage.success('环境更新成功') }
    else { await createEnv(customerId, form.value); ElMessage.success('环境创建成功') }
    showCreateDialog.value = false
    fetchEnvs()
  } catch { ElMessage.error(isEditing.value ? '更新失败' : '创建失败') } finally { submitting.value = false }
}
const handleDelete = async (envId: string) => { try { await deleteEnv(customerId, envId); ElMessage.success('删除成功'); fetchEnvs() } catch { ElMessage.error('删除失败') } }
const viewConfigs = (env: Env) => router.push(`/envs/${env.id}/configs`)
const copyEnvKey = (key: string) => { navigator.clipboard.writeText(key); ElMessage.success('已复制环境密钥') }

const showComponentDialog = ref(false)
const componentTarget = ref<Env | null>(null)
const componentOptions = ref<any[]>([])
const selectedComponents = ref<Record<string, boolean>>({})
const componentLoading = ref(false)
const componentSaving = ref(false)
const showCloneDialog = ref(false)
const cloneTarget = ref<Env | null>(null)
const cloneSource = ref('')
const cloneLoading = ref(false)

const openComponentDialog = async (env: Env) => {
  componentTarget.value = env
  showComponentDialog.value = true
  componentLoading.value = true
  selectedComponents.value = {}
  try {
    const [componentsRes, envComponentsRes]: any[] = await Promise.all([getComponents(), getEnvComponents(env.id)])
    componentOptions.value = (componentsRes.data || []).filter((comp: any) => ['postgresql', 'java-app', 'nginx'].includes(comp.template_dir || comp.name))
    ;(envComponentsRes.data || []).forEach((item: any) => { selectedComponents.value[item.component_id] = item.enabled !== false })
  } catch { ElMessage.error('获取环境组件失败') } finally { componentLoading.value = false }
}
const saveEnvComponents = async () => {
  if (!componentTarget.value) return
  componentSaving.value = true
  try {
    const components = componentOptions.value.filter((comp: any) => selectedComponents.value[comp.id]).map((comp: any, index: number) => ({ component_id: comp.id, enabled: true, deploy_order: index + 1 }))
    await replaceEnvComponents(componentTarget.value.id, { components })
    ElMessage.success('环境组件保存成功')
    showComponentDialog.value = false
  } catch { ElMessage.error('环境组件保存失败') } finally { componentSaving.value = false }
}
const openCloneDialog = (env: Env) => { cloneTarget.value = env; cloneSource.value = ''; showCloneDialog.value = true }
const handleClone = async () => {
  if (!cloneSource.value) return ElMessage.warning('请选择源环境')
  if (!cloneTarget.value) return ElMessage.warning('目标环境不存在')
  cloneLoading.value = true
  try { await cloneEnv(cloneTarget.value.id, { from_env_id: cloneSource.value, operator: 'admin' }); ElMessage.success('环境克隆成功'); showCloneDialog.value = false }
  catch { ElMessage.error('克隆失败') } finally { cloneLoading.value = false }
}
const exportPkg = async (env: Env) => {
  ElMessage.info('正在打包部署包，请稍候...')
  try {
    const res: any = await exportPackage(env.id)
    const blob = res.data || res
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url; a.download = `${env.env_key}-deploy.tar.gz`; document.body.appendChild(a); a.click(); document.body.removeChild(a); URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch { ElMessage.error('导出失败') }
}
const resetForm = () => { isEditing.value = false; editingId.value = ''; form.value = { env_name: '生产环境', env_key: '', description: '' }; formRef.value?.resetFields() }
onMounted(fetchEnvs)
</script>

<style scoped>
.envs-page { max-width: 1440px; margin: 0 auto; }
.env-key-wrap { display: flex; align-items: center; gap: 6px; }
.env-key { padding: 4px 8px; color: #1d4ed8; border-radius: 8px; background: #eff6ff; font-size: 12px; }
.component-name { font-weight: 650; }
.dialog-alert { margin-bottom: 16px; }
</style>
