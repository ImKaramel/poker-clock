package httpapi

import (
	"time"

	"github.com/pridecrm/app-backend/internal/domain"
)

func userToMap(u *domain.User) map[string]any {
	if u == nil {
		return nil
	}
	m := map[string]any{
		"user_id":            u.UserID,
		"username":           u.Username,
		"points":             u.Points,
		"total_games_played": u.TotalGamesPlayed,
		"is_admin":           u.IsAdmin,
		"is_banned":          u.IsBanned,
		"created_at":         u.CreatedAt.UTC().Format(time.RFC3339),
	}
	if u.NickName != nil {
		m["nick_name"] = *u.NickName
	}
	if u.FirstName != nil {
		m["first_name"] = *u.FirstName
	}
	if u.LastName != nil {
		m["last_name"] = *u.LastName
	}
	if u.PhoneNumber != nil {
		m["phone_number"] = *u.PhoneNumber
	}
	if u.Email != nil {
		m["email"] = *u.Email
	}
	if u.DateOfBirth != nil {
		m["date_of_birth"] = u.DateOfBirth.Format("2006-01-02")
	}
	return m
}

func gameToMap(g *domain.Game, participantsCount int64, details []map[string]any) map[string]any {
	if g == nil {
		return nil
	}
	m := map[string]any{
		"game_id":                      g.GameID,
		"date":                         g.Date.Format("2006-01-02"),
		"time":                         g.Time.Format("15:04:05"),
		"description":                  g.Description,
		"buyin":                        g.Buyin,
		"reentry_buyin":                g.ReentryBuyin,
		"location":                     g.Location,
		"is_active":                    g.IsActive,
		"participants_count":           participantsCount,
		"participants_details":         details,
		"base_points":                  g.BasePoints,
		"points_per_extra_player":      g.PointsPerExtraPlayer,
		"min_players_for_extra_points": g.MinPlayersForExtraPoints,
	}
	if g.Photo != nil {
		m["photo"] = *g.Photo
	} else {
		m["photo"] = nil
	}
	return m
}

func paymentMethodDisplay(pm *string) string {
	if pm == nil {
		return ""
	}
	switch *pm {
	case "cash_ivan":
		return "Наличные Иван"
	case "cash_petr":
		return "Наличные Петр"
	case "qr_code":
		return "QR код"
	case "card":
		return "Картой"
	default:
		return *pm
	}
}
