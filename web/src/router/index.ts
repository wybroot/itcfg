import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'Home',
      component: () => import('../views/Home.vue'),
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
      path: '/deploy-records',
      name: 'DeployRecords',
      component: () => import('../views/DeployRecords.vue'),
    },
  ],
})

export default router