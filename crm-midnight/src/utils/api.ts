import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'https://api.midnight-club.ru/api';

export const api = axios.create({
  baseURL: API_BASE_URL,
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token');
  console.log('🔐 API Request - Token:', token ? 'YES' : 'NO');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

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
  telegramInitAuth: async (initData: string) => {
    console.log("📤 Sending initData:", initData.slice(0, 100) + "...");
    
    try {
      const response = await api.post("/auth/telegram/validate/", { initData });
      
      // 🔑 СОХРАНЯЕМ ТОКЕН - сервер возвращает его в поле "token"
      const token = response.data?.token;
      
      if (token) {
        localStorage.setItem('auth_token', token);
        console.log('✅ Token saved to localStorage');
        console.log('📝 Token value:', token.slice(0, 50) + '...');
      } else {
        console.error('❌ No token in response! Response data:', response.data);
      }
      
      return response;
    } catch (error) {
      console.error('❌ Auth error:', error);
      throw error;
    }
  },
};

export const adminAPI = {
  dashboard: () => {
    return api.get("/admin/dashboard/")
  }
};

export const gamesAPI = {
  getGames: () => api.get('/games/'),
  getGame: (id: number) => api.get(`/games/${id}/`),
  registerForGame: (gameId: number) => 
    api.post('/participants/register/', { game_id: gameId }),
  discardRegisterForGame: (gameId: number) => 
    api.delete('/participants/unregister/', { data: { game_id: gameId } }),
};

export const ratingAPI = {
  getRating: () => api.get('/rating/'),
};

export const profileAPI = {
  getProfile: () => api.get('/profile/'),
  updateProfile: (data: any) => api.patch('/profile/', data),
};

export const supportAPI = {
  createTicket: (data: { subject: string; message: string }) =>
    api.post('/support-tickets/', data),
  getTickets: () => api.get('/support-tickets/'),
};