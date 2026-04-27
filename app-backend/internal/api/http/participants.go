package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pridecrm/app-backend/internal/domain"
	infraauth "github.com/pridecrm/app-backend/internal/infrastructure/auth"
	"github.com/pridecrm/app-backend/internal/usecase"
)

func (h *Handlers) ListParticipants(c *gin.Context) {
	uid, ok := infraauth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}
	isAdmin := infraauth.IsAdminFromContext(c)
	var parts []domain.Participant
	var err error
	if isAdmin {
		parts, err = h.Repo.Participants.ListAll(c.Request.Context())
	} else {
		parts, err = h.Repo.Participants.ListByUser(c.Request.Context(), uid)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]map[string]any, 0, len(parts))
	for _, p := range parts {
		u, _ := h.Repo.Users.GetByID(c.Request.Context(), p.UserID)
		out = append(out, map[string]any{
			"id": p.ID, "user": userToMap(u), "game": p.GameID,
			"entries": p.Entries, "rebuys": p.Rebuys, "addons": p.Addons,
			"final_points": p.FinalPoints, "position": p.Position, "arrived": p.Arrived, "is_out": p.IsOut,
			"joined_at": p.JoinedAt.UTC().Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handlers) GetParticipant(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	uid, ok := infraauth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}
	found, err := h.Repo.Participants.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if found == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if !infraauth.IsAdminFromContext(c) && found.UserID != uid {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	u, _ := h.Repo.Users.GetByID(c.Request.Context(), found.UserID)
	c.JSON(http.StatusOK, map[string]any{
		"id": found.ID, "user": userToMap(u), "game": found.GameID,
		"entries": found.Entries, "rebuys": found.Rebuys, "addons": found.Addons,
		"final_points": found.FinalPoints, "position": found.Position,
		"arrived": found.Arrived, "is_out": found.IsOut,
		"joined_at": found.JoinedAt.UTC().Format(time.RFC3339),
	})
}

func (h *Handlers) CreateParticipant(c *gin.Context) {
	var body struct {
		UserID  string `json:"user_id"`
		GameID  int64  `json:"game"`
		Entries int    `json:"entries"`
		Rebuys  int    `json:"rebuys"`
		Addons  int    `json:"addons"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.UserID == "" || body.GameID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and game required"})
		return
	}
	ent := body.Entries
	if ent <= 0 {
		ent = 1
	}
	p := &domain.Participant{UserID: body.UserID, GameID: body.GameID, Entries: ent, Rebuys: body.Rebuys, Addons: body.Addons}
	if err := h.Repo.Participants.Create(c.Request.Context(), p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, _ := h.Repo.Users.GetByID(c.Request.Context(), p.UserID)
	c.JSON(http.StatusCreated, map[string]any{
		"id": p.ID, "user": userToMap(u), "game": p.GameID,
		"entries": p.Entries, "rebuys": p.Rebuys, "addons": p.Addons,
		"final_points": p.FinalPoints, "position": p.Position, "arrived": p.Arrived, "is_out": p.IsOut,
		"joined_at": p.JoinedAt.UTC().Format(time.RFC3339),
	})
}

func (h *Handlers) UpdateParticipant(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	p, err := h.Repo.Participants.GetByID(c.Request.Context(), id)
	if err != nil || p == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	var body struct {
		Entries     *int  `json:"entries"`
		Rebuys      *int  `json:"rebuys"` // delta: +1 / -1
		Addons      *int  `json:"addons"`
		Position    *int  `json:"position"`
		FinalPoints *int  `json:"final_points"`
		Arrived     *bool `json:"arrived"`
		IsOut       *bool `json:"is_out"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.Entries != nil {
		p.Entries = *body.Entries
	}
	rebuyDelta := 0
	if body.Rebuys != nil {
		rebuyDelta = *body.Rebuys
	}

	if body.Addons != nil {
		p.Addons = *body.Addons
	}

	if body.Position != nil {
		p.Position = body.Position
	}

	if body.FinalPoints != nil {
		p.FinalPoints = *body.FinalPoints
	}

	if body.Arrived != nil {
		p.Arrived = *body.Arrived
	}

	if body.IsOut != nil {
		p.IsOut = *body.IsOut
	}

	if err := h.Repo.Participants.Update(
		c.Request.Context(),
		p,
		rebuyDelta,
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.Repo.Participants.GetByID(c.Request.Context(), id)
	if err != nil || updated == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reload participant"})
		return
	}

	u, _ := h.Repo.Users.GetByID(c.Request.Context(), updated.UserID)

	c.JSON(http.StatusOK, map[string]any{
		"id":           updated.ID,
		"user":         userToMap(u),
		"game":         updated.GameID,
		"entries":      updated.Entries,
		"rebuys":       updated.Rebuys,
		"addons":       updated.Addons,
		"final_points": updated.FinalPoints,
		"position":     updated.Position,
		"arrived":      updated.Arrived,
		"is_out":       updated.IsOut,
		"joined_at":    updated.JoinedAt.UTC().Format(time.RFC3339),
	})
}

func (h *Handlers) DeleteParticipant(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.Repo.Participants.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handlers) RegisterParticipant(c *gin.Context) {
	uid, ok := infraauth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}
	var body struct {
		GameID int64 `json:"game_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game_id required"})
		return
	}
	already, err := h.UC.RegisterParticipant(c.Request.Context(), uid, body.GameID)
	if err == usecase.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}
	if err == usecase.ErrForbidden {
		c.JSON(http.StatusForbidden, gin.H{"error": "User is banned"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if already {
		c.JSON(http.StatusOK, gin.H{"status": "already registered"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "registered"})
}

func (h *Handlers) UnregisterParticipant(c *gin.Context) {
	uid, ok := infraauth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}
	var body struct {
		GameID int64 `json:"game_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game_id required"})
		return
	}
	err := h.UC.UnregisterParticipant(c.Request.Context(), uid, body.GameID)
	if err == usecase.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not registered for this game"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handlers) SetParticipantArrived(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Arrived bool `json:"arrived"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad request"})
		return
	}

	err = h.UC.SetParticipantArrived(c.Request.Context(), id, req.Arrived)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Status(204)
}
