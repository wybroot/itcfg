<template>
  <div>
    <el-page-header @back="$router.push('/')" title="返回">
      <template #content>
        <span style="font-size: 16px; font-weight: 500">组件模板管理</span>
      </template>
    </el-page-header>

    <el-card style="margin-top: 20px">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>模板列表 ({{ templates.length }})</span>
          <el-button size="small" @click="fetchTemplates" :loading="loading">刷新</el-button>
        </div>
      </template>

      <el-table :data="templates" stripe v-loading="loading" row-key="name">
        <el-table-column prop="name" label="模板名称" width="160">
          <template #default="{ row }">
            <el-tag type="primary">{{ row.name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="display_name" label="显示名称" width="140" />
        <el-table-column prop="category" label="分类" width="120">
          <template #default="{ row }">
            <el-tag size="small" :type="categoryType(row.category)">
              {{ row.category || '未分类' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="200">
          <template #default="{ row }">
            {{ row.description || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="var_count" label="变量数" width="90" align="center" />
        <el-table-column prop="file_count" label="配置文件" width="100" align="center" />
        <el-table-column prop="output_dir" label="输出目录" width="180">
          <template #default="{ row }">
            <code style="font-size: 12px">{{ row.output_dir || '-' }}</code>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getTemplates } from '../api'

const templates = ref<any[]>([])
const loading = ref(false)

const categoryType = (cat: string) => {
  const map: Record<string, string> = {
    '数据库': 'success',
    '缓存': 'warning',
    '消息队列': 'danger',
    'Web服务器': 'primary',
    '应用服务': '',
    '对象存储': 'info',
    '搜索引擎': 'danger',
    '办公套件': '',
  }
  return map[cat] || 'info'
}

const fetchTemplates = async () => {
  loading.value = true
  try {
    const res: any = await getTemplates()
    templates.value = res.data || []
  } catch {
    ElMessage.error('获取模板列表失败')
  } finally {
    loading.value = false
  }
}

onMounted(fetchTemplates)
</script>