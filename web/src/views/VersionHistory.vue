<template>
  <div>
    <el-page-header @back="$router.back()" title="返回">
      <template #content>
        <span style="font-size: 16px; font-weight: 500">配置版本历史</span>
      </template>
    </el-page-header>

    <el-card style="margin-top: 20px">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>版本列表</span>
          <el-button type="primary" @click="createSnapshotHandler">
            <el-icon><Camera /></el-icon> 保存当前版本
          </el-button>
        </div>
      </template>

      <el-table :data="versions" stripe v-loading="loading">
        <el-table-column prop="version" label="版本号" width="100">
          <template #default="{ row }">
            <el-tag type="primary">v{{ row.version }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="change_summary" label="变更说明" min-width="200" />
        <el-table-column prop="created_by" label="操作人" width="120" />
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ new Date(row.created_at).toLocaleString('zh-CN') }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="viewDiff(row)">对比差异</el-button>
            <el-popconfirm
              :title="`确认回滚到版本 ${row.version}?`"
              @confirm="handleRollback(row)"
            >
              <template #reference>
                <el-button size="small" type="warning">回滚</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 差异对比对话框 -->
    <el-dialog v-model="showDiff" title="版本差异对比" width="800px">
      <el-form inline>
        <el-form-item label="对比版本">
          <el-input-number v-model="diffFrom" :min="1" />
        </el-form-item>
        <el-form-item label="→">
          <el-input-number v-model="diffTo" :min="1" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadDiff">对比</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="diffList" stripe>
        <el-table-column prop="key" label="变量ID" width="300" />
        <el-table-column prop="oldValue" label="旧值">
          <template #default="{ row }">
            <span style="color: #F56C6C">{{ row.oldValue || '(空)' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="newValue" label="新值">
          <template #default="{ row }">
            <span style="color: #67C23A">{{ row.newValue || '(空)' }}</span>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getVersions, createSnapshot, diffVersions, rollbackVersion } from '../api'

interface VersionItem {
  version: number
  change_summary: string
  created_by: string
  created_at: string
}

const route = useRoute()
const envId = route.params.envId as string
const versions = ref<VersionItem[]>([])
const loading = ref(false)
const showDiff = ref(false)
const diffFrom = ref(0)
const diffTo = ref(0)
const diffList = ref<any[]>([])

const fetchVersions = async () => {
  loading.value = true
  try {
    const res: any = await getVersions(envId)
    versions.value = res.data || []
  } catch {
    ElMessage.error('获取版本列表失败')
  } finally {
    loading.value = false
  }
}

const createSnapshotHandler = async () => {
  try {
    await createSnapshot(envId, {
      created_by: 'admin',
      change_summary: prompt('请输入变更说明') || '手动保存',
    })
    ElMessage.success('快照保存成功')
    fetchVersions()
  } catch {
    ElMessage.error('保存快照失败')
  }
}

const viewDiff = (row: any) => {
  diffFrom.value = row.version - 1 > 0 ? row.version - 1 : 1
  diffTo.value = row.version
  showDiff.value = true
}

const loadDiff = async () => {
  try {
    const res: any = await diffVersions(envId, diffFrom.value, diffTo.value)
    const data = res.data || {}
    diffList.value = Object.entries(data).map(([key, val]: any) => ({
      key,
      oldValue: val.old_value,
      newValue: val.new_value,
    }))
  } catch {
    ElMessage.error('加载差异失败')
  }
}

const handleRollback = async (row: any) => {
  try {
    await rollbackVersion(envId, {
      target_version: row.version,
      operator: 'admin',
    })
    ElMessage.success(`已回滚到版本 ${row.version}`)
    fetchVersions()
  } catch {
    ElMessage.error('回滚失败')
  }
}

onMounted(fetchVersions)
</script>