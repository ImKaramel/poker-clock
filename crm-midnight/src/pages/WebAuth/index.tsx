import React, { useEffect } from "react";

const WebAuth: React.FC = () => {
  useEffect(() => {
    // callback обязательно на window
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

        localStorage.setItem("auth_token", data.token);
        window.location.href = "/";
      } catch {
        alert("Auth error");
      }
    };

    const container = document.getElementById("tg-login");

    if (!container) return;

    // ❗ очистка обязательна
    container.innerHTML = "";

    // создаём script
    const script = document.createElement("script");
    script.src = "https://telegram.org/js/telegram-widget.js?23";
    script.async = true;

    script.setAttribute("data-telegram-login", "Midnight_poker_bot");
    script.setAttribute("data-size", "large");
    script.setAttribute("data-onauth", "onTelegramAuth(user)");
    script.setAttribute("data-request-access", "write");

    container.appendChild(script);
  }, []);

  return (
    <div style={{ textAlign: "center", marginTop: 100, color: "white" }}>
      <h2>Вход через Telegram</h2>

      {/* ВАЖНО: Telegram сам сюда вставит кнопку */}
      <div id="tg-login" />

      <br />

      <a href="https://t.me/Midnight_poker_bot/app">
        Открыть приложение в Telegram
      </a>
    </div>
  );
};

export default WebAuth;