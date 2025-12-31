<script setup lang="ts">
import { ref, watch } from "vue";
import { useAuthStore } from "../../stores/auth";
import { useRoute, useRouter } from "vue-router";

const authStore = useAuthStore();
const route = useRoute();
const router = useRouter();
const loading = ref(false);

const handleGoogleLogin = () => {
  loading.value = true;
  const currentUrl = window.location.href;
  window.location.href = `http://localhost:8080/api/auth/google/login?redirect_to=${encodeURIComponent(
    currentUrl
  )}`;
};

watch(
  () => route.query.google_auth,
  (newVal) => {
    if (newVal === "success") {
      const { token, user_id, role, name, picture, email } = route.query;

      if (token && user_id) {
        console.group("🎉 Login Logic (via LoginModal)");
        console.table({
          Name: name,
          Email: email,
          ID: user_id,
          Place: "LoginModal.vue - Watcher",
        });
        console.groupEnd();

        authStore.login(
          {
            user_id: user_id as string,
            role: (role as string) || "USER",
            name: name as string,
            email: email as string,
            picture: picture as string,
          },
          token as string
        );

        // Remove params from URL
        const path = route.path;
        router.replace({ path });
      }
    }
  },
  { immediate: true }
);
</script>

<template>
  <Transition name="modal">
    <div
      v-if="authStore.showLoginModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4"
    >
      <!-- Backdrop: Removed backdrop-blur for performance, increased opacity -->
      <div
        class="absolute inset-0 bg-black/90 transition-opacity duration-200"
        @click="authStore.closeLoginModal"
      ></div>

      <!-- Modal Content: Removed transition-all, added transform-gpu -->
      <div
        class="relative bg-[#181818] w-full max-w-md rounded-2xl border border-white/10 shadow-2xl p-8 overflow-hidden transform-gpu"
      >
        <!-- Close Button -->
        <button
          @click="authStore.closeLoginModal"
          class="absolute top-4 right-4 text-gray-400 hover:text-white transition-colors"
        >
          <svg
            class="w-6 h-6"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M6 18L18 6M6 6l12 12"
            ></path>
          </svg>
        </button>

        <div class="text-center mb-8">
          <h2 class="text-3xl font-bold text-white mb-2">Sign In</h2>
          <p class="text-gray-400 text-sm">Welcome back to MovieTicket</p>
        </div>

        <button
          @click="handleGoogleLogin"
          :disabled="loading"
          class="w-full bg-white text-black font-bold py-3.5 px-4 rounded-xl hover:bg-gray-200 transition-colors flex items-center justify-center gap-3 group relative overflow-hidden active:scale-95 duration-100"
        >
          <span v-if="!loading" class="flex items-center gap-3">
            <svg class="w-5 h-5" viewBox="0 0 24 24">
              <path
                d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                fill="#4285F4"
              />
              <path
                d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                fill="#34A853"
              />
              <path
                d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
                fill="#FBBC05"
              />
              <path
                d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                fill="#EA4335"
              />
            </svg>
            Continue with Google
          </span>
          <span v-else class="flex items-center gap-2">
            <svg
              class="animate-spin h-5 w-5 text-black"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            Signing in...
          </span>
        </button>

        <p class="mt-6 text-center text-xs text-gray-500">
          By continuing, you agree to our Terms of Service and Privacy Policy.
        </p>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
/* Modal Transition Wrapper */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease-out;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

/* Specific Content Animation (Scale + Opacity ONLY) */
/* Isolate specific properties to unnecessary repaints */
.modal-enter-active .relative {
  transition: opacity 0.25s ease-out,
    transform 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}
.modal-leave-active .relative {
  transition: opacity 0.2s ease-in, transform 0.2s ease-in;
}

.modal-enter-from .relative {
  opacity: 0;
  transform: scale(0.95) translateY(5px);
}
.modal-leave-to .relative {
  opacity: 0;
  transform: scale(0.98);
}
</style>
