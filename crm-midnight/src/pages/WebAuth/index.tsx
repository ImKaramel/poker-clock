import React, { useEffect, useState } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import background from "../../assets/background.jpg";

const WebAuth: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [hasToken, setHasToken] = useState(false);

  useEffect(() => {
    // Получаем параметры из URL
    const params = new URLSearchParams(location.search);
    let token = params.get('token');
    
    // Если токен не найден, проверяем весь URL (на случай проблем с роутером)
    if (!token && window.location.href.includes('token=')) {
      const match = window.location.href.match(/token=([^&]+)/);
      if (match) token = match[1];
    }
    
    const error = params.get('error');
    
    console.log("📍 WebAuth - Full URL:", window.location.href);
    console.log("📍 WebAuth - Token found:", token ? "YES" : "NO");
    
    if (token) {
      console.log("✅ Saving token to localStorage");
      localStorage.setItem("auth_token", token);
      setHasToken(true);
      
      // Очищаем URL от токена и перенаправляем на главную
      console.log("✅ Redirecting to main page");
      navigate("/", { replace: true });
      return;
    }
    
    if (error) {
      console.error("❌ Auth error:", error);
      alert(`Ошибка авторизации: ${error}`);
    }
  }, [location, navigate]);

  // Загружаем виджет только если нет токена
  useEffect(() => {
    if (hasToken) return;
    
    const container = document.getElementById("tg-login");
    if (!container) return;

    // Очищаем контейнер
    container.innerHTML = "";

    // Удаляем старые скрипты
    const existingScripts = document.querySelectorAll('script[src*="telegram-widget.js"]');
    existingScripts.forEach(script => script.remove());

    // Создаем новый скрипт
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
    
    console.log("Telegram widget loaded");
  }, [hasToken]);

  return (
    <div style={{
      textAlign: "center",
      marginTop: 50,
      color: "white",
      minHeight: "100vh",
      background: `linear-gradient(rgba(0, 0, 0, 0.7), rgba(0, 0, 0, 0.7)), url(${background})`,
      backgroundSize: "cover",
      backgroundPosition: "center",
      padding: "20px"
    }}>
      <h2>Вход через Telegram</h2>

      <div style={{
        margin: "30px 0",
        padding: "20px",
        borderRadius: "10px",
        backgroundColor: "rgba(0,0,0,0.5)",
        backdropFilter: "blur(5px)"
      }}>
        <div id="tg-login" style={{
          display: "flex",
          justifyContent: "center",
          minHeight: "100px",
          alignItems: "center"
        }}>
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