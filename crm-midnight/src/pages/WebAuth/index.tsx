import React, { useEffect } from "react";

const WebAuth: React.FC = () => {
  useEffect(() => {
    // 1. callback ДОЛЖЕН быть на window
    (window as any).onTelegramAuth = async (user: any) => {
      try {
        const res = await fetch("/api/auth/telegram/callback", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(user),
        });

        const data = await res.json();

        if (!data?.token) {
          throw new Error("No token received");
        }

        localStorage.setItem("auth_token", data.token);
        window.location.href = "/";
      } catch (e) {
        alert("Auth error");
      }
    };

    // 2. удалить старый widget (важно при HMR / dev)
    const oldScript = document.getElementById("telegram-login-script");
    if (oldScript) oldScript.remove();

    // 3. создать контейнер для Telegram (ВАЖНО)
    const container = document.getElementById("tg-login-container");

    if (container) {
      container.innerHTML = "";

      const script = document.createElement("script");
      script.id = "telegram-login-script";
      script.src = "https://telegram.org/js/telegram-widget.js?23";
      script.async = true;

      // ⚠️ это должен быть username бота БЕЗ https и БЕЗ /app
      script.setAttribute("data-telegram-login", "Midnight_poker_bot");

      script.setAttribute("data-size", "large");
      script.setAttribute("data-onauth", "onTelegramAuth(user)");
      script.setAttribute("data-request-access", "write");

      container.appendChild(script);
    }
  }, []);

  return (
    <div style={{ textAlign: "center", marginTop: "100px", color: "white" }}>
      <h2>Вход через Telegram</h2>

      {/* Telegram сам сюда отрисует кнопку */}
      <div id="tg-login-container" />

      <br />

      {/* это ТОЛЬКО fallback для открытия Mini App */}
      <a href="https://t.me/Midnight_poker_bot/app">
        Открыть приложение в Telegram
      </a>
    </div>
  );
};

export default WebAuth;