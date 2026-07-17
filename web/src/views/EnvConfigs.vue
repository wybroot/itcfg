<template>
  <div class="page configs-page">
    <PageHeader title="配置管理" :subtitle="`环境 ID：${envId}`" back>
      <template #actions>
        <el-space wrap>
          <el-button @click="$router.push(`/envs/${envId}/versions`)">
            <el-icon><Clock /></el-icon> 版本历史
          </el-button>
          <el-button type="warning" plain @click="$router.push(`/envs/${envId}/artifacts`)">
            <el-icon><Box /></el-icon> 制品版本
          </el-button>
          <el-button type="info" plain @click="$router.push(`/envs/${envId}/deploy-records`)">
            <el-icon><Clock /></el-icon> 部署记录
          </el-button>
          <el-button type="primary" plain :disabled="!activeComponent" :loading="previewLoading" @click="previewConfig">
            <el-icon><View /></el-icon> 预览
          </el-button>
          <el-button type="success" :disabled="!activeComponent" :loading="saving" @click="saveConfigs">
            <el-icon><Check /></el-icon> 保存配置
          </el-button>
        </el-space>
      </template>
    </PageHeader>

    <PageCard title="组件配置" subtitle="按当前环境已启用组件维护变量，保存时只提交当前组件配置">
      <el-empty v-if="!loading && components.length === 0" description="当前环境还没有启用组件">
        <el-button type="primary" @click="$router.back()">返回环境管理</el-button>
      </el-empty>

      <div v-else v-loading="loading" class="config-workbench">
        <el-tabs v-model="activeComponent" class="component-tabs" @tab-change="onComponentChange">
          <el-tab-pane v-for="comp in components" :key="comp.id" :label="comp.display_name" :name="comp.id" />
        </el-tabs>

        <div v-if="currentComponent" class="component-summary">
          <div>
            <div class="summary-title">{{ currentComponent.display_name }}</div>
            <div class="summary-subtitle code-text">{{ currentComponent.template_dir || currentComponent.name }}</div>
          </div>
          <el-tag effect="light" type="primary">{{ currentVariables.length }} 个变量</el-tag>
        </div>

        <el-empty v-if="currentComponent && currentVariables.length === 0" description="该组件暂无可配置变量" />

        <div v-else class="variable-groups">
          <section v-for="group in variableGroups" :key="group" class="variable-group">
            <div class="group-header">
              <div>
                <h3>{{ group }}</h3>
                <p>{{ getVariablesByGroup(group).length }} 个配置项</p>
              </div>
            </div>

            <el-form label-width="160px" label-position="left" class="config-form">
              <el-form-item v-for="v in getVariablesByGroup(group)" :key="v.id" :label="v.var_label" :required="v.required">
                <div class="field-control">
                  <el-input v-if="v.var_type === 'string'" v-model="configValues[v.id]" :placeholder="`请输入${v.var_label}`" />

                  <template v-else-if="v.var_type === 'password'">
                    <div v-if="configValues[v.id] === '***ENCRYPTED***'" class="encrypted-field">
                      <el-tag type="warning" size="small">已加密存储</el-tag>
                      <el-button size="small" text type="primary" @click="configValues[v.id] = ''">设置新密码</el-button>
                    </div>
                    <el-input v-else v-model="configValues[v.id]" type="password" show-password :placeholder="`请输入${v.var_label}`" />
                  </template>

                  <el-input-number v-else-if="v.var_type === 'number'" v-model="configValues[v.id]" :min="getMin(v)" :max="getMax(v)" />

                  <el-switch v-else-if="v.var_type === 'boolean' || v.var_type === 'bool'" v-model="configValues[v.id]" active-value="true" inactive-value="false" />

                  <el-select v-else-if="v.var_type === 'select'" v-model="configValues[v.id]" :placeholder="`请选择${v.var_label}`" style="width: 100%">
                    <el-option v-for="opt in getOptions(v)" :key="opt" :label="opt" :value="opt" />
                  </el-select>

                  <el-input v-else v-model="configValues[v.id]" :placeholder="`请输入${v.var_label}`" />

                  <div v-if="v.description" class="field-description">{{ v.description }}</div>
                </div>
              </el-form-item>
            </el-form>
          </section>
        </div>
      </div>
    </PageCard>

    <el-dialog v-model="showPreview" title="配置预览" width="86vw" class="preview-dialog">
      <el-empty v-if="Object.keys(previewData).length === 0" description="暂无预览内容" />
      <el-tabs v-else v-model="previewTab" type="card">
        <el-tab-pane v-for="(content, path) in previewData" :key="path" :label="path" :name="path">
          <pre class="preview-code"><code>{{ content }}</code></pre>
        </el-tab-pane>
      </el-tabs>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import PageHeader from '../components/PageHeader.vue'
import PageCard from '../components/PageCard.vue'
import { getEnvComponents, getEnvConfigs, updateEnvConfigs, previewConfigs } from '../api'

const route = useRoute()
const envId = route.params.envId as string

const components = ref<any[]>([])
const activeComponent = ref('')
const currentVariables = ref<any[]>([])
const configValues = ref<Record<string, string>>({})
const showPreview = ref(false)
const previewTab = ref('')
const previewData = ref<Record<string, string>>({})
const loading = ref(false)
const saving = ref(false)
const previewLoading = ref(false)

