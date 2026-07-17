<template>
  <div id="app">
    <!-- 登录页：纯全屏，无侧边栏/顶栏 -->
    <template v-if="isLoginPage">
      <router-view />
    </template>

    <!-- 管理后台：侧边栏 + 顶栏 + 内容区 -->
    <el-container v-else style="height: 100vh">
      <el-aside width="220px" class="sidebar">
        <div class="logo" @click="$router.push('/')">
          <el-icon :size="22"><Setting /></el-icon>
          <span class="logo-text">ITCFG 配置中台</span>
        </div>
        <el-menu
          :default-active="activeMenu"
          router
          background-color="transparent"
          text-color="#bfcbd9"
          active-text-color="#fff"
        >
          <el-menu-item index="/">
            <el-icon><HomeFilled /></el-icon>
            <span>首页</span>
          </el-menu-item>
          <el-menu-item index="/customers">
            <el-icon><OfficeBuilding /></el-icon>
            <span>客户管理</span>
          </el-menu-item>
          <el-menu-item index="/users">
            <el-icon><UserFilled /></el-icon>
            <span>用户管理</span>
          </el-menu-item>
          <el-menu-item index="/components">
            <el-icon><Grid /></el-icon>
            <span>组件管理</span>
          </el-menu-item>
          <el-menu-item index="/templates">
            <el-icon><Document /></el-icon>
            <span>模板管理</span>
          </el-menu-item>
          <el-menu-item index="/notify-configs">
            <el-icon><Bell /></el-icon>
            <span>通知配置</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <el-container>
        <el-header class="topbar">
          <div class="topbar-left">
            <el-breadcrumb separator="/">
              <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
              <el-breadcrumb-item v-if="pageTitle !== '首页'">{{ pageTitle }}</el-breadcrumb-item>
            </el-breadcrumb>
          </div>
          <div class="topbar-right">
            <el-dropdown trigger="click">
              <span class="user-dropdown">
                <el-avatar :size="32" icon="UserFilled" />
                <span class="user-name">{{ currentUser.nickname || currentUser.username || '管理员' }}</span>
                <el-icon><ArrowDown /></el-icon>
              </span>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item disabled>
                    <div>
                      <div style="font-weight:500">{{ currentUser.nickname || currentUser.username }}</div>
                      <div style="font-size:12px;color:#909399">{{ currentUser.role === 'admin' ? '管理员' : '用户' }}</div>
                    </div>
                  </el-dropdown-item>
                  <el-dropdown-item divided @click="handleLogout">
                    <el-icon><SwitchButton /></el-icon> 退出登录
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
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
const activeMenu = computed(() => route.path)

// 响应式用户信息：监听路由变化重新读取
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

const pageTitle = computed(() => {
  const titles: Record<string, string> = {
    '/': '首页',
    '/customers': '客户管理',
    '/components': '组件管理',
    '/templates': '模板管理',
    '/users': '用户管理',
    '/notify-configs': '通知配置',
  }
  if (titles[route.path]) return titles[route.path]
  if (route.path.includes('/envs/') && route.path.endsWith('/configs')) return '配置管理'
  if (route.path.includes('/envs/') && route.path.endsWith('/versions')) return '配置版本历史'
  if (route.path.includes('/envs/') && route.path.endsWith('/artifacts')) return '制品版本管理'
  if (route.path.includes('/envs/') && route.path.endsWith('/deploy-records')) return '部署记录'
  if (route.path.includes('/customers/') && route.path.endsWith('/envs')) return '客户环境'
  return 'ITCFG 配置中台'
})
</script>

<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body {
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', Arial, sans-serif;
}
#app { height: 100vh; }
</style>

<style scoped>
.sidebar {
  background: linear-gradient(180deg, #1e3a5f 0%, #152238 100%);
  overflow-y: auto;
}
.logo {
  display: flex; align-items: center; gap: 10px;
  padding: 18px 20px; cursor: pointer;
  border-bottom: 1px solid rgba(255,255,255,0.08);
}
.logo-text {
  color: #fff; font-size: 17px; font-weight: 600; letter-spacing: 1px;
}
.sidebar .el-menu {
  border-right: none;
}
.sidebar .el-menu-item {
  margin: 4px 8px;
  border-radius: 8px;
  transition: all 0.2s;
}
.sidebar .el-menu-item:hover {
  background-color: rgba(255,255,255,0.08) !important;
}
.sidebar .el-menu-item.is-active {
  background-color: #409EFF !important;
  color: #fff !important;
}

.topbar {
  background: #fff;
  display: flex; align-items: center; justify-content: space-between;
  border-bottom: 1px solid #ebeef5;
  box-shadow: 0 1px 4px rgba(0,0,0,0.04);
  padding: 0 24px;
  z-index: 10;
}
.user-dropdown {
  display: flex; align-items: center; gap: 8px;
  cursor: pointer; padding: 4px 8px;
  border-radius: 6px; transition: background 0.2s;
}
.user-dropdown:hover { background: #f5f7fa; }
.user-name { font-size: 14px; color: #303133; max-width: 120px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.main-content {
  background-color: #f0f2f5;
  min-height: 0;
}
</style>
