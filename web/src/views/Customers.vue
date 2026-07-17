<template>
  <div class="page customers-page">
    <PageHeader title="客户管理" subtitle="维护客户信息，并进入客户环境配置">
      <template #actions>
        <el-button type="primary" @click="showCreateDialog = true">
          <el-icon><Plus /></el-icon> 新建客户
        </el-button>
      </template>
    </PageHeader>

    <PageCard title="客户列表" :subtitle="`共 ${customers.length} 个客户`">
      <el-table :data="customers" stripe v-loading="loading" empty-text="暂无客户，请先新建客户">
        <el-table-column label="客户" min-width="180">
          <template #default="{ row }">
            <div class="customer-name">{{ row.name }}</div>
            <div class="customer-code code-text">{{ row.code }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="contact" label="联系人" width="140">
          <template #default="{ row }">
            <span v-if="row.contact">{{ row.contact }}</span>
            <span v-else class="muted">未填写</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'" effect="light">
              {{ row.status === 'active' ? '正常' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160">
          <template #default="{ row }">
            {{ new Date(row.created_at).toLocaleDateString('zh-CN') }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-space>
              <el-button size="small" type="primary" plain @click="viewEnvs(row)">
                <el-icon><Setting /></el-icon> 环境
              </el-button>
              <el-button size="small" @click="editCustomer(row)">编辑</el-button>
              <el-popconfirm title="确认删除该客户?" @confirm="handleDelete(row.id)">
                <template #reference>
                  <el-button size="small" type="danger" plain>删除</el-button>
                </template>
              </el-popconfirm>
            </el-space>
          </template>
        </el-table-column>
      </el-table>
    </PageCard>

    <el-dialog :title="isEditing ? '编辑客户' : '新建客户'" v-model="showCreateDialog" width="500px" @close="resetForm">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="客户名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入客户名称" />
        </el-form-item>
        <el-form-item label="客户编码" prop="code">
          <el-input v-model="form.code" placeholder="请输入唯一编码" :disabled="isEditing" />
        </el-form-item>
        <el-form-item label="联系人" prop="contact">
          <el-input v-model="form.contact" placeholder="请输入联系人" />
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
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import PageHeader from '../components/PageHeader.vue'
import PageCard from '../components/PageCard.vue'
import { getCustomers, createCustomer, updateCustomer, deleteCustomer } from '../api'

interface CustomerItem {
  id: string
  name: string
  code: string
  contact: string
  status: string
  created_at: string
}

const router = useRouter()
const customers = ref<CustomerItem[]>([])
const loading = ref(false)
const showCreateDialog = ref(false)
const isEditing = ref(false)
const editingId = ref('')
const submitting = ref(false)
const formRef = ref()
const form = ref({ name: '', code: '', contact: '' })
const rules = {
  name: [{ required: true, message: '请输入客户名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入客户编码', trigger: 'blur' }],
}

const fetchCustomers = async () => {
  loading.value = true
  try {
    const res: any = await getCustomers()
    customers.value = res.data || []
  } catch {
    ElMessage.error('获取客户列表失败')
  } finally {
    loading.value = false
  }
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    if (isEditing.value) {
      await updateCustomer(editingId.value, form.value)
      ElMessage.success('客户更新成功')
    } else {
      await createCustomer(form.value)
      ElMessage.success('客户创建成功')
    }
    showCreateDialog.value = false
    fetchCustomers()
  } catch {
    ElMessage.error('操作失败')
  } finally {
    submitting.value = false
  }
}

const editCustomer = (row: CustomerItem) => {
  isEditing.value = true
  editingId.value = row.id
  form.value = { name: row.name, code: row.code, contact: row.contact || '' }
  showCreateDialog.value = true
}

const handleDelete = async (id: string) => {
  try {
    await deleteCustomer(id)
    ElMessage.success('删除成功')
    fetchCustomers()
  } catch {
    ElMessage.error('删除失败')
  }
}

const viewEnvs = (row: CustomerItem) => router.push(`/customers/${row.id}/envs`)
const resetForm = () => {
  isEditing.value = false
  editingId.value = ''
  form.value = { name: '', code: '', contact: '' }
  formRef.value?.resetFields()
}

onMounted(fetchCustomers)
</script>

<style scoped>
.customers-page { max-width: 1320px; margin: 0 auto; }
.customer-name { font-weight: 650; color: var(--itcfg-text-primary); }
.customer-code { margin-top: 4px; color: var(--itcfg-text-secondary); font-size: 12px; }
</style>
