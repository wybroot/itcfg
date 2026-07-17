<template>
  <div class="dashboard">
    <!-- 欢迎横幅 -->
    <div class="welcome-banner">
      <div class="welcome-info">
        <h2>👋 欢迎回来，{{ userName }}</h2>
        <p>{{ greetMsg }}</p>
      </div>
      <div class="welcome-stats">
        <div class="mini-stat">
          <span class="mini-num">{{ stats.customers }}</span>
          <span class="mini-label">客户</span>
        </div>
        <div class="mini-stat">
          <span class="mini-num">{{ stats.components }}</span>
          <span class="mini-label">组件</span>
        </div>
        <div class="mini-stat">
          <span class="mini-num">{{ stats.todayDeploys }}</span>
          <span class="mini-label">今日部署</span>
        </div>
      </div>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="6" v-for="card in statCards" :key="card.label">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-card-inner">
            <div class="stat-icon" :style="{ background: card.bg }">
              <el-icon :size="28" :color="card.color"><component :is="card.icon" /></el-icon>
            </div>
            <div class="stat-body">
              <div class="stat-value" :style="{ color: card.color }">{{ card.value }}</div>
              <div class="stat-label">{{ card.label }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 快捷操作 + 健康状态 -->
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="16">
        <el-card shadow="never">
          <template #header>
            <span style="font-weight:600">快捷操作</span>
          </template>
          <div class="quick-actions">
            <div class="action-card" @click="$router.push('/customers')">
              <el-icon :size="32" color="#409EFF"><OfficeBuilding /></el-icon>
              <span>客户管理</span>
              <small>创建和管理客户</small>
            </div>
            <div class="action-card" @click="$router.push('/components')">
              <el-icon :size="32" color="#67C23A"><Grid /></el-icon>
              <span>组件管理</span>
              <small>查看系统组件定义</small>
            </div>
            <div class="action-card" @click="$router.push('/templates')">
              <el-icon :size="32" color="#E6A23C"><Document /></el-icon>
              <span>模板管理</span>
              <small>浏览配置模板</small>
            </div>
            <div class="action-card" @click="$router.push('/users')">
              <el-icon :size="32" color="#F56C6C"><UserFilled /></el-icon>
              <span>用户管理</span>
              <small>管理系统用户</small>
            </div>
            <div class="action-card" @click="$router.push('/notify-configs')">
              <el-icon :size="32" color="#909399"><Bell /></el-icon>
              <span>通知配置</span>
              <small>Webhook 通知设置</small>
            </div>
            <div class="action-card" @click="refreshAll">
              <el-icon :size="32" color="#409EFF"><Refresh /></el-icon>
              <span>刷新数据</span>
              <small>重新加载面板数据</small>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never">
          <template #header>
            <div style="display:flex;justify-content:space-between;align-items:center">
              <span style="font-weight:600">系统健康</span>
              <el-tag :type="healthStatus.type" size="small">{{ healthStatus.text }}</el-tag>
            </div>
          </template>
          <div class="health-list">
            <div class="health-item" v-for="item in healthItems" :key="item.name">
              <div class="health-dot" :class="item.status"></div>
              <span>{{ item.label }}</span>
              <span style="margin-left:auto;color:#909399;font-size:13px">{{ item.status === 'ok' ? '正常' : '异常' }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 使用流程 -->
    <el-card shadow="never" style="margin-top: 20px">
      <template #header><span style="font-weight:600">使用流程</span></template>
      <el-steps :active="5" align-center finish-status="success">
        <el-step title="创建客户" description="录入客户信息" />
        <el-step title="配置环境" description="生产/测试/灾备" />
        <el-step title="填写配置" description="按组件填写变量" />
        <el-step title="导出部署包" description="一键导出" />
        <el-step title="Agent 部署" description="自动部署上线" />
      </el-steps>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, onMounted } from 'vue'
import axios from 'axios'
import { getDashboardStats, getUser } from '../api'

const user = getUser()
const userName = user?.nickname || user?.username || '管理员'

const greetMsg = computed(() => {
  const h = new Date().getHours()
  if (h < 9) return '新的一天开始了，看看今天的部署计划吧 ☀️'
  if (h < 12) return '上午好，今天的工作开始了 💪'
  if (h < 14) return '午安，别忘了检查环境状态 👀'
  if (h < 18) return '下午好，继续保持高效节奏 🚀'
  return '辛苦了，回顾一下今天的工作成果 📊'
})

const stats = reactive({
  customers: 0,
  components: 0,
  todayDeploys: 0,
  successRate: 100,
})

const statCards = computed(() => [
  { label: '客户数量', value: stats.customers, icon: 'OfficeBuilding', color: '#409EFF', bg: '#ecf5ff' },
  { label: '组件模板', value: stats.components, icon: 'Grid', color: '#67C23A', bg: '#f0f9eb' },
  { label: '今日部署', value: stats.todayDeploys, icon: 'Promotion', color: '#E6A23C', bg: '#fdf6ec' },
  { label: '部署成功率', value: stats.successRate + '%', icon: 'CircleCheck', color: '#F56C6C', bg: '#fef0f0' },
])

const fetchStats = async () => {
  try {
    const res: any = await getDashboardStats()
    const d = res.data || res
    stats.customers = d.customers || 0
    stats.components = d.components || 0
    stats.todayDeploys = d.todayDeploys || 0
    stats.successRate = d.successRate ?? 100
  } catch { /* use defaults */ }
}

const health = reactive<Record<string, string>>({ service: '', database: '', templates: '' })
const healthLoading = ref(false)

const healthStatus = computed(() => {
  const vals = [health.service, health.database, health.templates]
  const allOk = vals.every(v => v === 'ok')
  const allErr = vals.some(v => v === 'error')
  if (allOk) return { type: 'success' as const, text: '运行正常' }
  if (allErr) return { type: 'danger' as const, text: '服务异常' }
  return { type: 'warning' as const, text: '部分异常' }
})

const healthItems = computed(() => [
  { name: 'service', label: '服务状态', status: health.service === 'ok' ? 'ok' : 'err' },
  { name: 'database', label: '数据库', status: health.database === 'ok' ? 'ok' : 'err' },
  { name: 'templates', label: '模板引擎', status: health.templates === 'ok' ? 'ok' : 'err' },
])

const fetchHealth = async () => {
  healthLoading.value = true
  try {
    const res = await axios.get('/health')
    health.service = res.data.status || 'error'
    health.database = res.data.checks?.database || 'error'
    health.templates = res.data.checks?.templates || 'error'
  } catch {
    health.service = 'error'; health.database = 'error'; health.templates = 'error'
  } finally {
    healthLoading.value = false
  }
}

const refreshAll = () => { fetchStats(); fetchHealth() }

onMounted(refreshAll)
</script>

<style scoped>
.dashboard { max-width: 1400px; margin: 0 auto; }

.welcome-banner {
  background: linear-gradient(135deg, #1e3a5f 0%, #2a5298 50%, #409EFF 100%);
  border-radius: 12px;
  padding: 28px 36px;
  display: flex; justify-content: space-between; align-items: center;
  color: #fff;
}
.welcome-info h2 { font-size: 22px; margin: 0 0 6px; font-weight: 600; }
.welcome-info p { margin: 0; opacity: 0.8; font-size: 14px; }
.welcome-stats { display: flex; gap: 32px; }
.mini-stat { text-align: center; }
.mini-num { display: block; font-size: 28px; font-weight: 700; }
.mini-label { font-size: 13px; opacity: 0.7; }

.stat-card { border-radius: 10px; }
.stat-card-inner { display: flex; align-items: center; gap: 16px; padding: 4px 0; }
.stat-icon { width: 56px; height: 56px; border-radius: 12px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.stat-value { font-size: 28px; font-weight: 700; }
.stat-label { font-size: 13px; color: #909399; margin-top: 2px; }

.quick-actions { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; }
.action-card {
  display: flex; flex-direction: column; align-items: center; gap: 8px;
  padding: 24px 12px;
  border-radius: 10px;
  border: 1px solid #ebeef5;
  cursor: pointer;
  transition: all 0.2s;
}
.action-card:hover { border-color: #409EFF; box-shadow: 0 2px 12px rgba(64,158,255,0.15); transform: translateY(-2px); }
.action-card span { font-size: 14px; font-weight: 500; color: #303133; }
.action-card small { font-size: 12px; color: #909399; }

.health-list { display: flex; flex-direction: column; gap: 16px; }
.health-item { display: flex; align-items: center; gap: 10px; }
.health-dot { width: 10px; height: 10px; border-radius: 50%; flex-shrink: 0; }
.health-dot.ok { background: #67C23A; box-shadow: 0 0 6px rgba(103,194,58,0.5); }
.health-dot.err { background: #F56C6C; box-shadow: 0 0 6px rgba(245,108,108,0.5); }
</style>
