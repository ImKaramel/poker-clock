package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/url"
	"time"

	"github.com/pridecrm/app-backend/internal/domain"
	"github.com/pridecrm/app-backend/internal/infrastructure/auth"
	"github.com/pridecrm/app-backend/internal/repository"
	"github.com/pridecrm/app-backend/internal/services"
)

var ErrNotFound = errors.New("not found")
var ErrForbidden = errors.New("forbidden")

type Service struct {
	Users        repository.UserRepository
	Games        repository.GameRepository
	Participants repository.ParticipantRepository
	Tickets      repository.SupportTicketRepository
	Tournaments  repository.TournamentRepository
	JWT          *auth.JWTService
	Log          *slog.Logger
	Clock        *services.Clock
}

func (s *Service) issueToken(u *domain.User) (string, error) {
	return s.JWT.Issue(u.UserID, u.IsAdmin)
}

func (s *Service) TelegramAuth(ctx context.Context, telegramData map[string]any) (token string, user *domain.User, isNew bool, err error) {
	idVal, ok := telegramData["id"]
	if !ok {
		return "", nil, false, fmt.Errorf("missing id")
	}
	userID := fmt.Sprint(idVal)
	username, _ := telegramData["username"].(string)
	firstName, _ := telegramData["first_name"].(string)
	lastName, _ := telegramData["last_name"].(string)
	if username == "" {
		username = userID
	}

	existing, err := s.Users.GetByID(ctx, userID)
	if err != nil {
		return "", nil, false, err
	}
	if existing == nil {
		u := &domain.User{
			UserID:    userID,
			Username:  username,
			FirstName: strPtr(firstName),
			LastName:  strPtr(lastName),
			IsActive:  true,
		}
		if err := s.Users.Create(ctx, u); err != nil {
			return "", nil, false, err
		}
		token, err := s.issueToken(u)
		return token, u, true, err
	}
	existing.FirstName = strPtr(firstName)
	existing.LastName = strPtr(lastName)
	if username != "" {
		existing.Username = username
	}
	if err := s.Users.Update(ctx, existing); err != nil {
		return "", nil, false, err
	}
	token, err = s.issueToken(existing)
	return token, existing, false, err
}

func (s *Service) TelegramValidateInitData(ctx context.Context, initData string) (token string, user *domain.User, err error) {
	vals, err := url.ParseQuery(initData)
	if err != nil {
		return "", nil, fmt.Errorf("parse initData: %w", err)
	}
	userJSON := vals.Get("user")
	if userJSON == "" {
		return "", nil, fmt.Errorf("missing user")
	}
	var ud struct {
		ID        json.RawMessage `json:"id"`
		Username  string          `json:"username"`
		FirstName string          `json:"first_name"`
		LastName  string          `json:"last_name"`
	}
	if err := json.Unmarshal([]byte(userJSON), &ud); err != nil {
		return "", nil, err
	}
	telegramID := string(ud.ID)
	if len(telegramID) > 0 && telegramID[0] == '"' {
		var sid string
		_ = json.Unmarshal(ud.ID, &sid)
		telegramID = sid
	} else {
		var nid int64
		_ = json.Unmarshal(ud.ID, &nid)
		telegramID = fmt.Sprint(nid)
	}
	if telegramID == "" {
		return "", nil, fmt.Errorf("missing telegram id")
	}
	username := ud.Username
	if username == "" {
		username = "user_" + telegramID
	}

	existing, err := s.Users.GetByID(ctx, telegramID)
	if err != nil {
		return "", nil, err
	}
	if existing == nil {
		u := &domain.User{
			UserID:    telegramID,
			Username:  username,
			FirstName: strPtr(ud.FirstName),
			LastName:  strPtr(ud.LastName),
			IsActive:  true,
		}
		if err := s.Users.Create(ctx, u); err != nil {
			return "", nil, err
		}
		token, err := s.issueToken(u)
		return token, u, err
	}
	token, err = s.issueToken(existing)
	return token, existing, err
}

// RegisterParticipant creates a registration row; returns already=true if row existed.
func (s *Service) RegisterParticipant(ctx context.Context, userID string, gameID int64) (already bool, err error) {
	u, err := s.Users.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}
	if u == nil {
		return false, ErrNotFound
	}
	if u.IsBanned {
		return false, ErrForbidden
	}
	g, err := s.Games.GetByID(ctx, gameID)
	if err != nil {
		return false, err
	}
	if g == nil || !g.IsActive {
		return false, ErrNotFound
	}
	existing, err := s.Participants.GetByUserAndGame(ctx, userID, gameID)
	if err != nil {
		return false, err
	}
	if existing != nil {
		return true, nil
	}
	p := &domain.Participant{UserID: userID, GameID: gameID, Entries: 1}
	if err := s.Participants.Create(ctx, p); err != nil {
		return false, err
	}
	return false, nil
}

