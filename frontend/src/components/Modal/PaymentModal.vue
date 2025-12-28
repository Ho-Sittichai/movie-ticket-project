<script setup lang="ts">
import { ref, computed, onUnmounted, watch } from "vue";

const props = defineProps<{
  isOpen: boolean;
  movieTitle: string;
  totalPrice: number;
  selectedSeats: any[];
  loading?: boolean;
}>();

const emit = defineEmits(["close", "confirm"]);

const paymentMethod = ref("credit"); // 'qr' | 'credit'

// Mock Data for QR
const qrCodeUrl =
  "https://upload.wikimedia.org/wikipedia/commons/d/d0/QR_code_for_mobile_English_Wikipedia.svg"; // Placeholder

const MAX_DISPLAY_SEATS = 15;

const displayedSeats = computed(() => {
  if (props.selectedSeats.length <= MAX_DISPLAY_SEATS + 1) {
    // If we have 16 seats, we can just show them all (4 rows * 4 cols = 16)
    // So distinct limit might be 16 if no overflow needed.
    // But if > 16, we show 15 and 1 overflow button.
    if (props.selectedSeats.length <= 16) return props.selectedSeats;
    return props.selectedSeats.slice(0, MAX_DISPLAY_SEATS);
  }
  return props.selectedSeats.slice(0, MAX_DISPLAY_SEATS);
});

const remainingSeatsCount = computed(() => {
  if (props.selectedSeats.length <= 16) return 0;
  return props.selectedSeats.length - MAX_DISPLAY_SEATS;
});

// Timer Logic
const timeLeft = ref(300); // 5 minutes in seconds
let timerInterval: number;

const formattedTime = computed(() => {
  const m = Math.floor(timeLeft.value / 60);
  const s = timeLeft.value % 60;
  return `${m}:${s < 10 ? "0" + s : s}`;
});

const startTimer = () => {
  timeLeft.value = 300;
  clearInterval(timerInterval);
  timerInterval = setInterval(() => {
    if (timeLeft.value > 0) {
      timeLeft.value--;
    } else {
      clearInterval(timerInterval);
      alert("Payment time expired!");
      emit("close");
    }
  }, 1000);
};

// Reset timer when modal opens
watch(
  () => props.isOpen,
  (newVal) => {
    if (newVal) {
      startTimer();
    } else {
      clearInterval(timerInterval);
    }
  }
);

onUnmounted(() => {
  clearInterval(timerInterval);
});

const onConfirm = () => {
  emit("confirm");
};
</script>

