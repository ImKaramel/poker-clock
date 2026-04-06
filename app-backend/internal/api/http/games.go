package httpapi

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/pridecrm/app-backend/internal/domain"
	infraauth "github.com/pridecrm/app-backend/internal/infrastructure/auth"
	"github.com/pridecrm/app-backend/internal/usecase"
	"net/http"
	"strconv"
	"time"
)

func (h *Handlers) listGamesQuery(ctx context.Context, isAdmin bool) ([]domain.Game, error) {
	if isAdmin {
		return h.Repo.Games.ListAll(ctx)
	}
	return h.Repo.Games.ListActive(ctx)
}

func (h *Handlers) gameParticipantDetails(ctx context.Context, gameID int64) ([]map[string]any, error) {
	parts, err := h.Repo.Participants.ListByGame(ctx, gameID)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(parts))
	for _, p := range parts {
		u, err := h.Repo.Users.GetByID(ctx, p.UserID)
		if err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"id":           p.ID,
			"user":         userToMap(u),
			"position":     p.Position,
			"final_points": p.FinalPoints,
		})
	}
	return out, nil
}

func (h *Handlers) ListGames(c *gin.Context) {
	isAdmin := infraauth.IsAdminFromContext(c)
	games, err := h.listGamesQuery(c.Request.Context(), isAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]map[string]any, 0, len(games))
	for i := range games {
		g := &games[i]
		n, _ := h.Repo.Participants.CountByGame(c.Request.Context(), g.GameID)
		details, _ := h.gameParticipantDetails(c.Request.Context(), g.GameID)
		out = append(out, gameToMap(g, n, details))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handlers) GetGame(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	g, err := h.Repo.Games.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if g == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	isAdmin := infraauth.IsAdminFromContext(c)
	if !isAdmin && !g.IsActive {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	n, _ := h.Repo.Participants.CountByGame(c.Request.Context(), g.GameID)
	details, _ := h.gameParticipantDetails(c.Request.Context(), g.GameID)
	c.JSON(http.StatusOK, gameToMap(g, n, details))
}

type gameMutationBody struct {
	Date                     string  `json:"date"`
	Time                     string  `json:"time"`
	Description              string  `json:"description"`
	Buyin                    float64 `json:"buyin"`
	ReentryBuyin             float64 `json:"reentry_buyin"`
	Location                 string  `json:"location"`
	Photo                    *string `json:"photo"`
	BasePoints               int     `json:"base_points"`
	PointsPerExtraPlayer     int     `json:"points_per_extra_player"`
	MinPlayersForExtraPoints int     `json:"min_players_for_extra_points"`
}

func parseGameBody(body gameMutationBody) (*domain.Game, error) {
	d, err := time.Parse("2006-01-02", body.Date)
	if err != nil {
		return nil, err
	}
	t, err := time.Parse("15:04:05", body.Time)
	if err != nil {
		t, err = time.Parse(time.RFC3339, body.Time)
		if err != nil {
			return nil, err
		}
	}
	g := &domain.Game{
		Date:                     d,
		Time:                     t,
		Description:              body.Description,
		Buyin:                    body.Buyin,
		ReentryBuyin:             body.ReentryBuyin,
		Location:                 body.Location,
		Photo:                    body.Photo,
		IsActive:                 true,
		BasePoints:               body.BasePoints,
		PointsPerExtraPlayer:     body.PointsPerExtraPlayer,
		MinPlayersForExtraPoints: body.MinPlayersForExtraPoints,
	}
	if g.BasePoints == 0 {
		g.BasePoints = 100
		g.PointsPerExtraPlayer = 10
		g.MinPlayersForExtraPoints = 10
	}
	return g, nil
}

func (h *Handlers) CreateGame(c *gin.Context) {
	var body gameMutationBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	g, err := parseGameBody(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date/time"})
		return
	}
	if err := h.Repo.Games.Create(c.Request.Context(), g); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	n, _ := h.Repo.Participants.CountByGame(c.Request.Context(), g.GameID)
	c.JSON(http.StatusCreated, gameToMap(g, n, nil))
}

func (h *Handlers) UpdateGame(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	g, err := h.Repo.Games.GetByID(c.Request.Context(), id)
	if err != nil || g == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var body gameMutationBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ng, err := parseGameBody(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date/time"})
		return
	}
	g.Date = ng.Date
	g.Time = ng.Time
	g.Description = ng.Description
	g.Buyin = ng.Buyin
	g.ReentryBuyin = ng.ReentryBuyin
	g.Location = ng.Location
	if body.Photo != nil {
		g.Photo = body.Photo
	}
	if body.BasePoints != 0 || ng.BasePoints != 0 {
		g.BasePoints = ng.BasePoints
		g.PointsPerExtraPlayer = ng.PointsPerExtraPlayer
		g.MinPlayersForExtraPoints = ng.MinPlayersForExtraPoints
	}
	if err := h.Repo.Games.Update(c.Request.Context(), g); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	n, _ := h.Repo.Participants.CountByGame(c.Request.Context(), g.GameID)
	details, _ := h.gameParticipantDetails(c.Request.Context(), g.GameID)
	c.JSON(http.StatusOK, gameToMap(g, n, details))
}

func (h *Handlers) DeleteGame(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.Repo.Games.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

type gameAdminUserBody struct {
	UserID  string `json:"user_id"`
	Entries int    `json:"entries"`
	Rebuys  int    `json:"rebuys"`
	Addons  int    `json:"addons"`
}

func (h *Handlers) GameParticipantsAdmin(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	parts, err := h.Repo.Participants.ListByGame(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]map[string]any, 0, len(parts))
	for _, p := range parts {
		u, _ := h.Repo.Users.GetByID(c.Request.Context(), p.UserID)
		var ui map[string]any
		if u != nil {
			ui = map[string]any{
				"user_id": u.UserID, "username": u.Username,
				"first_name": derefStrPtr(u.FirstName), "last_name": derefStrPtr(u.LastName),
			}
		}
		out = append(out, map[string]any{
			"id": p.ID, "user": p.UserID, "user_info": ui, "game": p.GameID,
			"entries": p.Entries, "rebuys": p.Rebuys, "addons": p.Addons,
			"final_points": p.FinalPoints, "position": p.Position,
			"joined_at": p.JoinedAt.UTC().Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handlers) GameAddParticipantAdmin(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var body gameAdminUserBody
	if err := c.ShouldBindJSON(&body); err != nil || body.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id required"})
		return
	}
	game, err := h.Repo.Games.GetByID(c.Request.Context(), id)
	if err != nil || game == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}
	u, err := h.Repo.Users.GetByID(c.Request.Context(), body.UserID)
	if err != nil || u == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	p, err := h.Repo.Participants.GetByUserAndGame(c.Request.Context(), body.UserID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ent, reb, add := body.Entries, body.Rebuys, body.Addons
	if ent <= 0 {
		ent = 1
	}
	if p == nil {
		p = &domain.Participant{UserID: body.UserID, GameID: id, Entries: ent, Rebuys: reb, Addons: add}
		if err := h.Repo.Participants.Create(c.Request.Context(), p); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		if body.Entries != 0 {
			p.Entries = ent
		}
		p.Rebuys = reb
		p.Addons = add
		if err := h.Repo.Participants.Update(c.Request.Context(), p); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	uFull, _ := h.Repo.Users.GetByID(c.Request.Context(), p.UserID)
	c.JSON(http.StatusOK, map[string]any{
		"id": p.ID, "user": userToMap(uFull), "game": p.GameID,
		"entries": p.Entries, "rebuys": p.Rebuys, "addons": p.Addons,
		"final_points": p.FinalPoints, "position": p.Position,
		"joined_at": p.JoinedAt.UTC().Format(time.RFC3339),
	})
}

func (h *Handlers) GameRemoveParticipantAdmin(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var body struct {
		UserID string `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id required"})
		return
	}
	u, err := h.Repo.Users.GetByID(c.Request.Context(), body.UserID)
	if err != nil || u == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	p, err := h.Repo.Participants.GetByUserAndGame(c.Request.Context(), body.UserID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if p == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Participant not found"})
		return
	}
	if err := h.Repo.Participants.DeleteByUserAndGame(c.Request.Context(), body.UserID, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "participant removed"})
}

type completeBody struct {
	Participants []usecase.CompleteParticipantInput `json:"participants"`
}

func (h *Handlers) GameComplete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var body completeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	th, err := h.UC.CompleteGame(c.Request.Context(), id, body.Participants)
	if err != nil {
		if err == usecase.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		h.Log.Error("complete game", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete game: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, tournamentHistoryToMap(th))
}

func (h *Handlers) GameUpdateParticipantAdmin(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var body struct {
		UserID      string `json:"user_id"`
		Entries     *int   `json:"entries"`
		Rebuys      *int   `json:"rebuys"`
		Addons      *int   `json:"addons"`
		Position    *int   `json:"position"`
		FinalPoints *int   `json:"final_points"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id required"})
		return
	}
	u, err := h.Repo.Users.GetByID(c.Request.Context(), body.UserID)
	if err != nil || u == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	p, err := h.Repo.Participants.GetByUserAndGame(c.Request.Context(), body.UserID, id)
	if err != nil || p == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Participant not found"})
		return
	}
	if body.Entries != nil {
		p.Entries = *body.Entries
	}
	if body.Rebuys != nil {
		p.Rebuys = *body.Rebuys
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
	if err := h.Repo.Participants.Update(c.Request.Context(), p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uFull, _ := h.Repo.Users.GetByID(c.Request.Context(), p.UserID)
	c.JSON(http.StatusOK, map[string]any{
		"id": p.ID, "user": userToMap(uFull), "game": p.GameID,
		"entries": p.Entries, "rebuys": p.Rebuys, "addons": p.Addons,
		"final_points": p.FinalPoints, "position": p.Position,
		"joined_at": p.JoinedAt.UTC().Format(time.RFC3339),
	})
}
