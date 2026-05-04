package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pridecrm/app-backend/internal/domain"
	"github.com/pridecrm/app-backend/internal/infrastructure/auth"
	"github.com/pridecrm/app-backend/internal/repository"
	"github.com/pridecrm/app-backend/internal/services"
)

var ErrNotFound = errors.New("not found")
var ErrForbidden = errors.New("forbidden")

type Service struct {
	Users            repository.UserRepository
	Games            repository.GameRepository
	Participants     repository.ParticipantRepository
	Tickets          repository.SupportTicketRepository
	Tournaments      repository.TournamentRepository
	JWT              *auth.JWTService
	Log              *slog.Logger
	Clock            *services.Clock
	Storage          Storage
	AdminTelegramIDs map[string]bool
}

type Storage interface {
	UploadAvatar(ctx context.Context, userID string, data []byte) (string, error)
}

func (s *Service) issueToken(u *domain.User) (string, error) {
	isAdmin := s.AdminTelegramIDs[u.UserID]

	s.Log.Info("ISSUING TOKEN",
		"user_id", u.UserID,
		"is_admin", isAdmin,
	)

	return s.JWT.Issue(u.UserID, isAdmin)
}

func (s *Service) TelegramAuthUnsafe(
	ctx context.Context,
	user map[string]any,
) (token string, dbUser *domain.User, isNew bool, err error) {
	isNew = false
	idVal := user["id"]
	username, _ := user["username"].(string)
	firstName, _ := user["first_name"].(string)
	lastName, _ := user["last_name"].(string)
	photoURL, _ := user["photo_url"].(string)

	s.Log.Info("🚀 TELEGRAM AUTH START",
		"id_raw", idVal,
		"username", username,
	)

	userID, err := normalizeTelegramID(idVal)
	if err != nil {
		return "", nil, false, err
	}

	isAdmin := s.AdminTelegramIDs[userID]

	s.Log.Info("TELEGRAM AUTH USER IDENTIFIED",
		"user_id", userID,
		"username", username,
		"is_admin", isAdmin,
	)

	if username == "" {
		username = userID
		s.Log.Info("⚠️ Username empty → using userID as username", "user_id", userID)
	}

	existing, err := s.Users.GetByID(ctx, userID)
	if err != nil {
		s.Log.Error("❌ DB GetByID FAILED", "user_id", userID, "err", err)
		return "", nil, false, err
	}

	if existing == nil {
		s.Log.Info("🆕 USER NOT FOUND → CREATING NEW", "user_id", userID)
		isNew = true
		newUser := &domain.User{
			UserID:    userID,
			Username:  username,
			FirstName: strPtr(firstName),
			LastName:  strPtr(lastName),
			IsActive:  true,
		}

		if photoURL != "" {
			newUser.PhotoURL = strPtr(photoURL)
		}

		if err := s.Users.Create(ctx, newUser); err != nil {
			s.Log.Error("❌ CREATE USER FAILED", "user_id", userID, "err", err)
			return "", nil, isNew, err
		}

		token, err = s.issueToken(newUser)
		if err != nil {
			s.Log.Error("❌ ISSUE TOKEN FAILED (create)", "user_id", userID, "err", err)
			return "", nil, isNew, err
		}

		s.Log.Info("✅ NEW USER CREATED SUCCESSFULLY",
			"user_id", userID,
			"has_photo", photoURL != "",
		)

		return token, newUser, isNew, err
	}

	s.Log.Info("USER FOUND → UPDATING", "user_id", userID)

	oldPhoto := derefStrPtr(existing.PhotoURL)

	existing.Username = username
	existing.FirstName = strPtr(firstName)
	existing.LastName = strPtr(lastName)

	if photoURL != "" {
		existing.PhotoURL = strPtr(photoURL)
		s.Log.Info(" PHOTO UPDATED", "user_id", userID, "new_photo_url", photoURL)
	} else {
		s.Log.Info("No new photo_url → keeping existing photo",
			"user_id", userID,
			"kept_photo", oldPhoto,
		)
	}

	if err := s.Users.Update(ctx, existing); err != nil {
		s.Log.Error("UPDATE USER FAILED", "user_id", userID, "err", err)
		return "", nil, isNew, err
	}

	token, err = s.issueToken(existing)
	if err != nil {
		s.Log.Error("ISSUE TOKEN FAILED (update)", "user_id", userID, "err", err)
		return "", nil, isNew, err
	}

	return token, existing, isNew, nil
}
func derefStrPtr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

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

	liveParts, err := s.Participants.ListByGame(ctx, gameID)
	if err != nil {
		return nil, err
	}
	completedByUser := make(map[string]CompleteParticipantInput)
	for _, p := range liveParts {
		if !p.Arrived {
			continue
		}
		completedByUser[p.UserID] = CompleteParticipantInput{
			UserID:  p.UserID,
			Entries: p.Entries,
			Rebuys:  p.Rebuys,
			Addons:  p.Addons,
		}
	}
	for _, p := range parts {
		existing, ok := completedByUser[p.UserID]
		if !ok {
			continue
		}
		if p.Entries > 0 {
			existing.Entries = p.Entries
		}
		if p.Rebuys > 0 {
			existing.Rebuys = p.Rebuys
		}
		if p.Addons > 0 {
			existing.Addons = p.Addons
		}
		existing.PaymentMethod = p.PaymentMethod
		completedByUser[p.UserID] = existing
	}
	completedParts := make([]CompleteParticipantInput, 0, len(completedByUser))
	for _, p := range completedByUser {
		completedParts = append(completedParts, p)
	}
	sort.Slice(completedParts, func(i, j int) bool {
		return completedParts[i].UserID < completedParts[j].UserID
	})

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
		ParticipantsCount: len(completedParts),
	}
	if err := s.Tournaments.CreateHistory(ctx, h); err != nil {
		return nil, err
	}

	totalRev := 0
	for _, p := range completedParts {
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
	p, err := s.Participants.GetByID(ctx, participantID)
	if err != nil {
		s.Log.Error("SetParticipantArrived: failed to get participant",
			"participant_id", participantID,
			"err", err)
		return err
	}
	if p == nil {
		s.Log.Error("SetParticipantArrived: participant not found",
			"participant_id", participantID)
		return ErrNotFound
	}

	// Update arrived status
	if err := s.Participants.SetArrived(ctx, participantID, arrived); err != nil {
		s.Log.Error("SetParticipantArrived: failed to update arrived status",
			"participant_id", participantID,
			"arrived", arrived,
			"err", err)
		return err
	}

	s.Log.Info("SetParticipantArrived: successfully updated arrived status",
		"participant_id", participantID,
		"user_id", p.UserID,
		"game_id", p.GameID,
		"arrived", arrived)

	// Update game stats
	//players, chips, err := s.calculateStats(ctx, p.GameID)
	//if err != nil {
	//	s.Log.Error("SetParticipantArrived: failed to calculate stats",
	//		"game_id", p.GameID,
	//		"err", err)
	//	// Don't return error here, the main operation succeeded
	//} else if s.Clock != nil {
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

func normalizeTelegramID(v any) (string, error) {
	switch id := v.(type) {
	case float64:
		return strconv.FormatInt(int64(id), 10), nil
	case int64:
		return strconv.FormatInt(id, 10), nil
	case int:
		return strconv.Itoa(id), nil
	case string:
		if f, err := strconv.ParseFloat(id, 64); err == nil {
			return strconv.FormatInt(int64(f), 10), nil
		}
		return id, nil
	default:
		return "", fmt.Errorf("unsupported telegram id type")
	}
}

// validateTelegramWebAuthHash проверяет hash от Telegram Login Widget
func (s *Service) validateTelegramWebAuthHash(queryParams url.Values, botToken string) error {
	// Получаем hash из параметров
	hash := queryParams.Get("hash")
	if hash == "" {
		return fmt.Errorf("hash not found")
	}

	// Создаем копию параметров без hash
	authData := make(url.Values)
	for key, values := range queryParams {
		if key != "hash" {
			authData[key] = values
		}
	}

	// Собираем data_check_string в алфавитном порядке
	var keys []string
	for key := range authData {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var dataCheckStrings []string
	for _, key := range keys {
		dataCheckStrings = append(dataCheckStrings, fmt.Sprintf("%s=%s", key, authData.Get(key)))
	}
	dataCheckString := strings.Join(dataCheckStrings, "\n")

	// Создаем secret_key = SHA256(bot_token)
	secretKey := sha256.Sum256([]byte(botToken))

	// Создаем HMAC-SHA256(data_check_string, secret_key)
	h := hmac.New(sha256.New, secretKey[:])
	h.Write([]byte(dataCheckString))
	calculatedHash := hex.EncodeToString(h.Sum(nil))

	// Сравниваем hash
	if !hmac.Equal([]byte(calculatedHash), []byte(hash)) {
		return fmt.Errorf("invalid hash")
	}

	return nil
}

func (s *Service) TelegramWebAuth(
	ctx context.Context,
	queryParams url.Values,
	botToken string,
) (token string, dbUser *domain.User, isNew bool, err error) {
	if err := s.validateTelegramWebAuthHash(queryParams, botToken); err != nil {
		s.Log.Error("❌ TELEGRAM WEB AUTH HASH VALIDATION FAILED",
			"err", err,
			"query_params", sanitizeQueryParams(queryParams),
		)
		return "", nil, false, fmt.Errorf("invalid telegram auth data: %w", err)
	}

	s.Log.Info("✅ TELEGRAM WEB AUTH HASH VALIDATION SUCCESS")

	user := make(map[string]any)
	user["id"] = queryParams.Get("id")
	user["username"] = queryParams.Get("username")
	user["first_name"] = queryParams.Get("first_name")
	user["last_name"] = queryParams.Get("last_name")
	user["photo_url"] = queryParams.Get("photo_url")

	s.Log.Info("🔐 TELEGRAM WEB AUTH USER DATA EXTRACTED",
		"id", user["id"],
		"username", user["username"],
		"first_name", user["first_name"],
	)

	return s.TelegramAuthUnsafe(ctx, user)
}

func sanitizeQueryParams(params url.Values) map[string]string {
	sanitized := make(map[string]string)
	for key, values := range params {
		if key == "hash" {
			sanitized[key] = "***"
		} else {
			sanitized[key] = strings.Join(values, ",")
		}
	}
	return sanitized
}
