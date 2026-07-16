<template>
  <div id="app">
    <el-container style="height: 100vh">
      <!-- 侧边栏 -->
      <el-aside width="220px" style="background-color: #304156">
        <div class="logo">
          <h2 style="color: #fff; text-align: center; padding: 16px 0; margin: 0">
            ITCFG 配置中台
          </h2>
        </div>
        <el-menu
          :default-active="activeMenu"
          router
          background-color="#304156"
          text-color="#bfcbd9"
          active-text-color="#409EFF"
        >
          <el-menu-item index="/">
            <el-icon><HomeFilled /></el-icon>
            <span>首页</span>
          </el-menu-item>
          <el-menu-item index="/customers">
            <el-icon><User /></el-icon>
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

      <!-- 主内容区 -->
      <el-container>
        <el-header style="border-bottom: 1px solid #e6e6e6; background: #fff">
          <div style="display: flex; justify-content: space-between; align-items: center; height: 100%">
            <h3 style="margin: 0">{{ pageTitle }}</h3>
            <div style="display: flex; align-items: center; gap: 12px">
              <span style="color: #606266; font-size: 14px">{{ userName }}</span>
              <el-tag size="small">{{ userRole === 'admin' ? '管理员' : '用户' }}</el-tag>
              <el-button size="small" text @click="handleLogout">退出</el-button>
            </div>
          </div>
        </el-header>
        <el-main style="background-color: #f0f2f5">
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getUser, removeToken, removeUser } from './api'

const route = useRoute()
const router = useRouter()
const activeMenu = computed(() => route.path)

const user = getUser()
const userName = user?.nickname || user?.username || '未登录'
const userRole = user?.role || 'user'

const handleLogout = () => {
  removeToken()
  removeUser()
  router.push('/login')
}

const pageTitle = computed(() => {
  const path = route.path
  const titles: Record<string, string> = {
    '/': '首页',
    '/customers': '客户管理',
    '/components': '组件管理',
    '/templates': '模板管理',
    '/users': '用户管理',
    '/notify-configs': '通知配置',
  }
  if (titles[path]) return titles[path]
  if (path.includes('/envs/') && path.endsWith('/configs')) return '配置管理'
  if (path.includes('/envs/') && path.endsWith('/versions')) return '配置版本历史'
  if (path.includes('/envs/') && path.endsWith('/artifacts')) return '制品版本管理'
  if (path.includes('/envs/') && path.endsWith('/deploy-records')) return '部署记录'
  if (path.includes('/customers/') && path.endsWith('/envs')) return '客户环境'
  return 'ITCFG 配置中台'
})
</script>

<style>
body {
  margin: 0;
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', Arial, sans-serif;
}
#app {
  height: 100vh;
}
</style>