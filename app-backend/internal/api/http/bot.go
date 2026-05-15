package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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
