package httpapi

import (
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pridecrm/app-backend/internal/domain"
	infraauth "github.com/pridecrm/app-backend/internal/infrastructure/auth"
)

func (h *Handlers) Rating(c *gin.Context) {
	month := time.Now()
	if raw := c.Query("month"); raw != "" {
		parsed, err := time.Parse("2006-01", raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month"})
			return
		}
		month = parsed
	}
	users, err := h.Repo.Users.ListForRatingByMonth(c.Request.Context(), month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(users) == 0 {
		users, err = h.Repo.Users.ListForRating(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	out := make([]map[string]any, 0, len(users))
	for rank, u := range users {
		out = append(out, map[string]any{
			"rank":         rank + 1,
			"user":         userToMap(&u),
			"points":       u.Points,
			"games_played": u.TotalGamesPlayed,
		})
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handlers) ProfileGet(c *gin.Context) {
	uid, ok := infraauth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}
	u, err := h.Repo.Users.GetByID(c.Request.Context(), uid)
	if err != nil || u == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	now := time.Now().In(time.Local)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	parts, err := h.Repo.Participants.ListUpcomingForUser(c.Request.Context(), uid, today)
	if err != nil {
		h.Log.Error("profile upcoming", slog.Any("err", err))
		parts = nil
	}
	games := make([]map[string]any, 0, len(parts))
	for _, p := range parts {
		g, err := h.Repo.Games.GetByID(c.Request.Context(), p.GameID)
		if err != nil || g == nil {
			continue
		}
		n, _ := h.Repo.Participants.CountByGame(c.Request.Context(), g.GameID)
		details, _ := h.gameParticipantDetails(c.Request.Context(), g.GameID)
		games = append(games, gameToMap(g, n, details))
	}
	history, err := h.Repo.Tournaments.ListHistoryByUser(c.Request.Context(), uid)
	if err != nil {
		h.Log.Error("profile past games", slog.Any("err", err))
		history = nil
	}
	pastGames := make([]map[string]any, 0, len(history))
	for i := range history {
		pastGames = append(pastGames, tournamentHistoryGameToMap(&history[i]))
	}
	c.JSON(http.StatusOK, gin.H{
		"user":           userToMap(u),
		"upcoming_games": games,
		"past_games":     pastGames,
	})
}

func tournamentHistoryGameToMap(h *domain.TournamentHistory) map[string]any {
	m := map[string]any{
		"history_id":           h.ID,
		"game_id":              h.GameID,
		"name":                 h.TournamentName,
		"date":                 h.Date.Format("2006-01-02"),
		"description":          h.TournamentName,
		"location":             h.Location,
		"buyin":                h.Buyin,
		"reentry_buyin":        h.ReentryBuyin,
		"participants_count":   h.ParticipantsCount,
		"is_active":            false,
		"completed_at":         h.CompletedAt.UTC().Format(time.RFC3339),
		"total_revenue":        h.TotalRevenue,
		"participants_details": []map[string]any{},
		"photo":                nil,
	}
	if h.Time != nil {
		m["time"] = h.Time.Format("15:04:05")
	} else {
		m["time"] = nil
	}
	return m
}

func (h *Handlers) ProfilePatch(c *gin.Context) {
	uid, ok := infraauth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}
	u, err := h.Repo.Users.GetByID(c.Request.Context(), uid)
	if err != nil || u == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var body userPatch
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Username != nil {
		u.Username = *body.Username
	}
	if body.NickName != nil {
		u.NickName = body.NickName
	}
	if body.FirstName != nil {
		u.FirstName = body.FirstName
	}
	if body.LastName != nil {
		u.LastName = body.LastName
	}
	if body.Phone != nil {
		u.PhoneNumber = body.Phone
	}
	if body.Email != nil {
		u.Email = body.Email
	}
	if body.DOB != nil && *body.DOB != "" {
		t, err := time.Parse("2006-01-02", *body.DOB)
		if err == nil {
			u.DateOfBirth = &t
		}
	}
	if err := h.Repo.Users.Update(c.Request.Context(), u); err != nil {
		if strings.Contains(err.Error(), "unique") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "nickname already taken",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, userToMap(u))
}

func (h *Handlers) ProfileAvatarUpload(c *gin.Context) {
	uid, ok := infraauth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}
	u, err := h.Repo.Users.GetByID(c.Request.Context(), uid)
	if err != nil || u == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "avatar required"})
		return
	}
	if fileHeader.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "avatar too large"})
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid avatar"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, 5*1024*1024+1))
	if err != nil || len(data) == 0 || len(data) > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid avatar"})
		return
	}
	if h.UC.Storage == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "storage unavailable"})
		return
	}
	url, err := h.UC.Storage.UploadAvatar(c.Request.Context(), uid, data)
	if err != nil {
		h.Log.Error("avatar upload", slog.Any("err", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload avatar"})
		return
	}
	u.PhotoURL = &url
	if err := h.Repo.Users.Update(c.Request.Context(), u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": userToMap(u), "photo_url": url})
}

func (h *Handlers) AdminDashboard(c *gin.Context) {
	uid, ok := infraauth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}
	u, err := h.Repo.Users.GetByID(c.Request.Context(), uid)
	if err != nil || u == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	tu, _ := h.Repo.Users.Count(c.Request.Context())
	tg, _ := h.Repo.Games.Count(c.Request.Context())
	ag, _ := h.Repo.Games.CountActive(c.Request.Context())
	bu, _ := h.Repo.Users.CountBanned(c.Request.Context())
	tp, _ := h.Repo.Participants.Count(c.Request.Context())
	ot, _ := h.Repo.Tickets.CountOpen(c.Request.Context())
	rg, _ := h.Repo.Games.ListRecent(c.Request.Context(), 5)
	ru, _ := h.Repo.Users.ListRecent(c.Request.Context(), 5)
	gamesOut := make([]map[string]any, 0, len(rg))
	for i := range rg {
		g := &rg[i]
		n, _ := h.Repo.Participants.CountByGame(c.Request.Context(), g.GameID)
		details, _ := h.gameParticipantDetails(c.Request.Context(), g.GameID)
		gamesOut = append(gamesOut, gameToMap(g, n, details))
	}
	usersOut := make([]map[string]any, 0, len(ru))
	for i := range ru {
		usersOut = append(usersOut, userToMap(&ru[i]))
	}
	c.JSON(http.StatusOK, gin.H{
		"admin_info": gin.H{"is_admin": u.IsAdmin},
		"stats": gin.H{
			"total_users": tu, "total_games": tg, "active_games": ag,
			"banned_users": bu, "total_participants": tp, "open_tickets": ot,
		},
		"recent_games": gamesOut,
		"recent_users": usersOut,
	})
}
