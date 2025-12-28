<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import api from "../services/api";

const router = useRouter();
const movies = ref<any[]>([]);
const loading = ref(true);
const error = ref<string | null>(null);

const fetchMovies = async () => {
  try {
    loading.value = true;
    error.value = null; // Reset error state on retry
    const response = await api.get("/movies");
    movies.value = response.data;
  } catch (err: any) {
    console.error("Failed to fetch movies:", err);
    error.value = `Failed to load movies: ${err.message || err}`;
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  fetchMovies();
});

const goToBooking = (movieId: string, startTime: string) => {
  // Pass startTime as query for easier lookup (or could be param if simplified)
  router.push({
    name: "booking",
    params: { movieId },
    query: { time: startTime },
  });
};

const formatTime = (dateStr: string) => {
  const date = new Date(dateStr);
  return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
};
</script>

<template>
  <div class="container mx-auto px-6 py-10 min-h-screen">
    <div class="flex items-center justify-between mb-8">
      <h1 class="text-3xl font-bold flex items-center gap-3">
        <span
          class="w-1.5 h-8 bg-brand-red rounded-full shadow-[0_0_10px_theme('colors.brand.red')]"
        ></span>
        Now Showing
      </h1>
      <span class="text-gray-500 text-sm"
        >Showing {{ movies.length }} Movies</span
      >
    </div>

    <div
      v-if="loading"
      class="flex flex-col items-center justify-center py-20 bg-black/20 rounded-3xl border border-white/5 backdrop-blur-sm"
    >
      <div
        class="w-12 h-12 border-4 border-brand-red border-t-transparent rounded-full animate-spin mb-4 shadow-[0_0_15px_rgba(229,9,20,0.5)]"
      ></div>
      <p class="text-gray-400 font-medium animate-pulse">
        Loading amazing movies...
      </p>
    </div>

    <div
      v-else-if="error"
      class="bg-red-500/10 border border-red-500/20 p-6 rounded-2xl text-center max-w-lg mx-auto backdrop-blur-md"
    >
      <p class="text-red-400 font-medium mb-4">{{ error }}</p>
      <button
        @click="fetchMovies"
        class="px-6 py-2 bg-brand-red text-white rounded-xl hover:bg-red-700 transition-all font-semibold shadow-lg shadow-red-900/20"
      >
        Try Again
      </button>
    </div>

    <div
      v-else
      class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-6"
    >
      <div
        v-for="movie in movies"
        :key="movie.id"
        class="group relative bg-[#1A1A1A] rounded-2xl overflow-hidden hover:scale-[1.03] transition-all duration-300 shadow-xl hover:shadow-2xl hover:shadow-brand-red/10 border border-white/5 flex flex-col h-full"
      >
        <!-- Poster Area -->
        <div class="aspect-[2/3] relative overflow-hidden bg-gray-900">
          <img
            :src="movie.poster_url"
            :alt="movie.title"
            class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
          />

          <!-- Overlay Gradient on Hover -->
          <div
            class="absolute inset-0 bg-gradient-to-t from-[#1A1A1A] via-transparent to-transparent opacity-60 group-hover:opacity-90 transition-opacity duration-300"
          ></div>

          <!-- Top Badge -->
          <div
            class="absolute top-3 left-3 bg-black/60 backdrop-blur-md px-2 py-1 rounded-md border border-white/10 text-[10px] font-bold text-white uppercase tracking-wider shadow-lg"
          >
            {{ movie.genre.split("/")[0] }}
          </div>
        </div>

        <!-- Content Info -->
        <div class="p-4 flex flex-col flex-grow relative -mt-20 z-10">
          <h3
            class="text-lg font-bold text-white mb-2 leading-tight group-hover:text-brand-red transition-colors line-clamp-2 h-[3.5rem]"
            :title="movie.title"
          >
            {{ movie.title }}
          </h3>

          <div class="flex items-center gap-3 text-xs text-gray-400 mb-4">
            <span class="flex items-center gap-1">
              <svg
                class="w-3 h-3"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                ></path>
              </svg>
              {{ movie.duration_min }}m
            </span>
            <span class="w-1 h-1 bg-gray-600 rounded-full"></span>
            <span>{{ movie.genre }}</span>
          </div>

          <!-- Description (Optional, maybe hidden on small cards) -->
          <!-- <p class="text-xs text-gray-500 line-clamp-2 mb-4">{{ movie.description }}</p> -->

          <!-- Showtimes -->
          <div class="mt-auto pt-4 border-t border-white/5">
            <p
              class="text-[10px] font-bold text-gray-500 uppercase mb-2 tracking-wider"
            >
              Select Showtime
            </p>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="screening in movie.screenings"
                :key="screening.id"
                @click="goToBooking(movie.id, screening.start_time)"
                class="px-3 py-1.5 bg-white/5 hover:bg-brand-red text-gray-300 hover:text-white text-xs font-medium rounded-lg transition-all border border-white/5 hover:border-transparent hover:shadow-lg hover:shadow-red-900/40"
              >
                {{ formatTime(screening.start_time) }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
