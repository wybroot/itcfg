<template>
  <div class="page dashboard">
    <section class="hero-card">
      <div>
        <div class="eyebrow">配置交付操作台</div>
        <h1>欢迎回来，{{ userName }}</h1>
        <p>{{ greetMsg }}</p>
      </div>
      <div class="hero-stats">
        <div>
          <strong>{{ stats.customers }}</strong>
          <span>客户</span>
        </div>
        <div>
          <strong>{{ stats.components }}</strong>
          <span>组件</span>
        </div>
        <div>
          <strong>{{ stats.todayDeploys }}</strong>
          <span>今日部署</span>
        </div>
      </div>
    </section>

    <div class="metric-grid">
      <div v-for="card in statCards" :key="card.label" class="metric-card">
        <div class="metric-icon" :style="{ color: card.color, background: card.bg }">
          <el-icon :size="24"><component :is="card.icon" /></el-icon>
        </div>
        <div>
          <div class="metric-value">{{ card.value }}</div>
          <div class="metric-label">{{ card.label }}</div>
        </div>
      </div>
    </div>

    <el-row :gutter="18">
      <el-col :lg="16" :md="24">
        <PageCard title="快捷操作" subtitle="从这里进入日常配置交付流程">
          <div class="quick-actions">
            <div v-for="action in quickActions" :key="action.title" class="action-card" @click="$router.push(action.path)">
              <div class="action-icon" :style="{ color: action.color, background: action.bg }">
                <el-icon :size="26"><component :is="action.icon" /></el-icon>
              </div>
              <span>{{ action.title }}</span>
              <small>{{ action.desc }}</small>
            </div>
            <div class="action-card" @click="refreshAll">
              <div class="action-icon refresh"><el-icon :size="26"><Refresh /></el-icon></div>
              <span>刷新数据</span>
              <small>重新加载面板数据</small>
            </div>
          </div>
        </PageCard>
      </el-col>
      <el-col :lg="8" :md="24">
        <PageCard title="系统健康" subtitle="后端服务、数据库与模板状态">
          <template #actions>
            <el-tag :type="healthStatus.type" size="small">{{ healthStatus.text }}</el-tag>
          </template>
          <div v-loading="healthLoading" class="health-list">
            <div v-for="item in healthItems" :key="item.name" class="health-item">
              <span class="health-dot" :class="item.status"></span>
              <span>{{ item.label }}</span>
              <span class="health-state">{{ item.status === 'ok' ? '正常' : '异常' }}</span>
            </div>
          </div>
        </PageCard>
      </el-col>
    </el-row>

    <PageCard title="交付流程" subtitle="从客户环境到离线部署包的标准闭环">
      <el-steps :active="5" align-center finish-status="success">
        <el-step title="创建客户" description="录入客户信息" />
        <el-step title="配置环境" description="选择组件" />
        <el-step title="填写配置" description="按组件填写变量" />
        <el-step title="绑定制品" description="维护镜像版本" />
        <el-step title="导出部署" description="Agent 一键安装" />
      </el-steps>
    </PageCard>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, onMounted } from 'vue'
import axios from 'axios'
import PageCard from '../components/PageCard.vue'
import { getDashboardStats, getUser } from '../api'

const user = getUser()
const userName = user?.nickname || user?.username || '管理员'

const greetMsg = computed(() => {
  const h = new Date().getHours()
  if (h < 12) return '上午好，建议先检查环境状态和近期部署记录。'
  if (h < 18) return '下午好，可以继续推进客户环境配置和部署包导出。'
  return '辛苦了，回顾今日交付进度并确认部署状态。'
})

const stats = reactive({ customers: 0, components: 0, todayDeploys: 0, successRate: 100 })

const statCards = computed(() => [
  { label: '客户数量', value: stats.customers, icon: 'OfficeBuilding', color: '#2563eb', bg: '#eff6ff' },
  { label: '组件模板', value: stats.components, icon: 'Grid', color: '#16a34a', bg: '#f0fdf4' },
  { label: '今日部署', value: stats.todayDeploys, icon: 'Promotion', color: '#d97706', bg: '#fffbeb' },
  { label: '部署成功率', value: stats.successRate + '%', icon: 'CircleCheck', color: '#dc2626', bg: '#fef2f2' },
])

