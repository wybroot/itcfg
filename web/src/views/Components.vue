<template>
  <div>
    <el-card>
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <span>组件列表</span>
          <el-button type="primary" size="small" @click="openCreateDialog">新建组件</el-button>
        </div>
      </template>
      <el-table :data="components" stripe v-loading="loading">
        <el-table-column prop="name" label="组件标识" width="150" />
        <el-table-column prop="display_name" label="组件名称" width="200" />
        <el-table-column prop="category" label="分类" width="120">
          <template #default="{ row }">
            <el-tag>{{ getCategoryLabel(row.category) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="250" />
        <el-table-column prop="is_active" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_active ? 'success' : 'info'">
              {{ row.is_active ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" link @click="openEditDialog(row)">编辑</el-button>
            <el-popconfirm title="确定删除此组件?" @confirm="handleDelete(row.id)">
              <template #reference>
                <el-button size="small" type="danger" link>删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑组件' : '新建组件'"
      width="520px"
      @closed="resetForm"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="组件标识" prop="name">
          <el-input v-model="form.name" :disabled="isEditing" placeholder="如 nginx, redis" />
        </el-form-item>
        <el-form-item label="组件名称" prop="display_name">
          <el-input v-model="form.display_name" placeholder="如 Nginx 反向代理" />
        </el-form-item>
        <el-form-item label="分类" prop="category">
          <el-select v-model="form.category" placeholder="选择分类" style="width: 100%">
            <el-option v-for="item in categoryOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="组件功能描述" />
        </el-form-item>
        <el-form-item label="模板目录" prop="template_dir">
          <el-input v-model="form.template_dir" placeholder="templates 下的目录名" />
        </el-form-item>
        <el-form-item label="启用状态">
          <el-switch v-model="form.is_active" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getComponents, createComponent, updateComponent, deleteComponent } from '../api'

interface ComponentItem {
  id: string
  name: string
  display_name: string
  category: string
  description: string
  template_dir: string
  is_active: boolean
}

const components = ref<ComponentItem[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const isEditing = ref(false)
const editingId = ref('')
const submitting = ref(false)
const formRef = ref()

const categoryOptions = [
  { value: 'web-server', label: 'Web 服务器' },
  { value: 'application', label: '应用服务' },
  { value: 'database', label: '数据库' },
  { value: 'cache', label: '缓存' },
  { value: 'message-queue', label: '消息队列' },
  { value: 'object-storage', label: '对象存储' },
  { value: 'coordination', label: '协调服务' },
  { value: 'search-engine', label: '搜索引擎' },
  { value: 'office', label: '办公套件' },
  { value: 'file-service', label: '文件服务' },
]

const form = ref({
  name: '',
  display_name: '',
  category: '',
  description: '',
  template_dir: '',
  is_active: true,
})

const rules = {
  name: [{ required: true, message: '请输入组件标识', trigger: 'blur' }],
  display_name: [{ required: true, message: '请输入组件名称', trigger: 'blur' }],
}

const getCategoryLabel = (category: string) => {
  const found = categoryOptions.find(c => c.value === category)
  return found ? found.label : category
}

const fetchComponents = async () => {
  loading.value = true
  try {
    const res: any = await getComponents()
    components.value = res.data || []
  } catch {
    ElMessage.error('获取组件列表失败')
  } finally {
    loading.value = false
  }
}

const openCreateDialog = () => {
  isEditing.value = false
  editingId.value = ''
  form.value = { name: '', display_name: '', category: '', description: '', template_dir: '', is_active: true }
  dialogVisible.value = true
}

const openEditDialog = (row: ComponentItem) => {
  isEditing.value = true
  editingId.value = row.id
  form.value = {
    name: row.name,
    display_name: row.display_name,
    category: row.category,
    description: row.description,
    template_dir: row.template_dir,
    is_active: row.is_active,
  }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (isEditing.value) {
      await updateComponent(editingId.value, form.value)
      ElMessage.success('组件更新成功')
    } else {
      await createComponent(form.value)
      ElMessage.success('组件创建成功')
    }
    dialogVisible.value = false
    fetchComponents()
  } catch {
    ElMessage.error(isEditing.value ? '更新失败' : '创建失败')
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (id: string) => {
  try {
    await deleteComponent(id)
    ElMessage.success('删除成功')
    fetchComponents()
  } catch {
    ElMessage.error('删除失败')
  }
}

const resetForm = () => {
  formRef.value?.resetFields()
}

onMounted(fetchComponents)
</script>
