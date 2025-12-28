<script setup lang="ts">
import { useToast } from "../../composables/useToast";

const { isVisible, message, type, hideToast } = useToast();
</script>

<template>
  <Transition name="toast">
    <div
      v-if="isVisible"
      class="fixed bottom-6 left-6 z-[100] flex items-center p-4 rounded-xl shadow-2xl border backdrop-blur-md min-w-[300px]"
      :class="{
        'bg-red-500/10 border-red-500/20 text-red-200': type === 'error',
        'bg-blue-500/10 border-blue-500/20 text-blue-200': type === 'info',
        'bg-green-500/10 border-green-500/20 text-green-200':
          type === 'success',
      }"
    >
      <!-- Icon -->
      <div class="mr-3 shrink-0">
        <svg
          v-if="type === 'error'"
          xmlns="http://www.w3.org/2000/svg"
          class="h-6 w-6 text-red-400"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <svg
          v-else-if="type === 'success'"
          xmlns="http://www.w3.org/2000/svg"
          class="h-6 w-6 text-green-400"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <svg
          v-else
          xmlns="http://www.w3.org/2000/svg"
          class="h-6 w-6 text-blue-400"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
      </div>

      <!-- Message -->
      <div class="flex-1 text-sm font-medium">{{ message }}</div>

      <!-- Close Button -->
      <button
        @click="hideToast"
        class="ml-3 text-white/40 hover:text-white transition-colors"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="h-4 w-4"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M6 18L18 6M6 6l12 12"
          />
        </svg>
      </button>
    </div>
  </Transition>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

.toast-enter-from {
  opacity: 0;
  transform: translateY(20px) scale(0.95);
}

.toast-leave-to {
  opacity: 0;
  transform: translateY(20px) scale(0.95);
}
</style>
