<script setup lang="ts">
import { RouterView } from 'vue-router'
import NavBar from './components/layout/NavBar.vue'
import Footer from './components/layout/Footer.vue'
import LoginModal from './components/Modal/LoginModal.vue'
</script>

<template>
  <div class="min-h-screen flex flex-col">
    <NavBar />
    
    <main class="flex-grow pt-12">
      <RouterView />
    </main>

    <Footer />
    <LoginModal />
  </div>
</template>
