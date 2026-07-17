<template>
  <div class="page templates-page">
    <PageHeader title="组件模板目录" subtitle="从内置模板同步真实组件变量、依赖和配置文件输出" back>
      <template #actions>
        <el-space wrap>
          <el-button @click="fetchTemplates" :loading="loading">刷新</el-button>
          <el-button type="primary" @click="syncAllTemplates" :loading="syncing">
            <el-icon><Refresh /></el-icon> 同步全部模板
          </el-button>
        </el-space>
      </template>
    </PageHeader>

    <PageCard title="可用组件模板" :subtitle="`共 ${templates.length} 个模板，已注册 ${registeredCount} 个`">
      <el-table :data="templates" stripe v-loading="loading" row-key="template_dir" empty-text="未发现有效模板">
        <el-table-column label="模板" min-width="220">
          <template #default="{ row }">
            <div class="template-name">{{ row.display_name || row.name }}</div>
            <div class="muted code-text">{{ row.template_dir || row.name }}</div>
            <div v-if="row.description" class="template-desc">{{ row.description }}</div>
          </template>
        </el-table-column>
        <el-table-column label="分类" width="130">
          <template #default="{ row }">
            <el-tag size="small" effect="light">{{ getCategoryLabel(row.category) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="变量分组" min-width="190">
          <template #default="{ row }">
            <el-space wrap v-if="row.variable_groups?.length">
              <el-tag v-for="group in row.variable_groups" :key="group" size="small" type="info" effect="plain">{{ group }}</el-tag>
            </el-space>
            <span v-else class="muted">无分组</span>
          </template>
        </el-table-column>
        <el-table-column label="依赖" min-width="170">
          <template #default="{ row }">
            <el-space wrap v-if="row.dependencies?.length">
              <el-tag v-for="dep in row.dependencies" :key="dep.component" size="small" type="warning" effect="light">{{ dep.component }}</el-tag>
            </el-space>
            <span v-else class="muted">无</span>
          </template>
        </el-table-column>
        <el-table-column label="规模" width="150" align="center">
          <template #default="{ row }">
            <div>{{ row.var_count || 0 }} 个变量</div>
            <div class="muted">{{ row.file_count || 0 }} 个配置文件</div>
          </template>
        </el-table-column>
        <el-table-column label="注册状态" width="120">
          <template #default="{ row }">
            <el-tag :type="row.registered ? 'success' : 'info'" effect="light">{{ row.registered ? '已注册' : '未注册' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="230" fixed="right">
          <template #default="{ row }">
            <el-space wrap>
              <el-button size="small" type="primary" plain @click="openDetail(row)">查看</el-button>
              <el-button size="small" type="success" @click="syncTemplate(row)" :loading="syncingTemplate === row.template_dir">
                {{ row.registered ? '同步' : '导入' }}
              </el-button>
            </el-space>
          </template>
        </el-table-column>
      </el-table>
    </PageCard>

    <el-drawer v-model="detailVisible" :title="detailTemplate ? `${detailTemplate.display_name || detailTemplate.name} 模板详情` : '模板详情'" size="820px">
      <div v-if="detailTemplate" class="template-detail">
        <div class="detail-summary">
          <div>
            <div class="template-name">{{ detailTemplate.display_name || detailTemplate.name }}</div>
            <div class="muted code-text">{{ detailTemplate.template_dir }}</div>
            <div v-if="detailTemplate.description" class="template-desc">{{ detailTemplate.description }}</div>
          </div>
          <el-space wrap>
            <el-tag type="primary" effect="light">{{ detailTemplate.var_count || 0 }} 个变量</el-tag>
            <el-tag type="success" effect="light">{{ detailTemplate.file_count || 0 }} 个文件</el-tag>
            <el-tag :type="detailTemplate.registered ? 'success' : 'info'" effect="light">{{ detailTemplate.registered ? '已注册' : '未注册' }}</el-tag>
          </el-space>
        </div>

        <PageCard title="变量定义" :subtitle="variablesLoading ? '正在读取模板变量' : `共 ${templateVariables.length} 个变量`">
          <el-table :data="templateVariables" stripe size="small" v-loading="variablesLoading" empty-text="该模板暂无变量定义">
            <el-table-column prop="name" label="变量名" min-width="150">
              <template #default="{ row }"><span class="code-text">{{ row.name }}</span></template>
            </el-table-column>
            <el-table-column prop="label" label="标签" min-width="150" />
            <el-table-column prop="type" label="类型" width="90">
              <template #default="{ row }"><el-tag size="small" effect="light">{{ row.type || 'string' }}</el-tag></template>
            </el-table-column>
            <el-table-column prop="group" label="分组" width="120" />
            <el-table-column prop="default" label="默认值" min-width="120">
              <template #default="{ row }">
                <span v-if="row.default" class="code-text">{{ row.default }}</span>
                <span v-else class="muted">空</span>
              </template>
            </el-table-column>
            <el-table-column prop="required" label="必填" width="70">
              <template #default="{ row }">{{ row.required ? '是' : '否' }}</template>
            </el-table-column>
            <el-table-column prop="description" label="说明" min-width="220" />
          </el-table>
        </PageCard>

        <PageCard title="配置文件输出" subtitle="模板渲染后会生成这些目标配置文件">
          <el-table :data="detailTemplate.config_files || []" size="small" empty-text="暂无配置文件">
            <el-table-column prop="path" label="模板文件" min-width="180">
              <template #default="{ row }"><span class="code-text">{{ row.path }}</span></template>
            </el-table-column>
            <el-table-column prop="target" label="目标路径" min-width="220">
              <template #default="{ row }"><span class="code-text">{{ row.target }}</span></template>
            </el-table-column>
            <el-table-column prop="owner" label="Owner" width="100" />
            <el-table-column prop="mode" label="Mode" width="90" />
          </el-table>
        </PageCard>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import PageHeader from '../components/PageHeader.vue'
import PageCard from '../components/PageCard.vue'
import { getTemplates, getTemplateVariables, syncTemplates } from '../api'

const templates = ref<any[]>([])
const loading = ref(false)
const syncing = ref(false)
const syncingTemplate = ref('')
const detailVisible = ref(false)
const detailTemplate = ref<any | null>(null)
const templateVariables = ref<any[]>([])
const variablesLoading = ref(false)

const registeredCount = computed(() => templates.value.filter(t => t.registered).length)

const categoryLabels: Record<string, string> = {
  'web-server': 'Web 服务器',
  application: '应用服务',
  database: '数据库',
  cache: '缓存',
  'message-queue': '消息队列',
  'object-storage': '对象存储',
  coordination: '协调服务',
  'search-engine': '搜索引擎',
  office: '办公套件',
  'file-service': '文件服务',
}

const getCategoryLabel = (category: string) => categoryLabels[category] || category || '未分类'

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

const syncAllTemplates = async () => {
  syncing.value = true
  try {
    await syncTemplates({ overwrite: false })
    ElMessage.success('全部模板已同步为组件')
    await fetchTemplates()
  } catch {
    ElMessage.error('同步模板失败')
  } finally {
    syncing.value = false
  }
}

const syncTemplate = async (row: any) => {
  syncingTemplate.value = row.template_dir
  try {
    await syncTemplates({ templates: [row.template_dir], overwrite: false })
    ElMessage.success(`${row.display_name || row.name} 已同步`)
    await fetchTemplates()
  } catch {
    ElMessage.error('同步模板失败')
  } finally {
    syncingTemplate.value = ''
  }
}

const openDetail = async (row: any) => {
  detailTemplate.value = row
  detailVisible.value = true
  variablesLoading.value = true
  templateVariables.value = []
  try {
    const res: any = await getTemplateVariables(row.template_dir)
    templateVariables.value = res.data || []
  } catch {
    ElMessage.error('获取模板变量失败')
  } finally {
    variablesLoading.value = false
  }
}

onMounted(fetchTemplates)
</script>

<style scoped>
.templates-page { max-width: 1440px; margin: 0 auto; }
.template-name { font-weight: 700; color: var(--itcfg-text-primary); }
.template-desc { margin-top: 6px; color: var(--itcfg-text-secondary); font-size: 13px; line-height: 1.5; }
.template-detail { display: flex; flex-direction: column; gap: 16px; }
.detail-summary { display: flex; align-items: center; justify-content: space-between; gap: 16px; padding: 16px; border: 1px solid var(--itcfg-border); border-radius: 16px; background: var(--itcfg-surface-soft); }
</style>
