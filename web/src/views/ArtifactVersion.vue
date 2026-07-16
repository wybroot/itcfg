<template>
  <div>
    <el-page-header @back="$router.back()" title="返回">
      <template #content>
        <span style="font-size: 16px; font-weight: 500">制品版本管理</span>
      </template>
    </el-page-header>

    <el-card style="margin-top: 20px">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>制品版本列表</span>
          <el-button type="primary" @click="openCreateDialog">
            <el-icon><Plus /></el-icon> 添加制品版本
          </el-button>
        </div>
      </template>

      <el-table :data="artifacts" stripe v-loading="loading">
        <el-table-column label="组件" min-width="120">
          <template #default="{ row }">
            {{ getComponentName(row.component_id) }}
          </template>
        </el-table-column>
        <el-table-column prop="artifact_type" label="制品类型" width="110">
          <template #default="{ row }">
            <el-tag size="small">{{ row.artifact_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="artifact_name" label="制品名称" min-width="150" />
        <el-table-column prop="artifact_version" label="制品版本" width="120">
          <template #default="{ row }">
            <el-tag type="success">{{ row.artifact_version }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="registry_url" label="仓库地址" min-width="200">
          <template #default="{ row }">
            <span v-if="row.registry_url" style="font-family: monospace; font-size: 12px">
              {{ row.registry_url }}
            </span>
            <span v-else style="color: #c0c4cc">未设置</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEditDialog(row)">编辑</el-button>
            <el-popconfirm title="确认删除此制品版本?" @confirm="handleDelete(row.id)">
              <template #reference>
                <el-button size="small" type="danger">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && artifacts.length === 0" description="暂无制品版本，请添加" />
    </el-card>

    <!-- 添加/编辑对话框 -->
    <el-dialog
      v-model="showDialog"
      :title="isEditing ? '编辑制品版本' : '添加制品版本'"
      width="550px"
      @close="resetForm"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="组件" prop="component_id">
          <el-select v-model="form.component_id" placeholder="请选择组件" style="width: 100%">
            <el-option
              v-for="comp in components"
              :key="comp.id"
              :label="comp.display_name"
              :value="comp.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="制品类型" prop="artifact_type">
          <el-select v-model="form.artifact_type" placeholder="请选择制品类型" style="width: 100%">
            <el-option label="Docker 镜像" value="docker" />
            <el-option label="Jar 包" value="jar" />
            <el-option label="二进制" value="binary" />
            <el-option label="Helm Chart" value="helm" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="制品名称" prop="artifact_name">
          <el-input v-model="form.artifact_name" placeholder="例如: nginx, java-app" />
        </el-form-item>
        <el-form-item label="制品版本" prop="artifact_version">
          <el-input v-model="form.artifact_version" placeholder="例如: 1.25.0, 2.3.1" />
        </el-form-item>
        <el-form-item label="仓库地址">
          <el-input v-model="form.registry_url" placeholder="例如: harbor.example.com/library/nginx:1.25.0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          {{ isEditing ? '保存' : '添加' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { getArtifacts, createArtifact, updateArtifact, deleteArtifact } from '../api'
import { getComponents } from '../api'

const route = useRoute()
const envId = route.params.envId as string

const artifacts = ref<any[]>([])
const components = ref<any[]>([])
const loading = ref(false)
const showDialog = ref(false)
const isEditing = ref(false)
const submitting = ref(false)
const editingId = ref('')
const formRef = ref<FormInstance>()

const form = ref({
  component_id: '',
  artifact_type: '',
  artifact_name: '',
  artifact_version: '',
  registry_url: '',
})

const rules: FormRules = {
  component_id: [{ required: true, message: '请选择组件', trigger: 'change' }],
  artifact_type: [{ required: true, message: '请选择制品类型', trigger: 'change' }],
  artifact_name: [{ required: true, message: '请输入制品名称', trigger: 'blur' }],
  artifact_version: [{ required: true, message: '请输入制品版本', trigger: 'blur' }],
}

const getComponentName = (id: string) => {
  const comp = components.value.find((c: any) => c.id === id)
  return comp ? comp.display_name : id
}

const fetchArtifacts = async () => {
  loading.value = true
  try {
    const res: any = await getArtifacts(envId)
    artifacts.value = res.data || []
  } catch {
    ElMessage.error('获取制品版本列表失败')
  } finally {
    loading.value = false
  }
}

const fetchComponents = async () => {
  try {
    const res: any = await getComponents()
    components.value = res.data || []
  } catch {
    ElMessage.error('获取组件列表失败')
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
    component_id: row.component_id,
    artifact_type: row.artifact_type,
    artifact_name: row.artifact_name,
    artifact_version: row.artifact_version,
    registry_url: row.registry_url || '',
  }
  showDialog.value = true
}

const resetForm = () => {
  form.value = {
    component_id: '',
    artifact_type: '',
    artifact_name: '',
    artifact_version: '',
    registry_url: '',
  }
  formRef.value?.resetFields()
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (isEditing.value) {
      await updateArtifact(envId, editingId.value, form.value)
      ElMessage.success('更新成功')
    } else {
      await createArtifact(envId, form.value)
      ElMessage.success('添加成功')
    }
    showDialog.value = false
    fetchArtifacts()
  } catch {
    ElMessage.error(isEditing.value ? '更新失败' : '添加失败')
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (id: string) => {
  try {
    await deleteArtifact(envId, id)
    ElMessage.success('删除成功')
    fetchArtifacts()
  } catch {
    ElMessage.error('删除失败')
  }
}

onMounted(() => {
  fetchComponents()
  fetchArtifacts()
})
</script>