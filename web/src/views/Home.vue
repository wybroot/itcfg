<template>
  <div>
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center">
              <span>客户数量</span>
              <el-icon :size="24" color="#409EFF"><User /></el-icon>
            </div>
          </template>
          <div style="font-size: 32px; font-weight: bold; color: #409EFF">{{ stats.customers }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center">
              <span>组件模板</span>
              <el-icon :size="24" color="#67C23A"><Grid /></el-icon>
            </div>
          </template>
          <div style="font-size: 32px; font-weight: bold; color: #67C23A">{{ stats.components }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center">
              <span>今日部署</span>
              <el-icon :size="24" color="#E6A23C"><Clock /></el-icon>
            </div>
          </template>
          <div style="font-size: 32px; font-weight: bold; color: #E6A23C">{{ stats.todayDeploys }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center">
              <span>部署成功率</span>
              <el-icon :size="24" color="#F56C6C"><CircleCheck /></el-icon>
            </div>
          </template>
          <div style="font-size: 32px; font-weight: bold; color: #F56C6C">{{ stats.successRate }}%</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card style="margin-top: 20px">
      <template #header>
        <span>快速操作</span>
      </template>
      <el-space wrap>
        <el-button type="primary" @click="$router.push('/customers')">
          <el-icon><Plus /></el-icon> 新建客户
        </el-button>
        <el-button type="success" @click="$router.push('/components')">
          <el-icon><Grid /></el-icon> 查看组件
        </el-button>
        <el-button type="warning" @click="$router.push('/deploy-records')">
          <el-icon><Clock /></el-icon> 部署记录
        </el-button>
      </el-space>
    </el-card>

    <el-card style="margin-top: 20px">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>系统健康状态</span>
          <el-button size="small" @click="fetchHealth" :loading="healthLoading">刷新</el-button>
        </div>
      </template>
      <el-row :gutter="16">
        <el-col :span="8">
          <div style="text-align: center; padding: 12px">
            <div style="font-size: 24px; margin-bottom: 8px">
              <el-icon :size="24" :color="health.service === 'ok' ? '#67C23A' : '#F56C6C'">
                <CircleCheck v-if="health.service === 'ok'" />
                <CircleClose v-else />
              </el-icon>
            </div>
            <div style="font-size: 14px; color: #606266">服务状态</div>
            <el-tag :type="health.service === 'ok' ? 'success' : 'danger'" size="small">{{ health.service || '检测中...' }}</el-tag>
          </div>
        </el-col>
        <el-col :span="8">
          <div style="text-align: center; padding: 12px">
            <div style="font-size: 24px; margin-bottom: 8px">
              <el-icon :size="24" :color="health.database === 'ok' ? '#67C23A' : '#F56C6C'">
                <CircleCheck v-if="health.database === 'ok'" />
                <CircleClose v-else />
              </el-icon>
            </div>
            <div style="font-size: 14px; color: #606266">数据库</div>
            <el-tag :type="health.database === 'ok' ? 'success' : 'danger'" size="small">{{ health.database || '检测中...' }}</el-tag>
          </div>
        </el-col>
        <el-col :span="8">
          <div style="text-align: center; padding: 12px">
            <div style="font-size: 24px; margin-bottom: 8px">
              <el-icon :size="24" :color="health.templates === 'ok' ? '#67C23A' : '#F56C6C'">
                <CircleCheck v-if="health.templates === 'ok'" />
                <CircleClose v-else />
              </el-icon>
            </div>
            <div style="font-size: 14px; color: #606266">模板引擎</div>
            <el-tag :type="health.templates === 'ok' ? 'success' : 'danger'" size="small">{{ health.templates || '检测中...' }}</el-tag>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <el-card style="margin-top: 20px">
      <template #header>
        <span>使用流程</span>
      </template>
      <el-steps :active="0" align-center>
        <el-step title="创建客户" description="录入客户信息" />
        <el-step title="配置环境" description="创建生产/测试环境" />
        <el-step title="填写配置" description="按组件填写配置变量" />
        <el-step title="导出部署包" description="一键导出完整部署包" />
        <el-step title="现场部署" description="Agent 一键部署" />
      </el-steps>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import axios from 'axios'
import { getDashboardStats } from '../api'

const stats = reactive({
  customers: 0,
  components: 0,
  todayDeploys: 0,
  successRate: 100,
})

const fetchStats = async () => {
  try {
    const res: any = await getDashboardStats()
    const data = res.data
    stats.customers = data.customers || 0
    stats.components = data.components || 0
    stats.todayDeploys = data.todayDeploys || 0
    stats.successRate = data.successRate || 100
  } catch {
    // 使用默认值
  }
}

const health = reactive<Record<string, string>>({
  service: '',
  database: '',
  templates: '',
})
const healthLoading = ref(false)

const fetchHealth = async () => {
  healthLoading.value = true
  try {
    const res = await axios.get('/health')
    const data = res.data
    health.service = data.status || 'unknown'
    health.database = data.checks?.database || 'unknown'
    health.templates = data.checks?.templates || 'unknown'
  } catch {
    health.service = 'error'
    health.database = 'error'
    health.templates = 'error'
  } finally {
    healthLoading.value = false
  }
}

onMounted(() => {
  fetchStats()
  fetchHealth()
})
</script>