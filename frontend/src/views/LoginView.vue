<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import api from '../services/api'

const router = useRouter()
const loading = ref(false)
const currentRole = ref('')

const handleLogin = async (role: string) => {
  loading.value = true
  currentRole.value = role
  try {
    const res = await api.get(`/auth/login?role=${role}`)
    const { token, user_id, role: userRole } = res.data
    
    localStorage.setItem('token', token)
    localStorage.setItem('user', JSON.stringify({ id: user_id, role: userRole }))
    
    if (userRole === 'ADMIN') {
      router.push('/admin')
    } else {
      router.push('/')
    }
  } catch (error) {
    console.error("Login failed", error)
    alert("Login failed")
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex items-center justify-center min-h-[80vh]">
    <div class="w-full max-w-md bg-brand-gray/50 backdrop-blur-sm p-8 rounded-2xl border border-white/10 shadow-2xl relative overflow-hidden">
      <!-- Decor -->
      <div class="absolute -top-20 -right-20 w-40 h-40 bg-brand-red/20 rounded-full blur-3xl"></div>
      <div class="absolute -bottom-20 -left-20 w-40 h-40 bg-blue-500/10 rounded-full blur-3xl"></div>

      <div class="relative z-10 text-center">
        <h2 class="text-3xl font-bold mb-2">Welcome Back</h2>
        <p class="text-gray-400 mb-8 text-sm">Please sign in to continue</p>
        
        <div class="space-y-4">
          <button 
            @click="handleLogin('USER')" 
            :disabled="loading"
            class="w-full bg-white text-black font-bold py-3 px-4 rounded-xl hover:bg-gray-200 transition-all flex items-center justify-center gap-3 group"
          >
             <span class="w-5 h-5 bg-black rounded-full text-white flex items-center justify-center text-xs">G</span>
             Continue with Google
             <span v-if="loading && currentRole==='USER'" class="animate-spin ml-2">...</span>
          </button>
          
          <div class="relative flex py-2 items-center">
             <div class="flex-grow border-t border-gray-700"></div>
             <span class="flex-shrink mx-4 text-gray-600 text-xs uppercase">Or for Demo</span>
             <div class="flex-grow border-t border-gray-700"></div>
          </div>

          <button 
            @click="handleLogin('ADMIN')" 
            :disabled="loading"
            class="w-full bg-brand-gray border border-white/10 text-gray-300 font-bold py-3 px-4 rounded-xl hover:bg-gray-800 hover:text-white transition-all"
          >
            Admin Access
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