const currentComponent = computed(() => components.value.find(c => c.id === activeComponent.value))

const variableGroups = computed(() => {
  const groups = new Set<string>()
  currentVariables.value.forEach(v => groups.add(v.var_group || '基础配置'))
  return Array.from(groups)
})

const getVariablesByGroup = (group: string) => {
  return currentVariables.value
    .filter(v => (v.var_group || '基础配置') === group)
    .sort((a, b) => a.sort_order - b.sort_order)
}

const getOptions = (v: any) => {
  if (typeof v.options === 'string') {
    try { return JSON.parse(v.options) } catch { return [] }
  }
  return v.options || []
}

const getMin = (v: any) => {
  if (typeof v.validation_rule === 'string') {
    try { return JSON.parse(v.validation_rule).min } catch { return undefined }
  }
  return v.validation_rule?.min
}

const getMax = (v: any) => {
  if (typeof v.validation_rule === 'string') {
    try { return JSON.parse(v.validation_rule).max } catch { return undefined }
  }
  return v.validation_rule?.max
}

const fetchComponents = async () => {
  loading.value = true
  try {
    const res: any = await getEnvComponents(envId)
    components.value = (res.data || []).map((item: any) => item.component).filter(Boolean)
    if (components.value.length > 0) {
      activeComponent.value = components.value[0].id
      currentVariables.value = components.value[0].variables || []
      await fetchConfigValues()
    } else {
      activeComponent.value = ''
      currentVariables.value = []
      configValues.value = {}
    }
  } catch {
    ElMessage.error('获取环境组件失败')
  } finally {
    loading.value = false
  }
}

const fetchVariables = async () => {
  if (!activeComponent.value) return
  const component = currentComponent.value
  currentVariables.value = component?.variables || []
  await fetchConfigValues()
}

const fetchConfigValues = async () => {
  try {
    const res: any = await getEnvConfigs(envId)
    const values: Record<string, string> = {}
    currentVariables.value.forEach(v => { values[v.id] = v.default_value || '' })
    ;(res.data || []).forEach((item: any) => { values[item.variable_id] = item.var_value })
    configValues.value = values
  } catch {
    configValues.value = Object.fromEntries(currentVariables.value.map(v => [v.id, v.default_value || '']))
  }
}

const onComponentChange = () => { fetchVariables() }

const saveConfigs = async () => {
  if (!activeComponent.value) return
  const cleanValues: Record<string, string> = {}
  for (const [key, val] of Object.entries(configValues.value)) {
    if (val !== '***ENCRYPTED***') cleanValues[key] = String(val ?? '')
  }
  saving.value = true
  try {
    await updateEnvConfigs(envId, activeComponent.value, { values: cleanValues, updated_by: 'admin' })
    ElMessage.success('配置保存成功')
  } catch {
    ElMessage.error('配置保存失败')
  } finally {
    saving.value = false
  }
}

const previewConfig = async () => {
  const comp = currentComponent.value
  if (!comp) return
  previewLoading.value = true
  try {
    const res: any = await previewConfigs(envId, comp.name)
    previewData.value = res.data || {}
    previewTab.value = Object.keys(previewData.value)[0] || ''
    showPreview.value = true
  } catch {
    ElMessage.error('配置预览失败')
  } finally {
    previewLoading.value = false
  }
}

onMounted(fetchComponents)
</script>

<style scoped>
.configs-page { max-width: 1440px; margin: 0 auto; }
.config-workbench { min-height: 360px; }
.component-tabs { margin-bottom: 18px; }
.component-summary { display: flex; align-items: center; justify-content: space-between; gap: 16px; margin-bottom: 18px; padding: 16px 18px; border: 1px solid var(--itcfg-border); border-radius: 16px; background: var(--itcfg-surface-soft); }
.summary-title { font-size: 18px; font-weight: 700; color: var(--itcfg-text-primary); }
.summary-subtitle { margin-top: 4px; color: var(--itcfg-text-secondary); }
.variable-groups { display: flex; flex-direction: column; gap: 16px; }
.variable-group { padding: 18px; border: 1px solid var(--itcfg-border); border-radius: 18px; background: #fff; }
.group-header { display: flex; justify-content: space-between; gap: 16px; margin-bottom: 14px; padding-bottom: 12px; border-bottom: 1px solid var(--itcfg-border); }
.group-header h3 { margin: 0; color: var(--itcfg-text-primary); font-size: 16px; }
.group-header p { margin: 5px 0 0; color: var(--itcfg-text-secondary); font-size: 12px; }
.config-form :deep(.el-form-item) { margin-bottom: 18px; }
.field-control { width: min(640px, 100%); }
.field-description { margin-top: 6px; color: var(--itcfg-text-secondary); font-size: 12px; line-height: 1.5; }
.encrypted-field { display: flex; align-items: center; gap: 8px; }
.preview-code { min-height: 520px; max-height: 68vh; margin: 0; padding: 18px; overflow: auto; color: #d4d4d8; border-radius: 14px; background: #0f172a; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: 13px; line-height: 1.65; }
</style>
