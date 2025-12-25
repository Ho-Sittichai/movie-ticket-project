<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

// Mock Data (10 items as requested)
const mockMovies = ref([
  {
    id: 1,
    title: "Avatar: The Way of Water",
    description: "Jake Sully lives with his newfound family formed on the extrasolar moon Pandora. Once a familiar threat returns to finish what was previously started, Jake must work with Neytiri and the army of the Na'vi race to protect their home.",
    genre: "Sci-Fi / Action",
    duration_min: 192,
    poster_url: "https://upload.wikimedia.org/wikipedia/en/5/54/Avatar_The_Way_of_Water_poster.jpg",
    screenings: [
       { id: 's1', start_time: '2025-12-26T10:00:00' },
       { id: 's2', start_time: '2025-12-26T14:00:00' },
       { id: 's3', start_time: '2025-12-26T18:00:00' }
    ]
  },
  {
    id: 2,
    title: "Oppenheimer",
    description: "The story of American scientist J. Robert Oppenheimer and his role in the development of the atomic bomb.",
    genre: "Biography / Drama",
    duration_min: 180,
     poster_url: "https://upload.wikimedia.org/wikipedia/en/4/4a/Oppenheimer_%28film%29.jpg",
    screenings: [
      { id: 's4', start_time: '2025-12-26T11:30:00' },
      { id: 's5', start_time: '2025-12-26T15:30:00' }
    ]
  },
  {
    id: 3,
    title: "Spider-Man: Across the Spider-Verse",
    description: "Miles Morales catapults across the Multiverse, where he encounters a team of Spider-People charged with protecting its very existence.",
    genre: "Animation / Action",
    duration_min: 140,
     poster_url: "https://upload.wikimedia.org/wikipedia/en/b/b4/Spider-Man-_Across_the_Spider-Verse_poster.jpg",
    screenings: [
      { id: 's6', start_time: '2025-12-26T12:00:00' },
      { id: 's7', start_time: '2025-12-26T16:00:00' }
    ]
  },
  {
    id: 4,
    title: "The Batman",
    description: "When a sadistic serial killer begins murdering key political figures in Gotham, Batman is forced to investigate the city's hidden corruption and question his family's involvement.",
    genre: "Action / Crime",
    duration_min: 176,
     poster_url: "https://upload.wikimedia.org/wikipedia/en/f/ff/The_Batman_%28film%29_poster.jpg",
    screenings: [
      { id: 's8', start_time: '2025-12-26T19:00:00' },
      { id: 's9', start_time: '2025-12-26T22:00:00' }
    ]
  },
  {
    id: 5,
    title: "Guardians of the Galaxy Vol. 3",
    description: "Still reeling from the loss of Gamora, Peter Quill rallies his team to defend the universe and one of their own - a mission that could mean the end of the Guardians if not successful.",
    genre: "Action / Adventure",
    duration_min: 150,
    poster_url: "https://upload.wikimedia.org/wikipedia/en/7/74/Guardians_of_the_Galaxy_Vol._3_poster.jpg",
    screenings: [
      { id: 's10', start_time: '2025-12-26T13:00:00' }
    ]
  },
  {
    id: 6,
    title: "Dune: Part Two",
    description: "Paul Atreides unites with Chani and the Fremen while on a warpath of revenge against the conspirators who destroyed his family.",
    genre: "Sci-Fi / Adventure",
    duration_min: 166,
    poster_url: "https://www.siamzone.com/movie/pic/2024/duneparttwo/poster1.jpg",
    screenings: [
      { id: 's11', start_time: '2025-12-26T14:30:00' },
      { id: 's12', start_time: '2025-12-26T18:30:00' }
    ]
  },
  {
    id: 7,
    title: "Mission: Impossible - Dead Reckoning",
    description: "Ethan Hunt and his IMF team must track down a dangerous new weapon before it falls into the wrong hands.",
    genre: "Action / Thriller",
    duration_min: 163,
    poster_url: "https://theatrgwaun.com/wp-content/uploads/2023/07/TG-Aug23-web-700px-mission-768x768.jpg",
    screenings: [
       { id: 's13', start_time: '2025-12-26T10:30:00' },
       { id: 's14', start_time: '2025-12-26T15:00:00' }
    ]
  },
  {
    id: 8,
    title: "Barbie",
    description: "Barbie suffers a crisis that leads her to question her world and her existence.",
    genre: "Adventure / Comedy",
    duration_min: 114,
    poster_url: "https://upload.wikimedia.org/wikipedia/en/0/0b/Barbie_2023_poster.jpg",
    screenings: [
       { id: 's15', start_time: '2025-12-26T11:00:00' },
       { id: 's16', start_time: '2025-12-26T13:30:00' },
       { id: 's17', start_time: '2025-12-26T16:00:00' }
    ]
  },
  {
    id: 9,
    title: "John Wick: Chapter 4",
    description: "John Wick uncovers a path to defeating The High Table. But before he can earn his freedom, Wick must face off against a new enemy with powerful alliances across the globe.",
    genre: "Action / Crime",
    duration_min: 169,
    poster_url: "https://assets-prd.ignimgs.com/2023/02/08/jw4-2025x3000-online-character-1sht-keanu-v187-1675886090936.jpg",
    screenings: [
       { id: 's18', start_time: '2025-12-26T20:00:00' }
    ]
  },
  {
    id: 10,
    title: "Inside Out 2",
    description: "Joy, Sadness, Anger, Fear and Disgust have been running a successful operation by all accounts. However, when Anxiety shows up, they aren't sure how to feel.",
    genre: "Animation / Family",
    duration_min: 100,
    poster_url: "https://upload.wikimedia.org/wikipedia/en/f/f7/Inside_Out_2_poster.jpg",
    screenings: [
       { id: 's19', start_time: '2025-12-26T09:00:00' },
       { id: 's20', start_time: '2025-12-26T11:00:00' }
    ]
  }
]) // End Mock

