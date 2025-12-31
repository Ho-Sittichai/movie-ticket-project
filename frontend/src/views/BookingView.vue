<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useToast } from "vue-toastification";
import api, { paymentApi, seatApi } from "../services/api";
import { useAuthStore } from "../stores/auth";
import PaymentModal from "../components/Modal/PaymentModal.vue";

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();
const toast = useToast();

const isPaymentModalOpen = ref(false);

// Tooltip State for Seats
const hoveredSeat = ref<any>(null);
const tooltipStyle = ref({
  top: "0px",
  left: "0px",
});

const handleSeatHover = (seat: any, event: MouseEvent) => {
  hoveredSeat.value = seat;
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect();
  tooltipStyle.value = {
    top: `${rect.top - 8}px`, // Slightly above the seat
    left: `${rect.left + rect.width / 2}px`,
  };
};

const handleSeatLeave = () => {
  hoveredSeat.value = null;
};

const loading = ref(true);
const movie = ref<any>({
  id: "",
  title: "Loading...",
  time: "",
  price: 0,
});

const rows = ref<string[]>(["A", "B", "C", "D", "E"]); // Default rows for skeleton
const seats = ref<any[]>([]);

// Initialize Skeleton Seats
const initSkeleton = () => {
  const skelSeats = [];
  for (const r of rows.value) {
    for (let i = 1; i <= 8; i++) {
      skelSeats.push({ id: `${r}${i}`, row: r, number: i, status: "LOADING" });
    }
  }
  seats.value = skelSeats;
};
initSkeleton();

const fetchScreening = async () => {
  try {
    loading.value = true;
    // initSkeleton() // Reset to skeleton on refetch

    const movieId = route.params.movieId as string;
    const startTime = route.query.time as string; // Passed as query
    console.log(
      "Fetching screening for movie:",
      movieId,
      "at time:",
      startTime
    );

    // POST Request with object
    const res = await api.post(`/screenings/details`, {
      movie_id: movieId,
      start_time: startTime,
    });
    const data = res.data;

    const screeningData = data.screening || data;
    const movieData = data.movie || {};

    // Update Movie Info
    movie.value = {
      id: movieData.id || "",
      title: movieData.title || "Unknown Movie",
      time: new Date(screeningData.start_time).toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
      }),
      price: screeningData.price,
    };

    // Process Seats
    const rawSeats = screeningData.seats || [];

    const uniqueRows = Array.from(
      new Set(rawSeats.map((s: any) => s.row))
    ).sort() as string[];
    rows.value = uniqueRows;

    seats.value = rawSeats.map((s: any) => {
      let status = s.status;
      // If locked by ME, show as SELECTED
      if (
        status === "LOCKED" &&
        authStore.user &&
        s.locked_by === authStore.user.user_id
      ) {
        status = "SELECTED";
      }
      return {
        id: s.id,
        row: s.row,
        number: s.number,
        status: status,
        locked_by: s.locked_by, // Store it just in case
      };
    });
  } catch (error) {
    console.error("Failed to fetch screening:", error);
    // Keep skeleton/seats visible for a moment, then show error
    setTimeout(() => {
      toast.error("Failed to load screening data. Please try again.");
      router.push("/");
    }, 1000);
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  fetchScreening();
});

const selectedSeats = computed(() =>
  seats.value.filter((s) => s.status === "SELECTED")
);
const totalPrice = computed(
  () => selectedSeats.value.length * movie.value.price
);

const toggleSeat = async (seat: any) => {
  // If it's effectively LOCKED by someone else, we can't touch it.
  // But if we mapped it to "SELECTED" above, we CAN touch it.
  // So we only block if status is "LOCKED" (meaning locked by others) or "BOOKED" or "LOADING"
  if (
    seat.status === "BOOKED" ||
    seat.status === "LOCKED" ||
    seat.status === "LOADING"
  )
    return;

  // Previously we returned here if SELECTED. Now we proceed to call API to unlock.

  // Note: We need UserID context. For now, assume authStore has it.
  // If not logged in, prompt login
  if (!authStore.user) {
    authStore.openLoginModal();
    return;
  }

  try {
    const originalStatus = seat.status;
    seat.status = "LOADING"; // Optimistic UI

    const movieId = route.params.movieId as string;
    const startTime = route.query.time as string;

    const res = await api.post("/seats/lock", {
      user_id: authStore.user.user_id, // Ensure this matches Store structure
      movie_id: movieId,
      start_time: startTime,
      seat_id: seat.id,
    });

    if (res.status === 200) {
      // Backend returns "status": "LOCKED" or "AVAILABLE"
      if (res.data.status === "LOCKED") {
        seat.status = "SELECTED";
      } else if (res.data.status === "AVAILABLE") {
        seat.status = "AVAILABLE";
      }
    } else {
      seat.status = originalStatus; // Revert
      toast.error("Failed to update seat");
    }
  } catch (e: any) {
    console.error("Lock error", e);
    seat.status = "AVAILABLE"; // Simplify revert
    if (e.response && e.response.status === 409) {
      toast.warning(e.response.data.error);
    } else {
      toast.error("Error updating seat");
    }
  }
};

