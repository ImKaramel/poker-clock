import { useEffect, useState } from "react";

const TELEGRAM_SCRIPT_ID = "telegram-web-app-sdk";
const TELEGRAM_SCRIPT_SRC = "https://telegram.org/js/telegram-web-app.js";
const TELEGRAM_HOSTS = new Set([
  "midnight-club.ru",
  "www.midnight-club.ru",
  "localhost",
  "127.0.0.1",
]);

const shouldLoadTelegramSdk = () => {
  if (typeof window === "undefined") {
    return false;
  }

  const launchParams = `${window.location.search}${window.location.hash}`;
  return (
    TELEGRAM_HOSTS.has(window.location.hostname.toLowerCase()) ||
    launchParams.includes("tgWebAppData")
  );
};

const loadTelegramSdk = async (): Promise<void> => {
  if (typeof window === "undefined") {
    return;
  }

  if ((window as any).Telegram?.WebApp) {
    return;
  }

  const existingScript = document.getElementById(TELEGRAM_SCRIPT_ID) as HTMLScriptElement | null;
  if (existingScript) {
    await new Promise<void>((resolve) => {
      if (existingScript.dataset.loaded === "true") {
        resolve();
        return;
      }

      existingScript.addEventListener("load", () => resolve(), { once: true });
      existingScript.addEventListener("error", () => resolve(), { once: true });
    });
    return;
  }

  await new Promise<void>((resolve) => {
    const script = document.createElement("script");
    script.id = TELEGRAM_SCRIPT_ID;
    script.src = TELEGRAM_SCRIPT_SRC;
    script.async = true;
    script.onload = () => {
      script.dataset.loaded = "true";
      resolve();
    };
    script.onerror = () => resolve();
    document.head.appendChild(script);
  });
};

export const useTelegram = () => {
  const [webApp, setWebApp] = useState<any>(null);
  const [user, setUser] = useState<any>(undefined);
  const [initData, setInitData] = useState<string>("");
  const [isTelegram, setIsTelegram] = useState<boolean>(false);
  const [isReady, setIsReady] = useState(false);

  useEffect(() => {
    let isMounted = true;
    let timeout: ReturnType<typeof setTimeout> | undefined;

    const init = async () => {
      if (shouldLoadTelegramSdk()) {
        await loadTelegramSdk();
      }

      if (!isMounted) {
        return;
      }

      const tg = (window as any).Telegram?.WebApp;

      // Browser mode: no Telegram SDK or no initData from Telegram shell.
      if (!tg || !tg.initData) {
        setIsTelegram(false);
        setIsReady(true);
        return;
      }

      setIsTelegram(true);
      setWebApp(tg);
      try {
        tg.ready();
        tg.expand();
        tg.disableVerticalSwipes?.();
      } catch (error) {
        console.warn("Telegram WebApp init error:", error);
      }

      const waitForInitData = (attempt = 0) => {
        if (!isMounted) {
          return;
        }

        const rawInitData = tg.initData || "";
        const tgUser = tg.initDataUnsafe?.user;

        if (rawInitData) {
          setInitData(rawInitData);
          setUser(tgUser || null);
          setIsReady(true);
          return;
        }

        if (attempt < 30) {
          timeout = setTimeout(() => waitForInitData(attempt + 1), 200);
          return;
        }

        setUser(tgUser || null);
        setIsReady(true);
      };

      waitForInitData();
    };

    init();

    return () => {
      isMounted = false;
      if (timeout) {
        clearTimeout(timeout);
      }
    };
  }, []);

  return {
    webApp,
    user,
    initData,
    isTelegram,
    isReady,
  };
};
