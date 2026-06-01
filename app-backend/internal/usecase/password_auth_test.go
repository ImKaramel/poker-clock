package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/pridecrm/app-backend/internal/domain"
	infraauth "github.com/pridecrm/app-backend/internal/infrastructure/auth"
)

type fakeUserRepo struct {
	users map[string]*domain.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[string]*domain.User)}
}

func cloneUser(u *domain.User) *domain.User {
	if u == nil {
		return nil
	}
	cp := *u
	if u.NickName != nil {
		v := *u.NickName
		cp.NickName = &v
	}
	if u.FirstName != nil {
		v := *u.FirstName
		cp.FirstName = &v
	}
	if u.LastName != nil {
		v := *u.LastName
		cp.LastName = &v
	}
	if u.PhoneNumber != nil {
		v := *u.PhoneNumber
		cp.PhoneNumber = &v
	}
	if u.Email != nil {
		v := *u.Email
		cp.Email = &v
	}
	if u.DateOfBirth != nil {
		v := *u.DateOfBirth
		cp.DateOfBirth = &v
	}
	if u.LastLogin != nil {
		v := *u.LastLogin
		cp.LastLogin = &v
	}
	if u.PhotoURL != nil {
		v := *u.PhotoURL
		cp.PhotoURL = &v
	}
	return &cp
}

func (r *fakeUserRepo) Create(_ context.Context, u *domain.User) error {
	if _, exists := r.users[u.UserID]; exists {
		return errors.New("duplicate user id")
	}
	r.users[u.UserID] = cloneUser(u)
	return nil
}

func (r *fakeUserRepo) GetByID(_ context.Context, userID string) (*domain.User, error) {
	return cloneUser(r.users[userID]), nil
}

func (r *fakeUserRepo) GetByUsername(_ context.Context, username string) (*domain.User, error) {
	for _, u := range r.users {
		if normalizeFallbackUsername(u.Username) == normalizeFallbackUsername(username) {
			return cloneUser(u), nil
		}
	}
	return nil, nil
}

func (r *fakeUserRepo) GetByNickname(_ context.Context, nickname string) (*domain.User, error) {
	for _, u := range r.users {
		if u.NickName != nil && *u.NickName == nickname {
			return cloneUser(u), nil
		}
	}
	return nil, nil
}

func (r *fakeUserRepo) ListByUsername(_ context.Context, username string) ([]domain.User, error) {
	var users []domain.User
	for _, u := range r.users {
		if normalizeFallbackUsername(u.Username) == normalizeFallbackUsername(username) {
			users = append(users, *cloneUser(u))
		}
	}
	return users, nil
}

func (r *fakeUserRepo) Update(_ context.Context, u *domain.User) error {
	r.users[u.UserID] = cloneUser(u)
	return nil
}

func (r *fakeUserRepo) Delete(_ context.Context, userID string) error {
	delete(r.users, userID)
	return nil
}

func (r *fakeUserRepo) List(_ context.Context) ([]domain.User, error) {
	users := make([]domain.User, 0, len(r.users))
	for _, u := range r.users {
		users = append(users, *cloneUser(u))
	}
	return users, nil
}
func (r *fakeUserRepo) Count(_ context.Context) (int64, error)                     { return 0, nil }
func (r *fakeUserRepo) CountBanned(_ context.Context) (int64, error)               { return 0, nil }
func (r *fakeUserRepo) ListRecent(_ context.Context, _ int) ([]domain.User, error) { return nil, nil }
func (r *fakeUserRepo) ListForRating(_ context.Context) ([]domain.User, error)     { return nil, nil }
func (r *fakeUserRepo) ListForRatingByMonth(_ context.Context, _ time.Time) ([]domain.User, error) {
	return nil, nil
}

