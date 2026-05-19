package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	infraauth "github.com/pridecrm/app-backend/internal/infrastructure/auth"
	"github.com/pridecrm/app-backend/internal/usecase"
)

const authGenericError = "неверные данные"

var passwordAuthLimiter = newAuthLimiter(5, 10*time.Minute)

const telegramVerificationMessage = "Код подтверждения для установки пароля в Midnight Club: %s\n\nЕсли это были не вы, просто проигнорируйте сообщение."

type authLimiter struct {
	mu       sync.Mutex
	limit    int
	lockTime time.Duration
	attempts map[string]authAttempt
}

type authAttempt struct {
	Count        int
	BlockedUntil time.Time
}

func newAuthLimiter(limit int, lockTime time.Duration) *authLimiter {
	return &authLimiter{
		limit:    limit,
		lockTime: lockTime,
		attempts: make(map[string]authAttempt),
	}
}

func (l *authLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	a := l.attempts[key]
	if time.Now().Before(a.BlockedUntil) {
		return false
	}
	return true
}

func (l *authLimiter) success(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, key)
}

func (l *authLimiter) fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	a := l.attempts[key]
	a.Count++
	if a.Count >= l.limit {
		a.Count = 0
		a.BlockedUntil = time.Now().Add(l.lockTime)
	}
	l.attempts[key] = a
}

func authLimitKey(c *gin.Context, username string) string {
	ip := c.ClientIP()
	return ip + ":" + strings.ToLower(strings.TrimSpace(strings.TrimPrefix(username, "@")))
}

func buildFrontendWebAuthURL(frontendURL string, values url.Values) string {
	base := strings.TrimRight(frontendURL, "/")
	if base == "" {
		base = "https://www.midnight-club-app.ru"
	}

	if len(values) == 0 {
		return base + "/web-auth"
	}
	return base + "/web-auth?" + values.Encode()
}

func (h *Handlers) sendTelegramMessage(ctx context.Context, chatID string, text string) error {
	if h.TelegramBotToken == "" {
		return errors.New("telegram bot token is empty")
	}

	payload := map[string]any{
		"chat_id": chatID,
		"text":    text,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.telegram.org/bot"+h.TelegramBotToken+"/sendMessage",
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("telegram send failed: %s", strings.TrimSpace(string(respBody)))
	}

	var apiResp struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err == nil && !apiResp.OK {
		if apiResp.Description == "" {
			apiResp.Description = "unknown telegram error"
		}
		return errors.New(apiResp.Description)
	}

	return nil
}

func (h *Handlers) TelegramAuth(c *gin.Context) {
	var body struct {
		User map[string]any `json:"user"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		h.Log.Error("INVALID JSON",
			"err", err,
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid json",
			"details": err.Error(),
		})
		return
	}

	if body.User == nil {
		h.Log.Error("USER IS NULL")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user is null",
		})
		return
	}

	h.Log.Info("TELEGRAM AUTH REQUEST OK",
		"user", body.User,
	)

	token, u, isNew, err := h.UC.TelegramAuthUnsafe(c.Request.Context(), body.User)
	if err != nil {
		h.Log.Error("❌ AUTH FAILED", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.Log.Info("isNew TELEGRAM AUTH REQUEST OK",
		"user", body.User,
		"isNew", isNew,
		"token", token,
	)

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  userToMap(u),
		"isNew": isNew,
	})
}

func (h *Handlers) RegisterPassword(c *gin.Context) {
	var body struct {
		TelegramUsername string `json:"telegram_username"`
		Nickname         string `json:"nickname"`
		Password         string `json:"password"`
		ConfirmPassword  string `json:"confirm_password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": authGenericError})
		return
	}

	key := authLimitKey(c, body.TelegramUsername)
	if !passwordAuthLimiter.allow(key) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "слишком много попыток, попробуйте позже"})
		return
	}
	if body.Password != body.ConfirmPassword {
		passwordAuthLimiter.fail(key)
		c.JSON(http.StatusBadRequest, gin.H{"error": authGenericError})
		return
	}

	authenticatedUserID, _ := infraauth.UserIDFromContext(c)
	result, err := h.UC.RegisterPasswordUser(
		c.Request.Context(),
		body.TelegramUsername,
		body.Nickname,
		body.Password,
		authenticatedUserID,
	)
	if err != nil {
		passwordAuthLimiter.fail(key)
		status := http.StatusBadRequest
		if errors.Is(err, usecase.ErrUserAlreadyExists) || errors.Is(err, usecase.ErrAccountMismatch) || errors.Is(err, usecase.ErrPasswordLinked) {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": authGenericError})
		return
	}

	if result.RequiresVerification && result.VerificationChallenge != nil {
		if err := h.sendTelegramMessage(
			c.Request.Context(),
			result.VerificationChallenge.TelegramUserID,
			fmt.Sprintf(telegramVerificationMessage, result.VerificationChallenge.Code),
		); err != nil {
			h.Log.Error("telegram verification send failed", "err", err)
			c.JSON(http.StatusAccepted, gin.H{
				"requires_verification": true,
				"telegram_code_sent":    false,
				"message":               "Не удалось отправить код в Telegram. Подтвердите аккаунт через Telegram ниже или откройте бота и нажмите /start.",
			})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"requires_verification": true,
			"telegram_code_sent":    true,
			"message":               "Мы отправили код подтверждения в Telegram. Если код не приходит, можно подтвердить аккаунт через Telegram ниже.",
		})
		return
	}

	passwordAuthLimiter.success(key)
	c.JSON(http.StatusCreated, gin.H{
		"token": result.Token,
		"user":  userToMap(result.User),
		"isNew": true,
	})
}

