package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/pridecrm/app-backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAuthInput   = errors.New("invalid auth input")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrPasswordLinked     = errors.New("password already linked")
	ErrAccountMismatch    = errors.New("account mismatch")
	ErrVerificationCode   = errors.New("invalid verification code")
)

var telegramUsernameRE = regexp.MustCompile(`^[a-z0-9_]{5,32}$`)
var passwordRegistrationCodes = newPasswordRegistrationCodeStore(10 * time.Minute)

type PasswordRegistrationChallenge struct {
	TelegramUserID string
	Code           string
	Username       string
}

type PasswordRegistrationResult struct {
	Token                 string
	User                  *domain.User
	RequiresVerification  bool
	VerificationChallenge *PasswordRegistrationChallenge
}

type pendingPasswordRegistration struct {
	UserID       string
	Username     string
	Nickname     string
	PasswordHash string
	ExpiresAt    time.Time
	Code         string
}

type passwordRegistrationCodeStore struct {
	mu      sync.Mutex
	ttl     time.Duration
	pending map[string]pendingPasswordRegistration
}

func newPasswordRegistrationCodeStore(ttl time.Duration) *passwordRegistrationCodeStore {
	return &passwordRegistrationCodeStore{
		ttl:     ttl,
		pending: make(map[string]pendingPasswordRegistration),
	}
}

func (s *passwordRegistrationCodeStore) issue(userID, username, nickname, passwordHash string) (pendingPasswordRegistration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	code, err := generateVerificationCode()
	if err != nil {
		return pendingPasswordRegistration{}, err
	}

	entry := pendingPasswordRegistration{
		UserID:       userID,
		Username:     username,
		Nickname:     nickname,
		PasswordHash: passwordHash,
		ExpiresAt:    time.Now().Add(s.ttl),
		Code:         code,
	}
	s.pending[username] = entry
	return entry, nil
}

func (s *passwordRegistrationCodeStore) consume(username, code string) (pendingPasswordRegistration, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.pending[username]
	if !ok {
		return pendingPasswordRegistration{}, false
	}
	if time.Now().After(entry.ExpiresAt) || strings.TrimSpace(code) != entry.Code {
		if time.Now().After(entry.ExpiresAt) {
			delete(s.pending, username)
		}
		return pendingPasswordRegistration{}, false
	}

	delete(s.pending, username)
	return entry, true
}

func generateVerificationCode() (string, error) {
	var digits strings.Builder
	for i := 0; i < 6; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		digits.WriteByte(byte('0' + n.Int64()))
	}
	return digits.String(), nil
}

func normalizeFallbackUsername(username string) string {
	return strings.TrimPrefix(strings.ToLower(strings.TrimSpace(username)), "@")
}

func validateFallbackUsername(username string) bool {
	return telegramUsernameRE.MatchString(username)
}

func validateNickname(nickname string) bool {
	l := len([]rune(strings.TrimSpace(nickname)))
	return l >= 2 && l <= 24
}

func validatePassword(password string) bool {
	if len([]rune(password)) < 8 {
		return false
	}
	hasLetter := false
	hasDigit := false
	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	return hasLetter && hasDigit
}

func hashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func checkPassword(hash string, password string) bool {
	if hash == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func fallbackUserID(username string) (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "fallback_" + username + "_" + hex.EncodeToString(b[:]), nil
}

func (s *Service) RegisterPasswordUser(ctx context.Context, username string, nickname string, password string, authenticatedUserID string) (*PasswordRegistrationResult, error) {
	username = normalizeFallbackUsername(username)
	nickname = strings.TrimSpace(nickname)

	if !validateFallbackUsername(username) || !validateNickname(nickname) || !validatePassword(password) {
		return nil, ErrInvalidAuthInput
	}

	existing, err := s.Users.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if authenticatedUserID != "" {
			if existing.UserID != authenticatedUserID {
				return nil, ErrAccountMismatch
			}
			if existing.Password != "" {
				return nil, ErrPasswordLinked
			}

			passwordHash, err := hashPassword(password)
			if err != nil {
				return nil, err
			}
			existing.Password = passwordHash
			existing.NickName = &nickname
			if err := s.Users.Update(ctx, existing); err != nil {
				return nil, err
			}

			token, err := s.issueToken(existing)
			if err != nil {
				return nil, err
			}
			return &PasswordRegistrationResult{
				Token: token,
				User:  existing,
			}, nil
		}
		if existing.Password != "" {
			return nil, ErrPasswordLinked
		}

		passwordHash, err := hashPassword(password)
		if err != nil {
			return nil, err
		}
		pending, err := passwordRegistrationCodes.issue(existing.UserID, username, nickname, passwordHash)
		if err != nil {
			return nil, err
		}
		return &PasswordRegistrationResult{
			RequiresVerification: true,
			VerificationChallenge: &PasswordRegistrationChallenge{
				TelegramUserID: pending.UserID,
				Code:           pending.Code,
				Username:       pending.Username,
			},
		}, nil
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}
	userID, err := fallbackUserID(username)
	if err != nil {
		return nil, err
	}

	u := &domain.User{
		UserID:   userID,
		Username: username,
		NickName: &nickname,
		Password: passwordHash,
		IsActive: true,
	}
	if err := s.Users.Create(ctx, u); err != nil {
		return nil, err
	}

	token, err := s.issueToken(u)
	if err != nil {
		return nil, err
	}
	return &PasswordRegistrationResult{
		Token: token,
		User:  u,
	}, nil
}

func (s *Service) VerifyPasswordRegistrationCode(ctx context.Context, username string, code string) (string, *domain.User, error) {
	username = normalizeFallbackUsername(username)
	entry, ok := passwordRegistrationCodes.consume(username, code)
	if !ok {
		return "", nil, ErrVerificationCode
	}

	u, err := s.Users.GetByID(ctx, entry.UserID)
	if err != nil {
		return "", nil, err
	}
	if u == nil {
		return "", nil, ErrNotFound
	}
	if u.Password != "" {
		return "", nil, ErrPasswordLinked
	}

	u.Password = entry.PasswordHash
	u.NickName = &entry.Nickname
	if err := s.Users.Update(ctx, u); err != nil {
		return "", nil, err
	}

	token, err := s.issueToken(u)
	if err != nil {
		return "", nil, err
	}
	return token, u, nil
}

func (s *Service) LoginPasswordUser(ctx context.Context, username string, password string) (string, *domain.User, error) {
	username = normalizeFallbackUsername(username)
	if !validateFallbackUsername(username) || password == "" {
		return "", nil, ErrInvalidCredentials
	}

	u, err := s.Users.GetByUsername(ctx, username)
	if err != nil {
		return "", nil, err
	}
	if u == nil || u.IsBanned || !checkPassword(u.Password, password) {
		return "", nil, ErrInvalidCredentials
	}

	token, err := s.issueToken(u)
	if err != nil {
		return "", nil, err
	}
	return token, u, nil
}

func (s *Service) LinkPassword(ctx context.Context, userID string, password string) (string, *domain.User, error) {
	if !validatePassword(password) {
		return "", nil, ErrInvalidAuthInput
	}

	u, err := s.Users.GetByID(ctx, userID)
	if err != nil {
		return "", nil, err
	}
	if u == nil {
		return "", nil, ErrNotFound
	}
	if u.Password != "" {
		return "", nil, ErrPasswordLinked
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		return "", nil, err
	}
	u.Password = passwordHash
	if err := s.Users.Update(ctx, u); err != nil {
		return "", nil, err
	}

	token, err := s.issueToken(u)
	if err != nil {
		return "", nil, err
	}
	return token, u, nil
}