const quickActions = [
  { title: '客户管理', desc: '创建和管理客户', path: '/customers', icon: 'OfficeBuilding', color: '#2563eb', bg: '#eff6ff' },
  { title: '组件管理', desc: '维护系统组件', path: '/components', icon: 'Grid', color: '#16a34a', bg: '#f0fdf4' },
  { title: '模板管理', desc: '浏览配置模板', path: '/templates', icon: 'Document', color: '#d97706', bg: '#fffbeb' },
  { title: '用户管理', desc: '管理系统用户', path: '/users', icon: 'UserFilled', color: '#7c3aed', bg: '#f5f3ff' },
  { title: '通知配置', desc: 'Webhook 通知设置', path: '/notify-configs', icon: 'Bell', color: '#475467', bg: '#f8fafc' },
]

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
  const anyErr = vals.some(v => v === 'error')
  if (allOk) return { type: 'success' as const, text: '运行正常' }
  if (anyErr) return { type: 'danger' as const, text: '服务异常' }
  return { type: 'warning' as const, text: '检查中' }
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
.dashboard { max-width: 1440px; margin: 0 auto; }
.hero-card { display: flex; justify-content: space-between; gap: 24px; padding: 30px 34px; color: #fff; border-radius: 24px; background: radial-gradient(circle at 78% 18%, rgba(34, 211, 238, .35), transparent 28%), linear-gradient(135deg, #0f172a, #1d4ed8); box-shadow: 0 24px 60px rgba(37, 99, 235, .22); }
.eyebrow { color: rgba(255,255,255,.72); font-size: 13px; font-weight: 600; letter-spacing: .08em; }
.hero-card h1 { margin: 8px 0 8px; font-size: 28px; }
.hero-card p { margin: 0; color: rgba(255,255,255,.78); }
.hero-stats { display: flex; align-items: center; gap: 18px; }
.hero-stats div { min-width: 92px; padding: 14px 16px; text-align: center; border: 1px solid rgba(255,255,255,.18); border-radius: 16px; background: rgba(255,255,255,.1); }
.hero-stats strong { display: block; font-size: 28px; }
.hero-stats span { color: rgba(255,255,255,.72); font-size: 12px; }
.metric-card { display: flex; align-items: center; gap: 14px; }
.metric-icon, .action-icon { display: grid; place-items: center; width: 50px; height: 50px; border-radius: 16px; }
.metric-value { font-size: 28px; font-weight: 750; }
.metric-label { color: var(--itcfg-text-secondary); font-size: 13px; }
.quick-actions { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 14px; }
.action-card { padding: 20px; cursor: pointer; border: 1px solid var(--itcfg-border); border-radius: 18px; background: var(--itcfg-surface); transition: .2s ease; }
.action-card:hover { transform: translateY(-3px); border-color: #bfdbfe; box-shadow: var(--itcfg-shadow-soft); }
.action-card span { display: block; margin-top: 12px; font-weight: 650; }
.action-card small { display: block; margin-top: 5px; color: var(--itcfg-text-secondary); }
.action-icon.refresh { color: #2563eb; background: #eff6ff; }
.health-list { display: flex; flex-direction: column; gap: 16px; min-height: 170px; }
.health-item { display: flex; align-items: center; gap: 10px; }
.health-dot { width: 9px; height: 9px; border-radius: 50%; }
.health-dot.ok { background: var(--itcfg-success); box-shadow: 0 0 0 4px rgba(22, 163, 74, .12); }
.health-dot.err { background: var(--itcfg-danger); box-shadow: 0 0 0 4px rgba(220, 38, 38, .12); }
.health-state { margin-left: auto; color: var(--itcfg-text-secondary); font-size: 13px; }
@media (max-width: 1000px) { .hero-card { flex-direction: column; } .quick-actions { grid-template-columns: repeat(2, 1fr); } }
</style>
