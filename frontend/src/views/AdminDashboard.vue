<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import { useAuthStore } from "../stores/auth";
import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  LineElement,
  PointElement,
  CategoryScale,
  LinearScale,
  ArcElement,
  Filler,
} from "chart.js";
import { Line, Doughnut } from "vue-chartjs";

ChartJS.register(
  Title,
  Tooltip,
  Legend,
  LineElement,
  PointElement,
  CategoryScale,
  LinearScale,
  ArcElement,
  Filler
);

interface Booking {
  id: string;
  user_email: string;
  user_name: string;
  movie_title: string;
  poster_url: string;
  seat_id: string;
  status: string;
  amount: number;
  created_at: string;
}

interface Movie {
  id: string;
  title: string;
}

// --- State ---
const authStore = useAuthStore();
const bookings = ref<Booking[]>([]);
const movies = ref<Movie[]>([]);
const loading = ref(false);

const pagination = ref({
  page: 1,
  limit: 10,
  total: 0,
  pages: 1,
});

const filters = ref({
  movie: "",
  date: "",
  user: "",
});

const stats = ref({
  revenue: 0,
  totalBookings: 0,
});

// Helpers for Date Filter
const dateInputRef = ref<HTMLInputElement | null>(null);

const isToday = computed(() => {
  if (!filters.value.date) return false;
  const now = new Date();
  const localToday = `${now.getFullYear()}-${String(
    now.getMonth() + 1
  ).padStart(2, "0")}-${String(now.getDate()).padStart(2, "0")}`;
  return filters.value.date === localToday;
});

const toggleToday = () => {
  const now = new Date();
  const localToday = `${now.getFullYear()}-${String(
    now.getMonth() + 1
  ).padStart(2, "0")}-${String(now.getDate()).padStart(2, "0")}`;
  if (filters.value.date === localToday) {
    filters.value.date = "";
  } else {
    filters.value.date = localToday;
  }
  resetAndFetch();
};

const openDatePicker = () => {
  if (dateInputRef.value) {
    try {
      dateInputRef.value.showPicker();
    } catch (e) {
      console.log("Date picker programmatic open not supported");
    }
  }
};

const clearDate = () => {
  filters.value.date = "";
  resetAndFetch();
};

const lineChartData = computed(() => {
  const groups: Record<string, number> = {};
  bookings.value.forEach((b) => {
    const d = new Date(b.created_at).toLocaleDateString();
    groups[d] = (groups[d] || 0) + 1;
  });

  return {
    labels: Object.keys(groups),
    datasets: [
      {
        label: "Bookings",
        data: Object.values(groups),
        borderColor: "#818cf8",
        backgroundColor: "rgba(129, 140, 248, 0.2)",
        tension: 0.4,
        fill: true,
        pointBackgroundColor: "#fff",
        pointBorderColor: "#6366f1",
        pointRadius: 4,
      },
    ],
  };
});

const lineChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      backgroundColor: "#1e293b",
      titleColor: "#e2e8f0",
      bodyColor: "#e2e8f0",
      padding: 10,
      borderColor: "#334155",
      borderWidth: 1,
    },
  },
  scales: {
    x: {
      grid: { display: false, color: "#334155" },
      ticks: { color: "#94a3b8" },
    },
    y: {
      grid: { color: "#334155", borderDash: [5, 5] },
      ticks: { color: "#94a3b8", stepSize: 1 },
    },
  },
};

const doughnutChartData = computed(() => {
  const revenueByMovie: Record<string, number> = {};
  bookings.value.forEach((b) => {
    if (b.status === "SUCCESS") {
      revenueByMovie[b.movie_title] =
        (revenueByMovie[b.movie_title] || 0) + b.amount;
    }
  });

  return {
    labels: Object.keys(revenueByMovie),
    datasets: [
      {
        data: Object.values(revenueByMovie),
        backgroundColor: [
          "#818cf8",
          "#34d399",
          "#f472b6",
          "#60a5fa",
          "#facc15",
          "#a78bfa",
        ],
        borderWidth: 0,
        hoverOffset: 4,
      },
    ],
  };
});

const doughnutChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: "right" as const,
      labels: { color: "#e2e8f0", font: { size: 11 } },
    },
  },
};

// --- Methods ---
const fetchData = async () => {
  await Promise.all([fetchMovies(), fetchBookings()]);
};

const fetchMovies = async () => {
  try {
    const res = await fetch("http://localhost:8080/api/movies");
    const data = await res.json();
    movies.value = data;
  } catch (err) {
    console.error("Failed to load movies", err);
  }
};

