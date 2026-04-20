import React, { useEffect, useRef } from "react";

const WebAuth: React.FC = () => {
  const isScriptLoaded = useRef(false);

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
    
    if (!container) {
      console.error("Container not found");
      return;
    }

    // ❗ очистка обязательна
    container.innerHTML = "";

    // Проверяем, не загружен ли уже скрипт
    if (document.querySelector('script[src*="telegram-widget.js"]') && !isScriptLoaded.current) {
      isScriptLoaded.current = true;
      // Если скрипт уже есть, но виджет не отображается, вызываем принудительную перерисовку
      const existingScript = document.querySelector('script[src*="telegram-widget.js"]');
      if (existingScript) {
        existingScript.remove();
      }
    }

    // создаём script
    const script = document.createElement("script");
    script.src = "https://telegram.org/js/telegram-widget.js?23";
    script.async = true;
    script.setAttribute("data-telegram-login", "Midnight_poker_bot");
    script.setAttribute("data-size", "large");
    script.setAttribute("data-onauth", "onTelegramAuth(user)");
    script.setAttribute("data-request-access", "write");
    
    // Добавляем обработчик ошибок
    script.onerror = () => {
      console.error("Failed to load Telegram widget");
      container.innerHTML = '<div style="color: red;">Ошибка загрузки виджета Telegram. Пожалуйста, обновите страницу.</div>';
    };

    container.appendChild(script);
    isScriptLoaded.current = true;

    // Очистка при размонтировании
    return () => {
      if (container) {
        container.innerHTML = "";
      }
      // Не удаляем onTelegramAuth, так как он может понадобиться при повторном монтировании
    };
  }, []);

  return (
    <div style={{ 
      textAlign: "center", 
      marginTop: 100, 
      color: "white",
      minHeight: "100vh",
      background: "black",
      padding: "20px"
    }}>
      <h2>Вход через Telegram</h2>

      {/* Контейнер для кнопки */}
      <div 
        id="tg-login" 
        style={{
          display: "flex",
          justifyContent: "center",
          margin: "20px 0",
          minHeight: "100px"
        }} 
      />

      <br />

      <div style={{ marginTop: "20px" }}>
        <a 
          href="https://t.me/Midnight_poker_bot/app" 
          target="_blank"
          rel="noopener noreferrer"
          style={{ 
            color: "#24A1DE",
            textDecoration: "none",
            fontSize: "16px"
          }}
        >
          📱 Открыть приложение в Telegram
        </a>
      </div>

      {/* Кнопка для ручного обновления, если виджет не загрузился */}
      <div style={{ marginTop: "30px" }}>
        <button
          onClick={() => window.location.reload()}
          style={{
            padding: "10px 20px",
            background: "#24A1DE",
            color: "white",
            border: "none",
            borderRadius: "5px",
            cursor: "pointer"
          }}
        >
          🔄 Обновить страницу
        </button>
      </div>
    </div>
  );
};

export default WebAuth;