package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/pridecrm/app-backend/internal/domain"
	"github.com/pridecrm/app-backend/internal/repository"
	"github.com/pridecrm/app-backend/internal/usecase"
)

type Handlers struct {
	Log              *slog.Logger
	UC               *usecase.Service
	TelegramBotToken string
	FrontendURL      string
	Repo             struct {
		Users        repository.UserRepository
		Games        repository.GameRepository
		Participants repository.ParticipantRepository
		Tickets      repository.SupportTicketRepository
		Tournaments  repository.TournamentRepository
	}
}

func (h *Handlers) Health(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func derefStrPtr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

type userCreateBody struct {
	UserID    string  `json:"user_id"`
	Username  string  `json:"username"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Phone     *string `json:"phone_number"`
	Email     *string `json:"email"`
}

type userPatch struct {
	Username  *string `json:"username"`
	NickName  *string `json:"nick_name"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Phone     *string `json:"phone_number"`
	Email     *string `json:"email"`
	DOB       *string `json:"date_of_birth"`
}

func tournamentHistoryToMap(h *domain.TournamentHistory) map[string]any {
	if h == nil {
		return nil
	}
	parts := make([]map[string]any, 0, len(h.Participants))
	for _, p := range h.Participants {
		parts = append(parts, map[string]any{
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
			"position":               p.Position,
			"final_points":           p.FinalPoints,
		})
	}
	m := map[string]any{
		"id":                 h.ID,
		"game":               h.GameID,
		"date":               h.Date.Format("2006-01-02"),
		"tournament_name":    h.TournamentName,
		"location":           h.Location,
		"buyin":              h.Buyin,
		"reentry_buyin":      h.ReentryBuyin,
		"total_revenue":      h.TotalRevenue,
		"participants_count": h.ParticipantsCount,
		"completed_at":       h.CompletedAt.UTC().Format(time.RFC3339),
		"participants":       parts,
	}
	if h.Time != nil {
		m["time"] = h.Time.Format("15:04:05")
	} else {
		m["time"] = nil
	}
	return m
}

func tournamentHistoryListMap(h *domain.TournamentHistory) map[string]any {
	m := map[string]any{
		"id":                 h.ID,
		"date":               h.Date.Format("2006-01-02"),
		"tournament_name":    h.TournamentName,
		"location":           h.Location,
		"participants_count": h.ParticipantsCount,
		"total_revenue":      h.TotalRevenue,
		"completed_at":       h.CompletedAt.UTC().Format(time.RFC3339),
	}
	if h.Time != nil {
		m["time"] = h.Time.Format("15:04:05")
	}
	return m
}
