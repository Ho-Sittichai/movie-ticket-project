<template>
  <div class="flex min-h-[calc(100vh-80px)] bg-slate-950 font-sans text-slate-100">
    <!-- Sidebar -->
    <!-- Sticky positioning to stay visible while scrolling the page -->
    <aside class="w-64 bg-slate-900/50 backdrop-blur-xl border-r border-slate-800 flex flex-col sticky top-20 h-[calc(100vh-80px)]">
      <div class="p-6 flex items-center space-x-3">
        <div class="w-10 h-10 rounded-xl bg-gradient-to-tr from-violet-600 to-indigo-600 flex items-center justify-center shadow-lg shadow-indigo-500/20">
          <i class="fas fa-ticket-alt text-white text-lg"></i>
        </div>
        <div>
          <h2 class="text-xl font-bold tracking-tight text-white">MovieTicket</h2>
          <span class="text-[10px] bg-slate-800 text-slate-400 px-2 py-0.5 rounded-full border border-slate-700 tracking-wide font-semibold">ADMIN</span>
        </div>
      </div>

      <nav class="flex-1 px-4 space-y-1.5 overflow-y-auto custom-scrollbar mt-4">
        <a href="#" class="flex items-center space-x-3 px-4 py-3 rounded-xl bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 shadow-sm transition-all hover:bg-indigo-500/20">
          <i class="fas fa-chart-pie w-5 text-center"></i>
          <span class="font-medium">Overview</span>
        </a>
        <a href="#" class="flex items-center space-x-3 px-4 py-3 rounded-xl text-slate-400 hover:text-slate-100 hover:bg-slate-800/50 transition-all group">
          <i class="fas fa-film w-5 text-center group-hover:text-indigo-400 transition-colors"></i>
          <span class="font-medium">Movies Manager</span>
        </a>
        <a href="#" class="flex items-center space-x-3 px-4 py-3 rounded-xl text-slate-400 hover:text-slate-100 hover:bg-slate-800/50 transition-all group">
          <i class="fas fa-users w-5 text-center group-hover:text-indigo-400 transition-colors"></i>
          <span class="font-medium">Customers</span>
        </a>
        <a href="#" class="flex items-center space-x-3 px-4 py-3 rounded-xl text-slate-400 hover:text-slate-100 hover:bg-slate-800/50 transition-all group">
          <i class="fas fa-calendar-check w-5 text-center group-hover:text-indigo-400 transition-colors"></i>
          <span class="font-medium">Bookings</span>
        </a>
        <a href="#" class="flex items-center space-x-3 px-4 py-3 rounded-xl text-slate-400 hover:text-slate-100 hover:bg-slate-800/50 transition-all group">
          <i class="fas fa-cog w-5 text-center group-hover:text-indigo-400 transition-colors"></i>
          <span class="font-medium">Settings</span>
        </a>
      </nav>
      
      <!-- Removed "Exit to Site" as we have global navbar now -->
    </aside>

    <!-- Main Content -->
    <!-- Removed fixed height/overflow to allow global page scroll -->
    <main class="flex-1 flex flex-col relative bg-[url('https://tailwindcss.com/_next/static/media/docs@tinypng.d9e4cddb.png')] bg-cover">
      <div class="absolute inset-0 bg-slate-950/90 pointer-events-none z-0"></div>
      
      <!-- Minimal Header -->
      <header class="relative z-10 flex justify-between items-center px-8 py-6">
        <div>
          <h1 class="text-3xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-white to-slate-400">Dashboard</h1>
          <p class="text-slate-400 text-sm mt-1">Overview of your cinema performance.</p>
        </div>
        <div class="flex items-center space-x-4">
          <div class="flex items-center gap-3 bg-slate-900/50 backdrop-blur-md border border-slate-700/50 rounded-full pl-1 pr-4 py-1">
             <div class="w-9 h-9 rounded-full bg-gradient-to-br from-indigo-500 via-purple-500 to-pink-500 p-[1px]">
               <div class="w-full h-full rounded-full bg-slate-900 flex items-center justify-center">
                 <img src="https://ui-avatars.com/api/?name=Admin+User&background=random&color=fff" alt="Admin" class="w-full h-full rounded-full opacity-90">
               </div>
             </div>
             <div class="text-sm">
                <span class="block font-semibold text-white leading-none">Admin User</span>
                <span class="text-[10px] text-slate-400 uppercase tracking-widest">Manager</span>
             </div>
          </div>
        </div>
      </header>

      <!-- Content Area -->
      <div class="flex-1 px-8 pb-8 relative z-10">
        
        <!-- Stats Grid (Expanded) -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <!-- Total Revenue -->
          <div class="relative group">
            <div class="absolute -inset-0.5 bg-gradient-to-r from-violet-600 to-fuchsia-600 rounded-2xl opacity-60 group-hover:opacity-100 transition duration-500 blur-sm"></div>
            <div class="relative bg-slate-900/90 backdrop-blur-xl border border-white/10 p-6 rounded-2xl h-full flex items-center justify-between shadow-2xl">
              <div>
                <p class="text-slate-400 text-xs font-bold uppercase tracking-widest mb-1.5 flex items-center gap-2">
                  <span class="w-2 h-2 rounded-full bg-fuchsia-500"></span> Total Revenue
                </p>
                <h3 class="text-4xl font-black text-white tracking-tight">${{ totalRevenue.toLocaleString() }}</h3>
              </div>
              <div class="w-14 h-14 rounded-2xl bg-gradient-to-br from-violet-500/20 to-fuchsia-500/20 flex items-center justify-center border border-white/5">
                <i class="fas fa-wallet text-2xl text-fuchsia-400 drop-shadow-[0_0_10px_rgba(232,121,249,0.5)]"></i>
              </div>
            </div>
          </div>

          <!-- Total Bookings -->
          <div class="relative group">
             <div class="absolute -inset-0.5 bg-gradient-to-r from-cyan-500 to-blue-600 rounded-2xl opacity-60 group-hover:opacity-100 transition duration-500 blur-sm"></div>
            <div class="relative bg-slate-900/90 backdrop-blur-xl border border-white/10 p-6 rounded-2xl h-full flex items-center justify-between shadow-2xl">
              <div>
                <p class="text-slate-400 text-xs font-bold uppercase tracking-widest mb-1.5 flex items-center gap-2">
                  <span class="w-2 h-2 rounded-full bg-cyan-500"></span> Bookings
                </p>
                <h3 class="text-4xl font-black text-white tracking-tight">{{ bookings.length }}</h3>
              </div>
              <div class="w-14 h-14 rounded-2xl bg-gradient-to-br from-cyan-500/20 to-blue-500/20 flex items-center justify-center border border-white/5">
                <i class="fas fa-ticket-alt text-2xl text-cyan-400 drop-shadow-[0_0_10px_rgba(34,211,238,0.5)]"></i>
              </div>
            </div>
          </div>

           <!-- Active Movies -->
           <div class="relative group">
             <div class="absolute -inset-0.5 bg-gradient-to-r from-emerald-500 to-teal-500 rounded-2xl opacity-60 group-hover:opacity-100 transition duration-500 blur-sm"></div>
            <div class="relative bg-slate-900/90 backdrop-blur-xl border border-white/10 p-6 rounded-2xl h-full flex items-center justify-between shadow-2xl">
              <div>
                <p class="text-slate-400 text-xs font-bold uppercase tracking-widest mb-1.5 flex items-center gap-2">
                  <span class="w-2 h-2 rounded-full bg-emerald-500"></span> Showing Now
                </p>
                <h3 class="text-4xl font-black text-white tracking-tight">{{ movies.length }}</h3>
              </div>
              <div class="w-14 h-14 rounded-2xl bg-gradient-to-br from-emerald-500/20 to-teal-500/20 flex items-center justify-center border border-white/5">
                <i class="fas fa-film text-2xl text-emerald-400 drop-shadow-[0_0_10px_rgba(52,211,153,0.5)]"></i>
              </div>
            </div>
          </div>
        </div>

        <!-- Content Block -->
        <div class="bg-slate-900/60 backdrop-blur-md rounded-3xl border border-white/5 overflow-hidden">
          
          <!-- Filter Bar -->
          <div class="p-5 border-b border-white/5 flex flex-wrap items-center justify-between gap-4 bg-white/5">
             <div class="flex items-center gap-3">
                <h3 class="text-lg font-bold text-white mr-2">Recent Bookings</h3>
                
                <!-- Movie Select -->
                <div class="relative group">
                   <select v-model="filters.movie" @change="fetchBookings" class="appearance-none pl-9 pr-8 py-2 bg-slate-900 border border-slate-700 rounded-lg text-xs font-medium text-slate-300 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-all hover:border-slate-500 cursor-pointer">
                     <option value="">All Movies</option>
                     <option v-for="m in movies" :key="m.id" :value="m.id">{{ m.title }}</option>
                   </select>
                   <i class="fas fa-filter absolute left-3 top-1/2 -translate-y-1/2 text-slate-500 text-xs pointer-events-none"></i>
                </div>

                <!-- Date Picker -->
                <div class="relative group">
                  <input type="date" v-model="filters.date" @change="fetchBookings" class="pl-9 pr-4 py-2 bg-slate-900 border border-slate-700 rounded-lg text-xs font-medium text-slate-300 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-all hover:border-slate-500 cursor-pointer uppercase">
                  <i class="fas fa-calendar absolute left-3 top-1/2 -translate-y-1/2 text-slate-500 text-xs pointer-events-none"></i>
                </div>
             </div>

             <div class="flex items-center gap-3 w-full sm:w-auto">
                <div class="relative flex-1">
                   <input type="text" v-model="filters.user" placeholder="Search customer..." @input="debounceSearch" class="pl-9 pr-4 py-2 w-full sm:w-64 bg-slate-900 border border-slate-700 rounded-lg text-xs font-medium text-slate-300 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-all focus:w-72 placeholder-slate-600">
                   <i class="fas fa-search absolute left-3 top-1/2 -translate-y-1/2 text-slate-500 text-xs pointer-events-none"></i>
                </div>
                <button @click="resetFilters" class="p-2 bg-slate-800 hover:bg-slate-700 text-slate-400 hover:text-white rounded-lg border border-slate-700 transition-all" title="Reset Filters">
                   <i class="fas fa-sync-alt text-xs"></i>
                </button>
             </div>
          </div>

          <!-- Table -->
          <div class="overflow-x-auto">
            <table class="w-full text-left border-collapse">
              <thead>
                <tr class="border-b border-white/5 text-xs text-slate-400 bg-slate-900/50">
                  <th class="px-6 py-4 font-semibold uppercase tracking-wider">Booking Info</th>
                  <th class="px-6 py-4 font-semibold uppercase tracking-wider">Customer</th>
                  <th class="px-6 py-4 font-semibold uppercase tracking-wider">Screening Details</th>
                  <th class="px-6 py-4 font-semibold uppercase tracking-wider text-right">Amount</th>
                  <th class="px-6 py-4 font-semibold uppercase tracking-wider text-center">Status</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-white/5 text-sm">
                <tr v-if="loading" class="animate-pulse bg-slate-900/30">
                  <td colspan="5" class="px-6 py-12 text-center text-slate-500">
                    <i class="fas fa-circle-notch fa-spin mr-2 text-indigo-500"></i> Updating list...
                  </td>
                </tr>
                <tr v-else-if="bookings.length === 0">
                   <td colspan="5" class="px-6 py-16 text-center text-slate-500">
                    <div class="flex flex-col items-center opacity-50">
                       <i class="fas fa-folder-open text-4xl mb-3"></i>
                       <span>No bookings found.</span>
                    </div>
                  </td>
                </tr>
                <tr v-for="b in bookings" :key="b.id" class="group hover:bg-white/[0.02] transition-colors">
                  <td class="px-6 py-4">
                     <div class="flex flex-col">
                        <span class="text-white font-semibold group-hover:text-indigo-300 transition-colors">{{ b.movie_title }}</span>
                        <span class="text-xs text-slate-500 font-mono mt-0.5">ID: #{{ b.id.slice(-6) }}</span>
                     </div>
                  </td>
                  <td class="px-6 py-4">
                    <div class="flex items-center">
                       <div class="w-8 h-8 rounded-full bg-slate-800 border border-slate-700 flex items-center justify-center text-xs font-bold text-slate-400 mr-3">
                          {{ b.user_name.charAt(0).toUpperCase() }}
                       </div>
                       <div>
                          <div class="text-slate-200 font-medium">{{ b.user_name }}</div>
                          <div class="text-xs text-slate-500">{{ b.user_email }}</div>
                       </div>
                    </div>
                  </td>
                  <td class="px-6 py-4">
                     <div class="flex items-center gap-4">
                        <div>
                           <div class="text-slate-300 text-xs font-semibold uppercase tracking-wide mb-0.5">Time</div>
                           <div class="text-slate-400 text-xs flex items-center gap-1.5">
                              <i class="far fa-clock"></i> {{ formatTime(b.screening_time) }}
                              <span class="text-slate-600">|</span>
                              {{ formatDate(b.screening_time) }}
                           </div>
                        </div>
                        <div class="pl-4 border-l border-white/5">
                           <div class="text-slate-300 text-xs font-semibold uppercase tracking-wide mb-0.5">Seat</div>
                           <div class="px-2 py-0.5 rounded bg-slate-800 border border-slate-700 text-indigo-300 text-xs font-mono font-bold text-center">
                              {{ b.seat_id }}
                           </div>
                        </div>
                     </div>
                  </td>
                  <td class="px-6 py-4 text-right">
                    <span class="font-bold text-slate-200">${{ b.amount }}</span>
                  </td>
                  <td class="px-6 py-4 text-center">
                    <span :class="[
                      'px-2.5 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wide border shadow-sm',
                      b.status === 'SUCCESS' ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20 shadow-emerald-500/10' : 
                      b.status === 'PENDING' ? 'bg-amber-500/10 text-amber-400 border-amber-500/20 shadow-amber-500/10' : 
                      'bg-red-500/10 text-red-400 border-red-500/20 shadow-red-500/10'
                    ]">
                      {{ b.status }}
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          
          <!-- Pagination -->
          <div class="px-6 py-4 border-t border-white/5 bg-slate-900/50 flex items-center justify-between">
              <span class="text-xs text-slate-500">Showing {{ bookings.length }} records</span>
              <div class="flex gap-2">
                 <button class="w-8 h-8 flex items-center justify-center rounded-lg bg-slate-800 border border-slate-700 text-slate-400 hover:text-white hover:border-slate-500 transition-all disabled:opacity-30" disabled>
                    <i class="fas fa-chevron-left text-xs"></i>
                 </button>
                 <button class="w-8 h-8 flex items-center justify-center rounded-lg bg-slate-800 border border-slate-700 text-slate-400 hover:text-white hover:border-slate-500 transition-all">
                    <i class="fas fa-chevron-right text-xs"></i>
                 </button>
              </div>
          </div>
        </div>

      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { movieApi, adminApi } from '../services/api'
