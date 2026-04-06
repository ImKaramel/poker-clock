package handlers

import (
	"encoding/json"
	"net/http"
	"os"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if req.Password != os.Getenv("ADMIN_PASSWORD") {
		http.Error(w, "invalid password", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}
