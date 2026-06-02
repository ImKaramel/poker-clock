import React, { useEffect, useRef, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import background from "../../assets/background.jpg";
import { authAPI } from "../../utils/api";

const AUTH_TOKEN_CHANGED_EVENT = "auth-token-changed";
const CONTACT_AUTH_TOKEN_KEY = "contact_auth_token";
const CONTACT_AUTH_BOT_LINK_KEY = "contact_auth_bot_link";
const CONTACT_AUTH_TG_LINK_KEY = "contact_auth_tg_link";

type ContactAuthState = "idle" | "starting" | "pending" | "completed";

const WebAuth: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const locationState = location.state as { authError?: string } | null;
  const [challengeToken, setChallengeToken] = useState(
    () => localStorage.getItem(CONTACT_AUTH_TOKEN_KEY) || "",
  );
  const [botLink, setBotLink] = useState(() => localStorage.getItem(CONTACT_AUTH_BOT_LINK_KEY) || "");
  const [tgLink, setTgLink] = useState(() => localStorage.getItem(CONTACT_AUTH_TG_LINK_KEY) || "");
  const [status, setStatus] = useState<ContactAuthState>(challengeToken ? "pending" : "idle");
  const [error, setError] = useState(locationState?.authError || "");
  const isPollingRef = useRef(false);

  const loginCommand = challengeToken ? `/login ${challengeToken}` : "";

  useEffect(() => {
    const params = new URLSearchParams(location.search);
    let token = params.get("token");

    if (!token && window.location.href.includes("token=")) {
      const match = window.location.href.match(/token=([^&]+)/);
      if (match) token = match[1];
    }

    if (token) {
      localStorage.setItem("auth_token", token);
      localStorage.removeItem(CONTACT_AUTH_TOKEN_KEY);
      localStorage.removeItem(CONTACT_AUTH_BOT_LINK_KEY);
      localStorage.removeItem(CONTACT_AUTH_TG_LINK_KEY);
      window.dispatchEvent(new Event(AUTH_TOKEN_CHANGED_EVENT));
      navigate("/", { replace: true });
      return;
    }

    if (params.get("error")) {
      setError("Telegram не смог авторизовать вход. Подтвердите номер через бота.");
    }
  }, [location, navigate]);

  useEffect(() => {
    if (!challengeToken || status !== "pending") {
      return;
    }

    let isMounted = true;
    const poll = async () => {
      if (isPollingRef.current) return;
      isPollingRef.current = true;

      try {
        const response = await authAPI.pollContactAuth(challengeToken);
        if (!isMounted) return;

        if (response.data?.status === "completed" && response.data?.token) {
          localStorage.setItem("auth_token", response.data.token);
          localStorage.removeItem(CONTACT_AUTH_TOKEN_KEY);
          localStorage.removeItem(CONTACT_AUTH_BOT_LINK_KEY);
          localStorage.removeItem(CONTACT_AUTH_TG_LINK_KEY);
          window.dispatchEvent(new Event(AUTH_TOKEN_CHANGED_EVENT));
          setStatus("completed");
          navigate("/", { replace: true });
        }
      } catch (err: any) {
        if (!isMounted) return;
        const statusCode = err?.response?.status;
        if (statusCode === 410 || statusCode === 404 || statusCode === 409) {
          localStorage.removeItem(CONTACT_AUTH_TOKEN_KEY);
          localStorage.removeItem(CONTACT_AUTH_BOT_LINK_KEY);
          localStorage.removeItem(CONTACT_AUTH_TG_LINK_KEY);
          setChallengeToken("");
          setBotLink("");
          setTgLink("");
          setStatus("idle");
          setError("Ссылка подтверждения устарела. Запустите вход заново.");
        }
      } finally {
        isPollingRef.current = false;
      }
    };

    poll();
    const interval = window.setInterval(poll, 2500);
    return () => {
      isMounted = false;
      window.clearInterval(interval);
    };
  }, [challengeToken, navigate, status]);

  const startContactAuth = async () => {
    try {
      setStatus("starting");
      setError("");
      const response = await authAPI.startContactAuth();
      const nextToken = response.data.token;
      const nextBotLink = response.data.bot_link;
      const nextTgLink = response.data.tg_link;

      localStorage.setItem(CONTACT_AUTH_TOKEN_KEY, nextToken);
      localStorage.setItem(CONTACT_AUTH_BOT_LINK_KEY, nextBotLink);
      if (nextTgLink) {
        localStorage.setItem(CONTACT_AUTH_TG_LINK_KEY, nextTgLink);
      }
      setChallengeToken(nextToken);
      setBotLink(nextBotLink);
      setTgLink(nextTgLink || "");
      setStatus("pending");
      openBot(nextBotLink, nextTgLink);
    } catch (err: any) {
      setStatus("idle");
      setError(err?.response?.data?.error || err?.message || "Не удалось начать вход через Telegram.");
    }
  };

  const openBot = (httpsLink = botLink, telegramLink = tgLink) => {
    if (telegramLink) {
      window.location.href = telegramLink;
      return;
    }
    if (httpsLink) {
      window.open(httpsLink, "_blank", "noopener,noreferrer");
    }
  };

  const copyLoginCommand = async () => {
    if (!loginCommand) return;
    try {
      await navigator.clipboard.writeText(loginCommand);
    } catch {
      window.prompt("Скопируйте команду и отправьте её боту", loginCommand);
    }
  };

  return (
    <div style={{
      minHeight: "100vh",
      color: "white",
      background: `linear-gradient(rgba(0, 0, 0, 0.68), rgba(0, 0, 0, 0.72)), url(${background})`,
      backgroundSize: "cover",
      backgroundPosition: "center",
      padding: "28px 18px 48px",
      boxSizing: "border-box",
    }}>
      <main style={{
        width: "100%",
        maxWidth: 420,
        margin: "0 auto",
        display: "flex",
        flexDirection: "column",
        gap: 18,
      }}>
        <h1 style={{ fontSize: 28, lineHeight: 1.15, margin: "12px 0 4px", textAlign: "center" }}>
          Авторизация
        </h1>

        <section style={{
          padding: 18,
          borderRadius: 8,
          backgroundColor: "rgba(7, 9, 12, 0.82)",
          border: "1px solid rgba(255,255,255,0.14)",
          backdropFilter: "blur(6px)",
        }}>
          <div style={{ fontSize: 18, fontWeight: 700, marginBottom: 8 }}>
            Вход через Telegram
          </div>
          <div style={{ fontSize: 14, color: "rgba(255,255,255,0.72)", marginBottom: 16 }}>
            Откройте бота и поделитесь номером. После подтверждения эта страница войдёт автоматически.
          </div>

          {error && <ErrorBox>{error}</ErrorBox>}

          <button
            type="button"
            onClick={startContactAuth}
            disabled={status === "starting"}
            style={{ ...primaryButtonStyle, marginTop: error ? 12 : 0 }}
          >
            {status === "starting" ? "Готовим вход..." : status === "pending" ? "Открыть бота ещё раз" : "Войти через бота"}
          </button>

          {status === "pending" && (
            <div style={pendingBoxStyle}>
              <div>Ожидаю подтверждение номера в Telegram.</div>
              {(botLink || tgLink) && (
                <button
                  type="button"
                  onClick={() => openBot()}
                  style={secondaryButtonStyle}
                >
                  Открыть бота
                </button>
              )}
              {loginCommand && (
                <div style={commandBoxStyle}>
                  <div style={{ color: "rgba(255,255,255,0.72)" }}>
                    Если Telegram открыл только /start, отправьте боту команду:
                  </div>
                  <code style={commandStyle}>{loginCommand}</code>
                  <button type="button" onClick={copyLoginCommand} style={copyButtonStyle}>
                    Скопировать команду
                  </button>
                </div>
              )}
            </div>
          )}
        </section>

        <a
          href={botLink || "https://t.me/Midnight_poker_bot"}
          target="_blank"
          rel="noopener noreferrer"
          style={{ color: "#54bde8", textAlign: "center", textDecoration: "none", fontWeight: 600 }}
        >
          Открыть бота в Telegram
        </a>
      </main>
    </div>
  );
};

