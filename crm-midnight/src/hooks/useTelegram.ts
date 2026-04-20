import { useEffect, useState } from "react";

export const useTelegram = () => {
  const [webApp, setWebApp] = useState<any>(null);
  const [user, setUser] = useState<any>(undefined);
  const [isTelegram, setIsTelegram] = useState<boolean>(false);
  const [isReady, setIsReady] = useState(false);

  useEffect(() => {
    const tg = (window as any).Telegram?.WebApp;

    // 🌐 BROWSER MODE
    if (!tg) {
      setIsTelegram(false);
      setIsReady(true);
      return;
    }

    // 📱 TELEGRAM MODE
    setIsTelegram(true);

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
      setUser(null);
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
    isTelegram,
    isReady,
  };
};