import { createRouter, createWebHistory } from 'vue-router'
import { isLoggedIn } from '../api'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/Login.vue'),
      meta: { public: true, title: '登录' },
    },
    {
      path: '/',
      name: 'Home',
      component: () => import('../views/Home.vue'),
      meta: { title: '首页', icon: 'HomeFilled', menu: true },
    },
    {
      path: '/users',
      name: 'Users',
      component: () => import('../views/Users.vue'),
      meta: { title: '用户管理', icon: 'UserFilled', menu: true },
    },
    {
      path: '/notify-configs',
      name: 'NotifyConfigs',
      component: () => import('../views/NotifyConfigs.vue'),
      meta: { title: '通知配置', icon: 'Bell', menu: true },
    },
    {
      path: '/customers',
      name: 'Customers',
      component: () => import('../views/Customers.vue'),
      meta: { title: '客户管理', icon: 'OfficeBuilding', menu: true },
    },
    {
      path: '/customers/:id/envs',
      name: 'CustomerEnvs',
      component: () => import('../views/CustomerEnvs.vue'),
      meta: { title: '环境管理', parentTitle: '客户管理', activeMenu: '/customers' },
    },
    {
      path: '/envs/:envId/configs',
      name: 'EnvConfigs',
      component: () => import('../views/EnvConfigs.vue'),
      meta: { title: '配置管理', parentTitle: '客户管理', activeMenu: '/customers' },
    },
    {
      path: '/components',
      name: 'Components',
      component: () => import('../views/Components.vue'),
      meta: { title: '组件管理', icon: 'Grid', menu: true },
    },
    {
      path: '/templates',
      name: 'Templates',
      component: () => import('../views/Templates.vue'),
      meta: { title: '模板管理', icon: 'Document', menu: true },
    },
    {
      path: '/envs/:envId/deploy-records',
      name: 'EnvDeployRecords',
      component: () => import('../views/DeployRecords.vue'),
      meta: { title: '部署记录', parentTitle: '配置管理', activeMenu: '/customers' },
    },
    {
      path: '/envs/:envId/versions',
      name: 'VersionHistory',
      component: () => import('../views/VersionHistory.vue'),
      meta: { title: '配置版本历史', parentTitle: '配置管理', activeMenu: '/customers' },
    },
    {
      path: '/envs/:envId/artifacts',
      name: 'ArtifactVersion',
      component: () => import('../views/ArtifactVersion.vue'),
      meta: { title: '制品版本管理', parentTitle: '配置管理', activeMenu: '/customers' },
    },
  ],
})

// 路由守卫
router.beforeEach((to, _from, next) => {
  if (to.meta.public) {
    // 已登录用户访问登录页，直接跳首页
    if (to.path === '/login' && isLoggedIn()) {
      next('/')
    } else {
      next()
    }
  } else if (isLoggedIn()) {
    next()
  } else {
    next('/login')
  }
})

export default router
