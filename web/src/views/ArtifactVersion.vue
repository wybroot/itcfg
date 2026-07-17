<template>
  <div class="page artifacts-page">
    <PageHeader title="制品版本管理" :subtitle="`环境 ID：${envId}`" back>
      <template #actions>
        <el-button type="primary" @click="openCreateDialog" :disabled="components.length === 0">
          <el-icon><Plus /></el-icon> 添加制品版本
        </el-button>
      </template>
    </PageHeader>

    <PageCard title="制品版本列表" subtitle="为环境已启用组件绑定镜像、Jar 或其他部署制品">
      <el-alert v-if="components.length === 0" type="warning" :closable="false" class="page-alert">
        当前环境还没有启用组件，请先回到环境管理选择需要部署的组件。
      </el-alert>

      <el-table :data="artifacts" stripe v-loading="loading" empty-text="暂无制品版本，请添加">
        <el-table-column label="组件" min-width="170">
          <template #default="{ row }">
            <div class="component-cell">
              <div>{{ getComponentName(row.component_id) }}</div>
              <div class="muted code-text">{{ getComponentKey(row.component_id) }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="artifact_type" label="制品类型" width="120">
          <template #default="{ row }">
            <el-tag effect="light">{{ getArtifactTypeLabel(row.artifact_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="artifact_name" label="制品名称" min-width="160" />
        <el-table-column prop="artifact_version" label="制品版本" width="150">
          <template #default="{ row }">
            <el-tag type="success" effect="light" class="code-text">{{ row.artifact_version }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="registry_url" label="仓库地址 / 本地文件" min-width="260">
          <template #default="{ row }">
            <span v-if="row.registry_url" class="code-text registry-url">{{ row.registry_url }}</span>
            <span v-else class="muted">未设置</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="170" fixed="right">
          <template #default="{ row }">
            <el-space>
              <el-button size="small" @click="openEditDialog(row)">编辑</el-button>
              <el-popconfirm title="确认删除此制品版本?" @confirm="handleDelete(row.id)">
                <template #reference>
                  <el-button size="small" type="danger" plain>删除</el-button>
                </template>
              </el-popconfirm>
            </el-space>
          </template>
        </el-table-column>
      </el-table>
    </PageCard>

    <el-dialog v-model="showDialog" :title="isEditing ? '编辑制品版本' : '添加制品版本'" width="560px" @close="resetForm">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="组件" prop="component_id">
          <el-select v-model="form.component_id" placeholder="请选择组件" style="width: 100%">
            <el-option v-for="comp in components" :key="comp.id" :label="comp.display_name" :value="comp.id">
              <span>{{ comp.display_name }}</span>
              <span class="option-meta">{{ comp.template_dir || comp.name }}</span>
            </el-option>
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
          <el-input v-model="form.registry_url" placeholder="镜像地址或本地 tar 文件路径" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">{{ isEditing ? '保存' : '添加' }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import PageHeader from '../components/PageHeader.vue'
import PageCard from '../components/PageCard.vue'
import { getArtifacts, createArtifact, updateArtifact, deleteArtifact, getEnvComponents } from '../api'

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

const form = ref({ component_id: '', artifact_type: '', artifact_name: '', artifact_version: '', registry_url: '' })
const typeLabels: Record<string, string> = { docker: 'Docker 镜像', jar: 'Jar 包', binary: '二进制', helm: 'Helm Chart', other: '其他' }

const rules: FormRules = {
  component_id: [{ required: true, message: '请选择组件', trigger: 'change' }],
  artifact_type: [{ required: true, message: '请选择制品类型', trigger: 'change' }],
  artifact_name: [{ required: true, message: '请输入制品名称', trigger: 'blur' }],
  artifact_version: [{ required: true, message: '请输入制品版本', trigger: 'blur' }],
}

const getComponent = (id: string) => components.value.find((c: any) => c.id === id)
const getComponentName = (id: string) => getComponent(id)?.display_name || id
const getComponentKey = (id: string) => {
  const comp = getComponent(id)
  return comp ? (comp.template_dir || comp.name) : id
}
const getArtifactTypeLabel = (type: string) => typeLabels[type] || type

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
    const res: any = await getEnvComponents(envId)
    components.value = (res.data || []).map((item: any) => item.component).filter(Boolean)
  } catch {
    ElMessage.error('获取环境组件失败')
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
  form.value = { component_id: '', artifact_type: '', artifact_name: '', artifact_version: '', registry_url: '' }
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

<style scoped>
.artifacts-page { max-width: 1440px; margin: 0 auto; }
.page-alert { margin-bottom: 16px; }
.component-cell { line-height: 1.45; }
.registry-url { color: #1d4ed8; font-size: 12px; }
.option-meta { float: right; color: var(--itcfg-text-muted); font-size: 12px; }
</style>
