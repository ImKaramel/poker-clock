package httpapi

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

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
