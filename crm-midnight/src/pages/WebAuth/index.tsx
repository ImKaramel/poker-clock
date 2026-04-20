import React, { useEffect, useRef } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import background from "../../assets/background.jpg";

const WebAuth: React.FC = () => {
  const isScriptLoaded = useRef(false);
  const navigate = useNavigate();
  const location = useLocation();

  // ✅ САМОЕ ВАЖНОЕ: обрабатываем параметры URL при загрузке
  useEffect(() => {
    const urlParams = new URLSearchParams(location.search);
    const token = urlParams.get('token');
    const error = urlParams.get('error');
    
    console.log("📍 WebAuth URL params:", { token: !!token, error });
    
    if (token) {
      console.log("✅ Token found, saving to localStorage");
      localStorage.setItem("auth_token", token);
      console.log("✅ Token saved, redirecting to main page");
      navigate("/");
      return;
    }
    
    if (error) {
      console.error("❌ Auth error from backend:", error);
      alert(`Ошибка авторизации: ${error}. Пожалуйста, попробуйте снова.`);
    }
  }, [location, navigate]);

  // Загружаем виджет только если нет токена в URL
  useEffect(() => {
    // Если уже есть токен в URL, не загружаем виджет заново
    const urlParams = new URLSearchParams(location.search);
    if (urlParams.get('token')) {
      return;
    }
    
    console.log("Loading Telegram widget...");
    const container = document.getElementById("tg-login");
    if (!container) return;

    container.innerHTML = "";

    const existingScripts = document.querySelectorAll(
      'script[src*="telegram-widget.js"]'
    );
    existingScripts.forEach((script) => script.remove());

    const script = document.createElement("script");
    script.src = "https://telegram.org/js/telegram-widget.js?22";
    script.async = true;
    script.setAttribute("data-telegram-login", "Midnight_poker_bot");
    script.setAttribute("data-size", "large");
    script.setAttribute("data-auth-url", "https://api.midnight-club-app.ru/api/auth/telegram/callback");
    script.setAttribute("data-request-access", "write");
    script.setAttribute("data-radius", "10");

    script.onerror = () => {
      container.innerHTML = '<div style="color: red; padding: 20px;">❌ Ошибка загрузки виджета Telegram</div>';
    };

    container.appendChild(script);
  }, [location.search]); // Добавили зависимость от location.search

  return (
    <div
      style={{
        textAlign: "center",
        marginTop: 50,
        color: "white",
        minHeight: "100vh",
        background: `linear-gradient(rgba(0, 0, 0, 0.7), rgba(0, 0, 0, 0.7)), url(${background})`,
        backgroundSize: "cover",
        backgroundPosition: "center",
        padding: "20px",
      }}
    >
      <h2>Вход через Telegram</h2>

      <div
        style={{
          margin: "30px 0",
          padding: "20px",
          borderRadius: "10px",
          backgroundColor: "rgba(0,0,0,0.5)",
          backdropFilter: "blur(5px)",
        }}
      >
        <div id="tg-login" style={{ display: "flex", justifyContent: "center", minHeight: "100px", alignItems: "center" }}>
          <div>Загрузка кнопки входа...</div>
        </div>
      </div>

      <div>
        <a href="https://t.me/Midnight_poker_bot" target="_blank" rel="noopener noreferrer" style={{ color: "#24A1DE" }}>
          📱 Открыть бота в Telegram
        </a>
      </div>
    </div>
  );
};

export default WebAuth;