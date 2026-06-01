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
var ErrInvalidTelegramAuth = errors.New("invalid telegram auth data")
var ErrTournamentResultsRequired = errors.New("tournament results required")
var ErrTournamentResultsUnresolved = errors.New("tournament results unresolved")
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
	var b [32]byte
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

type CompleteResultInput struct {
	Position  int    `json:"position"`
	Nickname  string `json:"nickname"`
	UserID    string `json:"user_id"`
	KOCount   int    `json:"ko_count"`
	Knockouts int    `json:"knockouts"`
	Bonus     int    `json:"bonus"`
}

type CompleteGameInput struct {
	Participants []CompleteParticipantInput `json:"participants"`
	Results      []CompleteResultInput      `json:"results"`
}

type TournamentResultUser struct {
	UserID    string  `json:"user_id"`
	Username  string  `json:"username"`
	NickName  *string `json:"nick_name,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
}

type TournamentResultPreviewRow struct {
	Position    int                    `json:"position"`
	Nickname    string                 `json:"nickname"`
	KOCount     int                    `json:"ko_count"`
	BasePoints  int                    `json:"base_points"`
	KOPoints    int                    `json:"ko_points"`
	Bonus       int                    `json:"bonus"`
	TotalPoints int                    `json:"total_points"`
	Status      string                 `json:"status"`
	User        *TournamentResultUser  `json:"user,omitempty"`
	Candidates  []TournamentResultUser `json:"candidates,omitempty"`
}

type CompleteGamePreview struct {
	PlayersCount int                          `json:"players_count"`
	KOValue      int                          `json:"ko_value"`
	Results      []TournamentResultPreviewRow `json:"results"`
	Unresolved   []TournamentResultPreviewRow `json:"unresolved"`
}

type TournamentResultConflictError struct {
	Preview *CompleteGamePreview
}

func (e *TournamentResultConflictError) Error() string {
	return ErrTournamentResultsUnresolved.Error()
}

const koRatingBonus = 100

var baselineRatingPoints = []float64{
	3100.00,
	2738.43,
	2233.03,
	2108.00,
	1983.11,
	1612.00,
	1378.38,
	1351.35,
	1240.00,
	1219.51,
	1161.17,
	1116.00,
	1013.51,
	992.00,
	927.89,
	878.05,
	868.00,
	829.27,
	829.27,
	784.62,
	744.00,
	702.70,
	664.54,
	634.15,
	634.15,
	620.00,
	527.03,
	496.00,
	487.80,
	487.80,
	486.49,
	461.54,
	439.02,
	439.02,
	405.41,
	405.41,
	390.24,
	390.24,
	372.00,
	364.86,
	124.00,
	124.00,
	124.00,
	124.00,
	124.00,
	124.00,
	124.00,
	124.00,
	124.00,
	124.00,
	124.00,
	124.00,
	124.00,
	124.00,
	124.00,
}

func scaledPlacePoints(playersCount int, position int) int {
	if playersCount <= 0 || position <= 0 {
		return 0
	}
	baseIndex := position - 1
	base := baselineRatingPoints[len(baselineRatingPoints)-1]
	if baseIndex < len(baselineRatingPoints) {
		base = baselineRatingPoints[baseIndex]
	}
	return int(math.Round(base * float64(playersCount) / 55.0))
}

func normalizeTournamentName(value string) string {
	value = strings.TrimSpace(strings.TrimPrefix(value, "@"))
	value = strings.ToLower(value)
	return strings.Join(strings.Fields(value), " ")
}

func tournamentResultUserFromDomain(u *domain.User) TournamentResultUser {
	return TournamentResultUser{
		UserID:    u.UserID,
		Username:  u.Username,
		NickName:  u.NickName,
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}
}

func userTournamentNames(u domain.User) []string {
	names := []string{u.UserID, u.Username}
	if u.NickName != nil {
		names = append(names, *u.NickName)
	}
	if u.FirstName != nil {
		names = append(names, *u.FirstName)
	}
	if u.FirstName != nil && u.LastName != nil {
		names = append(names, strings.TrimSpace(*u.FirstName+" "+*u.LastName))
	}
	return names
}

func (s *Service) resolveTournamentResultUser(
	users []domain.User,
	row CompleteResultInput,
) (*TournamentResultUser, []TournamentResultUser) {
	if row.UserID != "" {
		for i := range users {
			if users[i].UserID == row.UserID {
				user := tournamentResultUserFromDomain(&users[i])
				return &user, nil
			}
		}
		return nil, nil
	}

	target := normalizeTournamentName(row.Nickname)
	if target == "" {
		return nil, nil
	}

	matches := make([]TournamentResultUser, 0, 2)
	for i := range users {
		for _, name := range userTournamentNames(users[i]) {
			if normalizeTournamentName(name) == target {
				matches = append(matches, tournamentResultUserFromDomain(&users[i]))
				break
			}
		}
	}

	if len(matches) == 1 {
		return &matches[0], nil
	}
	if len(matches) > 1 {
		return nil, matches
	}
	return nil, nil
}

func (s *Service) PreviewCompleteGame(
	ctx context.Context,
	gameID int64,
	input CompleteGameInput,
) (*CompleteGamePreview, error) {
	g, err := s.Games.GetByID(ctx, gameID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, ErrNotFound
	}
	return s.previewCompleteGame(ctx, input)
}

func (s *Service) previewCompleteGame(ctx context.Context, input CompleteGameInput) (*CompleteGamePreview, error) {
	if len(input.Results) == 0 {
		return nil, ErrTournamentResultsRequired
	}

	users, err := s.Users.List(ctx)
	if err != nil {
		return nil, err
	}

	playersCount := len(input.Results)
	preview := &CompleteGamePreview{
		PlayersCount: playersCount,
		KOValue:      koRatingBonus,
		Results:      make([]TournamentResultPreviewRow, 0, playersCount),
		Unresolved:   []TournamentResultPreviewRow{},
	}
	seenUsers := make(map[string]bool)

	for index, raw := range input.Results {
		position := raw.Position
		if position <= 0 {
			position = index + 1
		}
		koCount := raw.KOCount
		if koCount == 0 && raw.Knockouts > 0 {
			koCount = raw.Knockouts
		}
		if koCount < 0 {
			koCount = 0
		}
		basePoints := scaledPlacePoints(playersCount, position)
		row := TournamentResultPreviewRow{
			Position:    position,
			Nickname:    strings.TrimSpace(raw.Nickname),
			KOCount:     koCount,
			BasePoints:  basePoints,
			KOPoints:    koCount * koRatingBonus,
			Bonus:       raw.Bonus,
			TotalPoints: basePoints + koCount*koRatingBonus + raw.Bonus,
			Status:      "unresolved",
		}

		user, candidates := s.resolveTournamentResultUser(users, raw)
		switch {
		case user != nil && seenUsers[user.UserID]:
			row.User = user
			row.Status = "duplicate"
		case user != nil:
			row.User = user
			row.Status = "resolved"
			seenUsers[user.UserID] = true
		case len(candidates) > 0:
			row.Candidates = candidates
			row.Status = "ambiguous"
		}

		if row.Status != "resolved" {
			preview.Unresolved = append(preview.Unresolved, row)
		}
		preview.Results = append(preview.Results, row)
	}

	return preview, nil
}

func reentryPrice(g *domain.Game) float64 {
	if g.ReentryBuyin > 0 {
		return g.ReentryBuyin
	}
	return g.Buyin
}

func (s *Service) CompleteGame(ctx context.Context, gameID int64, input CompleteGameInput) (*domain.TournamentHistory, error) {
	g, err := s.Games.GetByID(ctx, gameID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, ErrNotFound
	}
	if len(input.Results) == 0 {
		return nil, ErrTournamentResultsRequired
	}

	preview, err := s.previewCompleteGame(ctx, input)
	if err != nil {
		return nil, err
	}
	if len(preview.Unresolved) > 0 {
		return nil, &TournamentResultConflictError{Preview: preview}
	}

	liveParts, err := s.Participants.ListByGame(ctx, gameID)
	if err != nil {
		return nil, err
	}
	completedByUser := make(map[string]CompleteParticipantInput)
	liveByUser := make(map[string]domain.Participant)
	for _, p := range liveParts {
		if !p.Arrived {
			continue
		}
		liveByUser[p.UserID] = p
		completedByUser[p.UserID] = CompleteParticipantInput{
			UserID:  p.UserID,
			Entries: p.Entries,
			Rebuys:  p.Rebuys,
			Addons:  p.Addons,
		}
	}
	for _, p := range input.Participants {
		existing := completedByUser[p.UserID]
		existing.UserID = p.UserID
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
		ParticipantsCount: len(preview.Results),
	}
	if err := s.Tournaments.CreateHistory(ctx, h); err != nil {
		return nil, err
	}

	totalRev := 0
	for _, result := range preview.Results {
		if result.User == nil {
			return nil, &TournamentResultConflictError{Preview: preview}
		}
		p := completedByUser[result.User.UserID]
		if p.UserID == "" {
			p.UserID = result.User.UserID
			p.Entries = 1
		}
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
			Position:            &result.Position,
			FinalPoints:         result.TotalPoints,
		}
		if live, ok := liveByUser[p.UserID]; ok {
			if tp.Position == nil {
				tp.Position = live.Position
			}
		}
		if err := s.Tournaments.AddTournamentParticipant(ctx, tp); err != nil {
			return nil, err
		}
		u.Points += result.TotalPoints
		u.TotalGamesPlayed++
		if err := s.Users.Update(ctx, u); err != nil {
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
