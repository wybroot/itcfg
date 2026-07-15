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
          <el-menu-item index="/components">
            <el-icon><Grid /></el-icon>
            <span>组件管理</span>
          </el-menu-item>
          <el-menu-item index="/templates">
            <el-icon><Document /></el-icon>
            <span>模板管理</span>
          </el-menu-item>
          <el-menu-item index="/deploy-records">
            <el-icon><Clock /></el-icon>
            <span>部署记录</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <!-- 主内容区 -->
      <el-container>
        <el-header style="border-bottom: 1px solid #e6e6e6; background: #fff">
          <div style="display: flex; justify-content: space-between; align-items: center; height: 100%">
            <h3 style="margin: 0">{{ pageTitle }}</h3>
            <div>
              <el-tag>管理员</el-tag>
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
import { useRoute } from 'vue-router'

const route = useRoute()
const activeMenu = computed(() => route.path)

const pageTitle = computed(() => {
  const titles: Record<string, string> = {
    '/': '首页',
    '/customers': '客户管理',
    '/components': '组件管理',
    '/templates': '模板管理',
    '/deploy-records': '部署记录',
  }
  return titles[route.path] || 'ITCFG 配置中台'
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