// WebSocket Connection
const connectWS = () => {
  const ws = new WebSocket("ws://localhost:8080/api/ws");

  ws.onopen = () => {
    console.log("WS Connected");
  };

  ws.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data);

      // Filter by Movie ID and Start Time
      const currentMovieId = route.params.movieId;
      const currentStartTime = route.query.time;
      if (
        msg.movie_id !== currentMovieId ||
        msg.start_time !== currentStartTime
      ) {
        return; // Ignore messages for other screenings
      }

      const targetSeat = seats.value.find((s) => s.id === msg.seat_id);
      if (targetSeat) {
        if (msg.status === "LOCKED") {
          // If I am the one locking it (check ID), show as SELECTED so I can toggle it off involved
          if (authStore.user && msg.user_id === authStore.user.user_id) {
            targetSeat.status = "SELECTED";
          } else {
            targetSeat.status = "LOCKED";
          }
        } else if (msg.status === "BOOKED") {
          targetSeat.status = "BOOKED";
        } else if (msg.status === "AVAILABLE") {
          targetSeat.status = "AVAILABLE";
        }
      }
    } catch (e) {
      console.error("WS Parse error", e);
    }
  };

  ws.onclose = () => {
    console.log("WS Disconnected");
    // Auto reconnect?
    setTimeout(connectWS, 3000);
  };
};

onMounted(() => {
  fetchScreening();
  connectWS();
});

const isBooking = ref(false);

const confirmBooking = async () => {
  if (!authStore.user) {
    authStore.openLoginModal();
    return;
  }

  if (selectedSeats.value.length === 0) return;

  try {
    isBooking.value = true; // Start loading
    const movieId = route.params.movieId as string;
    const startTime = route.query.time as string;
    const seatIds = selectedSeats.value.map((s: any) => s.id);
    const paymentId = `BILL-${Date.now()}-${Math.floor(Math.random() * 1000)}`;

    const res = await seatApi.book(
      authStore.user.user_id,
      movieId,
      startTime,
      seatIds,
      paymentId
    );

    if (res.status === 200) {
      toast.success("Booking Success!");
      // Loop to update local status if needed (though API/WS should handle it)
      selectedSeats.value.forEach((s: any) => (s.status = "BOOKED"));
      isPaymentModalOpen.value = false;
      router.push("/");
    }
  } catch (e: any) {
    console.error("Booking failed", e);
    toast.error(
      "Booking Failed: " + (e.response?.data?.error || "Unknown Error")
    );
  } finally {
    isBooking.value = false; // Stop loading
  }
};

const isExtending = ref(false);
const paymentExpireAt = ref(0);

const handleBookTicket = async () => {
  if (!authStore.user) {
    authStore.openLoginModal();
    return;
  }
  if (selectedSeats.value.length === 0) return;

  try {
    isExtending.value = true;
    const movieId = route.params.movieId as string;
    const startTime = route.query.time as string;
    const seatIds = selectedSeats.value.map((s: any) => s.id);

    // Start Payment (Extends locks + Sets payment lock)
    const { data } = await paymentApi.start(
      authStore.user.user_id,
      movieId,
      startTime,
      seatIds
    );
    paymentExpireAt.value = new Date(data.expire_at).getTime();

    // Success -> Open Modal
    isPaymentModalOpen.value = true;
  } catch (e: any) {
    console.error("Failed to extend lock", e);
    toast.warning(e.response?.data?.error || "Failed to start payment process");
    // Maybe refresh?
    fetchScreening();
  } finally {
    isExtending.value = false;
  }
};

