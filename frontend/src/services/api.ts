import axios from 'axios';

const api = axios.create({
  baseURL: 'http://localhost:8080/api', 
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

export const movieApi = {
  getAll: () => api.get('/movies'),
};

export const screeningApi = {
  getDetails: (movieId: string, startTime: string) => api.post('/screenings/details', { movie_id: movieId, start_time: startTime }),
};

export const seatApi = {
  lock: (userId: string, movieId: string, startTime: string, seatId: string) => 
    api.post('/seats/lock', { user_id: userId, movie_id: movieId, start_time: startTime, seat_id: seatId }),
  
  book: (userId: string, movieId: string, startTime: string, seatId: string) =>
    api.post('/seats/book', { user_id: userId, movie_id: movieId, start_time: startTime, seat_id: seatId }),
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
