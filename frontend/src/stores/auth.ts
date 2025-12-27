import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  interface User {
    user_id: string
    name: string
    email: string
    role: string
    picture: string
  }

  const user = ref<User | null>(null)
  const token = ref<string | null>(null)
  const showLoginModal = ref(false)

  // Check session validity (1 hour default)
  const SESSION_DURATION = 60 * 60 *1000

  const checkSession = () => {
    const expiry = localStorage.getItem('auth_expiry')
    if (!expiry) return false
    
    if (Date.now() > parseInt(expiry)) {
      logout()
      return false
    }
    return true
  }

  // Initialize from local storage
  const init = () => {
    const storedUser = localStorage.getItem('user')
    const storedToken = localStorage.getItem('token')
    
    if (storedUser && storedToken) {
      if (checkSession()) {
        user.value = JSON.parse(storedUser)
        token.value = storedToken
      }
    }
  }

  const login = (userData: User, userToken: string) => {
    user.value = userData
    token.value = userToken
    
    const expiryTime = Date.now() + SESSION_DURATION
    
    localStorage.setItem('user', JSON.stringify(userData))
    localStorage.setItem('token', userToken)
    localStorage.setItem('auth_expiry', expiryTime.toString())
    
    showLoginModal.value = false
  }

  const logout = () => {
    user.value = null
    token.value = null
    localStorage.removeItem('user')
    localStorage.removeItem('token')
    localStorage.removeItem('auth_expiry')
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
    checkSession,
    openLoginModal, 
    closeLoginModal 
  }
})
