import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'https://api.midnight-club-app.ru/api';

export const api = axios.create({
  baseURL: API_BASE_URL,
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token');
  console.log('🔐 API Request - Token:', token ? 'YES' : 'NO');
  // const token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiI0LjYzMDIxNTcyZSswOCIsImFkbSI6dHJ1ZSwic3ViIjoiNC42MzAyMTU3MmUrMDgiLCJleHAiOjE3NzYyNzk4MTYsImlhdCI6MTc3NTY3NTAxNn0.123htHIfVkHegEZaBpqYYqcrkqWpo7ubRgpPM7yhkRI'
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Добавьте обработку 401 ошибок
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      console.log('❌ 401 Unauthorized - clearing token');
      localStorage.removeItem('auth_token');
    }
    return Promise.reject(error);
  }
);

export const authAPI = {
  telegramInitAuth: (user: any) => {
    console.log("📤 Sending user:", user);
    return api.post("/auth/telegram", user);
  },
};
export const adminAPI = {
   dashboard: () => {
   return api.get("/admin/dashboard/")
  }
};

export const gamesAPI = {
  getGames: () => api.get('/games'),
  getGame: (id: number) => api.get(`/games/${id}`),
  registerForGame: (gameId: number) => 
    api.post('/participants/register', { game_id: gameId }),
  discardRegisterForGame: (gameId: number) => 
    api.delete('/participants/unregister', { data: { game_id: gameId } }),
  getParticipantsAdmin: (gameId: number) =>
    api.get(`/games/${gameId}/participants_admin`),
};

export const ratingAPI = {
  getRating: () => api.get('/rating'),
};

export const profileAPI = {
  getProfile: () => api.get('/profile'),
  updateProfile: (nick_name: string) => api.patch('/profile/', { nick_name }),
};

export const supportAPI = {
  createTicket: (data: { subject: string; message: string }) =>
    api.post('/support-tickets/', data),
  getTickets: () => api.get('/support-tickets/'),
};