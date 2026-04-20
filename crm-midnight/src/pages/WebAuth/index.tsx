import React, { useEffect } from "react";

const WebAuth: React.FC = () => {
  useEffect(() => {
    // 1. обязательно на window ДО загрузки скрипта
    (window as any).onTelegramAuth = (user: any) => {
      fetch("/api/auth/telegram/callback", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(user),
      })
        .then(res => res.json())
        .then(data => {
          localStorage.setItem("auth_token", data.token);
          window.location.href = "/";
        })
        .catch(() => alert("Auth error"));
    };

    // 2. удалить старый скрипт если есть
    const oldScript = document.getElementById("telegram-login-script");
    if (oldScript) oldScript.remove();

    // 3. создать новый script
    const script = document.createElement("script");
    script.id = "telegram-login-script";
    script.src = "https://telegram.org/js/telegram-widget.js?23";
    script.async = true;

    script.setAttribute("data-telegram-login", "Midnight_poker_bot");
    script.setAttribute("data-size", "large");
    script.setAttribute("data-onauth", "onTelegramAuth(user)");
    script.setAttribute("data-request-access", "write");

    // 4. вставка прямо в body (ВАЖНО)
    document.body.appendChild(script);
  }, []);

  return (
    <div style={{ textAlign: "center", marginTop: "100px", color: "white" }}>
      <h2>Вход через Telegram</h2>

      {/* сюда Telegram сам вставит кнопку */}
      <div />

      <br />

      <a href="https://t.me/Midnight_poker_bot">
        Открыть в Telegram
      </a>
    </div>
  );
};

export default WebAuth;