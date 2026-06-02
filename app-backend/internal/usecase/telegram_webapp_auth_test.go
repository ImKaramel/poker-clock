package usecase

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"
)

func signedTelegramInitData(t *testing.T, botToken string, now time.Time, userJSON string) string {
	t.Helper()

	values := url.Values{}
	values.Set("auth_date", fmt.Sprintf("%d", now.Unix()))
	values.Set("query_id", "AAHdF6IQAAAAAN0XohDhrOrc")
	values.Set("user", userJSON)

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", key, values.Get(key)))
	}

	secretMAC := hmac.New(sha256.New, []byte("WebAppData"))
	secretMAC.Write([]byte(botToken))

	hashMAC := hmac.New(sha256.New, secretMAC.Sum(nil))
	hashMAC.Write([]byte(strings.Join(parts, "\n")))
	values.Set("hash", hex.EncodeToString(hashMAC.Sum(nil)))

	return values.Encode()
}

func TestValidateTelegramWebAppInitData(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	initData := signedTelegramInitData(
		t,
		"123456:ABC-DEF",
		now,
		`{"id":123456789,"first_name":"Test","username":"tester"}`,
	)

	user, err := validateTelegramInitData(initData, "123456:ABC-DEF", now)
	if err != nil {
		t.Fatalf("validateTelegramInitData returned error: %v", err)
	}
	if got := user.Username; got != "tester" {
		t.Fatalf("username = %v, want tester", got)
	}
}

func TestValidateTelegramWebAppInitDataRejectsInvalidHash(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	initData := signedTelegramInitData(
		t,
		"123456:ABC-DEF",
		now,
		`{"id":123456789,"first_name":"Test","username":"tester"}`,
	)
	initData = strings.Replace(initData, "tester", "attacker", 1)

	if _, err := validateTelegramInitData(initData, "123456:ABC-DEF", now); err == nil {
		t.Fatal("expected invalid hash error")
	}
}

func TestValidateTelegramWebAppInitDataRejectsExpiredAuthDate(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	initData := signedTelegramInitData(
		t,
		"123456:ABC-DEF",
		now.Add(-telegramInitDataMaxAge-time.Second),
		`{"id":123456789,"first_name":"Test","username":"tester"}`,
	)

	if _, err := validateTelegramInitData(initData, "123456:ABC-DEF", now); err == nil {
		t.Fatal("expected expired init data error")
	}
}
