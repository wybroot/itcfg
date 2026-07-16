<template>
  <div>
    <el-page-header @back="$router.back()" title="返回">
      <template #content>
        <span style="font-size: 16px; font-weight: 500">配置管理</span>
      </template>
    </el-page-header>

    <el-card style="margin-top: 20px">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>组件配置</span>
          <el-space>
            <el-button type="info" @click="$router.push(`/envs/${envId}/versions`)">
              <el-icon><Clock /></el-icon> 版本历史
            </el-button>
            <el-button type="warning" @click="$router.push(`/envs/${envId}/artifacts`)">
              <el-icon><Box /></el-icon> 制品版本
            </el-button>
            <el-button type="danger" plain @click="$router.push(`/envs/${envId}/deploy-records`)">
              <el-icon><Clock /></el-icon> 部署记录
            </el-button>
            <el-button type="primary" @click="previewConfig">
              <el-icon><View /></el-icon> 预览配置
            </el-button>
            <el-button type="success" @click="saveConfigs">
              <el-icon><Check /></el-icon> 保存配置
            </el-button>
          </el-space>
        </div>
      </template>

      <el-tabs v-model="activeComponent" type="border-card" @tab-change="onComponentChange">
        <el-tab-pane
          v-for="comp in components"
          :key="comp.id"
          :label="comp.display_name"
          :name="comp.id"
        >
          <div v-if="currentComponent">
            <div
              v-for="group in variableGroups"
              :key="group"
              style="margin-bottom: 20px"
            >
              <h4 style="margin-bottom: 12px; border-bottom: 1px solid #eee; padding-bottom: 8px">
                {{ group }}
              </h4>
              <el-form label-width="150px" label-position="left">
                <el-form-item
                  v-for="v in getVariablesByGroup(group)"
                  :key="v.id"
                  :label="v.var_label"
                  :required="v.required"
                >
                  <!-- 字符串输入 -->
                  <el-input
                    v-if="v.var_type === 'string'"
                    v-model="configValues[v.id]"
                    :placeholder="`请输入${v.var_label}`"
                  />
                  <!-- 密码输入 -->
                  <template v-else-if="v.var_type === 'password'">
                    <div v-if="configValues[v.id] === '***ENCRYPTED***'" style="display: flex; align-items: center; gap: 8px">
                      <el-tag type="warning" size="small">已加密存储</el-tag>
                      <el-button size="small" text type="primary" @click="configValues[v.id] = ''">
                        设置新密码
                      </el-button>
                    </div>
                    <el-input
                      v-else
                      v-model="configValues[v.id]"
                      type="password"
                      show-password
                      :placeholder="`请输入${v.var_label}`"
                    />
                  </template>
                  <!-- 数字输入 -->
                  <el-input-number
                    v-else-if="v.var_type === 'number'"
                    v-model="configValues[v.id]"
                    :min="getMin(v)"
                    :max="getMax(v)"
                  />
                  <!-- 布尔值 -->
                  <el-switch
                    v-else-if="v.var_type === 'boolean'"
                    v-model="configValues[v.id]"
                    active-value="true"
                    inactive-value="false"
                  />
                  <!-- 下拉选择 -->
                  <el-select
                    v-else-if="v.var_type === 'select'"
                    v-model="configValues[v.id]"
                    :placeholder="`请选择${v.var_label}`"
                  >
                    <el-option
                      v-for="opt in getOptions(v)"
                      :key="opt"
                      :label="opt"
                      :value="opt"
                    />
                  </el-select>
                  <div v-if="v.description" style="color: #909399; font-size: 12px; margin-top: 4px">
                    {{ v.description }}
                  </div>
                </el-form-item>
              </el-form>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 配置预览对话框 -->
    <el-dialog v-model="showPreview" title="配置预览" width="800px" fullscreen>
      <el-tabs v-model="previewTab" type="card">
        <el-tab-pane
          v-for="(content, path) in previewData"
          :key="path"
          :label="path"
          :name="path"
        >
          <pre style="background: #1e1e1e; color: #d4d4d4; padding: 16px; border-radius: 4px; overflow: auto; max-height: 500px"><code>{{ content }}</code></pre>
        </el-tab-pane>
      </el-tabs>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getComponents, getComponentVariables, getEnvConfigs, updateEnvConfigs, previewConfigs } from '../api'

