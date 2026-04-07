import React, { useEffect, useState } from "react";
import { Routes, Route, HashRouter, Navigate } from "react-router-dom";
import styled from "styled-components";
import { useTelegram } from "./hooks/useTelegram";
import { authAPI } from "./utils/api";
import Schedule from "./pages/Tournaments/Schedule";
import About from "./pages/About/About";
// import StartPage from "./pages/StartPage/WelcomePage";
// import RatingPage from "./pages/StartPage/RatingPage";
// import CurrentTournament from "./pages/Tournaments/CurrentTournament";
import Rating from "./pages/Rating/Rating";
import Profile from "./pages/Profile/Profile";
import Main from "./pages/Main/Main";
import Menu from "./pages/Menu/Menu";
import Support from "./pages/Support/Support";

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
  const { initData, isTelegram, isReady } = useTelegram();

  const [authError, setAuthError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const runAuth = async () => {
      try {
        if (isTelegram && !isReady) {
          return;
        }

        const existing = localStorage.getItem("auth_token");
        if (existing) {
          setLoading(false);
          return;
        }

        if (isTelegram) {
          if (!initData)
            throw new Error(
              "initData is empty — Telegram did not provide auth payload"
            );

          const response = await authAPI.telegramInitAuth(initData);

          if (!response.data?.token) {
            throw new Error("No token in API response");
          }

          setTimeout(() => {
            try {
              localStorage.setItem("auth_token", response.data.token);
            } catch (e) {
              console.warn("localStorage error:", e);
            }
          }, 300);
          setLoading(false);
          return;
        }

        setLoading(false);
      } catch (err: any) {
        setAuthError(
          err.response?.data?.error || err.message || "Unknown error"
        );
        setLoading(false);
      }
    };

    runAuth();
  }, [initData, isReady, isTelegram]);

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

        <button
          onClick={() => window.location.reload()}
          style={{
            marginTop: 12,
            padding: "10px 18px",
            borderRadius: 8,
            background: "#2196F3",
            color: "white",
            border: "none",
            cursor: "pointer",
          }}
        >
          Перезапустить
        </button>
      </Loader>
    );
  }

  return (
    <HashRouter>
      <>
        <Routes>
          <Route path="/" element={<Main />} />
          <Route path="/rating" element={<Rating />} />
          <Route path="/profile" element={<Profile />} />
          <Route path="/about" element={<About />} />
          <Route path="/tournaments" element={<Schedule />} />
          <Route path="/support" element={<Support />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
        <Menu />
      </>
    </HashRouter>
  );
};

export default App;
