<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { useAuthStore } from '../../stores/auth'

const authStore = useAuthStore()
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
        <div v-if="authStore.user" class="flex items-center space-x-4">
          <span class="text-xs text-gray-400">Hi, {{ authStore.user.role }}</span>
          <button @click="authStore.logout" class="bg-white/10 hover:bg-white/20 text-white px-4 py-1.5 rounded-full text-xs font-semibold transition-all">
            Logout
          </button>
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
