package auth

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func performAuthRequest(t *testing.T, token string) map[string]any {
	t.Helper()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(MiddlewareJWT(NewJWTService("test-secret", time.Hour), testLogger()))
	router.GET("/check", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id": UserIDFromMust(c),
			"admin":   IsAdminFromContext(c),
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/check", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return body
}

func UserIDFromMust(c *gin.Context) string {
	uid, _ := UserIDFromContext(c)
	return uid
}

func TestMiddlewareJWTKeepsNonAdminUsersNonAdmin(t *testing.T) {
	jwtSvc := NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Issue("100", false)
	if err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}

	body := performAuthRequest(t, token)
	if body["user_id"] != "100" {
		t.Fatalf("expected user_id=100, got %#v", body["user_id"])
	}
	if body["admin"] != false {
		t.Fatalf("expected admin=false, got %#v", body["admin"])
	}
}

func TestMiddlewareJWTKeepsAdminUsersAdmin(t *testing.T) {
	jwtSvc := NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Issue("200", true)
	if err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}

	body := performAuthRequest(t, token)
	if body["user_id"] != "200" {
		t.Fatalf("expected user_id=200, got %#v", body["user_id"])
	}
	if body["admin"] != true {
		t.Fatalf("expected admin=true, got %#v", body["admin"])
	}
}
