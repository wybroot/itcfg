<template>
  <div>
    <el-card>
      <template #header>
        <span>组件列表</span>
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
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getComponents } from '../api'

interface ComponentItem {
  name: string
  display_name: string
  category: string
  description: string
  is_active: boolean
}

const components = ref<ComponentItem[]>([])
const loading = ref(false)

const getCategoryLabel = (category: string) => {
  const labels: Record<string, string> = {
    'web-server': 'Web 服务器',
    'application': '应用服务',
    'database': '数据库',
    'cache': '缓存',
    'message-queue': '消息队列',
    'object-storage': '对象存储',
    'coordination': '协调服务',
    'search-engine': '搜索引擎',
    'office': '办公套件',
    'file-service': '文件服务',
  }
  return labels[category] || category
}

onMounted(async () => {
  loading.value = true
  try {
    const res: any = await getComponents()
    components.value = res.data || []
  } catch {
    // ignore
  } finally {
    loading.value = false
  }
})
</script>