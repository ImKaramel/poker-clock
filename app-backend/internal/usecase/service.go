package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
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
var ErrInvalidTelegramAuth = errors.New("invalid telegram auth data")
var ErrContactAuthPending = errors.New("contact auth pending")
var ErrContactAuthExpired = errors.New("contact auth expired")
var ErrContactAuthConsumed = errors.New("contact auth consumed")

const telegramInitDataMaxAge = 24 * time.Hour
const contactAuthTTL = 10 * time.Minute

type Service struct {
	Users            repository.UserRepository
	Games            repository.GameRepository
	Participants     repository.ParticipantRepository
	Tickets          repository.SupportTicketRepository
	Tournaments      repository.TournamentRepository
	ContactAuth      repository.ContactAuthRepository
	JWT              *auth.JWTService
	Log              *slog.Logger
	Clock            *services.Clock
	Storage          Storage
	AdminTelegramIDs map[string]bool
}

type Storage interface {
	UploadAvatar(ctx context.Context, userID string, data []byte) (string, error)
	UploadTournamentPhoto(ctx context.Context, data []byte) (string, error)
}

type telegramInitDataUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	PhotoURL  string `json:"photo_url"`
}

func validateTelegramInitData(rawInitData string, botToken string, now time.Time) (*telegramInitDataUser, error) {
	values, err := url.ParseQuery(rawInitData)
	if err != nil {
		return nil, fmt.Errorf("%w: parse init data: %v", ErrInvalidTelegramAuth, err)
	}

	telegramHash := values.Get("hash")
	if telegramHash == "" {
		return nil, fmt.Errorf("%w: hash is empty", ErrInvalidTelegramAuth)
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		if key == "hash" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	checkParts := make([]string, 0, len(keys))
	for _, key := range keys {
		checkParts = append(checkParts, fmt.Sprintf("%s=%s", key, values.Get(key)))
	}
	dataCheckString := strings.Join(checkParts, "\n")

	secretHMAC := hmac.New(sha256.New, []byte("WebAppData"))
	secretHMAC.Write([]byte(botToken))
	secretKey := secretHMAC.Sum(nil)

	hashHMAC := hmac.New(sha256.New, secretKey)
	hashHMAC.Write([]byte(dataCheckString))
	calculatedHash := hex.EncodeToString(hashHMAC.Sum(nil))

	if !hmac.Equal([]byte(calculatedHash), []byte(telegramHash)) {
		return nil, fmt.Errorf("%w: invalid hash", ErrInvalidTelegramAuth)
	}

	authDateRaw := values.Get("auth_date")
	authDateUnix, err := strconv.ParseInt(authDateRaw, 10, 64)
	if err != nil || authDateUnix <= 0 {
		return nil, fmt.Errorf("%w: invalid auth_date", ErrInvalidTelegramAuth)
	}
	authDate := time.Unix(authDateUnix, 0)
	if now.Sub(authDate) > telegramInitDataMaxAge {
		return nil, fmt.Errorf("%w: init data expired", ErrInvalidTelegramAuth)
	}
	if authDate.After(now.Add(5 * time.Minute)) {
		return nil, fmt.Errorf("%w: init data from future", ErrInvalidTelegramAuth)
	}

	userRaw := values.Get("user")
	if userRaw == "" {
		return nil, fmt.Errorf("%w: user is empty", ErrInvalidTelegramAuth)
	}

	var user telegramInitDataUser
	if err := json.Unmarshal([]byte(userRaw), &user); err != nil {
		return nil, fmt.Errorf("%w: parse user: %v", ErrInvalidTelegramAuth, err)
	}
	if user.ID <= 0 {
		return nil, fmt.Errorf("%w: user id is empty", ErrInvalidTelegramAuth)
	}

	return &user, nil
}

func (s *Service) issueToken(u *domain.User) (string, error) {
	isAdmin := s.AdminTelegramIDs[u.UserID]

	s.Log.Info("ISSUING TOKEN",
		"user_id", u.UserID,
		"is_admin", isAdmin,
	)

	return s.JWT.Issue(u.UserID, isAdmin)
}

func generateContactAuthToken() (string, error) {
	var b [18]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b[:]), nil
}

func (s *Service) CreateContactAuthChallenge(ctx context.Context) (*domain.ContactAuthChallenge, error) {
	if s.ContactAuth == nil {
		return nil, errors.New("contact auth repository is not configured")
	}

	token, err := generateContactAuthToken()
	if err != nil {
		return nil, err
	}

	challenge := &domain.ContactAuthChallenge{
		Token:     token,
		Status:    "pending",
		ExpiresAt: time.Now().Add(contactAuthTTL),
	}
	if err := s.ContactAuth.Create(ctx, challenge); err != nil {
		return nil, err
	}
	return challenge, nil
}

func (s *Service) ConfirmContactAuthChallenge(
	ctx context.Context,
	token string,
	telegramUserID string,
	phoneNumber string,
) error {
	if s.ContactAuth == nil {
		return errors.New("contact auth repository is not configured")
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return ErrNotFound
	}
	return s.ContactAuth.Complete(ctx, token, telegramUserID, phoneNumber)
}

func (s *Service) ConsumeContactAuthChallenge(ctx context.Context, token string) (string, *domain.User, error) {
	if s.ContactAuth == nil {
		return "", nil, errors.New("contact auth repository is not configured")
	}

	challenge, err := s.ContactAuth.GetByToken(ctx, strings.TrimSpace(token))
	if err != nil {
		return "", nil, err
	}
	if challenge == nil {
		return "", nil, ErrNotFound
	}
	if challenge.ConsumedAt != nil || challenge.Status == "consumed" {
		return "", nil, ErrContactAuthConsumed
	}
	if time.Now().After(challenge.ExpiresAt) {
		return "", nil, ErrContactAuthExpired
	}
	if challenge.Status != "completed" || challenge.TelegramUserID == nil || *challenge.TelegramUserID == "" {
		return "", nil, ErrContactAuthPending
	}

	u, err := s.Users.GetByID(ctx, *challenge.TelegramUserID)
	if err != nil {
		return "", nil, err
	}
	if u == nil {
		return "", nil, ErrNotFound
	}

	tokenJWT, err := s.issueToken(u)
	if err != nil {
		return "", nil, err
	}
	if err := s.ContactAuth.Consume(ctx, challenge.Token); err != nil {
		return "", nil, err
	}
	return tokenJWT, u, nil
}

func (s *Service) TelegramAuthInitData(
	ctx context.Context,
	initData string,
	botToken string,
) (token string, dbUser *domain.User, isNew bool, err error) {
	user, err := validateTelegramInitData(initData, botToken, time.Now())
	if err != nil {
		s.Log.Error("TELEGRAM INIT DATA VALIDATION FAILED", "err", err)
		return "", nil, false, err
	}

	return s.TelegramAuthUnsafe(ctx, map[string]any{
		"id":         strconv.FormatInt(user.ID, 10),
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"photo_url":  user.PhotoURL,
	})
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
