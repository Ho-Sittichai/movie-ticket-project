import axios from 'axios';
import { useToast } from '../composables/useToast';
import { useAuthStore } from '../stores/auth';

const api = axios.create({
  baseURL: 'http://localhost:8080/api', 
  timeout: 10000, // 10 seconds timeout
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token'); // Simplest way to get token without importing store (avoid circular dependency)
    // Alternatively, if pinia persistance is used, we might parse it. 
    // Or we can just import the store inside the interceptor function to avoid circular dep at module level.
    
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor for global error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response && error.response.status === 401) {
      // 401 Unauthorized -> Session Expired or Invalid Token
      const authStore = useAuthStore();
      
      // Only clear if we currently think we are logged in
      if (authStore.token) {
        authStore.logout();
        
        // Trigger Toast
        const { showToast } = useToast();
        showToast('Session expired. Please log in again.', 'error');
        
        // Optional: Open login modal to prompt user
        authStore.openLoginModal();
      }
    }
    return Promise.reject(error);
  }
);

export const movieApi = {
  getAll: () => api.get('/movies'),
};

export const paymentApi = {
  start: (userId: string, movieId: string, startTime: string, seatIds: string[]) => 
    api.post('/payment/start', { user_id: userId, movie_id: movieId, start_time: startTime, seat_ids: seatIds }),
  cancel: (reason?: string) => api.post('/payment/cancel', { reason: reason || 'user_cancelled' }),
};
export const screeningApi = {
  getDetails: (movieId: string, startTime: string) => api.post('/screenings/details', { movie_id: movieId, start_time: startTime }),
};

export const seatApi = {
  lock: (userId: string, movieId: string, startTime: string, seatId: string) => 
    api.post('/seats/lock', { user_id: userId, movie_id: movieId, start_time: startTime, seat_id: seatId }),
  
  book: (userId: string, movieId: string, startTime: string, seatIds: string[], paymentId?: string) =>
    api.post('/seats/book', { user_id: userId, movie_id: movieId, start_time: startTime, seat_ids: seatIds, payment_id: paymentId }),

  extend: (userId: string, movieId: string, startTime: string, seatIds: string[]) =>
    api.post('/seats/extend', { user_id: userId, movie_id: movieId, start_time: startTime, seat_ids: seatIds }),
};

export const adminApi = {
  getBookings: (params: any) => {
    const queryParams = new URLSearchParams();
    if (params.movie) queryParams.append('movie_id', params.movie);
    if (params.date) queryParams.append('date', params.date);
    if (params.user) queryParams.append('user', params.user);
    return api.get(`/admin/bookings?${queryParams.toString()}`);
  }
};

export default api;
