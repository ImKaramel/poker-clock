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

	token, u, err := h.UC.RegisterPasswordUser(c.Request.Context(), body.TelegramUsername, body.Nickname, body.Password)
	if err != nil {
		passwordAuthLimiter.fail(key)
		status := http.StatusBadRequest
		if errors.Is(err, usecase.ErrUserAlreadyExists) {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": authGenericError})
		return
	}

	passwordAuthLimiter.success(key)
	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user":  userToMap(u),
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

	redirectURL := fmt.Sprintf(
		"https://www.midnight-club-app.ru/web-auth?token=%s",
		//h.FrontendURL,
		url.QueryEscape(token),
	)

	h.Log.Info("➡️ Redirecting to",
		"url", redirectURL,
	)

	c.Redirect(http.StatusFound, redirectURL)
}
