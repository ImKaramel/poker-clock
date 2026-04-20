import React, { useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import background from "../../assets/background.jpg";

const WebAuth: React.FC = () => {
  const isScriptLoaded = useRef(false);
  // const navigate = useNavigate();

  useEffect(() => {
    console.log("WebAuth component mounted");

    // Убираем onTelegramAuth, используем data-auth-url
    const container = document.getElementById("tg-login");
    console.log("Container found:", container);

    if (!container) {
      console.error("Container with id 'tg-login' not found");
      return;
    }

    container.innerHTML = "";

    const existingScripts = document.querySelectorAll(
      'script[src*="telegram-widget.js"]'
    );
    console.log("Existing telegram scripts:", existingScripts.length);

    existingScripts.forEach((script) => script.remove());

    const script = document.createElement("script");
    script.src = "https://telegram.org/js/telegram-widget.js?22";
    script.async = true;
    script.setAttribute("data-telegram-login", "Midnight_poker_bot");
    script.setAttribute("data-size", "large");
    script.setAttribute("data-auth-url", "https://api.midnight-club-app.ru/api/auth/telegram/callback");
    script.setAttribute("data-request-access", "write");
    script.setAttribute("data-radius", "10");

    script.onload = () => {
      console.log("Telegram widget script loaded successfully");
    };

    script.onerror = (error) => {
      console.error("Failed to load Telegram widget:", error);
      container.innerHTML =
        '<div style="color: red; padding: 20px;">❌ Ошибка загрузки виджета Telegram</div>';
    };

    console.log("Appending script to container");
    container.appendChild(script);
    isScriptLoaded.current = true;

    return () => {
      console.log("WebAuth component unmounting");
    };
  }, []);

  return (
    <div
      style={{
        textAlign: "center",
        marginTop: 50,
        color: "white",
        minHeight: "100vh",
        background: "black",
        padding: "20px",
      }}
    >
      <h2>Вход через Telegram</h2>

      <div
        style={{
          margin: "30px 0",
          padding: "20px",
          border: "1px solid #333",
          borderRadius: "10px",
          position: "relative",
          overflow: "hidden",
          minHeight: "300px",
        }}
      >
        <div
          style={{
            position: "absolute",
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            backgroundImage: `url(${background})`,
            backgroundSize: "cover",
            backgroundPosition: "center",
            transform: "rotate(90deg) scale(1.4)",
            transformOrigin: "center",
            zIndex: 0,
          }}
        />
        
        <div
          style={{
            position: "absolute",
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            backgroundColor: "rgba(0, 0, 0, 0.6)",
            zIndex: 1,
          }}
        />

        <div
          style={{
            position: "relative",
            zIndex: 2,
          }}
        >
          <div
            id="tg-login"
            style={{
              display: "flex",
              justifyContent: "center",
              minHeight: "100px",
              alignItems: "center",
            }}
          >
            <div>Загрузка кнопки входа...</div>
          </div>
        </div>
      </div>

      <div>
        <a
          href="https://t.me/Midnight_poker_bot"
          target="_blank"
          rel="noopener noreferrer"
          style={{ color: "#24A1DE" }}
        >
          📱 Открыть бота в Telegram
        </a>
      </div>
    </div>
  );
};

export default WebAuth;