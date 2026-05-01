import React, { FormEvent, useEffect, useMemo, useState } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import background from "../../assets/background.jpg";
import { authAPI } from "../../utils/api";

type Mode = "login" | "register";

const usernamePattern = /^[a-z0-9_]{5,32}$/;

const normalizeUsername = (value: string) =>
  value.trim().toLowerCase().replace(/^@+/, "");

const WebAuth: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const locationState = location.state as { authError?: string } | null;
  const [hasToken, setHasToken] = useState(false);
  const [mode, setMode] = useState<Mode>("login");
  const [telegramUsername, setTelegramUsername] = useState("");
  const [nickname, setNickname] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState(locationState?.authError || "");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showFallback, setShowFallback] = useState(true);

  const username = useMemo(() => normalizeUsername(telegramUsername), [telegramUsername]);

  useEffect(() => {
    const params = new URLSearchParams(location.search);
    let token = params.get("token");

    if (!token && window.location.href.includes("token=")) {
      const match = window.location.href.match(/token=([^&]+)/);
      if (match) token = match[1];
    }

    const authError = params.get("error");

    if (token) {
      localStorage.setItem("auth_token", token);
      setHasToken(true);
      navigate("/", { replace: true });
      return;
    }

    if (authError) {
      setError("Telegram не смог авторизовать вход. Войдите резервным способом.");
      setShowFallback(true);
    }
  }, [location, navigate]);

  useEffect(() => {
    if (hasToken) return;

    const container = document.getElementById("tg-login");
    if (!container) return;

    container.innerHTML = "";

    const existingScripts = document.querySelectorAll('script[src*="telegram-widget.js"]');
    existingScripts.forEach(script => script.remove());

    const script = document.createElement("script");
    script.src = "https://telegram.org/js/telegram-widget.js?22";
    script.async = true;
    script.setAttribute("data-telegram-login", "Midnight_poker_bot");
    script.setAttribute("data-size", "large");
    script.setAttribute("data-auth-url", "https://api.midnight-club-app.ru/api/auth/telegram/callback");
    script.setAttribute("data-request-access", "write");
    script.setAttribute("data-radius", "10");

    script.onerror = () => {
      setError("Не удалось загрузить вход через Telegram. Войдите резервным способом.");
      setShowFallback(true);
    };

    container.appendChild(script);
  }, [hasToken]);

  const validateForm = () => {
    if (!usernamePattern.test(username)) {
      return "Username: 5-32 символа, только a-z, 0-9 и _.";
    }
    if (password.length < 8 || !/[a-zа-яё]/i.test(password) || !/\d/.test(password)) {
      return "Пароль: минимум 8 символов, хотя бы 1 буква и 1 цифра.";
    }
    if (mode === "register") {
      const nicknameLength = Array.from(nickname.trim()).length;
      if (nicknameLength < 2 || nicknameLength > 24) {
        return "Nickname: 2-24 символа.";
      }
      if (password !== confirmPassword) {
        return "Пароли не совпадают.";
      }
    }
    return "";
  };

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    setError("");

    const validationError = validateForm();
    if (validationError) {
      setError(validationError);
      return;
    }

    setIsSubmitting(true);
    try {
      const response =
        mode === "login"
          ? await authAPI.login({ telegram_username: username, password })
          : await authAPI.register({
              telegram_username: username,
              nickname: nickname.trim(),
              password,
              confirm_password: confirmPassword,
            });

      localStorage.setItem("auth_token", response.data.token);
      navigate("/", { replace: true });
    } catch (err: any) {
      setError(err?.response?.data?.error || "Неверные данные.");
    } finally {
      setIsSubmitting(false);
    }
  };

  const switchMode = (nextMode: Mode) => {
    setMode(nextMode);
    setError("");
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
          backgroundColor: "rgba(7, 9, 12, 0.76)",
          border: "1px solid rgba(255,255,255,0.12)",
          backdropFilter: "blur(6px)",
        }}>
          <div id="tg-login" style={{
            display: "flex",
            justifyContent: "center",
            minHeight: 64,
            alignItems: "center",
          }}>
            <div>Загрузка кнопки входа...</div>
          </div>

        </section>

        {showFallback && (
          <section style={{
            padding: 18,
            borderRadius: 8,
            backgroundColor: "rgba(7, 9, 12, 0.82)",
            border: "1px solid rgba(255,255,255,0.14)",
            backdropFilter: "blur(6px)",
          }}>
            <div style={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr",
              gap: 8,
              marginBottom: 16,
            }}>
              <button type="button" onClick={() => switchMode("login")} style={tabStyle(mode === "login")}>
                Вход
              </button>
              <button type="button" onClick={() => switchMode("register")} style={tabStyle(mode === "register")}>
                Регистрация
              </button>
            </div>

            <form onSubmit={submit} style={{ display: "flex", flexDirection: "column", gap: 12 }}>
              <label style={labelStyle}>
                Telegram username
                <input
                  value={telegramUsername}
                  onChange={(e) => setTelegramUsername(e.target.value)}
                  placeholder="username"
                  autoCapitalize="none"
                  autoComplete="username"
                  style={inputStyle}
                />
              </label>

              {mode === "register" && (
                <label style={labelStyle}>
                  Nickname
                  <input
                    value={nickname}
                    onChange={(e) => setNickname(e.target.value)}
                    placeholder="Ваш ник"
                    autoComplete="nickname"
                    style={inputStyle}
                  />
                </label>
              )}

              <label style={labelStyle}>
                Пароль
                <input
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  type="password"
                  autoComplete={mode === "login" ? "current-password" : "new-password"}
                  style={inputStyle}
                />
              </label>

              {mode === "register" && (
                <label style={labelStyle}>
                  Повторите пароль
                  <input
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                    type="password"
                    autoComplete="new-password"
                    style={inputStyle}
                  />
                </label>
              )}

              {error && (
                <div style={{
                  color: "#ffb4b4",
                  background: "rgba(128, 0, 0, 0.22)",
                  border: "1px solid rgba(255, 130, 130, 0.24)",
                  borderRadius: 8,
                  padding: "10px 12px",
                  fontSize: 14,
                }}>
                  {error}
                </div>
              )}

              <button type="submit" disabled={isSubmitting} style={primaryButtonStyle}>
                {isSubmitting ? "Проверка..." : mode === "login" ? "Войти" : "Зарегистрироваться"}
              </button>
            </form>
          </section>
        )}

        <a
          href="https://t.me/Midnight_poker_bot"
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

const labelStyle: React.CSSProperties = {
  display: "flex",
  flexDirection: "column",
  gap: 7,
  fontSize: 14,
  color: "rgba(255,255,255,0.82)",
  textAlign: "left",
};

const inputStyle: React.CSSProperties = {
  minHeight: 44,
  borderRadius: 8,
  border: "1px solid rgba(255,255,255,0.16)",
  background: "rgba(255,255,255,0.08)",
  color: "white",
  padding: "0 12px",
  fontSize: 16,
  outline: "none",
  boxSizing: "border-box",
};

const primaryButtonStyle: React.CSSProperties = {
  minHeight: 46,
  borderRadius: 8,
  border: 0,
  background: "#24a1de",
  color: "white",
  fontWeight: 700,
  fontSize: 16,
  cursor: "pointer",
};

const tabStyle = (active: boolean): React.CSSProperties => ({
  minHeight: 40,
  borderRadius: 8,
  border: active ? "1px solid #24a1de" : "1px solid rgba(255,255,255,0.14)",
  background: active ? "rgba(36,161,222,0.22)" : "rgba(255,255,255,0.06)",
  color: "white",
  fontWeight: 700,
  cursor: "pointer",
});

export default WebAuth;