// import { useAuthStore } from '../stores/auth' // Removed unused import

// const authStore = useAuthStore() // Access Token (Handled by interceptor now)

interface Booking {
  id: string
  user_email: string
  user_name: string
  movie_title: string
  screening_time: string
  seat_id: string
  status: string
  amount: number
  created_at: string
}

interface Movie {
  id: string
  title: string
}

const bookings = ref<Booking[]>([])
const movies = ref<Movie[]>([])
const loading = ref(false)

const filters = ref({
  movie: '',
  date: '',
  user: ''
})

const totalRevenue = computed(() => {
  return bookings.value.reduce((sum, b) => sum + b.amount, 0)
})

let searchTimeout: any = null

const debounceSearch = () => {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    fetchBookings()
  }, 500)
}

const fetchMovies = async () => {
  try {
    const res = await movieApi.getAll()
    movies.value = res.data
  } catch (err) {
    console.error("Failed to load movies", err)
  }
}

const fetchBookings = async () => {
  loading.value = true
  try {
    const params = new URLSearchParams()
    if (filters.value.movie) params.append('movie_id', filters.value.movie)
    if (filters.value.date) params.append('date', filters.value.date)
    if (filters.value.user) params.append('user', filters.value.user)

    const res = await adminApi.getBookings(filters.value)
    
    bookings.value = res.data || []
  } catch (err: any) {
    console.error("Failed to load bookings", err)
    if (err.response && (err.response.status === 401 || err.response.status === 403)) {
       alert("Session expired or Unauthorized")
       window.location.href = '/'
    }
    bookings.value = []
  } finally {
    loading.value = false
  }
}

const resetFilters = () => {
  filters.value = { movie: '', date: '', user: '' }
  fetchBookings()
}

const formatDate = (dateStr: string) => {
  if (!dateStr || dateStr.startsWith('0001')) return '-'
  return new Date(dateStr).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
}

const formatTime = (dateStr: string) => {
  if (!dateStr || dateStr.startsWith('0001')) return '-'
  return new Date(dateStr).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })
}

onMounted(() => {
  fetchMovies()
  fetchBookings()
})
</script>

<style scoped>
/* Custom Scrollbar */
.custom-scrollbar::-webkit-scrollbar {
  width: 5px;
  height: 5px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent; 
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #475569; 
  border-radius: 10px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #64748b; 
}
</style>