func (s *Service) UnregisterParticipant(ctx context.Context, userID string, gameID int64) error {
	g, err := s.Games.GetByID(ctx, gameID)
	if err != nil {
		return err
	}
	if g == nil || !g.IsActive {
		return ErrNotFound
	}
	existing, err := s.Participants.GetByUserAndGame(ctx, userID, gameID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrNotFound
	}
	return s.Participants.DeleteByUserAndGame(ctx, userID, gameID)
}

type CompleteParticipantInput struct {
	UserID        string  `json:"user_id"`
	Entries       int     `json:"entries"`
	Rebuys        int     `json:"rebuys"`
	Addons        int     `json:"addons"`
	PaymentMethod *string `json:"payment_method"`
}

func reentryPrice(g *domain.Game) float64 {
	if g.ReentryBuyin > 0 {
		return g.ReentryBuyin
	}
	return g.Buyin
}

func (s *Service) CompleteGame(ctx context.Context, gameID int64, parts []CompleteParticipantInput) (*domain.TournamentHistory, error) {
	g, err := s.Games.GetByID(ctx, gameID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, ErrNotFound
	}
	re := reentryPrice(g)
	buyinI := int(math.Round(g.Buyin))
	reI := int(math.Round(re))
	var rePtr *int
	if g.ReentryBuyin > 0 {
		rePtr = &reI
	} else {
		rePtr = &buyinI
	}
	name := g.Description
	if name == "" {
		name = fmt.Sprintf("Турнир %s", g.Date.Format("2006-01-02"))
	}

	h := &domain.TournamentHistory{
		GameID:            gameID,
		Date:              g.Date,
		Time:              timePtr(g.Time),
		TournamentName:    name,
		Location:          g.Location,
		Buyin:             buyinI,
		ReentryBuyin:      rePtr,
		ParticipantsCount: len(parts),
	}
	if err := s.Tournaments.CreateHistory(ctx, h); err != nil {
		return nil, err
	}

	totalRev := 0
	for _, p := range parts {
		u, err := s.Users.GetByID(ctx, p.UserID)
		if err != nil {
			return nil, err
		}
		if u == nil {
			return nil, fmt.Errorf("user %s not found", p.UserID)
		}
		ent := p.Entries
		if ent <= 0 {
			ent = 1
		}
		spent := int(math.Round(float64(ent)*g.Buyin + float64(p.Rebuys)*re + float64(p.Addons)*re))
		totalRev += spent
		tp := &domain.TournamentParticipant{
			TournamentHistoryID: h.ID,
			UserID:              u.UserID,
			Username:            u.Username,
			FirstName:           derefStr(u.FirstName),
			LastName:            derefStr(u.LastName),
			Entries:             ent,
			Rebuys:              p.Rebuys,
			Addons:              p.Addons,
			TotalSpent:          spent,
			PaymentMethod:       p.PaymentMethod,
		}
		if err := s.Tournaments.AddTournamentParticipant(ctx, tp); err != nil {
			return nil, err
		}
	}
	h.TotalRevenue = totalRev
	if err := s.Tournaments.UpdateHistory(ctx, h); err != nil {
		return nil, err
	}

	now := time.Now()
	g.IsActive = false
	g.Completed = true
	g.CompletedAt = &now
	if err := s.Games.Update(ctx, g); err != nil {
		return nil, err
	}
	h.Participants = nil
	full, err := s.Tournaments.GetHistoryByID(ctx, h.ID)
	if s.Clock != nil {
		live, _ := s.Participants.ListByGame(ctx, gameID)
		s.Clock.SyncParticipants(fmt.Sprint(gameID), live)
	}
	return full, err
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func (s *Service) SetParticipantArrived(
	ctx context.Context,
	participantID int64,
	arrived bool,
) error {

	//p, err := s.Participants.GetByID(ctx, participantID)
	//if err != nil {
	//	return err
	//}
	//if p == nil {
	//	return ErrNotFound
	//}
	//
	//p.Arrived = arrived
	//
	//if err := s.Participants.Update(ctx, p); err != nil {
	//	return err
	//}
	//
	//
	//players, chips, err := s.calculateStats(ctx, p.GameID)
	//if err != nil {
	//	return err
	//}
	//
	//if s.Clock != nil {
	//	go s.Clock.UpdateStats(ctx, fmt.Sprint(p.GameID), players, chips)
	//}

	return nil
}

func (s *Service) calculateStats(
	ctx context.Context,
	gameID int64,
) (players int, chips int, err error) {

	participants, err := s.Participants.ListByGame(ctx, gameID)
	if err != nil {
		return 0, 0, err
	}

	for _, p := range participants {
		if !p.Arrived {
			continue
		}

		players++

		chips += (p.Entries + p.Rebuys + p.Addons) * 1000 //!!!1
	}

	return players, chips, nil
}
