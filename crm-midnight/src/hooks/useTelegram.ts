import { useEffect, useState } from "react";

export const useTelegram = () => {
  const [webApp, setWebApp] = useState<any>(null);
  const [user, setUser] = useState<any>(null);           // ← меняем название и тип
  const [isReady, setIsReady] = useState(false);

  useEffect(() => {
    const tg = (window as any).Telegram?.WebApp;
    if (!tg) return;

    console.log("Telegram WebApp initDataUnsafe.user:", tg.initDataUnsafe?.user);

    tg.ready();
    tg.expand();

    setWebApp(tg);

    // Пуллинг на случай задержки Telegram
    const interval = setInterval(() => {
      const tgUser = tg.initDataUnsafe?.user;

      if (tgUser && tgUser.id) {                    // проверяем по наличию id
        console.log("✅ Telegram user received:", tgUser);

        setUser(tgUser);                            // сохраняем объект пользователя
        setIsReady(true);
        clearInterval(interval);
      }
    }, 50);

    const timeout = setTimeout(() => {
      console.warn("⚠️ Timeout waiting for Telegram user");
      setIsReady(true);
      clearInterval(interval);
    }, 3000);

    return () => {
      clearInterval(interval);
      clearTimeout(timeout);
    };
  }, []);

  return {
    webApp,
    user,                    // ← теперь возвращаем user, а не initDataUnsafe
    isReady,
    isTelegram: !!webApp && !!user,
  };
}