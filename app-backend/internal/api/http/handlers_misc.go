package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/pridecrm/app-backend/internal/domain"
	infraauth "github.com/pridecrm/app-backend/internal/infrastructure/auth"
)

func ticketToMap(t *domain.SupportTicket) map[string]any {
	var userID any
	if t.UserID != "" {
		userID = t.UserID
	}
	return map[string]any{
		"id":         t.ID,
		"user":       userID,
		"subject":    t.Subject,
		"message":    t.Message,
		"status":     t.Status,
		"created_at": t.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at": t.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func (h *Handlers) ListTickets(c *gin.Context) {
	uid, ok := infraauth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}
	isAdmin := infraauth.IsAdminFromContext(c)
	var list []domain.SupportTicket
	var err error
	if isAdmin {
		list, err = h.Repo.Tickets.ListAll(c.Request.Context())
	} else {
		list, err = h.Repo.Tickets.ListByUser(c.Request.Context(), uid)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for i := range list {
		t := &list[i]
		u, _ := h.Repo.Users.GetByID(c.Request.Context(), t.UserID)
		m := ticketToMap(t)
		m["user"] = userToMap(u)
		out = append(out, m)
	}
	c.JSON(http.StatusOK, out)
}

type ticketCreateBody struct {
	Subject string `json:"subject"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

func (h *Handlers) CreateTicket(c *gin.Context) {
	uid, ok := infraauth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}
	var body ticketCreateBody
	if err := c.ShouldBindJSON(&body); err != nil || body.Subject == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subject required"})
		return
	}
	st := body.Status
	if st == "" {
		st = "open"
	}
	t := &domain.SupportTicket{UserID: uid, Subject: body.Subject, Message: body.Message, Status: st}
	if err := h.Repo.Tickets.Create(c.Request.Context(), t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, _ := h.Repo.Users.GetByID(c.Request.Context(), uid)
	m := ticketToMap(t)
	m["user"] = userToMap(u)
	c.JSON(http.StatusCreated, m)
}

func (h *Handlers) GetTicket(c *gin.Context) {
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
	t, err := h.Repo.Tickets.GetByID(c.Request.Context(), id)
	if err != nil || t == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if !infraauth.IsAdminFromContext(c) && t.UserID != uid {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	u, _ := h.Repo.Users.GetByID(c.Request.Context(), t.UserID)
	m := ticketToMap(t)
	m["user"] = userToMap(u)
	c.JSON(http.StatusOK, m)
}

func (h *Handlers) UpdateTicket(c *gin.Context) {
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
	t, err := h.Repo.Tickets.GetByID(c.Request.Context(), id)
	if err != nil || t == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if !infraauth.IsAdminFromContext(c) && t.UserID != uid {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var body ticketCreateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Subject != "" {
		t.Subject = body.Subject
	}
	if body.Message != "" {
		t.Message = body.Message
	}
	if body.Status != "" {
		t.Status = body.Status
	}
	if err := h.Repo.Tickets.Update(c.Request.Context(), t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, _ := h.Repo.Users.GetByID(c.Request.Context(), t.UserID)
	m := ticketToMap(t)
	m["user"] = userToMap(u)
	c.JSON(http.StatusOK, m)
}

func (h *Handlers) DeleteTicket(c *gin.Context) {
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
	t, err := h.Repo.Tickets.GetByID(c.Request.Context(), id)
	if err != nil || t == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if !infraauth.IsAdminFromContext(c) && t.UserID != uid {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err := h.Repo.Tickets.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handlers) ListTournamentHistory(c *gin.Context) {
	list, err := h.Repo.Tournaments.ListHistory(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for i := range list {
		out = append(out, tournamentHistoryListMap(&list[i]))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handlers) GetTournamentHistory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	th, err := h.Repo.Tournaments.GetHistoryByID(c.Request.Context(), id)
	if err != nil || th == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, tournamentHistoryToMap(th))
}

type tournamentCreateBody struct {
	GameID            int64   `json:"game"`
	Date              string  `json:"date"`
	Time              *string `json:"time"`
	TournamentName    string  `json:"tournament_name"`
	Location          string  `json:"location"`
	Buyin             int     `json:"buyin"`
	ReentryBuyin      *int    `json:"reentry_buyin"`
	TotalRevenue      int     `json:"total_revenue"`
	ParticipantsCount int     `json:"participants_count"`
}

func (h *Handlers) CreateTournamentHistory(c *gin.Context) {
	var body tournamentCreateBody
	if err := c.ShouldBindJSON(&body); err != nil || body.GameID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game required"})
		return
	}
	d, err := time.Parse("2006-01-02", body.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date"})
		return
	}
	var tm *time.Time
	if body.Time != nil && *body.Time != "" {
		t, err := time.Parse("15:04:05", *body.Time)
		if err == nil {
			tm = &t
		}
	}
	th := &domain.TournamentHistory{
		GameID:            body.GameID,
		Date:              d,
		Time:              tm,
		TournamentName:    body.TournamentName,
		Location:          body.Location,
		Buyin:             body.Buyin,
		ReentryBuyin:      body.ReentryBuyin,
		TotalRevenue:      body.TotalRevenue,
		ParticipantsCount: body.ParticipantsCount,
	}
	if err := h.Repo.Tournaments.CreateHistory(c.Request.Context(), th); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, tournamentHistoryToMap(th))
}

func (h *Handlers) UpdateTournamentHistory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	th, err := h.Repo.Tournaments.GetHistoryByID(c.Request.Context(), id)
	if err != nil || th == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var body tournamentCreateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Date != "" {
		d, err := time.Parse("2006-01-02", body.Date)
		if err == nil {
			th.Date = d
		}
	}
	if body.Time != nil && *body.Time != "" {
		t, err := time.Parse("15:04:05", *body.Time)
		if err == nil {
			th.Time = &t
		}
	}
	if body.TournamentName != "" {
		th.TournamentName = body.TournamentName
	}
	if body.Location != "" {
		th.Location = body.Location
	}
	if body.Buyin != 0 || body.TournamentName != "" {
		th.Buyin = body.Buyin
	}
	if body.ReentryBuyin != nil {
		th.ReentryBuyin = body.ReentryBuyin
	}
	if body.TotalRevenue != 0 || body.TournamentName != "" {
		th.TotalRevenue = body.TotalRevenue
	}
	if body.ParticipantsCount != 0 || body.TournamentName != "" {
		th.ParticipantsCount = body.ParticipantsCount
	}
	if err := h.Repo.Tournaments.UpdateHistory(c.Request.Context(), th); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	full, _ := h.Repo.Tournaments.GetHistoryByID(c.Request.Context(), id)
	c.JSON(http.StatusOK, tournamentHistoryToMap(full))
}

func (h *Handlers) DeleteTournamentHistory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.Repo.Tournaments.DeleteHistory(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handlers) TournamentParticipants(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	parts, err := h.Repo.Tournaments.ListTournamentParticipants(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]map[string]any, 0, len(parts))
	for _, p := range parts {
		out = append(out, map[string]any{
			"id":                     p.ID,
			"user_id":                p.UserID,
			"username":               p.Username,
			"first_name":             p.FirstName,
			"last_name":              p.LastName,
			"entries":                p.Entries,
			"rebuys":                 p.Rebuys,
			"addons":                 p.Addons,
			"total_spent":            p.TotalSpent,
			"payment_method":         p.PaymentMethod,
			"payment_method_display": paymentMethodDisplay(p.PaymentMethod),
		})
	}
	c.JSON(http.StatusOK, out)
}