<template>
  <div
    v-if="isOpen"
    class="fixed inset-0 z-50 flex items-center justify-center p-4"
  >
    <!-- Backdrop -->
    <div
      class="absolute inset-0 bg-black/80 backdrop-blur-sm"
      @click="$emit('close')"
    ></div>

    <!-- Modal Content -->
    <div
      class="relative bg-[#1e1e1e] border border-white/10 rounded-2xl w-full max-w-2xl shadow-2xl shadow-black animate-in fade-in zoom-in duration-300 flex flex-col"
    >
      <!-- Header -->
      <div
        class="p-6 border-b border-white/10 flex justify-between items-center bg-white/5 rounded-t-2xl"
      >
        <div>
          <h2 class="text-xl font-bold text-white">Confirm Booking</h2>
          <p class="text-sm text-gray-400">{{ movieTitle }}</p>
        </div>

        <!-- Combined Timer & Close -->
        <div
          class="flex items-center bg-red-500/10 border border-brand-red/20 rounded-lg overflow-hidden"
        >
          <div
            class="px-3 py-1 font-mono font-bold text-brand-red text-sm border-r border-brand-red/20"
          >
            {{ formattedTime }}
          </div>
          <button
            @click="!loading && $emit('close')"
            class="p-2 text-gray-400 transition-colors"
            :class="
              loading
                ? 'cursor-not-allowed opacity-50'
                : 'hover:text-white hover:bg-white/10'
            "
            :disabled="loading"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              class="h-5 w-5"
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
      </div>

      <div class="flex flex-col md:flex-row min-h-[400px]">
        <!-- Left: Summary -->
        <div
          class="w-full md:w-1/2 p-6 border-b md:border-b-0 md:border-r border-white/10 flex flex-col relative"
        >
          <h3
            class="text-sm font-bold text-gray-400 uppercase tracking-wider mb-4"
          >
            Seat Summary
          </h3>

          <div class="flex-1">
            <div class="grid grid-cols-4 gap-2 content-start">
              <div
                v-for="seat in displayedSeats"
                :key="seat.id"
                class="bg-gray-800 text-gray-300 text-xs font-medium py-2 px-2 rounded text-center border border-white/5"
              >
                {{ seat.id }}
              </div>

              <!-- Overflow Indicator -->
              <div
                v-if="remainingSeatsCount > 0"
                class="bg-brand-red/20 text-brand-red text-xs font-bold py-2 px-2 rounded text-center border border-brand-red/30 flex items-center justify-center cursor-help group relative"
              >
                +{{ remainingSeatsCount }}

                <!-- Tooltip for remaining seats -->
                <div
                  class="absolute bottom-full right-0 mb-2 w-56 bg-gray-900 border border-white/20 rounded-lg p-3 hidden group-hover:block z-[60] shadow-xl"
                >
                  <div
                    class="text-[10px] text-gray-500 mb-2 font-bold uppercase tracking-wider"
                  >
                    Additional Seats
                  </div>
                  <div class="flex flex-wrap gap-1.5">
                    <span
                      v-for="s in selectedSeats.slice(15)"
                      :key="s.id"
                      class="text-xs text-gray-300 bg-white/5 px-1.5 py-0.5 rounded border border-white/10"
                    >
                      {{ s.id }}
                    </span>
                  </div>
                </div>
              </div>
            </div>

            <div class="mt-8 pt-4 border-t border-white/10">
              <div class="flex justify-between items-center mb-2">
                <span class="text-gray-400 text-sm"
                  >Tickets ({{ selectedSeats.length }})</span
                >
                <span class="text-white font-medium">{{ totalPrice }} THB</span>
              </div>
              <div
                class="flex justify-between items-center text-gray-500 text-xs"
              >
                <span>VAT (included)</span>
                <span>7%</span>
              </div>
            </div>
          </div>

          <div class="mt-auto pt-6">
            <div class="flex justify-between items-end">
              <span class="text-gray-400 text-sm">Total Amount</span>
              <span class="text-3xl font-bold text-brand-red"
                >{{ totalPrice }}
                <span class="text-sm text-gray-500">THB</span></span
              >
            </div>
          </div>
        </div>

        <!-- Right: Payment -->
        <div class="w-full md:w-1/2 bg-[#121212] flex flex-col">
          <!-- Payment Tabs -->
          <div class="flex border-b border-white/10">
            <button
              @click="paymentMethod = 'credit'"
              class="flex-1 py-4 text-sm font-medium transition-colors border-b-2"
              :class="
                paymentMethod === 'credit'
                  ? 'text-white border-brand-red bg-white/5'
                  : 'text-gray-500 border-transparent hover:text-gray-300'
              "
            >
              Credit Card
            </button>
            <button
              @click="paymentMethod = 'qr'"
              class="flex-1 py-4 text-sm font-medium transition-colors border-b-2"
              :class="
                paymentMethod === 'qr'
                  ? 'text-white border-brand-red bg-white/5'
                  : 'text-gray-500 border-transparent hover:text-gray-300'
              "
            >
              QR PromptPay
            </button>
          </div>

          <div
            class="p-6 flex-1 flex flex-col items-center justify-center text-center"
          >
            <!-- QR Mode -->
            <transition name="fade" mode="out-in">
              <div
                v-if="paymentMethod === 'qr'"
                class="w-full h-full flex flex-col items-center justify-center space-y-4"
              >
                <div
                  class="bg-white p-4 rounded-xl shadow-lg border-2 border-brand-red ml-4 mr-4"
                >
                  <img
                    :src="qrCodeUrl"
                    alt="PromptPay QR"
                    class="w-40 h-40 object-contain mx-auto opacity-90"
                  />
                </div>
                <div class="space-y-1">
                  <p class="text-white font-medium">Scan to Pay</p>
                  <p class="text-xs text-gray-500">
                    Supported by all major Thai banks
                  </p>
                </div>
              </div>

              <!-- Credit Card Mode -->
              <div
                v-else
                class="w-full h-full flex flex-col justify-start space-y-4 pt-2"
              >
                <!-- Mock Card Info -->
                <div
                  class="bg-gradient-to-br from-gray-800 to-black p-5 rounded-xl border border-white/10 w-full mb-2 relative overflow-hidden group"
                >
                  <div class="absolute top-0 right-0 p-3 opacity-20">
                    <svg
                      class="w-12 h-12 text-white"
                      fill="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        d="M22 6H2a2 2 0 00-2 2v10a2 2 0 002 2h20a2 2 0 002-2V8a2 2 0 00-2-2zm0 12H2V10h20v8zm0-10H2V8h20v2z"
                      />
                    </svg>
                  </div>
                  <div class="text-left space-y-4 relative z-10">
                    <div class="w-10 h-6 bg-yellow-500/80 rounded"></div>
                    <div>
                      <div
                        class="text-[10px] text-gray-400 uppercase tracking-widest"
                      >
                        Card Number
                      </div>
                      <div class="text-lg text-white font-mono tracking-wider">
                        4123 •••• •••• 9012
                      </div>
                    </div>
                    <div class="flex justify-between">
                      <div>
                        <div
                          class="text-[10px] text-gray-400 uppercase tracking-widest"
                        >
                          Holder
                        </div>
                        <div class="text-xs text-white tracking-wide">
                          MovieTicket
                        </div>
                      </div>
                      <div>
                        <div
                          class="text-[10px] text-gray-400 uppercase tracking-widest"
                        >
                          Expires
                        </div>
                        <div class="text-xs text-white tracking-wide">
                          12/28
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="space-y-3 px-1">
                  <input
                    type="text"
                    disabled
                    value="4123 4567 8901 2345"
                    class="w-full bg-gray-800 border border-gray-700 rounded p-2.5 text-sm text-gray-400 cursor-not-allowed"
                    placeholder="Card Number"
                  />
                  <div class="flex gap-3">
                    <input
                      type="text"
                      disabled
                      value="12/28"
                      class="w-1/2 bg-gray-800 border border-gray-700 rounded p-2.5 text-sm text-gray-400 cursor-not-allowed"
                      placeholder="MM/YY"
                    />
                    <input
                      type="text"
                      disabled
                      value="***"
                      class="w-1/2 bg-gray-800 border border-gray-700 rounded p-2.5 text-sm text-gray-400 cursor-not-allowed"
                      placeholder="CVV"
                    />
                  </div>
                </div>
              </div>
            </transition>
          </div>

          <div
            v-if="paymentMethod === 'credit'"
            class="p-6 border-t border-white/10 bg-[#1e1e1e]"
          >
            <button
              @click="onConfirm"
              :disabled="loading"
              class="w-full bg-brand-red text-white font-bold py-3.5 rounded-xl shadow-lg shadow-red-900/20 transition-all flex items-center justify-center gap-2"
              :class="
                loading
                  ? 'opacity-50 cursor-not-allowed'
                  : 'hover:bg-red-600 hover:scale-[1.02] active:scale-[0.98]'
              "
            >
              <span
                v-if="loading"
                class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"
              ></span>
              {{ loading ? "Processing..." : "Confirm Payment" }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 4px;
}
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
