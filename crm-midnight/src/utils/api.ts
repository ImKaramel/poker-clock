import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'https://api.midnight-club-app.ru/api';
const API_FALLBACK_URL = process.env.REACT_APP_API_FALLBACK_URL;
const API_BASE_URLS = [
  API_BASE_URL,
  ...(API_FALLBACK_URL ? [API_FALLBACK_URL] : []),
];

export const api = axios.create({
  baseURL: API_BASE_URLS[0],
  timeout: 12000,
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token');

  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Добавьте обработку 401 ошибок
api.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error.response?.status;

    if (status === 401) {
      localStorage.removeItem('auth_token');
      return Promise.reject(error);
    }

    const isNetworkError = !error.response;
    const isTimeout = error.code === 'ECONNABORTED';
    const isRetryableStatus = [502, 503, 504].includes(status);
    const requestConfig = error.config as any;

    if (
      requestConfig &&
      !requestConfig.__retryWithFallback &&
      API_BASE_URLS.length > 1 &&
      (isNetworkError || isTimeout || isRetryableStatus)
    ) {
      const currentBaseUrl = requestConfig.baseURL || api.defaults.baseURL || API_BASE_URLS[0];
      const currentIndex = API_BASE_URLS.indexOf(currentBaseUrl);
      const nextIndex = currentIndex >= 0 ? currentIndex + 1 : 1;

      if (nextIndex < API_BASE_URLS.length) {
        requestConfig.__retryWithFallback = true;
        requestConfig.baseURL = API_BASE_URLS[nextIndex];
        return api(requestConfig);
      }
    }

    return Promise.reject(error);
  }
);

export const authAPI = {
  telegramInitAuth: (user: any) => {
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
  updateProfile: async (nick_name: string) => {
    try {
      return await api.patch('/profile/', { nick_name });
    } catch (error: any) {
      const status = error?.response?.status;
      if (status === 404 || status === 405) {
        return api.patch('/profile', { nick_name });
      }
      throw error;
    }
  },
};

export const supportAPI = {
  createTicket: (data: { subject: string; message: string }) =>
    api.post('/support-tickets/', data),
  getTickets: () => api.get('/support-tickets/'),
};
