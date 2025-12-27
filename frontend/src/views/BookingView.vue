<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import api from '../services/api'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const loading = ref(true)
const movie = ref<any>({
  id: '',
  title: "Loading...",
  time: "",
  price: 0
})

const rows = ref<string[]>(['A', 'B', 'C', 'D', 'E']) // Default rows for skeleton
const seats = ref<any[]>([])

// Initialize Skeleton Seats
const initSkeleton = () => {
  const skelSeats = []
  for (const r of rows.value) {
    for (let i = 1; i <= 8; i++) {
       skelSeats.push({ id: `${r}${i}`, row: r, number: i, status: 'LOADING' })
    }
  }
  seats.value = skelSeats
}
initSkeleton()

const fetchScreening = async () => {
  try {
    loading.value = true
    // initSkeleton() // Reset to skeleton on refetch
    
    const movieId = route.params.movieId as string
    const startTime = route.query.time as string // Passed as query
    console.log("Fetching screening for movie:", movieId, "at time:", startTime)

    // POST Request with object
    const res = await api.post(`/screenings/details`, {
      movie_id: movieId,
      start_time: startTime
    })
    const data = res.data
    
    const screeningData = data.screening || data
    const movieData = data.movie || {}

    // Update Movie Info
    movie.value = {
      id: movieData.id || '',
      title: movieData.title || "Unknown Movie",
      time: new Date(screeningData.start_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
      price: screeningData.price
    }

    // Process Seats
    const rawSeats = screeningData.seats || []
    
    const uniqueRows = Array.from(new Set(rawSeats.map((s: any) => s.row))).sort() as string[]
    rows.value = uniqueRows

    seats.value = rawSeats.map((s: any) => ({
      id: s.id,
      row: s.row,
      number: s.number,
      status: s.status 
    }))

  } catch (error) {
    console.error("Failed to fetch screening:", error)
    // Keep skeleton/seats visible for a moment, then show error
    setTimeout(() => {
       alert("Failed to load screening data. Please try again.")
       router.push('/')
    }, 1000)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchScreening()
})

const selectedSeats = computed(() => seats.value.filter(s => s.status === 'SELECTED'))
const totalPrice = computed(() => selectedSeats.value.length * movie.value.price)

const toggleSeat = (seat: any) => {
  if (seat.status === 'LOADING' || seat.status === 'BOOKED' || seat.status === 'LOCKED') return
  
  if (seat.status === 'SELECTED') {
    seat.status = 'AVAILABLE'
  } else {
    seat.status = 'SELECTED'
  }
}

const confirmBooking = () => {
  if (!authStore.user) {
    authStore.openLoginModal()
    return
  }
  // TODO: Call API to Book
  alert(`Booking Feature coming next!\nAmount: ${totalPrice.value} THB`)
}
</script>

<template>
  <div class="min-h-screen pb-32">
    <!-- Header Info -->
    <div class="container mx-auto px-6 py-6 flex items-center gap-4">
      <button @click="router.back()" class="w-10 h-10 rounded-full bg-white/5 hover:bg-white/10 flex items-center justify-center text-white transition-colors">
        &larr;
      </button>
      <div>
        <h1 class="text-2xl font-bold">{{ movie.title }}</h1>
        <p class="text-gray-400 text-sm">{{ movie.time }} &bull; Hall 4 (Laser)</p>
      </div>
    </div>

    <!-- Screen -->
    <div class="relative w-full max-w-3xl mx-auto mt-10 mb-16 perspective">
       <div class="w-2/3 mx-auto h-2 bg-white/20 rounded-full blur-[2px] shadow-[0_20px_60px_rgba(255,255,255,0.2)]"></div>
       <div class="text-center text-white/20 text-xs tracking-[0.5em] mt-8 font-bold">SCREEN</div>
       
       <!-- Ambient Light -->
       <div class="absolute top-0 left-1/2 -translate-x-1/2 w-3/4 h-32 bg-gradient-to-b from-brand-red/10 to-transparent blur-3xl rounded-full pointer-events-none"></div>
    </div>

    <!-- Seat Map -->
    <div class="flex flex-col items-center gap-4">
      <div v-for="row in rows" :key="row" class="flex items-center gap-4">
        <!-- Row Label -->
        <div class="w-6 text-center text-gray-500 text-xs font-bold">{{ row }}</div>
        
        <!-- Seats -->
        <div class="flex gap-2">
          <button 
            v-for="seat in seats.filter(s => s.row === row)" 
            :key="seat.id"
            @click="toggleSeat(seat)"
            class="w-10 h-10 rounded-lg flex items-center justify-center text-xs font-medium transition-all duration-300 relative group"
            :class="{
              'bg-gray-700 text-gray-300 hover:bg-gray-600': seat.status === 'AVAILABLE',
              'bg-brand-red text-white shadow-lg shadow-brand-red/40 scale-110': seat.status === 'SELECTED',
              'bg-white/5 text-gray-600 cursor-not-allowed': seat.status === 'BOOKED',
              'bg-red-900/50 text-red-400 cursor-not-allowed animate-pulse': seat.status === 'LOCKED',
              'bg-gray-800 animate-pulse cursor-wait': seat.status === 'LOADING'
            }"
            :disabled="seat.status === 'BOOKED' || seat.status === 'LOCKED' || seat.status === 'LOADING'"
          >
            <span v-if="seat.status !== 'LOADING'">{{ seat.number }}</span>
            <span v-else class="w-2 h-2 rounded-full bg-white/20"></span>
            
            <!-- Tooltip -->
            <span v-if="seat.status === 'AVAILABLE'" class="absolute -top-8 bg-black text-white text-[10px] px-2 py-1 rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none">
              {{ movie.price }} ฿
            </span>
          </button>
        </div>
      </div>
    </div>

    <!-- Legend -->
    <div class="flex justify-center gap-8 mt-12 text-sm text-gray-400">
      <div class="flex items-center gap-2">
        <div class="w-4 h-4 rounded bg-gray-700"></div> Available
      </div>
      <div class="flex items-center gap-2">
        <div class="w-4 h-4 rounded bg-brand-red shadow-lg shadow-brand-red/40"></div> Selected
      </div>
      <div class="flex items-center gap-2">
        <div class="w-4 h-4 rounded bg-white/5"></div> Booked
      </div>
    </div>

    <!-- Checkout Bar -->
    <transition name="slide-up">
      <div v-if="selectedSeats.length > 0" class="fixed bottom-0 left-0 w-full bg-[#121212] border-t border-white/10 p-4 pb-8 z-40">
        <div class="container mx-auto max-w-4xl flex items-center justify-between">
          <div>
            <div class="text-gray-400 text-xs uppercase tracking-wider mb-1">Total Price</div>
            <div class="text-3xl font-bold text-white flex items-end">
               {{ totalPrice }} <span class="text-base text-brand-red ml-1 font-medium">THB</span>
            </div>
            <div class="text-sm text-gray-500 mt-1">
              Seats: <span class="text-white">{{ selectedSeats.map(s => s.id).join(', ') }}</span>
            </div>
          </div>
          
          <button @click="confirmBooking" class="bg-brand-red hover:bg-red-600 text-white px-8 py-3 rounded-xl font-bold text-lg shadow-xl shadow-red-900/20 transition-all hover:scale-105 active:scale-95">
            Book Ticket
          </button>
        </div>
      </div>
    </transition>
  </div>
</template>

<style scoped>
.perspective {
  perspective: 1000px;
}
.slide-up-enter-active,
.slide-up-leave-active {
  transition: transform 0.3s ease, opacity 0.3s ease;
}

.slide-up-enter-from,
.slide-up-leave-to {
  transform: translateY(100%);
  opacity: 0;
}
</style>