func (h *Handlers) VerifyRegisterPasswordCode(c *gin.Context) {
	var body struct {
		TelegramUsername string `json:"telegram_username"`
		Code             string `json:"code"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": authGenericError})
		return
	}

	token, u, err := h.UC.VerifyPasswordRegistrationCode(c.Request.Context(), body.TelegramUsername, body.Code)
	if err != nil {
		status := http.StatusBadRequest
		message := authGenericError
		if errors.Is(err, usecase.ErrVerificationCode) {
			status = http.StatusUnauthorized
			message = "неверный или устаревший код"
		}
		if errors.Is(err, usecase.ErrPasswordLinked) {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": message})
		return
	}

	passwordAuthLimiter.success(authLimitKey(c, body.TelegramUsername))
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  userToMap(u),
		"isNew": false,
	})
}

func (h *Handlers) CompleteRegisterPassword(c *gin.Context) {
	uid, ok := infraauth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}

	var body struct {
		TelegramUsername string `json:"telegram_username"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": authGenericError})
		return
	}

	token, u, err := h.UC.CompletePasswordRegistration(c.Request.Context(), uid, body.TelegramUsername)
	if err != nil {
		status := http.StatusBadRequest
		message := authGenericError
		switch {
		case errors.Is(err, usecase.ErrAccountMismatch):
			status = http.StatusForbidden
			message = "подтверждён другой Telegram-аккаунт"
		case errors.Is(err, usecase.ErrVerificationCode):
			status = http.StatusUnauthorized
			message = "сессия подтверждения истекла, начните регистрацию заново"
		case errors.Is(err, usecase.ErrPasswordLinked):
			status = http.StatusConflict
		case errors.Is(err, usecase.ErrNotFound):
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": message})
		return
	}

	passwordAuthLimiter.success(authLimitKey(c, body.TelegramUsername))
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  userToMap(u),
		"isNew": false,
	})
}

func (h *Handlers) LoginPassword(c *gin.Context) {
	var body struct {
		TelegramUsername string `json:"telegram_username"`
		Password         string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": authGenericError})
		return
	}

	key := authLimitKey(c, body.TelegramUsername)
	if !passwordAuthLimiter.allow(key) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "слишком много попыток, попробуйте позже"})
		return
	}

	token, u, err := h.UC.LoginPasswordUser(c.Request.Context(), body.TelegramUsername, body.Password)
	if err != nil {
		passwordAuthLimiter.fail(key)
		c.JSON(http.StatusUnauthorized, gin.H{"error": authGenericError})
		return
	}

	passwordAuthLimiter.success(key)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  userToMap(u),
		"isNew": false,
	})
}

func (h *Handlers) LinkPassword(c *gin.Context) {
	uid, ok := infraauth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}

	var body struct {
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Password != body.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": authGenericError})
		return
	}

	token, u, err := h.UC.LinkPassword(c.Request.Context(), uid, body.Password)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, usecase.ErrNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": authGenericError})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  userToMap(u),
		"isNew": false,
	})
}

// TelegramWebAuthCallback - GET /api/auth/telegram/callback
func (h *Handlers) TelegramWebAuthCallback(c *gin.Context) {
	botToken := h.TelegramBotToken
	flow := c.Query("flow")
	username := strings.TrimSpace(c.Query("username"))

	if botToken == "" {
		h.Log.Error("telegram bot token is empty")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "server configuration error",
		})
		return
	}

	h.Log.Info("TELEGRAM WEB AUTH CALLBACK REQUEST",
		"query_params", c.Request.URL.RawQuery,
	)

	token, user, isNew, err := h.UC.TelegramWebAuth(
		c.Request.Context(),
		c.Request.URL.Query(),
		botToken,
	)

	if err != nil {
		h.Log.Error("❌ TELEGRAM WEB AUTH FAILED",
			"err", err,
		)

		redirectParams := url.Values{}
		redirectParams.Set("error", "telegram_auth_failed")
		if flow != "" {
			redirectParams.Set("flow", flow)
		}
		if username != "" {
			redirectParams.Set("username", username)
		}
		c.Redirect(http.StatusFound, buildFrontendWebAuthURL(h.FrontendURL, redirectParams))
		return
	}

	h.Log.Info("✅ TELEGRAM WEB AUTH SUCCESS",
		"user_id", user.UserID,
		"username", user.Username,
		"is_new", isNew,
	)

	redirectParams := url.Values{}
	redirectParams.Set("token", token)
	if flow != "" {
		redirectParams.Set("flow", flow)
	}
	if username != "" {
		redirectParams.Set("username", username)
	}
	redirectURL := buildFrontendWebAuthURL(h.FrontendURL, redirectParams)

	h.Log.Info("➡️ Redirecting to",
		"url", redirectURL,
	)

	c.Redirect(http.StatusFound, redirectURL)
}
