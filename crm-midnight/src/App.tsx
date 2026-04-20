import React, { useEffect, useState } from "react";
import { Routes, Route, Navigate, useLocation } from "react-router-dom";
import styled from "styled-components";
import { useTelegram } from "./hooks/useTelegram";
import { authAPI } from "./utils/api";

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

  const [loading, setLoading] = useState(true);
  const [authError, setAuthError] = useState<string | null>(null);
  const [initialRoute, setInitialRoute] = useState<string | null>(null);

  const location = useLocation();

const hideMenuRoutes = [
  "/start",
  "/useragreement",
  "/welcome-page",
  "/rating-page",
];

const hideMenu = hideMenuRoutes.includes(location.pathname);

  useEffect(() => {
    if (!isReady) return;

    if (!isTelegram) {
      setInitialRoute("/web-auth");
      setLoading(false);
      return;
    }

    if (!user) {
      setAuthError("Telegram user not found");
      setLoading(false);
      return;
    }

    if (!user || !isTelegram || !isReady) return;

    const runAuth = async () => {
      try {
        const response = await authAPI.telegramInitAuth({ user });

        if (!response.data?.token) {
          throw new Error("No token in API response");
        }

        localStorage.setItem("auth_token", response.data.token);

        // 👇 ВАЖНО: решаем маршрут ДО рендера
        setInitialRoute(response.data.isNew ? "/start" : "/");

      } catch (e: any) {
        setAuthError(e.message);
      } finally {
        setLoading(false);
      }
    };

    runAuth();
  }, [user, isTelegram, isReady]);

  if (loading) {
    return (
      <Loader>
        <div>⏳ Загрузка...</div>
        <div style={{ fontSize: 14, opacity: 0.7 }}>
          {isTelegram ? "Ожидание Telegram WebApp…" : "Ожидание…"}
        </div>
      </Loader>
    );
  }

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

  return (
    <>
      <Routes>
        <Route path="/" element={<Main />} />
        <Route path="/rating" element={<Rating />} />
        <Route path="/profile" element={<Profile />} />
        <Route path="/about" element={<About />} />
        <Route path="/games" element={<Schedule />} />
        <Route path="/games/:id" element={<CurrentTournament />} />
        <Route path="/support" element={<Support />} />

        <Route path="/start" element={<Welcome />} />
        <Route path="/useragreement" element={<UserAgreement />} />
        <Route path="/welcome-page" element={<StartPage />} />
        <Route path="/rating-page" element={<RatingPage />} />
        <Route path="/web-auth" element={<WebAuth />} />

        {initialRoute && (
          <Route path="*" element={<Navigate to={initialRoute} replace />} />
        )}
      </Routes>

      {!hideMenu && <Menu />}
    </>
  );
};

export default App;