<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { useAuthStore } from '../../stores/auth'
import { ref, onMounted, onUnmounted } from 'vue'

const authStore = useAuthStore()
const isProfileOpen = ref(false)

const toggleProfile = () => {
  isProfileOpen.value = !isProfileOpen.value
}

const closeProfile = (e: MouseEvent) => {
  const target = e.target as HTMLElement
  if (!target.closest('.profile-dropdown')) {
    isProfileOpen.value = false
  }
}

onMounted(() => {
  window.addEventListener('click', closeProfile)
})

onUnmounted(() => {
  window.removeEventListener('click', closeProfile)
})
</script>

<template>
  <nav class="fixed w-full z-50 bg-black/80 backdrop-blur-md border-b border-white/10">
    <div class="container mx-auto px-6 h-16 flex items-center justify-between">
      <!-- Left Side: Logo + Main Nav -->
      <div class="flex items-center gap-8">
        <!-- Logo -->
        <RouterLink to="/" class="text-2xl font-bold text-brand-red tracking-tighter hover:scale-105 transition-transform">
          MOVIE<span class="text-white">TICKET</span>
        </RouterLink>

        <!-- Main Nav Links -->
        <div class="hidden md:flex space-x-6">
          <RouterLink to="/" class="text-gray-300 hover:text-white transition-colors text-sm font-medium">
            Movies
          </RouterLink>
        </div>
      </div>

      <!-- Right Side: Auth / Admin -->
      <div class="flex items-center space-x-6">
        <RouterLink v-if="authStore.user?.role === 'ADMIN'" to="/admin" class="text-gray-300 hover:text-white transition-colors text-sm font-medium">
          Dashboard
        </RouterLink>

        <!-- Auth State -->
        <div v-if="authStore.user" class="relative profile-dropdown">
           <!-- Profile Button -->
           <!-- Profile Button -->
           <button 
             @click.stop="toggleProfile"
             class="flex items-center justify-between gap-3 pl-1 pr-3 sm:pr-4 py-1.5 rounded-full border border-white/10 bg-white/5 hover:bg-white/10 transition-all group min-w-[120px] max-w-[160px] sm:max-w-none sm:w-[220px]"
           >
              <div class="flex items-center gap-3 flex-1 min-w-0">
                <!-- Avatar -->
                <img 
                  v-if="authStore.user.picture" 
                  :src="authStore.user.picture" 
                  alt="Profile" 
                  class="w-8 h-8 rounded-full object-cover border border-white/10 flex-shrink-0"
                >
                <div v-else class="w-8 h-8 rounded-full bg-brand-red flex items-center justify-center text-white font-bold text-xs shadow-inner flex-shrink-0">
                  {{ authStore.user.name?.charAt(0).toUpperCase() }}
                </div>
                
                <!-- Name -->
                <span class="text-sm font-medium text-white truncate">
                  {{ authStore.user.name }}
                </span>
              </div>

              <!-- Chevron -->
              <svg 
                class="w-4 h-4 text-gray-400 group-hover:text-white transition-transform duration-200 flex-shrink-0"
                :class="{ 'rotate-180': isProfileOpen }"
                fill="none" stroke="currentColor" viewBox="0 0 24 24"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
              </svg>
           </button>

           <!-- Dropdown Menu -->
           <Transition
              enter-active-class="transition duration-100 ease-out"
              enter-from-class="transform scale-95 opacity-0"
              enter-to-class="transform scale-100 opacity-100"
              leave-active-class="transition duration-75 ease-in"
              leave-from-class="transform scale-100 opacity-100"
              leave-to-class="transform scale-95 opacity-0"
           >
             <div v-if="isProfileOpen" class="absolute right-0 mt-2 w-full bg-[#1A1A1A] rounded-xl shadow-xl border border-white/10 overflow-hidden py-1 z-50">
               <!-- User Info (Mobile Only or Extra Detail) -->
               <div class="px-4 py-3 border-b border-white/5">
                 <p class="text-sm text-white font-medium truncate">{{ authStore.user.email }}</p>
                 <p class="text-xs text-gray-400 truncate">{{ authStore.user.role }}</p>
               </div>

               <button 
                 @click="authStore.logout" 
                 class="w-full text-left px-4 py-2 text-sm text-red-400 hover:bg-white/5 transition-colors flex items-center gap-2"
               >
                 <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"></path></svg>
                 <span class="truncate">Sign Out</span>
               </button>
             </div>
           </Transition>
        </div>

        <button 
          v-else 
          @click="authStore.openLoginModal" 
          class="bg-brand-red hover:bg-red-700 text-white px-5 py-2 rounded-full text-sm font-semibold shadow-lg shadow-red-900/20 transition-all hover:shadow-red-900/40 transform hover:-translate-y-0.5"
        >
          Sign In
        </button>
      </div>
    </div>
  </nav>
</template>