const goToBooking = (screeningId: string) => {
  router.push(`/booking/${screeningId}`)
}

const formatTime = (dateStr: string) => {
  const date = new Date(dateStr)
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}
</script>

<template>
  <div class="container mx-auto px-6 py-10 min-h-screen">
    <div class="flex items-center justify-between mb-8">
      <h1 class="text-3xl font-bold flex items-center gap-3">
        <span class="w-1.5 h-8 bg-brand-red rounded-full shadow-[0_0_10px_theme('colors.brand.red')]"></span>
        Now Showing
      </h1>
      <span class="text-gray-500 text-sm">Showing 10 Movies</span>
    </div>
    
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 xl:grid-cols-5 gap-6">
      <div v-for="movie in mockMovies" :key="movie.id" class="group relative bg-[#1A1A1A] rounded-2xl overflow-hidden hover:scale-[1.03] transition-all duration-300 shadow-xl hover:shadow-2xl hover:shadow-brand-red/10 border border-white/5 flex flex-col h-full">
        
        <!-- Poster Area -->
        <div class="aspect-[2/3] relative overflow-hidden bg-gray-900">
          <img :src="movie.poster_url" :alt="movie.title" class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110">
          
          <!-- Overlay Gradient on Hover -->
          <div class="absolute inset-0 bg-gradient-to-t from-[#1A1A1A] via-transparent to-transparent opacity-60 group-hover:opacity-90 transition-opacity duration-300"></div>

          <!-- Top Badge -->
          <div class="absolute top-3 left-3 bg-black/60 backdrop-blur-md px-2 py-1 rounded-md border border-white/10 text-[10px] font-bold text-white uppercase tracking-wider shadow-lg">
             {{ movie.genre.split('/')[0] }}
          </div>
        </div>

        <!-- Content Info -->
        <div class="p-4 flex flex-col flex-grow relative -mt-20 z-10">
          <h3 class="text-lg font-bold text-white mb-2 leading-tight group-hover:text-brand-red transition-colors line-clamp-2 h-[3.5rem]" :title="movie.title">
            {{ movie.title }}
          </h3>
          
          <div class="flex items-center gap-3 text-xs text-gray-400 mb-4">
            <span class="flex items-center gap-1">
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
              {{ movie.duration_min }}m
            </span>
            <span class="w-1 h-1 bg-gray-600 rounded-full"></span>
            <span>{{ movie.genre }}</span>
          </div>

          <!-- Description (Optional, maybe hidden on small cards) -->
           <!-- <p class="text-xs text-gray-500 line-clamp-2 mb-4">{{ movie.description }}</p> -->

          <!-- Showtimes -->
          <div class="mt-auto pt-4 border-t border-white/5">
            <p class="text-[10px] font-bold text-gray-500 uppercase mb-2 tracking-wider">Select Showtime</p>
            <div class="flex flex-wrap gap-2">
              <button 
                v-for="screening in movie.screenings" 
                :key="screening.id"
                @click="goToBooking(screening.id)"
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
