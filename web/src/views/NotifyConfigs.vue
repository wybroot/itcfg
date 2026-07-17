<template>
  <div>
    <el-page-header @back="$router.push('/')" title="返回">
      <template #content>
        <span style="font-size: 16px; font-weight: 500">通知配置</span>
      </template>
    </el-page-header>

    <el-card style="margin-top: 20px">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>通知渠道</span>
          <el-button type="primary" @click="openCreateDialog">
            <el-icon><Plus /></el-icon> 添加通知
          </el-button>
        </div>
      </template>

      <el-table :data="configs" stripe v-loading="loading">
        <el-table-column prop="name" label="名称" width="150" />
        <el-table-column prop="type" label="类型" width="110">
          <template #default="{ row }">
            <el-tag :type="row.type === 'dingtalk' ? 'primary' : row.type === 'wecom' ? 'success' : 'info'" size="small">
              {{ row.type === 'dingtalk' ? '钉钉' : row.type === 'wecom' ? '企微' : 'Webhook' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="webhook_url" label="Webhook 地址" min-width="250">
          <template #default="{ row }">
            <span style="font-family: monospace; font-size: 12px">{{ row.webhook_url }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="events" label="触发事件" width="200">
          <template #default="{ row }">
            <el-tag v-for="ev in (row.events || '').split(',')" :key="ev" size="small" style="margin-right: 4px">
              {{ ev.trim() }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="is_active" label="状态" width="80">
          <template #default="{ row }">
            <el-switch :model-value="row.is_active" disabled />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEditDialog(row)">编辑</el-button>
            <el-button size="small" type="success" @click="handleTest(row)">测试</el-button>
            <el-popconfirm title="确认删除?" @confirm="handleDelete(row.id)">
              <template #reference>
                <el-button size="small" type="danger">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && configs.length === 0" description="暂无通知配置" />
    </el-card>

    <el-dialog v-model="showDialog" :title="isEditing ? '编辑通知' : '添加通知'" width="550px" @close="resetForm">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="如: 钉钉运维群" />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-select v-model="form.type" style="width: 100%">
            <el-option label="钉钉机器人" value="dingtalk" />
            <el-option label="企业微信机器人" value="wecom" />
            <el-option label="通用 Webhook" value="webhook" />
          </el-select>
        </el-form-item>
        <el-form-item label="Webhook URL" prop="webhook_url">
          <el-input v-model="form.webhook_url" placeholder="https://hooks.example.com/..." />
        </el-form-item>
        <el-form-item label="加签密钥">
          <el-input v-model="form.secret" placeholder="可选，钉钉加签用" />
        </el-form-item>
        <el-form-item label="触发事件">
          <el-checkbox-group v-model="eventList">
            <el-checkbox label="deploy_success">部署成功</el-checkbox>
            <el-checkbox label="deploy_failed">部署失败</el-checkbox>
            <el-checkbox label="config_updated">配置更新</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.is_active" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { getNotifyConfigs, createNotifyConfig, updateNotifyConfig, deleteNotifyConfig, testNotifyConfig } from '../api'

interface NotifyConfigItem {
  id: string
  name: string
  type: string
  webhook_url: string
  events: string
  is_active: boolean
}

const configs = ref<NotifyConfigItem[]>([])
const loading = ref(false)
const showDialog = ref(false)
const isEditing = ref(false)
const submitting = ref(false)
const editingId = ref('')
const formRef = ref<FormInstance>()

const form = ref({
  name: '',
  type: 'dingtalk',
  webhook_url: '',
  secret: '',
  is_active: true,
})

const eventList = ref(['deploy_success', 'deploy_failed'])

const rules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  webhook_url: [{ required: true, message: '请输入 Webhook 地址', trigger: 'blur' }],
}

const fetchConfigs = async () => {
  loading.value = true
  try {
    const res: any = await getNotifyConfigs()
    configs.value = res.data || []
  } catch {
    ElMessage.error('获取通知配置失败')
  } finally {
    loading.value = false
  }
}

const openCreateDialog = () => {
  isEditing.value = false
  editingId.value = ''
  resetForm()
  showDialog.value = true
}

const openEditDialog = (row: any) => {
  isEditing.value = true
  editingId.value = row.id
  form.value = {
    name: row.name,
    type: row.type,
    webhook_url: row.webhook_url,
    secret: '',
    is_active: row.is_active,
  }
  eventList.value = (row.events || '').split(',').map((e: string) => e.trim()).filter(Boolean)
  showDialog.value = true
}

const resetForm = () => {
  form.value = { name: '', type: 'dingtalk', webhook_url: '', secret: '', is_active: true }
  eventList.value = ['deploy_success', 'deploy_failed']
  formRef.value?.resetFields()
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  const data: any = {
    name: form.value.name,
    type: form.value.type,
    webhook_url: form.value.webhook_url,
    events: eventList.value.join(','),
    is_active: form.value.is_active,
  }
  if (form.value.secret) data.secret = form.value.secret

  try {
    if (isEditing.value) {
      await updateNotifyConfig(editingId.value, data)
      ElMessage.success('更新成功')
    } else {
      await createNotifyConfig(data)
      ElMessage.success('创建成功')
    }
    showDialog.value = false
    fetchConfigs()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '操作失败')
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (id: string) => {
  try {
    await deleteNotifyConfig(id)
    ElMessage.success('删除成功')
    fetchConfigs()
  } catch {
    ElMessage.error('删除失败')
  }
}

const handleTest = async (row: any) => {
  try {
    await testNotifyConfig(row.id)
    ElMessage.success('测试消息发送成功')
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '测试失败')
  }
}

onMounted(fetchConfigs)
</script>