import { useEffect, useState } from "react";

export const useTelegram = () => {
  const [webApp, setWebApp] = useState<any>(null);
  const [initData, setInitData] = useState<string | null>(null);
  const [isReady, setIsReady] = useState(false);

  useEffect(() => {
    const tg = (window as any).Telegram?.WebApp;

    if (!tg) return;

    if (!tg.initData || tg.initData.length === 0) {
      console.log("⏳ Telegram WebApp not ready yet...");
    } else {
      tg.ready();
    }

    setWebApp(tg);

    // Telegram иногда задерживает initData → запускаем пуллинг
    const interval = setInterval(() => {
      if (tg.initData && tg.initData.length > 0) {
        console.log("✅ initData received:", tg.initData);

        setInitData(tg.initData);
        setIsReady(true);

        // разворачиваем webview
        tg.expand();

        clearInterval(interval);
      }
    }, 50);

    // если Telegram так и НЕ передал initData → считаем не MiniApp
    const timeout = setTimeout(() => {
      console.warn("⚠️ initData timeout. Probably not a Mini App.");
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
    initData,
    isReady,
    isTelegram: !!webApp && !!initData,
  };
};