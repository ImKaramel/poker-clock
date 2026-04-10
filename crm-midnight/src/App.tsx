import React, { useEffect, useState } from "react";
import { Routes, Route, HashRouter, Navigate } from "react-router-dom";
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
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!user) return;
    if (!isTelegram) return;
    if (isReady === false) return;
    console.log(user)
    const runAuth = async () => {
      try {
        const response = await authAPI.telegramInitAuth(user);
  
        if (!response.data?.token) {
          throw new Error("No token in API response");
        }
  
        localStorage.setItem("auth_token", response.data.token);
        setLoading(false);
      } catch (e: any) {
        setAuthError(e.message);
        setLoading(false);
      }
    };
  
    runAuth();
  }, [user, isReady, isTelegram]);

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
          <Route path="/games" element={<Schedule />} />
          <Route path="/games/:id" element={<CurrentTournament />} />
          <Route path="/support" element={<Support />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
        <Menu />
      </>
    </HashRouter>
  );
};

export default App;
