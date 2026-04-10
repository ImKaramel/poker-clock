package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type telegramBody struct {
	TelegramData map[string]any `json:"telegram_data"`
}

func (h *Handlers) TelegramAuth(c *gin.Context) {
	var body telegramBody
	if err := c.ShouldBindJSON(&body); err != nil || body.TelegramData == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Telegram data required"})
		return
	}
	token, u, isNew, err := h.UC.TelegramAuth(c.Request.Context(), body.TelegramData)
	if err != nil {
		h.Log.Error("telegram auth", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		return
	}
	//photoURL, _ := telegramData["photo_url"].(string)
	c.JSON(http.StatusOK, gin.H{
		"token":  token,
		"user":   userToMap(u),
		"is_new": isNew,
	})
}

type validateBody struct {
	InitData string `json:"initData"`
}

func (h *Handlers) TelegramValidate(c *gin.Context) {
	var body validateBody
	if err := c.ShouldBindJSON(&body); err != nil || body.InitData == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing initData"})
		return
	}
	token, u, err := h.UC.TelegramValidateInitData(c.Request.Context(), body.InitData)
	if err != nil {
		h.Log.Error("telegram validate", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Auth failed: " + err.Error()})
		return
	}
	h.Log.Info("TOKEN GENERATED", "token", token, "user", u.UserID)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"username":    u.Username,
			"first_name":  derefStrPtr(u.FirstName),
			"last_name":   derefStrPtr(u.LastName),
			"telegram_id": u.UserID,
			"id":          u.UserID,
			"photo_url":   derefStrPtr(u.PhotoURL),
		},
	})
}
