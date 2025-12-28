<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import api, { seatApi } from "../services/api";
import { useAuthStore } from "../stores/auth";
import PaymentModal from "../components/Modal/PaymentModal.vue";

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();

const isPaymentModalOpen = ref(false);

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
      alert("Failed to load screening data. Please try again.");
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
      alert("Failed to update seat");
    }
  } catch (e: any) {
    console.error("Lock error", e);
    seat.status = "AVAILABLE"; // Simplify revert
    if (e.response && e.response.status === 409) {
      alert("Seat is already taken by another user!");
    } else {
      alert("Error updating seat");
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
      // Check if message belongs to current screening?
      // In production, we should filter on Backend or here.
      // For now, simple check: seat_id exists in our map

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
    const seatIds = selectedSeats.value.map((s) => s.id);

    const res = await seatApi.book(
      authStore.user.user_id,
      movieId,
      startTime,
      seatIds
    );

    if (res.status === 200) {
      alert("Booking Success!");
      // Loop to update local status if needed (though API/WS should handle it)
      selectedSeats.value.forEach((s) => (s.status = "BOOKED"));
      isPaymentModalOpen.value = false;
      router.push("/");
    }
  } catch (e: any) {
    console.error("Booking failed", e);
    alert("Booking Failed: " + (e.response?.data?.error || "Unknown Error"));
  } finally {
    isBooking.value = false; // Stop loading
  }
};

const isExtending = ref(false);

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
    const seatIds = selectedSeats.value.map((s) => s.id);

    // Call Extend API
    await api.post("/seats/extend", {
      user_id: authStore.user.user_id,
      movie_id: movieId,
      start_time: startTime,
      seat_ids: seatIds,
    });

    // Success -> Open Modal
    isPaymentModalOpen.value = true;
  } catch (e) {
    console.error("Failed to extend lock", e);
    alert("Failed to proceed. Your seat lock might have expired.");
    // Maybe refresh?
    fetchScreening();
  } finally {
    isExtending.value = false;
  }
};
</script>

<template>
  <div class="min-h-screen pb-32">
    <!-- Header Info -->
    <div class="container mx-auto px-6 py-6 flex items-center gap-4">
      <button
        @click="router.back()"
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
          >
            <span v-if="seat.status !== 'LOADING'">{{ seat.number }}</span>
            <span v-else class="w-2 h-2 rounded-full bg-white/20"></span>

            <!-- Tooltip -->
            <span
              v-if="seat.status === 'AVAILABLE'"
              class="hidden sm:block absolute -top-8 bg-black text-white text-[10px] px-2 py-1 rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none z-20"
            >
              {{ movie.price }} ฿
            </span>
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
      @close="isPaymentModalOpen = false"
      @confirm="confirmBooking"
    />
  </div>
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