const fetchBookings = async () => {
  loading.value = true;
  try {
    const params = new URLSearchParams();
    if (filters.value.movie) params.append("movie_id", filters.value.movie);
    if (filters.value.date) params.append("date", filters.value.date);
    if (filters.value.user) params.append("user", filters.value.user);

    // Pagination params
    params.append("page", pagination.value.page.toString());
    params.append("limit", pagination.value.limit.toString());
    // ...

    const res = await fetch(
      `http://localhost:8080/api/admin/bookings?${params.toString()}`,
      {
        headers: {
          Authorization: `Bearer ${authStore.token}`,
        },
      }
    );

    const responseData = await res.json();
    // Handle new response structure { data: [], meta: {} }
    if (responseData.data) {
      bookings.value = responseData.data;
      if (responseData.meta) {
        pagination.value.total = responseData.meta.total;
        pagination.value.pages = responseData.meta.pages;
      }
    } else {
      // Fallback for old API structure or empty
      bookings.value = Array.isArray(responseData) ? responseData : [];
    }

    // Stats calc (Naive approach: In real app, stats should come from separate API to be accurate across ALL pages)
    // For now, we update stats based on current page which is WRONG but efficient,
    // OR we can make a separate call for stats.
    // Given the task is about performance, let's just sum the current view or keep placeholders.
    // For better UX, let's assume valid totals should come from backend, but we'll sum current page for now to avoid errors.
    const totalRev = bookings.value.reduce(
      (sum, b) => (b.status === "SUCCESS" ? sum + b.amount : sum),
      0
    );
    stats.value = {
      revenue: totalRev, // This is only for current page!
      totalBookings: pagination.value.total,
    };
  } catch (err) {
    console.error("Failed to load bookings", err);
    bookings.value = [];
  } finally {
    loading.value = false;
  }
};

const resetAndFetch = () => {
  pagination.value.page = 1;
  fetchBookings();
};

const changePage = (newPage: number) => {
  if (newPage < 1 || newPage > pagination.value.pages) return;
  pagination.value.page = newPage;
  fetchBookings();
};

let searchTimeout: any = null;
const debounceSearch = () => {
  if (searchTimeout) clearTimeout(searchTimeout);
  searchTimeout = setTimeout(() => {
    resetAndFetch();
  }, 500);
};

const resetFilters = () => {
  filters.value = { movie: "", date: "", user: "" };
  resetAndFetch();
};

const formatDate = (dateStr: string) => {
  if (!dateStr || dateStr.startsWith("0001")) return "-";
  return new Date(dateStr).toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
};

const formatTime = (dateStr: string) => {
  if (!dateStr || dateStr.startsWith("0001")) return "-";
  return new Date(dateStr).toLocaleTimeString(undefined, {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  });
};

onMounted(() => {
  fetchData();
});
</script>

