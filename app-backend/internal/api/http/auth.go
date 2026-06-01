package httpapi

import (
	"errors"
	"fmt"
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
var contactAuthLimiter = newAuthLimiter(10, 10*time.Minute)

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

func contactAuthLimitKey(c *gin.Context) string {
	return c.ClientIP() + ":contact-auth"
}

func (h *Handlers) contactAuthBotLink(token string) string {
	username := strings.TrimSpace(strings.TrimPrefix(h.TelegramBotUsername, "@"))
	if username == "" {
		username = "Midnight_poker_bot"
	}
	return fmt.Sprintf("https://t.me/%s?start=contact_%s", url.QueryEscape(username), url.QueryEscape(token))
}

func (h *Handlers) ContactAuthStart(c *gin.Context) {
	key := contactAuthLimitKey(c)
	if !contactAuthLimiter.allow(key) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "слишком много попыток, попробуйте позже"})
		return
	}
	contactAuthLimiter.fail(key)

	challenge, err := h.UC.CreateContactAuthChallenge(c.Request.Context())
	if err != nil {
		h.Log.Error("contact auth start failed", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start contact auth"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token":      challenge.Token,
		"status":     challenge.Status,
		"bot_link":   h.contactAuthBotLink(challenge.Token),
		"expires_at": challenge.ExpiresAt.UTC().Format(time.RFC3339),
	})
}

func (h *Handlers) ContactAuthStatus(c *gin.Context) {
	token := strings.TrimSpace(c.Query("token"))
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token required"})
		return
	}

	jwtToken, u, err := h.UC.ConsumeContactAuthChallenge(c.Request.Context(), token)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrContactAuthPending):
			c.JSON(http.StatusOK, gin.H{"status": "pending"})
		case errors.Is(err, usecase.ErrContactAuthExpired):
			c.JSON(http.StatusGone, gin.H{"status": "expired", "error": "contact auth expired"})
		case errors.Is(err, usecase.ErrContactAuthConsumed):
			c.JSON(http.StatusConflict, gin.H{"status": "consumed", "error": "contact auth already used"})
		case errors.Is(err, usecase.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		default:
			h.Log.Error("contact auth status failed", "err", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check contact auth"})
		}
		return
	}

	contactAuthLimiter.success(contactAuthLimitKey(c))
	c.JSON(http.StatusOK, gin.H{
		"status": "completed",
		"token":  jwtToken,
		"user":   userToMap(u),
		"isNew":  false,
	})
}

func (h *Handlers) TelegramAuth(c *gin.Context) {
	var body struct {
		InitData      string `json:"init_data"`
		InitDataCamel string `json:"initData"`
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

	initData := strings.TrimSpace(body.InitData)
	if initData == "" {
		initData = strings.TrimSpace(body.InitDataCamel)
	}

	if initData == "" {
		h.Log.Error("TELEGRAM INIT DATA IS EMPTY")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "telegram init data is required",
		})
		return
	}

	if h.TelegramBotToken == "" {
		h.Log.Error("telegram bot token is empty")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server configuration error"})
		return
	}

	h.Log.Info("TELEGRAM INIT DATA AUTH REQUEST OK")

	token, u, isNew, err := h.UC.TelegramAuthInitData(c.Request.Context(), initData, h.TelegramBotToken)
	if err != nil {
		h.Log.Error("❌ AUTH FAILED", "err", err)
		status := http.StatusUnauthorized
		if errors.Is(err, usecase.ErrInvalidTelegramAuth) {
			status = http.StatusUnauthorized
		}
		c.JSON(status, gin.H{"error": "invalid telegram auth data"})
		return
	}
	h.Log.Info("isNew TELEGRAM AUTH REQUEST OK",
		"isNew", isNew,
		"user_id", u.UserID,
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

	result, err := h.UC.RegisterPasswordUser(
		c.Request.Context(),
		body.TelegramUsername,
		body.Nickname,
		body.Password,
	)
	if err != nil {
		passwordAuthLimiter.fail(key)
		status := http.StatusBadRequest
		message := authGenericError
		switch {
		case errors.Is(err, usecase.ErrUserAlreadyExists):
			status = http.StatusConflict
			message = "username уже занят"
		case errors.Is(err, usecase.ErrNicknameTaken):
			status = http.StatusConflict
			message = "nickname уже занят"
		}
		c.JSON(status, gin.H{"error": message})
		return
	}

	passwordAuthLimiter.success(key)
	c.JSON(http.StatusCreated, gin.H{
		"token": result.Token,
		"user":  userToMap(result.User),
		"isNew": true,
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

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	h.Log.Info("✅ TELEGRAM WEB AUTH SUCCESS",
		"user_id", user.UserID,
		"username", user.Username,
		"is_new", isNew,
	)

	frontendURL := strings.TrimRight(h.FrontendURL, "/")
	if frontendURL == "" {
		frontendURL = "https://midnight-club-app.ru"
	}
	redirectURL := fmt.Sprintf("%s/web-auth?token=%s", frontendURL, url.QueryEscape(token))

	h.Log.Info("➡️ Redirecting to",
		"url", redirectURL,
	)

	c.Redirect(http.StatusFound, redirectURL)
}
