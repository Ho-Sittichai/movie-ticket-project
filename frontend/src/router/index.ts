import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import BookingView from '../views/BookingView.vue'
import { useToast } from "vue-toastification"

import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView
    },
    {
      path: '/booking/:movieId',
      name: 'booking',
      component: BookingView
    },
    {
      path: '/admin',
      name: 'admin',
      component: () => import('../views/AdminDashboard.vue'),
      meta: { requiresAdmin: true }
    }
  ]
})

// Global Auth Guard
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  const toast = useToast()
  // Check session
  authStore.checkSession()
  
  // 1. Check if route requires auth
  if (to.meta.requiresAdmin) {
    // Check if user is logged in
    if (!authStore.user || !authStore.token) {
      next('/') // Not logged in -> Home (or login)
      return
    }
    // Check role
    if (authStore.user.role !== 'ADMIN') {
      toast.error("Access Denied: Admins only")
      next('/') // Wrong role -> Home
      return
    }
  }

  next()
})

export default router
