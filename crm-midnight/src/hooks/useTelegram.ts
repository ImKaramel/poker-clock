import { useEffect, useState } from "react";

export const useTelegram = () => {
  const [webApp, setWebApp] = useState<any>(null);
  const [initDataUnsafe, setInitData] = useState<string | null>(null);
  const [isReady, setIsReady] = useState(false);

  useEffect(() => {
    const tg = (window as any).Telegram?.WebApp;

    if (!tg) return;

    if (!tg.initDataUnsafe.user || tg.initDataUnsafe.length === 0) {
      console.log("⏳ Telegram WebApp not ready yet...");
    } else {
      tg.ready();
    }

    setWebApp(tg);

    // Telegram иногда задерживает initDataUnsafe → запускаем пуллинг
    const interval = setInterval(() => {
      if (tg.initDataUnsafe.user && tg.initDataUnsafe.user.length > 0) {
        console.log("✅ initDataUnsafe received:", tg.initDataUnsafe.user);

        setInitData(tg.initDataUnsafe.user);
        setIsReady(true);

        // разворачиваем webview
        tg.expand();

        clearInterval(interval);
      }
    }, 50);

    // если Telegram так и НЕ передал initDataUnsafe → считаем не MiniApp
    const timeout = setTimeout(() => {
      console.warn("⚠️ initDataUnsafe timeout. Probably not a Mini App.");
      setIsReady(true);
      clearInterval(interval);
    }, 3000);

    return () => {
      clearInterval(interval);
      clearTimeout(timeout);
    };
  }, []); // Оставляем пустой массив

  return {
    webApp,
    initDataUnsafe,
    isReady,
    isTelegram: !!webApp && !!initDataUnsafe,
  };
};