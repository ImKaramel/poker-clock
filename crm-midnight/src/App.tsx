import React, { Suspense, lazy, useEffect, useMemo, useState } from "react";
import { Routes, Route, Navigate, useLocation } from "react-router-dom";
import styled from "styled-components";
import { useTelegram } from "./hooks/useTelegram";
import { authAPI } from "./utils/api";

const Schedule = lazy(() => import("./pages/Tournaments/Schedule"));
const About = lazy(() => import("./pages/About/About"));
const CurrentTournament = lazy(() => import("./pages/Tournaments/CurrentTournament"));
const Rating = lazy(() => import("./pages/Rating/Rating"));
const Profile = lazy(() => import("./pages/Profile/Profile"));
const Main = lazy(() => import("./pages/Main/Main"));
const Menu = lazy(() => import("./pages/Menu/Menu"));
const Support = lazy(() => import("./pages/Support/Support"));
const Welcome = lazy(() => import("./pages/StartPage/MainPage"));
const UserAgreement = lazy(() => import("./pages/StartPage/UserAgreement/UserAgreement"));
const StartPage = lazy(() => import("./pages/StartPage/WelcomePage"));
const RatingPage = lazy(() => import("./pages/StartPage/RatingPage"));
const WebAuth = lazy(() => import("./pages/WebAuth"));

const HIDE_MENU_ROUTES = new Set([
  "/start",
  "/useragreement",
  "/welcome-page",
  "/rating-page",
  "/web-auth",
]);

const checkTokenValidity = (token: string | null): boolean => {
  if (!token) return false;

  try {
    const payload = JSON.parse(atob(token.split(".")[1]));
    const currentTime = Date.now() / 1000;

    if (payload.exp && payload.exp < currentTime) {
      localStorage.removeItem("auth_token");
      return false;
    }

    return true;
  } catch {
    localStorage.removeItem("auth_token");
    return false;
  }
};

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
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);
  const [isLoadingAuthCheck, setIsLoadingAuthCheck] = useState<boolean>(true);

  const location = useLocation();
  const hideMenu = useMemo(() => HIDE_MENU_ROUTES.has(location.pathname), [location.pathname]);

  useEffect(() => {
    if (!isReady) return;

    if (!isTelegram) {
      const token = localStorage.getItem("auth_token");
      setIsAuthenticated(checkTokenValidity(token));
      setIsLoadingAuthCheck(false);
      return;
    }

    if (!user?.id) {
      setAuthError("Не удалось получить данные пользователя из Telegram.");
      setIsAuthenticated(false);
      setIsLoadingAuthCheck(false);
      return;
    }

    let isMounted = true;

    const runAuth = async () => {
      setIsLoadingAuthCheck(true);
      setAuthError(null);

      try {
        const response = await authAPI.telegramInitAuth({ user });
        const newToken = response.data.token;

        if (!isMounted) return;

        if (checkTokenValidity(newToken)) {
          localStorage.setItem("auth_token", newToken);
          setIsAuthenticated(true);
        } else {
          setIsAuthenticated(false);
          setAuthError("Сервер вернул невалидный токен авторизации.");
        }
      } catch (error: any) {
        if (!isMounted) return;
        setAuthError(error?.message || "Ошибка авторизации Telegram.");
        setIsAuthenticated(false);
      } finally {
        if (isMounted) {
          setIsLoadingAuthCheck(false);
        }
      }
    };

    runAuth();

    return () => {
      isMounted = false;
    };
  }, [isReady, isTelegram, user, location.pathname]);

  if (!isReady || isLoadingAuthCheck) {
    return (
      <Loader>
        <div>⏳ Загрузка...</div>
        <div style={{ fontSize: 14, opacity: 0.7 }}>
          {isLoadingAuthCheck ? "Проверка авторизации..." : "Инициализация Telegram WebApp..."}
        </div>
      </Loader>
    );
  }

  if (authError) {
    if (location.pathname !== "/web-auth") {
      return <Navigate to="/web-auth" replace state={{ authError }} />;
    }
  }

  if (!isAuthenticated && !isTelegram && location.pathname !== "/web-auth") {
    return <Navigate to="/web-auth" replace />;
  }

  return (
    <>
      <Suspense fallback={<Loader>⏳ Загрузка страницы...</Loader>}>
        <Routes>
          <Route path="/start" element={<Welcome />} />
          <Route path="/useragreement" element={<UserAgreement />} />
          <Route path="/welcome-page" element={<StartPage />} />
          <Route path="/rating-page" element={<RatingPage />} />
          <Route path="/web-auth" element={<WebAuth />} />

          {isAuthenticated ? (
            <>
              <Route path="/" element={<Main />} />
              <Route path="/rating" element={<Rating />} />
              <Route path="/profile" element={<Profile />} />
              <Route path="/about" element={<About />} />
              <Route path="/games" element={<Schedule />} />
              <Route path="/games/:id" element={<CurrentTournament />} />
              <Route path="/support" element={<Support />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </>
          ) : (
            <Route path="*" element={<Navigate to={isTelegram ? "/start" : "/web-auth"} replace />} />
          )}
        </Routes>

        {!hideMenu && <Menu />}
      </Suspense>
    </>
  );
};

export default App;
