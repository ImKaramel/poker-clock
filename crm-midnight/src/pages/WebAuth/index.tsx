import React, { useEffect, useRef } from "react";
import background from '../../assets/2026-04-20 22.33.10.jpg'

const WebAuth: React.FC = () => {
  const isScriptLoaded = useRef(false);

  useEffect(() => {
    console.log("WebAuth component mounted");
    
    // callback обязательно на window
    (window as any).onTelegramAuth = async (user: any) => {
      console.log("Telegram auth callback triggered", user);
      try {
        const res = await fetch("/api/auth/telegram/callback", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(user),
        });

        const data = await res.json();
        console.log("Auth response:", data);

        localStorage.setItem("auth_token", data.token);
        window.location.href = "/";
      } catch (error) {
        console.error("Auth error:", error);
        alert("Auth error");
      }
    };

    const container = document.getElementById("tg-login");
    console.log("Container found:", container);
    
    if (!container) {
      console.error("Container with id 'tg-login' not found");
      return;
    }

    // ❗ очистка обязательна
    container.innerHTML = "";
    
    // Проверяем существующие скрипты
    const existingScripts = document.querySelectorAll('script[src*="telegram-widget.js"]');
    console.log("Existing telegram scripts:", existingScripts.length);
    
    existingScripts.forEach(script => script.remove());

    // создаём script
    const script = document.createElement("script");
    script.src = "https://telegram.org/js/telegram-widget.js?22";
    script.async = true;
    script.setAttribute("data-telegram-login", "Midnight_poker_bot");
    script.setAttribute("data-size", "large");
    script.setAttribute("data-onauth", "onTelegramAuth(user)");
    script.setAttribute("data-request-access", "write");
    
    // Добавляем обработчики событий скрипта
    script.onload = () => {
      console.log("Telegram widget script loaded successfully");
    };
    
    script.onerror = (error) => {
      console.error("Failed to load Telegram widget:", error);
      container.innerHTML = '<div style="color: red; padding: 20px;">❌ Ошибка загрузки виджета Telegram. Проверьте интернет-соединение.</div>';
    };

    console.log("Appending script to container");
    container.appendChild(script);
    isScriptLoaded.current = true;

    return () => {
      console.log("WebAuth component unmounting");
      // Не очищаем container при размонтировании, чтобы виджет остался
    };
  }, []);

  return (
    <div style={{ 
      textAlign: "center", 
      marginTop: 50, 
      color: "white",
      minHeight: "100vh",
      background: "black",
      padding: "20px",
      position: "relative"
    }}>
      <h2>Вход через Telegram</h2>
      
      <div style={{ 
        margin: "30px 0",
        padding: "20px",
        border: "1px solid #333",
        borderRadius: "10px",
        position: "relative",
        overflow: "hidden",
        minHeight: "300px"
      }}>
        {/* Фоновое изображение повернутое на 90 градусов и растянутое */}
        <div style={{
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
          zIndex: 0
        }} />
        
        {/* Затемнение для лучшей читаемости текста */}
        <div style={{
          position: "absolute",
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          backgroundColor: "rgba(0, 0, 0, 0.5)",
          zIndex: 1
        }} />
        
        {/* Контент поверх фона */}
        <div style={{ position: "relative", zIndex: 2 }}>
          <div id="tg-login" style={{ 
            display: "flex", 
            justifyContent: "center",
            minHeight: "100px",
            alignItems: "center"
          }}>
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
      
      <div style={{ marginTop: 20, fontSize: 12, opacity: 0.7 }}>
        Если кнопка не появилась через 10 секунд, проверьте консоль (F12)
      </div>
    </div>
  );
};

export default WebAuth;