const closePaymentModal = async (reason = "user_cancelled") => {
  isPaymentModalOpen.value = false;
  try {
    await paymentApi.cancel(reason);
  } catch (e) {
    console.error("Failed to cancel payment lock", e);
  }
};

// Browser Refresh/Close Protection
const handleBeforeUnload = (event: BeforeUnloadEvent) => {
  if (isPaymentModalOpen.value) {
    // 1. Prevent Default (Show Browser Confirmation)
    event.preventDefault();
    event.returnValue = "";

    // 2. Auto-Cancel using fetch + keepalive
    // Note: We use raw fetch because axios might be killed
    const token = localStorage.getItem("token");
    if (token) {
      fetch("http://localhost:8080/api/payment/cancel", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ reason: "tab_closed" }),
        keepalive: true,
      });
    }
  }
};

watch(isPaymentModalOpen, (isOpen) => {
  if (isOpen) {
    window.addEventListener("beforeunload", handleBeforeUnload);
  } else {
    window.removeEventListener("beforeunload", handleBeforeUnload);
  }
});

onUnmounted(() => {
  window.removeEventListener("beforeunload", handleBeforeUnload);
  if (isPaymentModalOpen.value) {
    // Fallback cleanup if component unmounts without closing modal (e.g. route change)
    closePaymentModal();
  }
});
</script>

