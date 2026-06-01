package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/pridecrm/app-backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAuthInput   = errors.New("invalid auth input")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrNicknameTaken      = errors.New("nickname already taken")
	ErrPasswordLinked     = errors.New("password already linked")
)

var telegramUsernameRE = regexp.MustCompile(`^[a-z0-9_]{5,32}$`)

type PasswordRegistrationResult struct {
	Token string
	User  *domain.User
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

func webUserID(username string) (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "web_" + username + "_" + hex.EncodeToString(b[:]), nil
}

func isPasswordAuthUserID(userID string) bool {
	return strings.HasPrefix(userID, "fallback_") || strings.HasPrefix(userID, "web_")
}

func sortUsersForUsernameResolution(users []domain.User) {
	sort.SliceStable(users, func(i, j int) bool {
		left := users[i]
		right := users[j]

		if isPasswordAuthUserID(left.UserID) != isPasswordAuthUserID(right.UserID) {
			return !isPasswordAuthUserID(left.UserID)
		}
		if left.IsBanned != right.IsBanned {
			return !left.IsBanned
		}
		if left.IsActive != right.IsActive {
			return left.IsActive
		}
		return left.CreatedAt.Before(right.CreatedAt)
	})
}

func (s *Service) RegisterPasswordUser(ctx context.Context, username string, nickname string, password string) (*PasswordRegistrationResult, error) {
	username = normalizeFallbackUsername(username)
	nickname = strings.TrimSpace(nickname)

	if !validateFallbackUsername(username) || !validateNickname(nickname) || !validatePassword(password) {
		return nil, ErrInvalidAuthInput
	}

	existingUsers, err := s.Users.ListByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if len(existingUsers) > 0 {
		return nil, ErrUserAlreadyExists
	}

	existingNickname, err := s.Users.GetByNickname(ctx, nickname)
	if err != nil {
		return nil, err
	}
	if existingNickname != nil {
		return nil, ErrNicknameTaken
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}
	userID, err := webUserID(username)
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

func (s *Service) LoginPasswordUser(ctx context.Context, username string, password string) (string, *domain.User, error) {
	username = normalizeFallbackUsername(username)
	if !validateFallbackUsername(username) || password == "" {
		return "", nil, ErrInvalidCredentials
	}

	users, err := s.Users.ListByUsername(ctx, username)
	if err != nil {
		return "", nil, err
	}
	sortUsersForUsernameResolution(users)
	for i := range users {
		u := &users[i]
		if u.IsBanned || !u.IsActive {
			continue
		}
		if checkPassword(u.Password, password) {
			now := time.Now().UTC()
			u.LastLogin = &now
			if err := s.Users.Update(ctx, u); err != nil {
				return "", nil, err
			}

			token, err := s.issueToken(u)
			if err != nil {
				return "", nil, err
			}
			return token, u, nil
		}
	}
	return "", nil, ErrInvalidCredentials
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