const ErrorBox = ({ children }: { children: React.ReactNode }) => (
  <div style={{
    padding: "10px 12px",
    borderRadius: 8,
    border: "1px solid rgba(255,98,98,0.4)",
    background: "rgba(122, 20, 20, 0.32)",
    color: "#ffc8c8",
    fontSize: 14,
  }}>
    {children}
  </div>
);

const primaryButtonStyle: React.CSSProperties = {
  minHeight: 46,
  width: "100%",
  borderRadius: 8,
  border: 0,
  background: "#24a1de",
  color: "white",
  fontWeight: 800,
  fontSize: 16,
  cursor: "pointer",
};

const pendingBoxStyle: React.CSSProperties = {
  marginTop: 12,
  padding: "12px",
  borderRadius: 8,
  border: "1px solid rgba(84,189,232,0.28)",
  background: "rgba(36,161,222,0.12)",
  color: "rgba(255,255,255,0.86)",
  fontSize: 14,
  display: "flex",
  flexDirection: "column",
  gap: 8,
};

const secondaryButtonStyle: React.CSSProperties = {
  minHeight: 38,
  borderRadius: 8,
  border: "1px solid rgba(84,189,232,0.6)",
  background: "rgba(84,189,232,0.14)",
  color: "#bfefff",
  fontWeight: 800,
  fontSize: 14,
  cursor: "pointer",
};

const commandBoxStyle: React.CSSProperties = {
  marginTop: 4,
  display: "flex",
  flexDirection: "column",
  gap: 8,
};

const commandStyle: React.CSSProperties = {
  display: "block",
  padding: "10px",
  borderRadius: 8,
  background: "rgba(0,0,0,0.28)",
  border: "1px solid rgba(255,255,255,0.14)",
  color: "white",
  overflowWrap: "anywhere",
  fontSize: 13,
};

const copyButtonStyle: React.CSSProperties = {
  minHeight: 36,
  borderRadius: 8,
  border: "1px solid rgba(255,255,255,0.2)",
  background: "rgba(255,255,255,0.08)",
  color: "white",
  fontWeight: 700,
  fontSize: 14,
  cursor: "pointer",
};

export default WebAuth;
