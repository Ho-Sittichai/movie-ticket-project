import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import BookingView from '../views/BookingView.vue'

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
    }
  ]
})

// Global Auth Guard
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  // Check auth validity on every navigation
  authStore.checkSession()
  next()
})

export default router
