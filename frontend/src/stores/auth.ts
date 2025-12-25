import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<any>(null)
  const token = ref<string | null>(null)
  const showLoginModal = ref(false)

  // Initialize from local storage
  const init = () => {
    const storedUser = localStorage.getItem('user')
    const storedToken = localStorage.getItem('token')
    if (storedUser && storedToken) {
      user.value = JSON.parse(storedUser)
      token.value = storedToken
    }
  }

  const login = (userData: any, userToken: string) => {
    user.value = userData
    token.value = userToken
    localStorage.setItem('user', JSON.stringify(userData))
    localStorage.setItem('token', userToken)
    showLoginModal.value = false
  }

  const logout = () => {
    user.value = null
    token.value = null
    localStorage.removeItem('user')
    localStorage.removeItem('token')
  }

  const openLoginModal = () => showLoginModal.value = true
  const closeLoginModal = () => showLoginModal.value = false

  init()

  return { 
    user, 
    token, 
    showLoginModal, 
    login, 
    logout, 
    openLoginModal, 
    closeLoginModal 
  }
})
