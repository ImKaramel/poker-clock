import React, { useEffect, useRef, useState } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import background from "../../assets/background.jpg";
import { authAPI } from "../../utils/api";

const AUTH_TOKEN_CHANGED_EVENT = "auth-token-changed";
const CONTACT_AUTH_TOKEN_KEY = "contact_auth_token";

type ContactAuthState = "idle" | "starting" | "pending" | "completed";

const WebAuth: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const locationState = location.state as { authError?: string } | null;
  const [challengeToken, setChallengeToken] = useState(() => localStorage.getItem(CONTACT_AUTH_TOKEN_KEY) || "");
  const [botLink, setBotLink] = useState("");
  const [status, setStatus] = useState<ContactAuthState>(challengeToken ? "pending" : "idle");
  const [error, setError] = useState(locationState?.authError || "");
  const isPollingRef = useRef(false);

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
          window.dispatchEvent(new Event(AUTH_TOKEN_CHANGED_EVENT));
          setStatus("completed");
          navigate("/", { replace: true });
          return;
        }
      } catch (err: any) {
        if (!isMounted) return;
        const statusCode = err?.response?.status;
        if (statusCode === 410 || statusCode === 404 || statusCode === 409) {
          localStorage.removeItem(CONTACT_AUTH_TOKEN_KEY);
          setChallengeToken("");
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

      localStorage.setItem(CONTACT_AUTH_TOKEN_KEY, nextToken);
      setChallengeToken(nextToken);
      setBotLink(nextBotLink);
      setStatus("pending");
      window.open(nextBotLink, "_blank", "noopener,noreferrer");
    } catch (err: any) {
      setStatus("idle");
      setError(err?.response?.data?.error || err?.message || "Не удалось начать вход через Telegram.");
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
            Вход через Telegram contact
          </div>
          <div style={{ fontSize: 14, color: "rgba(255,255,255,0.72)", marginBottom: 16 }}>
            Нажмите кнопку, откройте бота и поделитесь номером. После подтверждения эта страница войдёт автоматически.
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
              {botLink && (
                <a href={botLink} target="_blank" rel="noopener noreferrer" style={linkStyle}>
                  Открыть бота
                </a>
              )}
            </div>
          )}
        </section>
      </main>
    </div>
  );
};

const ErrorBox = ({ children }: { children: React.ReactNode }) => (
  <div style={{
    color: "#ffb4b4",
    background: "rgba(128, 0, 0, 0.22)",
    border: "1px solid rgba(255, 130, 130, 0.24)",
    borderRadius: 8,
    padding: "10px 12px",
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
  fontWeight: 700,
  fontSize: 16,
  cursor: "pointer",
};

const pendingBoxStyle: React.CSSProperties = {
  marginTop: 12,
  padding: "12px",
  borderRadius: 8,
  border: "1px solid rgba(36,161,222,0.35)",
  background: "rgba(36,161,222,0.12)",
  color: "rgba(255,255,255,0.86)",
  fontSize: 14,
  display: "flex",
  flexDirection: "column",
  gap: 8,
};

const linkStyle: React.CSSProperties = {
  color: "#54bde8",
  textDecoration: "none",
  fontWeight: 700,
};

export default WebAuth;