func newPasswordAuthService(repo *fakeUserRepo) *Service {
	return &Service{
		Users:            repo,
		JWT:              infraauth.NewJWTService("test-secret", time.Hour),
		Log:              slog.New(slog.NewTextHandler(io.Discard, nil)),
		AdminTelegramIDs: map[string]bool{},
	}
}

func TestRegisterPasswordUserCreatesUserAndAllowsLogin(t *testing.T) {
	repo := newFakeUserRepo()
	svc := newPasswordAuthService(repo)

	result, err := svc.RegisterPasswordUser(context.Background(), "@Test_User", "MagicNick", "qazxqazx1")
	if err != nil {
		t.Fatalf("RegisterPasswordUser returned error: %v", err)
	}
	if result == nil || result.User == nil || result.Token == "" {
		t.Fatalf("expected token and user, got %+v", result)
	}
	if result.User.Username != "test_user" {
		t.Fatalf("expected normalized username, got %q", result.User.Username)
	}
	if result.User.UserID == "" || result.User.UserID[:4] != "web_" {
		t.Fatalf("expected web user id, got %q", result.User.UserID)
	}

	stored, err := repo.GetByID(context.Background(), result.User.UserID)
	if err != nil || stored == nil {
		t.Fatalf("expected stored user, err=%v", err)
	}
	if stored.Password == "" || stored.Password == "qazxqazx1" {
		t.Fatalf("expected hashed password, got %q", stored.Password)
	}

	loginToken, loggedInUser, err := svc.LoginPasswordUser(context.Background(), "test_user", "qazxqazx1")
	if err != nil {
		t.Fatalf("LoginPasswordUser returned error: %v", err)
	}
	if loginToken == "" || loggedInUser == nil {
		t.Fatalf("expected login token and user")
	}
	if loggedInUser.UserID != result.User.UserID {
		t.Fatalf("expected same user after login, got %q", loggedInUser.UserID)
	}
	if loggedInUser.LastLogin == nil {
		t.Fatalf("expected last login to be set")
	}
}

func TestRegisterPasswordUserRejectsDuplicateUsername(t *testing.T) {
	repo := newFakeUserRepo()
	repo.users["100"] = &domain.User{
		UserID:    "100",
		Username:  "existing_user",
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	svc := newPasswordAuthService(repo)

	_, err := svc.RegisterPasswordUser(context.Background(), "existing_user", "AnotherNick", "qazxqazx1")
	if !errors.Is(err, ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestRegisterPasswordUserRejectsDuplicateNickname(t *testing.T) {
	repo := newFakeUserRepo()
	nick := "MagicNick"
	repo.users["100"] = &domain.User{
		UserID:    "100",
		Username:  "existing_user",
		NickName:  &nick,
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	svc := newPasswordAuthService(repo)

	_, err := svc.RegisterPasswordUser(context.Background(), "new_user", "MagicNick", "qazxqazx1")
	if !errors.Is(err, ErrNicknameTaken) {
		t.Fatalf("expected ErrNicknameTaken, got %v", err)
	}
}

func TestLoginPasswordUserFindsMatchingPasswordAmongLegacyDuplicates(t *testing.T) {
	repo := newFakeUserRepo()
	hash, err := hashPassword("qazxqazx1")
	if err != nil {
		t.Fatalf("hashPassword returned error: %v", err)
	}
	repo.users["200"] = &domain.User{
		UserID:    "200",
		Username:  "same_user",
		IsActive:  true,
		CreatedAt: time.Now().Add(-time.Hour),
	}
	repo.users["web_same_user_1"] = &domain.User{
		UserID:    "web_same_user_1",
		Username:  "same_user",
		Password:  hash,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	svc := newPasswordAuthService(repo)
	_, user, err := svc.LoginPasswordUser(context.Background(), "same_user", "qazxqazx1")
	if err != nil {
		t.Fatalf("LoginPasswordUser returned error: %v", err)
	}
	if user == nil || user.UserID != "web_same_user_1" {
		t.Fatalf("expected password auth user, got %+v", user)
	}
}