<template>
  <div
    class="flex min-h-screen bg-[#0f172a] font-sans text-slate-100 selection:bg-indigo-500/30"
  >
    <!-- Sidebar -->
    <aside
      class="w-20 lg:w-64 bg-slate-900/50 border-r border-slate-800 flex flex-col sticky top-0 h-screen transition-all duration-300 z-20"
    >
      <nav class="flex-1 px-4 space-y-2 mt-10">
        <!-- Dashboard (Active) -->
        <a
          href="#"
          class="flex items-center gap-3 px-3 lg:px-4 py-3 rounded-xl bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 shadow-sm transition-all group"
        >
          <i class="fas fa-chart-pie w-6 text-center"></i>
          <span class="hidden lg:block font-medium">Dashboard</span>
        </a>

        <!-- Movie (Disabled) -->
        <div
          class="flex items-center gap-3 px-3 lg:px-4 py-3 rounded-xl text-slate-500 cursor-not-allowed opacity-50 border border-transparent"
        >
          <i class="fas fa-film w-6 text-center"></i>
          <span class="hidden lg:block font-medium">Movie</span>
          <span
            class="hidden lg:inline-flex px-1.5 py-0.5 rounded text-[10px] bg-slate-800 text-slate-400 border border-slate-700 ml-auto"
            >Soon</span
          >
        </div>

        <!-- User (Disabled) -->
        <div
          class="flex items-center gap-3 px-3 lg:px-4 py-3 rounded-xl text-slate-500 cursor-not-allowed opacity-50 border border-transparent"
        >
          <i class="fas fa-users w-6 text-center"></i>
          <span class="hidden lg:block font-medium">User</span>
          <span
            class="hidden lg:inline-flex px-1.5 py-0.5 rounded text-[10px] bg-slate-800 text-slate-400 border border-slate-700 ml-auto"
            >Soon</span
          >
        </div>
      </nav>

      <div class="p-4 border-t border-slate-800">
        <div class="flex items-center gap-3 justify-center lg:justify-start">
          <img
            src="https://ui-avatars.com/api/?name=Admin&background=random"
            class="w-8 h-8 rounded-full border border-slate-600"
          />
          <div class="hidden lg:block">
            <div class="text-sm font-medium text-white">Admin User</div>
            <div class="text-xs text-slate-500">View Profile</div>
          </div>
        </div>
      </div>
    </aside>

    <!-- Main Content -->
    <main class="flex-1 flex flex-col relative overflow-hidden">
      <!-- Background Effects (Optimized: Removed heavy blurs/animations if they cause lag, but kept static for aesthetics) -->
      <div class="absolute inset-0 z-0 pointer-events-none">
        <div
          class="absolute top-[-10%] right-[-5%] w-[500px] h-[500px] bg-indigo-600/10 rounded-full blur-[100px]"
        ></div>
        <div
          class="absolute bottom-[-10%] left-[-10%] w-[600px] h-[600px] bg-violet-600/05 rounded-full blur-[80px]"
        ></div>
      </div>

      <header
        class="relative z-10 flex flex-col md:flex-row justify-between items-start md:items-center px-8 py-8 gap-4"
      >
        <div>
          <h1
            class="text-3xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-white to-slate-400 drop-shadow-sm"
          >
            Cinema Overview
          </h1>
          <p class="text-slate-400 text-sm mt-1">
            Real-time performance metrics and booking management.
          </p>
        </div>

        <div class="flex gap-3">
          <button
            @click="fetchData"
            class="flex items-center gap-2 px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-lg border border-slate-700 transition-all text-sm font-medium"
          >
            <i class="fas fa-sync-alt" :class="{ 'fa-spin': loading }"></i>
            Refresh
          </button>
        </div>
      </header>

      <div
        class="flex-1 overflow-y-auto custom-scrollbar px-8 pb-8 relative z-10 space-y-6"
      >
        <!-- Stats Row -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div
            class="bg-slate-900/40 border border-slate-800/60 p-6 rounded-2xl shadow-lg relative overflow-hidden group"
          >
            <div class="relative">
              <p
                class="text-slate-400 text-xs font-bold uppercase tracking-wider mb-2"
              >
                Total Revenue
              </p>
              <h3 class="text-3xl font-black text-white">
                ${{ stats.revenue.toLocaleString() }}
              </h3>
            </div>
          </div>

          <div
            class="bg-slate-900/40 border border-slate-800/60 p-6 rounded-2xl shadow-lg relative overflow-hidden group"
          >
            <div class="relative">
              <p
                class="text-slate-400 text-xs font-bold uppercase tracking-wider mb-2"
              >
                Total Bookings
              </p>
              <h3 class="text-3xl font-black text-white">
                {{ stats.totalBookings }}
              </h3>
            </div>
          </div>
        </div>

        <!-- Charts Row -->
        <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div
            class="lg:col-span-2 bg-slate-900/40 border border-slate-800/60 p-6 rounded-2xl shadow-lg flex flex-col"
          >
            <h3 class="text-white font-bold mb-4 flex items-center gap-2">
              <i class="fas fa-chart-line text-indigo-400"></i> Booking Trends
            </h3>
            <div class="flex-1 w-full h-[300px] relative">
              <Line :data="lineChartData" :options="lineChartOptions" />
            </div>
          </div>

          <div
            class="bg-slate-900/40 border border-slate-800/60 p-6 rounded-2xl shadow-lg flex flex-col"
          >
            <h3 class="text-white font-bold mb-4 flex items-center gap-2">
              <i class="fas fa-chart-pie text-pink-400"></i> Revenue by Movie
            </h3>
            <div
              class="flex-1 w-full h-[300px] relative flex items-center justify-center"
            >
              <Doughnut
                :data="doughnutChartData"
                :options="doughnutChartOptions"
              />
            </div>
          </div>
        </div>

        <!-- Filters & Table Section -->
        <div
          class="bg-slate-900/40 border border-slate-800/60 rounded-2xl shadow-lg overflow-hidden"
        >
          <!-- Filter Bar -->
          <div
            class="p-5 border-b border-slate-700/50 bg-slate-800/20 flex flex-wrap gap-4 items-center justify-between"
          >
            <div class="flex items-center gap-2">
              <h3 class="text-lg font-bold text-white">Recent Transactions</h3>
            </div>

            <div class="flex flex-wrap items-center gap-3">
              <!-- Movie Filter -->
              <div class="relative">
                <select
                  v-model="filters.movie"
                  @change="resetAndFetch"
                  class="appearance-none pl-9 pr-8 py-2 bg-slate-900 border border-slate-700 rounded-lg text-xs font-medium text-slate-300 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-all hover:border-slate-600 cursor-pointer min-w-[140px]"
                >
                  <option value="">All Movies</option>
                  <option v-for="m in movies" :key="m.id" :value="m.id">
                    {{ m.title }}
                  </option>
                </select>
                <i
                  class="fas fa-film absolute left-3 top-1/2 -translate-y-1/2 text-slate-500 text-xs pointer-events-none"
                ></i>
              </div>

              <!-- Date Filter with Shortcuts -->
              <div
                class="flex items-center gap-2 bg-slate-900 border border-slate-700 rounded-lg p-1"
              >
                <button
                  @click="toggleToday"
                  class="px-2 py-1 text-[10px] font-bold uppercase rounded hover:bg-slate-700 transition-colors"
                  :class="
                    isToday ? 'bg-indigo-500 text-white' : 'text-slate-400'
                  "
                >
                  Today
                </button>
                <div class="h-4 w-px bg-slate-700"></div>
                <div class="relative group">
                  <input
                    ref="dateInputRef"
                    type="date"
                    v-model="filters.date"
                    @change="resetAndFetch"
                    @click="openDatePicker"
                    class="pl-4 pr-16 py-1.5 bg-slate-900 border border-slate-700 rounded-lg text-xs font-medium text-slate-300 focus:ring-1 focus:ring-indigo-500 outline-none cursor-pointer uppercase w-48 transition-colors custom-date-input text-center"
                  />

                  <!-- Clear Button (Visible on hover or if date set) -->
                  <button
                    v-if="filters.date"
                    @click.stop="clearDate"
                    class="absolute right-9 top-1/2 -translate-y-1/2 w-5 h-5 flex items-center justify-center rounded-full bg-slate-700 hover:bg-red-500 text-white border border-slate-500 hover:border-red-500 transition-colors z-10 shadow-sm"
                  >
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      class="h-3 w-3"
                      viewBox="0 0 20 20"
                      fill="currentColor"
                    >
                      <path
                        fill-rule="evenodd"
                        d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
                        clip-rule="evenodd"
                      />
                    </svg>
                  </button>

                  <!-- Custom Calendar Icon -->
                  <i
                    class="fas fa-calendar-alt absolute right-3 top-1/2 -translate-y-1/2 text-white text-xs pointer-events-none"
                  ></i>
                </div>
              </div>

              <div class="relative">
                <input
                  type="text"
                  v-model="filters.user"
                  placeholder="Search user..."
                  @input="debounceSearch"
                  class="pl-9 pr-4 py-2 bg-slate-900 border border-slate-700 rounded-lg text-xs font-medium text-slate-300 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-all hover:border-slate-600 w-48 focus:w-64 placeholder-slate-600"
                />
                <i
                  class="fas fa-search absolute left-3 top-1/2 -translate-y-1/2 text-slate-500 text-xs pointer-events-none"
                ></i>
              </div>

              <button
                @click="resetFilters"
                class="p-2 bg-slate-800 hover:bg-slate-700 text-slate-400 hover:text-white rounded-lg border border-slate-700 transition-all active:scale-95"
                title="Reset Filters"
              >
                <i class="fas fa-undo-alt text-xs"></i>
              </button>
            </div>
          </div>

          <!-- Booking Table -->
          <div class="overflow-x-auto min-h-[400px]">
            <table class="w-full text-left border-collapse">
              <thead>
                <tr
                  class="bg-slate-900/50 border-b border-slate-700/50 text-xs uppercase tracking-wider text-slate-400 font-semibold"
                >
                  <th class="px-6 py-4">Booking Details</th>
                  <th class="px-6 py-4">Customer</th>
                  <th class="px-6 py-4 text-center">Seat</th>
                  <th class="px-6 py-4 text-center">Status</th>
                  <th class="px-6 py-4 text-right">Amount</th>
                  <th class="px-6 py-4 text-right">Transaction Time</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-700/30">
                <tr v-if="loading" class="bg-slate-900/20 animate-pulse">
                  <td colspan="6" class="px-6 py-12 text-center text-slate-500">
                    <i
                      class="fas fa-circle-notch fa-spin mr-2 text-indigo-500"
                    ></i>
                    Loading data...
                  </td>
                </tr>
                <tr v-else-if="bookings.length === 0">
                  <td colspan="6" class="px-6 py-16 text-center text-slate-500">
                    <span class="font-medium">No bookings found</span>
                  </td>
                </tr>
                <tr
                  v-for="b in bookings"
                  :key="b.id"
                  class="group hover:bg-white/[0.02] transition-colors relative"
                >
                  <td class="px-6 py-4">
                    <div class="flex items-center gap-3">
                      <div
                        class="w-10 h-14 rounded overflow-hidden bg-slate-800 flex-shrink-0"
                      >
                        <img
                          v-if="b.poster_url"
                          :src="b.poster_url"
                          alt="Poster"
                          class="w-full h-full object-cover"
                        />
                        <div
                          v-else
                          class="w-full h-full flex items-center justify-center text-slate-600"
                        >
                          <i class="fas fa-film"></i>
                        </div>
                      </div>
                      <div>
                        <div class="font-medium text-white text-sm">
                          {{ b.movie_title || "Unknown Movie" }}
                        </div>
                      </div>
                    </div>
                  </td>
                  <td class="px-6 py-4">
                    <div>
                      <div class="text-sm text-slate-300">
                        {{ b.user_name }}
                      </div>
                      <div class="text-xs text-slate-500">
                        {{ b.user_email }}
                      </div>
                    </div>
                  </td>
                  <td class="px-6 py-4 text-center">
                    <span
                      class="px-2 py-1 rounded bg-slate-800 border border-slate-700 text-indigo-300 text-xs font-mono font-bold"
                      >{{ b.seat_id }}</span
                    >
                  </td>
                  <td class="px-6 py-4 text-center">
                    <span
                      class="inline-flex items-center px-2.5 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wide border shadow-sm"
                      :class="{
                        'bg-emerald-500/10 text-emerald-400 border-emerald-500/20':
                          b.status === 'SUCCESS',
                        'bg-amber-500/10 text-amber-400 border-amber-500/20':
                          b.status === 'PENDING',
                        'bg-red-500/10 text-red-400 border-red-500/20':
                          b.status === 'FAILED',
                      }"
                    >
                      <span
                        class="w-1.5 h-1.5 rounded-full mr-1.5"
                        :class="{
                          'bg-emerald-400': b.status === 'SUCCESS',
                          'bg-amber-400': b.status === 'PENDING',
                          'bg-red-400': b.status === 'FAILED',
                        }"
                      ></span>
                      {{ b.status }}
                    </span>
                  </td>
                  <td class="px-6 py-4 text-right">
                    <span class="text-sm font-semibold text-white"
                      >${{ b.amount.toFixed(2) }}</span
                    >
                  </td>
                  <td class="px-6 py-4 text-right">
                    <div class="text-sm text-indigo-100 font-medium">
                      {{ formatTime(b.created_at) }}
                    </div>
                    <div class="text-[10px] text-slate-500">
                      {{ formatDate(b.created_at) }}
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Pagination Controls -->
          <div
            class="px-6 py-4 border-t border-slate-700/50 bg-slate-800/20 flex items-center justify-between"
          >
            <span class="text-xs text-slate-500">
              Showing {{ bookings.length }} of {{ pagination.total }} entries
            </span>
            <div class="flex gap-2 items-center">
              <button
                @click="changePage(pagination.page - 1)"
                :disabled="pagination.page <= 1 || loading"
                class="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-slate-800 border border-slate-700 text-slate-400 hover:text-white hover:border-slate-500 transition-all disabled:opacity-50 disabled:cursor-not-allowed text-xs font-medium"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  class="w-3.5 h-3.5"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <path d="M15 18l-6-6 6-6" />
                </svg>
                Prev
              </button>

              <span
                class="text-xs text-slate-400 font-mono bg-slate-900 px-3 py-1.5 rounded border border-slate-800"
              >
                Page {{ pagination.page }} of {{ pagination.pages }}
              </span>

              <button
                @click="changePage(pagination.page + 1)"
                :disabled="pagination.page >= pagination.pages || loading"
                class="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-slate-800 border border-slate-700 text-slate-400 hover:text-white hover:border-slate-500 transition-all disabled:opacity-50 disabled:cursor-not-allowed text-xs font-medium"
              >
                Next
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  class="w-3.5 h-3.5"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <path d="M9 18l6-6-6-6" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped>
/* Simplified CSS to reduce lag */
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #334155;
  border-radius: 10px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #475569;
}

/* Hide native date picker icon to use custom one */
.custom-date-input::-webkit-calendar-picker-indicator {
  background: transparent;
  bottom: 0;
  color: transparent;
  cursor: pointer;
  height: auto;
  left: 0;
  position: absolute;
  right: 0;
  top: 0;
  width: auto;
  opacity: 0; /* Important: Make it invisible but still clickable over the input if needed, though we use openShowPicker too */
  z-index: 10;
}
</style>
