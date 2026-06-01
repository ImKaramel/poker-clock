package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pridecrm/app-backend/internal/domain"
)

const botTokenHeader = "X-Bot-Token"

func botStringPtr(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

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

func (h *Handlers) BotContact(c *gin.Context) {
	if h.TelegramBotToken == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "telegram bot is not configured"})
		return
	}

	if c.GetHeader(botTokenHeader) != h.TelegramBotToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var body struct {
		UserID         string `json:"user_id"`
		Username       string `json:"username"`
		FirstName      string `json:"first_name"`
		LastName       string `json:"last_name"`
		PhoneNumber    string `json:"phone_number"`
		ChallengeToken string `json:"challenge_token"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	body.UserID = strings.TrimSpace(body.UserID)
	body.Username = strings.TrimSpace(strings.TrimPrefix(body.Username, "@"))
	body.ChallengeToken = strings.TrimSpace(body.ChallengeToken)
	phone, ok := normalizePhoneNumber(body.PhoneNumber)
	if body.UserID == "" || !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and phone_number are required"})
		return
	}
	body.PhoneNumber = phone
	if body.Username == "" {
		body.Username = body.UserID
	}

	u, err := h.Repo.Users.GetByID(c.Request.Context(), body.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if u == nil {
		u = &domain.User{
			UserID:      body.UserID,
			Username:    body.Username,
			FirstName:   botStringPtr(body.FirstName),
			LastName:    botStringPtr(body.LastName),
			PhoneNumber: &body.PhoneNumber,
			IsActive:    true,
		}
		if err := h.Repo.Users.Create(c.Request.Context(), u); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		u.PhoneNumber = &body.PhoneNumber
		if body.Username != "" {
			u.Username = body.Username
		}
		if body.FirstName != "" {
			u.FirstName = botStringPtr(body.FirstName)
		}
		if body.LastName != "" {
			u.LastName = botStringPtr(body.LastName)
		}
		if err := h.Repo.Users.Update(c.Request.Context(), u); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	if body.ChallengeToken != "" {
		if err := h.UC.ConfirmContactAuthChallenge(
			c.Request.Context(),
			body.ChallengeToken,
			body.UserID,
			body.PhoneNumber,
		); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired contact auth challenge"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"user": userToMap(u)})
}
