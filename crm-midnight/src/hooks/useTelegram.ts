import { useEffect, useState } from "react";

export const useTelegram = () => {
  const [webApp, setWebApp] = useState<any>(null);
  const [user, setUser] = useState<any>(undefined);
  const [isTelegram, setIsTelegram] = useState<boolean>(false);
  const [isReady, setIsReady] = useState(false);

  useEffect(() => {
    const tg = (window as any).Telegram?.WebApp;

    // 🌐 BROWSER MODE
    if (!tg || !tg.initData) {
      setIsTelegram(false);
      setIsReady(true);
      return;
    }

    // 📱 TELEGRAM MODE
    setIsTelegram(true);
    setWebApp(tg);
    try {
      tg.ready();
      tg.expand();
    } catch (error) {
      console.warn("Telegram WebApp init error:", error);
    }

    const tgUser = tg.initDataUnsafe?.user;
    if (tgUser?.id) {
      setUser(tgUser);
      setIsReady(true);
      return;
    }

    // Fallback for slow WebView environments where initDataUnsafe may lag.
    const timeout = setTimeout(() => {
      const delayedUser = tg.initDataUnsafe?.user;
      if (delayedUser?.id) {
        setUser(delayedUser);
      } else {
        setUser(null);
      }
      setIsReady(true);
    }, 1200);

    return () => {
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
