<template>
  <div>
    <el-page-header @back="$router.push('/customers')" title="返回">
      <template #content>
        <span style="font-size: 16px; font-weight: 500">环境管理 - {{ customerName }}</span>
      </template>
    </el-page-header>

    <el-card style="margin-top: 20px">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>环境列表</span>
          <el-button type="primary" @click="showCreateDialog = true">
            <el-icon><Plus /></el-icon> 新建环境
          </el-button>
        </div>
      </template>

      <el-table :data="envs" stripe v-loading="loading">
        <el-table-column prop="env_name" label="环境名称" width="150" />
        <el-table-column prop="env_key" label="环境密钥" width="250">
          <template #default="{ row }">
            <el-tag>{{ row.env_key }}</el-tag>
            <el-button size="small" text @click="copyEnvKey(row.env_key)">
              <el-icon><CopyDocument /></el-icon>
            </el-button>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="200" />
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ new Date(row.created_at).toLocaleDateString('zh-CN') }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="300" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="viewConfigs(row)">
              <el-icon><Edit /></el-icon> 配置管理
            </el-button>
            <el-button size="small" type="success" @click="exportPkg(row)">
              <el-icon><Download /></el-icon> 导出部署包
            </el-button>
            <el-popconfirm title="确认删除该环境?" @confirm="handleDelete(row.id)">
              <template #reference>
                <el-button size="small" type="danger">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新建环境对话框 -->
    <el-dialog
      title="新建环境"
      v-model="showCreateDialog"
      width="500px"
      @close="resetForm"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="环境名称" prop="env_name">
          <el-select v-model="form.env_name" placeholder="请选择环境">
            <el-option label="生产环境" value="生产环境" />
            <el-option label="测试环境" value="测试环境" />
            <el-option label="灾备环境" value="灾备环境" />
          </el-select>
        </el-form-item>
        <el-form-item label="环境密钥" prop="env_key">
          <el-input v-model="form.env_key" placeholder="请输入唯一密钥">
            <template #append>
              <el-button @click="generateKey">生成</el-button>
            </template>
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getEnvs, createEnv, deleteEnv } from '../api'

const route = useRoute()
const router = useRouter()
const customerId = route.params.id as string
const customerName = ref('环境管理')
const envs = ref([])
const loading = ref(false)
const showCreateDialog = ref(false)
const submitting = ref(false)
const formRef = ref()

const form = ref({
  env_name: '生产环境',
  env_key: '',
  description: '',
})

const rules = {
  env_name: [{ required: true, message: '请选择环境名称', trigger: 'change' }],
  env_key: [{ required: true, message: '请输入环境密钥', trigger: 'blur' }],
}

const generateKey = () => {
  const chars = 'abcdefghijklmnopqrstuvwxyz0123456789'
  let key = 'env-'
  for (let i = 0; i < 16; i++) {
    key += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  form.value.env_key = key
}

const fetchEnvs = async () => {
  loading.value = true
  try {
    const res: any = await getEnvs(customerId)
    envs.value = res.data || []
  } catch {
    ElMessage.error('获取环境列表失败')
  } finally {
    loading.value = false
  }
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    await createEnv(customerId, form.value)
    ElMessage.success('环境创建成功')
    showCreateDialog.value = false
    fetchEnvs()
  } catch {
    ElMessage.error('创建失败')
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (envId: string) => {
  try {
    await deleteEnv(customerId, envId)
    ElMessage.success('删除成功')
    fetchEnvs()
  } catch {
    ElMessage.error('删除失败')
  }
}

const viewConfigs = (env: any) => {
  router.push(`/envs/${env.id}/configs`)
}

const exportPkg = (env: any) => {
  ElMessage.info('部署包导出功能将在后续版本实现')
}

const copyEnvKey = (key: string) => {
  navigator.clipboard.writeText(key)
  ElMessage.success('已复制环境密钥')
}

const resetForm = () => {
  form.value = { env_name: '生产环境', env_key: '', description: '' }
  formRef.value?.resetFields()
}

onMounted(fetchEnvs)
</script>