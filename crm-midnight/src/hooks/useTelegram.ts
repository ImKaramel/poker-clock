import { useEffect, useState } from "react";

export const useTelegram = () => {
  const [webApp, setWebApp] = useState<any>(null);
  const [user, setUser] = useState<any>(null);
  const [isReady, setIsReady] = useState(false);

  useEffect(() => {
    const tg = (window as any).Telegram?.WebApp;

    if (!tg) {
      // 👉 браузер
      setIsReady(true);
      return;
    }

    tg.ready();
    tg.expand();

    setWebApp(tg);

    const interval = setInterval(() => {
      const tgUser = tg.initDataUnsafe?.user;

      if (tgUser?.id) {
        setUser(tgUser);
        setIsReady(true);
        clearInterval(interval);
      }
    }, 50);

    const timeout = setTimeout(() => {
      console.warn("⚠️ Telegram user timeout");
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
    user,
    isReady,
    isTelegram: !!webApp, // ✅ фикс
  };
};