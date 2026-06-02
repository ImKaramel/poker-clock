package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pridecrm/app-backend/internal/domain"
)

type fakeContactAuthRepo struct {
	challenges map[string]*domain.ContactAuthChallenge
}

func newFakeContactAuthRepo() *fakeContactAuthRepo {
	return &fakeContactAuthRepo{challenges: make(map[string]*domain.ContactAuthChallenge)}
}

func cloneContactAuth(c *domain.ContactAuthChallenge) *domain.ContactAuthChallenge {
	if c == nil {
		return nil
	}
	cp := *c
	if c.TelegramUserID != nil {
		v := *c.TelegramUserID
		cp.TelegramUserID = &v
	}
	if c.PhoneNumber != nil {
		v := *c.PhoneNumber
		cp.PhoneNumber = &v
	}
	if c.CompletedAt != nil {
		v := *c.CompletedAt
		cp.CompletedAt = &v
	}
	if c.ConsumedAt != nil {
		v := *c.ConsumedAt
		cp.ConsumedAt = &v
	}
	return &cp
}

func (r *fakeContactAuthRepo) Create(_ context.Context, challenge *domain.ContactAuthChallenge) error {
	r.challenges[challenge.Token] = cloneContactAuth(challenge)
	return nil
}

func (r *fakeContactAuthRepo) GetByToken(_ context.Context, token string) (*domain.ContactAuthChallenge, error) {
	return cloneContactAuth(r.challenges[token]), nil
}

func (r *fakeContactAuthRepo) Complete(_ context.Context, token string, telegramUserID string, phoneNumber string) error {
	challenge := r.challenges[token]
	if challenge == nil || challenge.Status != "pending" || time.Now().After(challenge.ExpiresAt) {
		return ErrNotFound
	}
	now := time.Now()
	challenge.Status = "completed"
	challenge.TelegramUserID = &telegramUserID
	challenge.PhoneNumber = &phoneNumber
	challenge.CompletedAt = &now
	return nil
}

func (r *fakeContactAuthRepo) Consume(_ context.Context, token string) error {
	challenge := r.challenges[token]
	if challenge == nil || challenge.Status != "completed" {
		return ErrNotFound
	}
	now := time.Now()
	challenge.Status = "consumed"
	challenge.ConsumedAt = &now
	return nil
}

func signedInitData(t *testing.T, botToken string, values url.Values) string {
	t.Helper()

	keys := make([]string, 0, len(values))
	for key := range values {
		if key == "hash" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+values.Get(key))
	}

	secretHMAC := hmac.New(sha256.New, []byte("WebAppData"))
	secretHMAC.Write([]byte(botToken))

	hashHMAC := hmac.New(sha256.New, secretHMAC.Sum(nil))
	hashHMAC.Write([]byte(strings.Join(parts, "\n")))
	values.Set("hash", hex.EncodeToString(hashHMAC.Sum(nil)))

	return values.Encode()
}

func TestValidateTelegramInitData(t *testing.T) {
	now := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
	botToken := "123456:test-token"
	raw := signedInitData(t, botToken, url.Values{
		"auth_date": {strconv.FormatInt(now.Add(-time.Minute).Unix(), 10)},
		"query_id":  {"AAE"},
		"user":      {`{"id":463021572,"username":"admin","first_name":"Admin"}`},
	})

	user, err := validateTelegramInitData(raw, botToken, now)
	if err != nil {
		t.Fatalf("validateTelegramInitData returned error: %v", err)
	}
	if user.ID != 463021572 || user.Username != "admin" || user.FirstName != "Admin" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestValidateTelegramInitDataRejectsExpiredHash(t *testing.T) {
	now := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
	botToken := "123456:test-token"
	raw := signedInitData(t, botToken, url.Values{
		"auth_date": {strconv.FormatInt(now.Add(-25*time.Hour).Unix(), 10)},
		"user":      {`{"id":463021572,"username":"admin"}`},
	})

	if _, err := validateTelegramInitData(raw, botToken, now); err == nil {
		t.Fatalf("expected expired init data error")
	}
}

func TestPreviewCompleteGameResolvesUsersAndAddsKO(t *testing.T) {
	nick := "Санчо"
	repo := newFakeUserRepo()
	repo.users["1"] = &domain.User{UserID: "1", Username: "sancho", NickName: &nick, IsActive: true}
	repo.users["2"] = &domain.User{UserID: "2", Username: "doc", IsActive: true}
	svc := &Service{Users: repo}
	game := &domain.Game{BasePoints: 100}

	preview, err := svc.buildCompletePreview(context.Background(), game, []CompleteResultInput{
		{Position: 1, Nickname: "Санчо", KOCount: 2},
		{Position: 2, Nickname: "unknown"},
	})
	if err != nil {
		t.Fatalf("buildCompletePreview returned error: %v", err)
	}
	if len(preview.Results) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(preview.Results))
	}
	if preview.Results[0].Status != "resolved" || preview.Results[0].User.UserID != "1" {
		t.Fatalf("expected first row resolved to user 1, got %+v", preview.Results[0])
	}
	expected := ratingPlacePoints(game, 2, 1) + 2*100
	if preview.Results[0].TotalPoints != expected {
		t.Fatalf("expected %d points, got %d", expected, preview.Results[0].TotalPoints)
	}
	if len(preview.Unresolved) != 1 || preview.Unresolved[0].Nickname != "unknown" {
		t.Fatalf("expected one unresolved row, got %+v", preview.Unresolved)
	}
}

func TestContactAuthChallengeFlow(t *testing.T) {
	ctx := context.Background()
	users := newFakeUserRepo()
	users.users["42"] = &domain.User{UserID: "42", Username: "player42", IsActive: true}
	svc := newPasswordAuthService(users)
	svc.ContactAuth = newFakeContactAuthRepo()

	challenge, err := svc.CreateContactAuthChallenge(ctx)
	if err != nil {
		t.Fatalf("CreateContactAuthChallenge returned error: %v", err)
	}
	if challenge.Token == "" {
		t.Fatalf("expected non-empty challenge token")
	}

	if _, _, err := svc.ConsumeContactAuthChallenge(ctx, challenge.Token); err != ErrContactAuthPending {
		t.Fatalf("expected pending error before contact confirmation, got %v", err)
	}

	if err := svc.ConfirmContactAuthChallenge(ctx, challenge.Token, "42", "+79991234567"); err != nil {
		t.Fatalf("ConfirmContactAuthChallenge returned error: %v", err)
	}

	token, user, err := svc.ConsumeContactAuthChallenge(ctx, challenge.Token)
	if err != nil {
		t.Fatalf("ConsumeContactAuthChallenge returned error: %v", err)
	}
	if token == "" || user == nil || user.UserID != "42" {
		t.Fatalf("expected jwt and user 42, token=%q user=%+v", token, user)
	}

	if _, _, err := svc.ConsumeContactAuthChallenge(ctx, challenge.Token); err != ErrContactAuthConsumed {
		t.Fatalf("expected consumed error on replay, got %v", err)
	}
}
