import { createRouter, createWebHistory } from 'vue-router'
import Landing from '@/views/landing/index.vue'
import UserLayout from '@/views/user/layout.vue'
import UserKeys from '@/views/user/keys.vue'
import UserUsage from '@/views/user/usage.vue'
import UserModels from '@/views/user/models.vue'
import AdminLayout from '@/views/admin/layout.vue'
import AdminUsers from '@/views/admin/users.vue'
import AdminModels from '@/views/admin/models.vue'
import AdminKeys from '@/views/admin/keys.vue'
import AdminUsage from '@/views/admin/usage.vue'
import AdminAlias from '@/views/admin/alias.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: Landing },
    {
      path: '/user',
      component: UserLayout,
      meta: { requiresAuth: true },
      children: [
        { path: '', redirect: '/user/keys' },
        { path: 'keys', component: UserKeys },
        { path: 'usage', component: UserUsage },
        { path: 'models', component: UserModels },
      ],
    },
    {
      path: '/admin',
      component: AdminLayout,
      meta: { requiresAuth: true, requiresAdmin: true },
      children: [
        { path: '', redirect: '/admin/users' },
        { path: 'users', component: AdminUsers },
        { path: 'models', component: AdminModels },
        { path: 'keys', component: AdminKeys },
        { path: 'usage', component: AdminUsage },
        { path: 'alias', component: AdminAlias },
      ],
    },
  ],
})

router.beforeEach((to) => {
  const token = localStorage.getItem('token')
  const role = localStorage.getItem('role')

  if (to.meta.requiresAuth && !token) {
    return '/'
  }
  if (to.meta.requiresAdmin && role !== 'admin') {
    return '/user'
  }
  if (to.path === '/' && token) {
    return role === 'admin' ? '/admin' : '/user'
  }
})

export default router
