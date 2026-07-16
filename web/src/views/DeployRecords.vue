<template>
  <div>
    <el-page-header @back="$router.back()" title="返回">
      <template #content>
        <span style="font-size: 16px; font-weight: 500">部署记录</span>
      </template>
    </el-page-header>

    <el-card style="margin-top: 20px">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>部署历史</span>
          <el-button size="small" @click="fetchRecords" :loading="loading">刷新</el-button>
        </div>
      </template>

      <el-table :data="records" stripe v-loading="loading" empty-text="暂无部署记录">
        <el-table-column prop="version_tag" label="版本标签" width="180">
          <template #default="{ row }">
            <el-tag>{{ row.version_tag }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'success' ? 'success' : row.status === 'failed' ? 'danger' : 'info'" size="small">
              {{ row.status === 'success' ? '成功' : row.status === 'failed' ? '失败' : row.status || '未知' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="deployed_by" label="部署人" width="120" />
        <el-table-column prop="notes" label="备注" min-width="200">
          <template #default="{ row }">
            {{ row.notes || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="deployed_at" label="部署时间" width="200">
          <template #default="{ row }">
            {{ row.deployed_at ? new Date(row.deployed_at).toLocaleString('zh-CN') : '-' }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getDeployRecords } from '../api'

const route = useRoute()
const envId = route.params.envId as string
const records = ref([])
const loading = ref(false)

const fetchRecords = async () => {
  loading.value = true
  try {
    const res: any = await getDeployRecords(envId)
    records.value = res.data || []
  } catch {
    ElMessage.error('获取部署记录失败')
  } finally {
    loading.value = false
  }
}

onMounted(fetchRecords)
</script>