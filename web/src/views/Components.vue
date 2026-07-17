<template>
  <div class="page components-page">
    <PageHeader title="组件管理" subtitle="维护可部署组件及其配置变量定义">
      <template #actions>
        <el-space wrap>
          <el-button @click="$router.push('/templates')">从模板添加组件</el-button>
          <el-button type="primary" @click="openCreateDialog">
            <el-icon><Plus /></el-icon> 新建自定义组件
          </el-button>
        </el-space>
      </template>
    </PageHeader>

    <PageCard title="组件列表" :subtitle="`共 ${components.length} 个组件，配置页面会按这里的变量定义生成表单`">
      <el-table :data="components" stripe v-loading="loading" empty-text="暂无组件，请先新建组件">
        <el-table-column label="组件" min-width="210">
          <template #default="{ row }">
            <div class="component-name">{{ row.display_name }}</div>
            <div class="muted code-text">{{ row.name }}</div>
            <div v-if="row.description" class="component-desc">{{ row.description }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="category" label="分类" width="130">
          <template #default="{ row }">
            <el-tag effect="light">{{ getCategoryLabel(row.category) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="template_dir" label="模板目录" width="160">
          <template #default="{ row }">
            <span class="code-text">{{ row.template_dir || row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column label="变量" width="100">
          <template #default="{ row }">
            <el-tag :type="row.variables?.length ? 'success' : 'info'" effect="light">{{ row.variables?.length || 0 }} 个</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="变量分组" min-width="180">
          <template #default="{ row }">
            <el-space wrap v-if="getVariableGroups(row).length">
              <el-tag v-for="group in getVariableGroups(row)" :key="group" size="small" type="info" effect="plain">{{ group }}</el-tag>
            </el-space>
            <span v-else class="muted">未同步变量</span>
          </template>
        </el-table-column>
        <el-table-column prop="is_active" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_active ? 'success' : 'info'" effect="light">{{ row.is_active ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="230" fixed="right">
          <template #default="{ row }">
            <el-space>
              <el-button size="small" type="primary" plain @click="openVariableDrawer(row)">变量</el-button>
              <el-button size="small" @click="openEditDialog(row)">编辑</el-button>
              <el-popconfirm title="确定删除此组件?" @confirm="handleDelete(row.id)">
                <template #reference>
                  <el-button size="small" type="danger" plain>删除</el-button>
                </template>
              </el-popconfirm>
            </el-space>
          </template>
        </el-table-column>
      </el-table>
    </PageCard>

    <el-dialog v-model="dialogVisible" :title="isEditing ? '编辑组件' : '新建组件'" width="520px" @closed="resetForm">
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

    <el-drawer v-model="variableDrawerVisible" :title="variableTarget ? `${variableTarget.display_name} - 变量管理` : '变量管理'" size="780px">
      <div v-if="variableTarget" class="variable-toolbar">
        <div>
          <div class="component-name">{{ variableTarget.display_name }}</div>
          <div class="muted code-text">{{ variableTarget.template_dir || variableTarget.name }}</div>
          <div class="muted">优先从模板同步变量；手工新增/编辑仅用于自定义组件或现场差异补充。</div>
        </div>
        <el-space wrap>
          <el-button type="success" @click="importVariables(false)" :loading="importing">同步模板变量</el-button>
          <el-button type="warning" plain @click="importVariables(true)" :loading="importing">覆盖同步</el-button>
          <el-button type="primary" plain @click="openVariableDialog()">高级：新增变量</el-button>
        </el-space>
      </div>

      <div v-loading="variableLoading" class="variable-groups">
        <el-empty v-if="!variableLoading && variables.length === 0" description="该组件暂无变量定义，请优先从模板同步变量" />
        <section v-for="group in variableGroups" :key="group" class="variable-group">
          <div class="group-header">
            <div>
              <h3>{{ group }}</h3>
              <p>{{ getVariablesByGroup(group).length }} 个配置项</p>
            </div>
          </div>
          <el-table :data="getVariablesByGroup(group)" stripe size="small">
            <el-table-column prop="var_name" label="变量名" min-width="150">
              <template #default="{ row }"><span class="code-text">{{ row.var_name }}</span></template>
            </el-table-column>
            <el-table-column prop="var_label" label="标签" min-width="150" />
            <el-table-column prop="var_type" label="类型" width="100">
              <template #default="{ row }"><el-tag effect="light" size="small">{{ row.var_type }}</el-tag></template>
            </el-table-column>
            <el-table-column prop="default_value" label="默认值" min-width="120">
              <template #default="{ row }">
                <span v-if="row.default_value" class="code-text">{{ row.default_value }}</span>
                <span v-else class="muted">空</span>
              </template>
            </el-table-column>
            <el-table-column prop="required" label="必填" width="80">
              <template #default="{ row }">{{ row.required ? '是' : '否' }}</template>
            </el-table-column>
            <el-table-column prop="description" label="说明" min-width="220">
              <template #default="{ row }">
                <div>{{ row.description || '-' }}</div>
                <div v-if="row.linked_to" class="muted">关联：<span class="code-text">{{ row.linked_to }}</span></div>
              </template>
            </el-table-column>
            <el-table-column label="高级操作" width="150" fixed="right">
              <template #default="{ row }">
                <el-button size="small" text type="primary" @click="openVariableDialog(row)">编辑</el-button>
                <el-popconfirm title="删除变量会同步删除该变量已保存的客户配置，确认删除?" @confirm="handleDeleteVariable(row.id)">
                  <template #reference>
                    <el-button size="small" text type="danger">删除</el-button>
                  </template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>
        </section>
      </div>
    </el-drawer>

    <el-dialog v-model="variableDialogVisible" :title="isVariableEditing ? '编辑变量' : '新增变量'" width="620px" @closed="resetVariableForm">
      <el-form ref="variableFormRef" :model="variableForm" :rules="variableRules" label-width="110px">
        <el-form-item label="变量名" prop="var_name">
          <el-input v-model="variableForm.var_name" placeholder="如 HTTP_PORT" />
        </el-form-item>
        <el-form-item label="显示标签" prop="var_label">
          <el-input v-model="variableForm.var_label" placeholder="如 HTTP 监听端口" />
        </el-form-item>
        <el-form-item label="变量类型" prop="var_type">
          <el-select v-model="variableForm.var_type" style="width: 100%">
            <el-option label="字符串" value="string" />
            <el-option label="密码" value="password" />
            <el-option label="数字" value="number" />
            <el-option label="布尔" value="bool" />
            <el-option label="下拉选择" value="select" />
          </el-select>
        </el-form-item>
        <el-form-item label="默认值">
          <el-input v-model="variableForm.default_value" placeholder="可选" />
        </el-form-item>
        <el-form-item label="必填">
          <el-switch v-model="variableForm.required" />
        </el-form-item>
        <el-form-item label="分组">
          <el-input v-model="variableForm.var_group" placeholder="如 基础配置" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="variableForm.sort_order" :min="0" />
        </el-form-item>
        <el-form-item label="选项 JSON">
          <el-input v-model="variableForm.options" type="textarea" :rows="2" placeholder='select 类型可填：["DEBUG","INFO","WARN"]' />
        </el-form-item>
        <el-form-item label="校验 JSON">
          <el-input v-model="variableForm.validation_rule" type="textarea" :rows="2" placeholder='如：{"min":1,"max":65535}' />
        </el-form-item>
        <el-form-item label="关联变量">
          <el-input v-model="variableForm.linked_to" placeholder="如 postgresql.PORT" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="variableForm.description" type="textarea" :rows="3" placeholder="变量说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="variableDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="variableSubmitting" @click="handleVariableSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, type FormRules } from 'element-plus'
import PageHeader from '../components/PageHeader.vue'
import PageCard from '../components/PageCard.vue'
import {
  getComponents,
  createComponent,
  updateComponent,
  deleteComponent,
  getComponentVariables,
  createComponentVariable,
  updateComponentVariable,
  deleteComponentVariable,
  importComponentVariables,
  type ComponentVariablePayload,
} from '../api'

interface ComponentVariableItem extends ComponentVariablePayload {
  id: string
  required: boolean
  sort_order: number
}

interface ComponentItem {
  id: string
  name: string
  display_name: string
  category: string
  description: string
  template_dir: string
  is_active: boolean
  variables?: ComponentVariableItem[]
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

const form = ref({ name: '', display_name: '', category: '', description: '', template_dir: '', is_active: true })
const rules = {
  name: [{ required: true, message: '请输入组件标识', trigger: 'blur' }],
  display_name: [{ required: true, message: '请输入组件名称', trigger: 'blur' }],
}

const variableDrawerVisible = ref(false)
const variableTarget = ref<ComponentItem | null>(null)
const variables = ref<ComponentVariableItem[]>([])
const variableLoading = ref(false)
const importing = ref(false)
const variableDialogVisible = ref(false)
const isVariableEditing = ref(false)
const editingVariableId = ref('')
const variableSubmitting = ref(false)
const variableFormRef = ref()
const variableForm = ref<ComponentVariablePayload>({
  var_name: '',
  var_label: '',
  var_type: 'string',
  default_value: '',
  required: false,
  validation_rule: '',
  var_group: '',
  sort_order: 0,
  description: '',
  options: '',
  linked_to: '',
})
const variableRules: FormRules = {
  var_name: [{ required: true, message: '请输入变量名', trigger: 'blur' }],
  var_label: [{ required: true, message: '请输入显示标签', trigger: 'blur' }],
  var_type: [{ required: true, message: '请选择变量类型', trigger: 'change' }],
}

const getCategoryLabel = (category: string) => categoryOptions.find(c => c.value === category)?.label || category || '未分类'
const getVariableGroups = (component: ComponentItem) => Array.from(new Set((component.variables || []).map(v => v.var_group || '基础配置')))
const variableGroups = computed(() => Array.from(new Set(variables.value.map(v => v.var_group || '基础配置'))))
const getVariablesByGroup = (group: string) => variables.value
  .filter(v => (v.var_group || '基础配置') === group)
  .sort((a, b) => (a.sort_order || 0) - (b.sort_order || 0))

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
  form.value = { name: row.name, display_name: row.display_name, category: row.category, description: row.description, template_dir: row.template_dir, is_active: row.is_active }
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

const resetForm = () => formRef.value?.resetFields()

const openVariableDrawer = async (row: ComponentItem) => {
  variableTarget.value = row
  variableDrawerVisible.value = true
  await fetchVariables()
}

const fetchVariables = async () => {
  if (!variableTarget.value) return
  variableLoading.value = true
  try {
    const res: any = await getComponentVariables(variableTarget.value.id)
    variables.value = res.data || []
  } catch {
    ElMessage.error('获取变量列表失败')
  } finally {
    variableLoading.value = false
  }
}

const openVariableDialog = (row?: ComponentVariableItem) => {
  isVariableEditing.value = !!row
  editingVariableId.value = row?.id || ''
  variableForm.value = row
    ? {
        var_name: row.var_name,
        var_label: row.var_label,
        var_type: row.var_type,
        default_value: row.default_value || '',
        required: row.required,
        validation_rule: row.validation_rule || '',
        var_group: row.var_group || '',
        sort_order: row.sort_order || 0,
        description: row.description || '',
        options: row.options || '',
        linked_to: row.linked_to || '',
      }
    : { var_name: '', var_label: '', var_type: 'string', default_value: '', required: false, validation_rule: '', var_group: '', sort_order: variables.value.length + 1, description: '', options: '', linked_to: '' }
  variableDialogVisible.value = true
}

const handleVariableSubmit = async () => {
  if (!variableTarget.value) return
  const valid = await variableFormRef.value?.validate().catch(() => false)
  if (!valid) return
  variableSubmitting.value = true
  try {
    if (isVariableEditing.value) {
      await updateComponentVariable(variableTarget.value.id, editingVariableId.value, variableForm.value)
      ElMessage.success('变量更新成功')
    } else {
      await createComponentVariable(variableTarget.value.id, variableForm.value)
      ElMessage.success('变量创建成功')
    }
    variableDialogVisible.value = false
    await fetchVariables()
    fetchComponents()
  } catch {
    ElMessage.error(isVariableEditing.value ? '变量更新失败' : '变量创建失败')
  } finally {
    variableSubmitting.value = false
  }
}

const handleDeleteVariable = async (variableId: string) => {
  if (!variableTarget.value) return
  try {
    await deleteComponentVariable(variableTarget.value.id, variableId)
    ElMessage.success('变量删除成功')
    await fetchVariables()
    fetchComponents()
  } catch {
    ElMessage.error('变量删除失败')
  }
}

const importVariables = async (overwrite: boolean) => {
  if (!variableTarget.value) return
  importing.value = true
  try {
    await importComponentVariables(variableTarget.value.id, { overwrite })
    ElMessage.success(overwrite ? '模板变量已覆盖同步' : '模板变量已导入')
    await fetchVariables()
    fetchComponents()
  } catch {
    ElMessage.error('导入模板变量失败')
  } finally {
    importing.value = false
  }
}

const resetVariableForm = () => variableFormRef.value?.resetFields()
onMounted(fetchComponents)
</script>

<style scoped>
.components-page { max-width: 1440px; margin: 0 auto; }
.component-name { font-weight: 650; color: var(--itcfg-text-primary); }
.component-desc { margin-top: 6px; color: var(--itcfg-text-secondary); font-size: 13px; line-height: 1.5; }
.variable-toolbar { display: flex; align-items: center; justify-content: space-between; gap: 16px; margin-bottom: 16px; padding: 14px 16px; border: 1px solid var(--itcfg-border); border-radius: 16px; background: var(--itcfg-surface-soft); }
.variable-groups { display: flex; flex-direction: column; gap: 16px; }
.variable-group { padding: 16px; border: 1px solid var(--itcfg-border); border-radius: 16px; background: #fff; }
.group-header { display: flex; justify-content: space-between; gap: 16px; margin-bottom: 12px; padding-bottom: 10px; border-bottom: 1px solid var(--itcfg-border); }
.group-header h3 { margin: 0; font-size: 16px; color: var(--itcfg-text-primary); }
.group-header p { margin: 5px 0 0; color: var(--itcfg-text-secondary); font-size: 12px; }
</style>
