import React, { useEffect, useState, useCallback } from "react"; // Добавляем useCallback
import { Routes, Route, Navigate, useLocation, useNavigate } from "react-router-dom";
import styled from "styled-components";
import { useTelegram } from "./hooks/useTelegram";
import { authAPI } from "./utils/api";

// ... (импорты страниц остаются без изменений)
import Schedule from "./pages/Tournaments/Schedule";
import About from "./pages/About/About";
import CurrentTournament from "./pages/Tournaments/CurrentTournament";
import Rating from "./pages/Rating/Rating";
import Profile from "./pages/Profile/Profile";
import Main from "./pages/Main/Main";
import Menu from "./pages/Menu/Menu";
import Support from "./pages/Support/Support";
import Welcome from "./pages/StartPage/MainPage";
import UserAgreement from "./pages/StartPage/UserAgreement/UserAgreement";
import StartPage from "./pages/StartPage/WelcomePage";
import RatingPage from "./pages/StartPage/RatingPage";
import WebAuth from "./pages/WebAuth";

const Loader = styled.div`
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100vh;
  font-size: 18px;
  gap: 15px;
  color: white;
  background: black;
`;

const App: React.FC = () => {
  const { user, isTelegram, isReady } = useTelegram();

  const [authError, setAuthError] = useState<string | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false); // Добавляем состояние авторизации
  const [isLoadingAuthCheck, setIsLoadingAuthCheck] = useState<boolean>(true); // Состояние для проверки токена при старте

  const location = useLocation();
  // const navigate = useNavigate();

  const hideMenuRoutes = [
    "/start",
    "/useragreement",
    "/welcome-page",
    "/rating-page",
    "/web-auth",
  ];

  const hideMenu = hideMenuRoutes.includes(location.pathname);

  // Функция для проверки токена
  const checkTokenValidity = useCallback((token: string | null): boolean => {
    if (!token) return false;
    try {
      const payload = JSON.parse(atob(token.split('.')[1])); // Декодируем payload JWT
      const currentTime = Date.now() / 1000; // Текущее время в секундах
      if (payload.exp && payload.exp < currentTime) {
        console.warn("Token expired. Removing from localStorage.");
        localStorage.removeItem("auth_token");
        return false;
      }
      return true;
    } catch (e) {
      console.error("Error decoding or validating token:", e);
      localStorage.removeItem("auth_token"); // Удаляем некорректный токен
      return false;
    }
  }, []);

  // 📱 TELEGRAM AUTH - Если пользователь через Telegram, получаем токен от бэкенда
  useEffect(() => {
    if (!isReady || !isTelegram || !user) {
      // Если не готово, не Telegram, или нет user, то Telegram auth не запускаем.
      // Но все равно устанавливаем isLoadingAuthCheck в false, если это не Telegram
      // и все готово, чтобы перейти к проверке браузерного токена.
      if (isReady && !isTelegram) {
          setIsLoadingAuthCheck(false); // Для браузерных пользователей, после isReady, сразу проверяем токен
      }
      return;
    }

    const runAuth = async () => {
      setIsLoadingAuthCheck(true); // Начинаем проверку auth
      try {
        const response = await authAPI.telegramInitAuth({ user });
        const newToken = response.data.token;
        if (checkTokenValidity(newToken)) { // Проверяем валидность токена из ответа
            localStorage.setItem("auth_token", newToken);
            setIsAuthenticated(true);
        } else {
            setIsAuthenticated(false);
            setAuthError("Received an invalid or expired token from Telegram auth.");
        }
      } catch (e: any) {
        setAuthError(e.message);
        setIsAuthenticated(false);
      } finally {
        setIsLoadingAuthCheck(false); // Завершили проверку auth
      }
    };

    runAuth();
  }, [user, isTelegram, isReady, checkTokenValidity]); // Добавляем checkTokenValidity в зависимости

  // 🌐 BROWSER AUTH CHECK - Проверяем наличие токена в localStorage для браузерных пользователей
  useEffect(() => {
    if (isLoadingAuthCheck) return; // Ждем завершения первичной проверки (например, от Telegram)
    if (isTelegram) return; // Для Telegram-пользователей эта логика не нужна, они обрабатываются выше.

    const token = localStorage.getItem("auth_token");
    console.log("App.tsx: Browser auth check. Token found:", token ? "YES" : "NO");

    if (checkTokenValidity(token)) {
      setIsAuthenticated(true);
    } else {
      setIsAuthenticated(false);
    }
    // Здесь setIsLoadingAuthCheck уже должен быть false, если не было isTelegram
  }, [isLoadingAuthCheck, isTelegram, checkTokenValidity]);


  // --- Экраны загрузки и ошибок ---

  // Если еще ждем готовности Telegram WebApp или идет проверка авторизации
  if (!isReady || isLoadingAuthCheck) {
    return (
      <Loader>
        <div>⏳ Загрузка...</div>
        <div style={{ fontSize: 14, opacity: 0.7 }}>
          {isLoadingAuthCheck ? "Проверка авторизации..." : "Ожидание Telegram WebApp..."}
        </div>
      </Loader>
    );
  }

  // Если возникла ошибка авторизации
  if (authError) {
    return (
      <Loader>
        <h2>❌ Ошибка авторизации</h2>
        <p>{authError}</p>
        <button onClick={() => window.location.reload()}>
          Перезапустить
        </button>
      </Loader>
    );
  }

  // --- Основная логика роутинга ---

  // Если пользователь не авторизован (и это не Telegram WebApp),
  // и он пытается попасть на защищенную страницу, перенаправляем на /web-auth.
  // Исключаем /web-auth из редиректа, чтобы не зациклиться
  if (!isAuthenticated && !isTelegram && location.pathname !== '/web-auth') {
    console.log("App.tsx: Unauthenticated browser user, redirecting to /web-auth");
    // Используем компонент Navigate для декларативного редиректа в рендере
    return <Navigate to="/web-auth" replace />;
  }

  return (
    <>
      <Routes>
        {/* Общедоступные маршруты (не требуют авторизации, или их auth-логика внутри компонентов) */}
        <Route path="/start" element={<Welcome />} />
        <Route path="/useragreement" element={<UserAgreement />} />
        <Route path="/welcome-page" element={<StartPage />} />
        <Route path="/rating-page" element={<RatingPage />} />
        <Route path="/web-auth" element={<WebAuth />} />

        {/* Защищенные маршруты */}
        {(isAuthenticated || isTelegram) ? ( // Если авторизован ИЛИ это пользователь Telegram
          <>
            <Route path="/" element={<Main />} />
            <Route path="/rating" element={<Rating />} />
            <Route path="/profile" element={<Profile />} />
            <Route path="/about" element={<About />} />
            <Route path="/games" element={<Schedule />} />
            <Route path="/games/:id" element={<CurrentTournament />} />
            <Route path="/support" element={<Support />} />
          </>
        ) : (
          // Если не авторизован и не Telegram, все остальные пути редиректятся на /web-auth
          // (хотя это уже должно быть поймано предыдущим Navigate, но как запасной вариант)
          <Route path="*" element={<Navigate to="/web-auth" replace />} />
        )}

        {/* Маршрут по умолчанию для всех остальных случаев, если не пойманы выше */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>

      {!hideMenu && <Menu />}
    </>
  );
};

export default App;