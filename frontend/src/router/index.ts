import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { requiresAuth: false },
    },
    {
      path: '/',
      component: () => import('@/layouts/MainLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          redirect: '/dashboard',
        },
        {
          path: 'dashboard',
          name: 'dashboard',
          component: () => import('@/views/DashboardView.vue'),
          meta: { title: '仪表盘' },
        },
        {
          path: 'inbounds',
          name: 'inbounds',
          component: () => import('@/views/InboundsView.vue'),
          meta: { title: '入站管理' },
        },
        {
          path: 'subscriptions',
          name: 'subscriptions',
          component: () => import('@/views/SubscriptionsView.vue'),
          meta: { title: '订阅管理' },
        },
        {
          path: 'outbounds',
          name: 'outbounds',
          component: () => import('@/views/OutboundsView.vue'),
          meta: { title: '出站管理' },
        },
        {
          path: 'rulesets',
          name: 'rulesets',
          component: () => import('@/views/RulesetsView.vue'),
          meta: { title: '规则集' },
        },
        {
          path: 'settings',
          name: 'settings',
          component: () => import('@/views/SettingsView.vue'),
          meta: { title: '系统设置' },
        },
      ],
    },
  ],
})

router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('token')

  if (to.meta.requiresAuth !== false && !token) {
    next('/login')
  } else if (to.path === '/login' && token) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router
