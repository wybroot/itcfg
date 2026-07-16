import { createRouter, createWebHistory } from 'vue-router'
import { isLoggedIn } from '../api'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/Login.vue'),
      meta: { public: true },
    },
    {
      path: '/',
      name: 'Home',
      component: () => import('../views/Home.vue'),
    },
    {
      path: '/users',
      name: 'Users',
      component: () => import('../views/Users.vue'),
    },
    {
      path: '/notify-configs',
      name: 'NotifyConfigs',
      component: () => import('../views/NotifyConfigs.vue'),
    },
    {
      path: '/customers',
      name: 'Customers',
      component: () => import('../views/Customers.vue'),
    },
    {
      path: '/customers/:id/envs',
      name: 'CustomerEnvs',
      component: () => import('../views/CustomerEnvs.vue'),
    },
    {
      path: '/envs/:envId/configs',
      name: 'EnvConfigs',
      component: () => import('../views/EnvConfigs.vue'),
    },
    {
      path: '/components',
      name: 'Components',
      component: () => import('../views/Components.vue'),
    },
    {
      path: '/templates',
      name: 'Templates',
      component: () => import('../views/Templates.vue'),
    },
    {
      path: '/envs/:envId/deploy-records',
      name: 'EnvDeployRecords',
      component: () => import('../views/DeployRecords.vue'),
    },
    {
      path: '/envs/:envId/versions',
      name: 'VersionHistory',
      component: () => import('../views/VersionHistory.vue'),
    },
    {
      path: '/envs/:envId/artifacts',
      name: 'ArtifactVersion',
      component: () => import('../views/ArtifactVersion.vue'),
    },
  ],
})

// 路由守卫：未登录重定向到登录页
router.beforeEach((to, _from, next) => {
  if (to.meta.public || isLoggedIn()) {
    next()
  } else {
    next('/login')
  }
})

export default router