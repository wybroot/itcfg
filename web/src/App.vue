<template>
  <div id="app">
    <template v-if="isLoginPage">
      <router-view />
    </template>

    <el-container v-else class="app-shell">
      <el-aside width="232px" class="sidebar">
        <div class="brand" @click="$router.push('/')">
          <div class="brand-icon">
            <el-icon><Setting /></el-icon>
          </div>
          <div>
            <div class="brand-title">ITCFG</div>
            <div class="brand-subtitle">配置交付操作台</div>
          </div>
        </div>

        <div class="nav-section-title">工作台</div>
        <el-menu
          :default-active="activeMenu"
          router
          background-color="transparent"
          text-color="rgba(255,255,255,0.68)"
          active-text-color="#fff"
        >
          <el-menu-item v-for="item in menuRoutes" :key="item.path" :index="item.path">
            <el-icon><component :is="item.meta.icon" /></el-icon>
            <span>{{ item.meta.title }}</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <el-container class="content-shell">
        <el-header class="topbar">
          <div class="topbar-left">
            <el-breadcrumb separator="/">
              <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
              <el-breadcrumb-item v-if="parentTitle">{{ parentTitle }}</el-breadcrumb-item>
              <el-breadcrumb-item v-if="pageTitle !== '首页'">{{ pageTitle }}</el-breadcrumb-item>
            </el-breadcrumb>
            <div class="topbar-title">{{ pageTitle }}</div>
          </div>

          <el-dropdown trigger="click">
            <span class="user-dropdown">
              <el-avatar :size="34" icon="UserFilled" class="user-avatar" />
              <span class="user-meta">
                <span class="user-name">{{ currentUser.nickname || currentUser.username || '管理员' }}</span>
                <span class="user-role">{{ currentUser.role === 'admin' ? '管理员' : '用户' }}</span>
              </span>
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item disabled>
                  <div>
                    <div style="font-weight: 600">{{ currentUser.nickname || currentUser.username }}</div>
                    <div style="font-size: 12px; color: #909399">{{ currentUser.role === 'admin' ? '管理员' : '用户' }}</div>
                  </div>
                </el-dropdown-item>
                <el-dropdown-item divided @click="handleLogout">
                  <el-icon><SwitchButton /></el-icon> 退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </el-header>

        <el-main class="main-content">
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getUser, removeToken, removeUser } from './api'

const route = useRoute()
const router = useRouter()

const isLoginPage = computed(() => route.path === '/login')
const activeMenu = computed(() => (route.meta.activeMenu as string) || route.path)
const pageTitle = computed(() => (route.meta.title as string) || 'ITCFG 配置中台')
const parentTitle = computed(() => route.meta.parentTitle as string | undefined)
const menuRoutes = computed(() =>
  router.getRoutes()
    .filter(r => r.meta.menu)
    .sort((a, b) => a.path.localeCompare(b.path))
)

const currentUser = reactive({
  nickname: '',
  username: '',
  role: '',
})

const refreshUser = () => {
  const u = getUser()
  if (u) {
    currentUser.nickname = u.nickname || ''
    currentUser.username = u.username || ''
    currentUser.role = u.role || 'user'
  } else {
    currentUser.nickname = ''
    currentUser.username = ''
    currentUser.role = ''
  }
}
refreshUser()
watch(() => route.path, refreshUser)

const handleLogout = () => {
  removeToken()
  removeUser()
  router.push('/login')
}
</script>

<style scoped>
.app-shell {
  height: 100vh;
  background: var(--itcfg-bg);
}

.sidebar {
  position: relative;
  overflow-y: auto;
  background:
    radial-gradient(circle at 20% 0%, rgba(59, 130, 246, 0.28), transparent 28%),
    linear-gradient(180deg, #0f172a 0%, #111827 52%, #0b1120 100%);
  border-right: 1px solid rgba(255, 255, 255, 0.08);
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 22px 20px 18px;
  cursor: pointer;
}

.brand-icon {
  display: grid;
  place-items: center;
  width: 42px;
  height: 42px;
  color: #fff;
  border-radius: 14px;
  background: linear-gradient(135deg, #3b82f6, #06b6d4);
  box-shadow: 0 12px 30px rgba(37, 99, 235, 0.35);
}

.brand-title {
  color: #fff;
  font-size: 20px;
  font-weight: 800;
  letter-spacing: 0.04em;
}

.brand-subtitle {
  margin-top: 2px;
  color: rgba(255, 255, 255, 0.58);
  font-size: 12px;
}

.nav-section-title {
  padding: 14px 20px 8px;
  color: rgba(255, 255, 255, 0.42);
  font-size: 12px;
  letter-spacing: 0.08em;
}

.sidebar :deep(.el-menu) {
  border-right: none;
  padding: 0 10px;
}

.sidebar :deep(.el-menu-item) {
  height: 44px;
  margin: 4px 0;
  border-radius: 12px;
  transition: all 0.2s ease;
}

.sidebar :deep(.el-menu-item:hover) {
  color: #fff !important;
  background: rgba(255, 255, 255, 0.08) !important;
}

.sidebar :deep(.el-menu-item.is-active) {
  color: #fff !important;
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.95), rgba(14, 165, 233, 0.85)) !important;
  box-shadow: 0 12px 24px rgba(37, 99, 235, 0.28);
}

.content-shell {
  min-width: 0;
}

.topbar {
  height: 72px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding: 0 28px;
  background: rgba(255, 255, 255, 0.86);
  border-bottom: 1px solid var(--itcfg-border);
  backdrop-filter: blur(14px);
}

.topbar-left {
  min-width: 0;
}

.topbar-title {
  margin-top: 6px;
  color: var(--itcfg-text-primary);
  font-size: 20px;
  font-weight: 700;
}

.user-dropdown {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  padding: 8px 10px;
  cursor: pointer;
  border: 1px solid transparent;
  border-radius: 14px;
  transition: all 0.2s ease;
}

.user-dropdown:hover {
  background: var(--itcfg-surface-soft);
  border-color: var(--itcfg-border);
}

.user-avatar {
  background: linear-gradient(135deg, #2563eb, #06b6d4);
}

.user-meta {
  display: flex;
  flex-direction: column;
  line-height: 1.2;
}

.user-name {
  max-width: 140px;
  overflow: hidden;
  color: var(--itcfg-text-primary);
  font-size: 14px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-role {
  margin-top: 3px;
  color: var(--itcfg-text-secondary);
  font-size: 12px;
}

.main-content {
  min-height: 0;
  padding: 26px 28px 32px;
  overflow: auto;
  background:
    radial-gradient(circle at 90% 0%, rgba(37, 99, 235, 0.08), transparent 30%),
    var(--itcfg-bg);
}
</style>