const route = useRoute()
const envId = route.params.envId as string

const components = ref<any[]>([])
const activeComponent = ref('')
const currentVariables = ref<any[]>([])
const configValues = ref<Record<string, string>>({})
const showPreview = ref(false)
const previewTab = ref('')
const previewData = ref<Record<string, string>>({})

const currentComponent = computed(() => {
  return components.value.find(c => c.id === activeComponent.value)
})

const variableGroups = computed(() => {
  const groups = new Set<string>()
  currentVariables.value.forEach(v => {
    if (v.var_group) groups.add(v.var_group)
  })
  return Array.from(groups)
})

const getVariablesByGroup = (group: string) => {
  return currentVariables.value
    .filter(v => v.var_group === group)
    .sort((a, b) => a.sort_order - b.sort_order)
}

const getOptions = (v: any) => {
  if (typeof v.options === 'string') {
    try {
      return JSON.parse(v.options)
    } catch {
      return []
    }
  }
  return v.options || []
}

const getMin = (v: any) => {
  if (typeof v.validation_rule === 'string') {
    try {
      return JSON.parse(v.validation_rule).min
    } catch {
      return undefined
    }
  }
  return v.validation_rule?.min
}

const getMax = (v: any) => {
  if (typeof v.validation_rule === 'string') {
    try {
      return JSON.parse(v.validation_rule).max
    } catch {
      return undefined
    }
  }
  return v.validation_rule?.max
}

const fetchComponents = async () => {
  try {
    const res: any = await getComponents()
    components.value = res.data || []
    if (components.value.length > 0) {
      activeComponent.value = components.value[0].id
      await fetchVariables()
    }
  } catch {
    ElMessage.error('获取组件列表失败')
  }
}

const fetchVariables = async () => {
  if (!activeComponent.value) return
  try {
    const res: any = await getComponentVariables(activeComponent.value)
    currentVariables.value = res.data || []
    await fetchConfigValues()
  } catch {
    ElMessage.error('获取变量定义失败')
  }
}

const fetchConfigValues = async () => {
  try {
    const res: any = await getEnvConfigs(envId)
    const values: Record<string, string> = {}
    // 先填默认值
    currentVariables.value.forEach(v => {
      values[v.id] = v.default_value || ''
    })
    // 再覆盖已配置的值
    ;(res.data || []).forEach((item: any) => {
      values[item.variable_id] = item.var_value
    })
    configValues.value = values
  } catch {
    // 配置值可能为空，使用默认值
  }
}

const onComponentChange = () => {
  fetchVariables()
}

const saveConfigs = async () => {
  // 过滤掩码值（用户未修改的加密密码）
  const cleanValues: Record<string, string> = {}
  for (const [key, val] of Object.entries(configValues.value)) {
    if (val !== '***ENCRYPTED***') {
      cleanValues[key] = val as string
    }
  }
  try {
    await updateEnvConfigs(envId, {
      values: cleanValues,
      updated_by: 'admin',
    })
    ElMessage.success('配置保存成功')
  } catch {
    ElMessage.error('配置保存失败')
  }
}

const previewConfig = async () => {
  const comp = currentComponent.value
  if (!comp) return
  try {
    const res: any = await previewConfigs(envId, comp.name)
    previewData.value = res.data || {}
    if (Object.keys(previewData.value).length > 0) {
      previewTab.value = Object.keys(previewData.value)[0]
    }
    showPreview.value = true
  } catch {
    ElMessage.error('配置预览失败')
  }
}

onMounted(fetchComponents)
</script>