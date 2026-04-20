import React, { useEffect } from "react";

const WebAuth: React.FC = () => {
  useEffect(() => {
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

    const script = document.createElement("script");
    script.src = "https://telegram.org/js/telegram-widget.js?22";
    script.async = true;

    script.setAttribute("data-telegram-login", "Midnight_poker_bot");
    script.setAttribute("data-size", "large");
    script.setAttribute("data-onauth", "onTelegramAuth(user)");
    script.setAttribute("data-request-access", "write");

    document.getElementById("tg-login")?.appendChild(script);
  }, []);

  return (
    <div style={{ textAlign: "center", marginTop: "100px", color: "white" }}>
      <h2>Вход через Telegram</h2>

      <div id="tg-login" />

      <br />

      <a href="https://t.me/Midnight_poker_bot">
        Открыть в Telegram
      </a>
    </div>
  );
};

export default WebAuth;