<template>
  <div class="min-h-screen pb-32">
    <!-- Header Info -->
    <div class="container mx-auto px-6 py-6 flex items-center gap-4">
      <button
        @click="router.push('/')"
        class="w-10 h-10 rounded-full bg-white/5 hover:bg-white/10 flex items-center justify-center text-white transition-colors"
      >
        &larr;
      </button>
      <div>
        <h1 class="text-2xl font-bold">{{ movie.title }}</h1>
        <p class="text-gray-400 text-sm">
          {{ movie.time }} &bull; Hall 4 (Laser)
        </p>
      </div>
    </div>

    <!-- Screen -->
    <div class="relative w-full max-w-3xl mx-auto mt-10 mb-16 perspective">
      <div
        class="w-2/3 mx-auto h-2 bg-white/20 rounded-full blur-[2px] shadow-[0_20px_60px_rgba(255,255,255,0.2)]"
      ></div>
      <div
        class="text-center text-white/20 text-xs tracking-[0.5em] mt-8 font-bold"
      >
        SCREEN
      </div>

      <!-- Ambient Light -->
      <div
        class="absolute top-0 left-1/2 -translate-x-1/2 w-3/4 h-32 bg-gradient-to-b from-brand-red/10 to-transparent blur-3xl rounded-full pointer-events-none"
      ></div>
    </div>

    <!-- Seat Map -->
    <div class="flex flex-col items-center gap-4 px-4 overflow-x-auto pb-8">
      <div
        v-for="row in rows"
        :key="row"
        class="flex items-center gap-2 sm:gap-4"
      >
        <!-- Row Label -->
        <div
          class="w-6 text-center text-gray-500 text-xs font-bold sticky left-0 z-10 bg-[#121212]"
        >
          {{ row }}
        </div>

        <!-- Seats -->
        <div class="flex gap-1.5 sm:gap-2">
          <button
            v-for="seat in seats.filter((s) => s.row === row)"
            :key="seat.id"
            @click="toggleSeat(seat)"
            class="w-8 h-8 sm:w-10 sm:h-10 rounded-lg flex items-center justify-center text-[10px] sm:text-xs font-medium transition-all duration-300 relative group shrink-0"
            :class="{
              'bg-gray-700 text-gray-300 hover:bg-gray-600':
                seat.status === 'AVAILABLE',
              'bg-red-900/50 text-red-400 cursor-not-allowed animate-pulse border border-red-900/30':
                seat.status === 'LOCKED',
              'bg-brand-red text-white shadow-lg shadow-brand-red/40 scale-110':
                seat.status === 'SELECTED',
              'bg-white/5 text-gray-600 cursor-not-allowed':
                seat.status === 'BOOKED',
              'bg-gray-800 animate-pulse cursor-wait':
                seat.status === 'LOADING',
            }"
            :disabled="
              seat.status === 'BOOKED' ||
              seat.status === 'LOCKED' ||
              seat.status === 'LOADING'
            "
            @mouseenter="handleSeatHover(seat, $event)"
            @mouseleave="handleSeatLeave"
          >
            <span v-if="seat.status !== 'LOADING'">{{ seat.number }}</span>
            <span v-else class="w-2 h-2 rounded-full bg-white/20"></span>
          </button>
        </div>
      </div>
    </div>

    <!-- Legend -->
    <div
      class="flex flex-wrap justify-center gap-4 sm:gap-8 mt-4 sm:mt-12 mb-20 text-xs sm:text-sm text-gray-400 px-4"
    >
      <div class="flex items-center gap-2">
        <div class="w-4 h-4 rounded bg-gray-700"></div>
        Available
      </div>
      <div class="flex items-center gap-2">
        <div
          class="w-4 h-4 rounded bg-brand-red shadow-lg shadow-brand-red/40"
        ></div>
        Selected
      </div>
      <div class="flex items-center gap-2">
        <div class="w-4 h-4 rounded bg-white/5"></div>
        Booked
      </div>
      <div class="flex items-center gap-2">
        <div
          class="w-4 h-4 rounded bg-red-900/50 border border-red-900/30"
        ></div>
        Locked
      </div>
    </div>

    <!-- Checkout Bar -->
    <transition name="slide-up">
      <div
        v-if="selectedSeats.length > 0"
        class="fixed bottom-0 left-0 w-full bg-[#121212]/95 backdrop-blur border-t border-white/10 p-4 pb-8 z-40"
      >
        <div
          class="container mx-auto max-w-4xl flex items-center justify-between"
        >
          <div>
            <div class="text-gray-400 text-xs uppercase tracking-wider mb-1">
              Total Price
            </div>
            <div
              class="text-2xl sm:text-3xl font-bold text-white flex items-end"
            >
              {{ totalPrice }}
              <span class="text-base text-brand-red ml-1 font-medium">THB</span>
            </div>
            <div
              class="text-sm text-gray-500 mt-1 truncate max-w-[150px] sm:max-w-none"
            >
              <span class="text-white">{{ selectedSeats.length }}</span> Seats
              selected
            </div>
          </div>

          <button
            @click="handleBookTicket"
            class="bg-brand-red hover:bg-red-600 text-white px-6 sm:px-8 py-3 rounded-xl font-bold text-base sm:text-lg shadow-xl shadow-red-900/20 transition-all hover:scale-105 active:scale-95 flex items-center gap-2"
            :disabled="isExtending"
          >
            <span
              v-if="isExtending"
              class="animate-spin h-4 w-4 border-2 border-white border-t-transparent rounded-full"
            ></span>
            Book Ticket
          </button>
        </div>
      </div>
    </transition>

    <!-- Payment Modal -->
    <PaymentModal
      :isOpen="isPaymentModalOpen"
      :movieTitle="movie.title"
      :totalPrice="totalPrice"
      :selectedSeats="selectedSeats"
      :loading="isBooking"
      :expireAt="paymentExpireAt"
      @close="closePaymentModal"
      @confirm="confirmBooking"
    />
  </div>

  <!-- Global Teleport Tooltip -->
  <Teleport to="body">
    <div
      v-if="hoveredSeat"
      class="fixed z-[9999] pointer-events-none -translate-x-1/2 -translate-y-full"
      :style="tooltipStyle"
    >
      <div
        class="mb-2 px-3 py-2 bg-slate-900 border border-white/10 rounded-xl shadow-2xl whitespace-nowrap animate-in fade-in zoom-in duration-200"
      >
        <div
          class="text-[10px] font-bold text-indigo-400 uppercase tracking-wider mb-0.5"
        >
          Seat {{ hoveredSeat.id }}
        </div>
        <div class="text-sm font-bold text-white">
          {{ movie.price }}
          <span class="text-[10px] text-gray-400">THB</span>
        </div>
        <!-- Arrow -->
        <div
          class="absolute top-full left-1/2 -translate-x-1/2 border-8 border-transparent border-t-slate-900"
        ></div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.perspective {
  perspective: 1000px;
}
.slide-up-enter-active,
.slide-up-leave-active {
  transition: transform 0.3s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.3s ease;
}

.slide-up-enter-from,
.slide-up-leave-to {
  transform: translateY(100%);
  opacity: 0;
}
</style>
