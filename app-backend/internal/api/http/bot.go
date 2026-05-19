package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pridecrm/app-backend/internal/usecase"
)

const botTokenHeader = "X-Bot-Token"

func (h *Handlers) BotRecipients(c *gin.Context) {
	if h.TelegramBotToken == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "telegram bot is not configured"})
		return
	}

	if c.GetHeader(botTokenHeader) != h.TelegramBotToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	users, err := h.Repo.Users.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	recipients := make([]map[string]any, 0, len(users))
	for i := range users {
		u := users[i]
		if strings.TrimSpace(u.UserID) == "" || u.IsBanned || !u.IsActive {
			continue
		}

		recipients = append(recipients, map[string]any{
			"user_id":    u.UserID,
			"username":   u.Username,
			"nick_name":  derefStr(u.NickName),
			"first_name": derefStr(u.FirstName),
		})
	}

	c.JSON(http.StatusOK, gin.H{"users": recipients})
}

func (h *Handlers) BotPasswordRegistrationCode(c *gin.Context) {
	if h.TelegramBotToken == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "telegram bot is not configured"})
		return
	}
	if c.GetHeader(botTokenHeader) != h.TelegramBotToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var body struct {
		TelegramUserID   string `json:"telegram_user_id"`
		TelegramUsername string `json:"telegram_username"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	code, err := h.UC.PendingPasswordRegistrationCode(
		c.Request.Context(),
		strings.TrimSpace(body.TelegramUserID),
		body.TelegramUsername,
	)
	if err != nil {
		status := http.StatusBadRequest
		message := "verification unavailable"
		switch {
		case errors.Is(err, usecase.ErrVerificationCode):
			status = http.StatusNotFound
			message = "registration request not found or expired"
		case errors.Is(err, usecase.ErrAccountMismatch):
			status = http.StatusForbidden
			message = "telegram account mismatch"
		case errors.Is(err, usecase.ErrPasswordLinked):
			status = http.StatusConflict
			message = "password already linked"
		case errors.Is(err, usecase.ErrNotFound):
			status = http.StatusNotFound
			message = "user not found"
		case errors.Is(err, usecase.ErrInvalidAuthInput):
			status = http.StatusBadRequest
			message = "invalid payload"
		}
		c.JSON(status, gin.H{"error": message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": code})
}
