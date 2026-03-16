package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"backend/internal/api/dto"
	"backend/internal/auth"
	"backend/internal/service"
	"backend/internal/validation"
)

type TournamentHandler struct {
	tournamentService *service.TournamentService
}

func NewTournamentHandler(tournamentService *service.TournamentService) *TournamentHandler {
	return &TournamentHandler{
		tournamentService: tournamentService,
	}
}

type CreateTournamentRequest struct {
	Name string `json:"name"`
}

type CreateLevelRequest struct {
	SmallBlind      int `json:"small_blind"`
	BigBlind        int `json:"big_blind"`
	DurationMinutes int `json:"duration_minutes"`
}

func (h *TournamentHandler) CreateTournament(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateTournamentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := validation.ValidateTournamentName(req.Name); err != nil {
		http.Error(w, "invalid tournament name", http.StatusBadRequest)
		return
	}

	t, err := h.tournamentService.CreateTournament(r.Context(), req.Name, userID)
	if err != nil {
		http.Error(w, "could not create tournament", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.ToTournamentResponse(t))
}

func (h *TournamentHandler) ListTournaments(w http.ResponseWriter, r *http.Request) {
	ts, err := h.tournamentService.ListTournaments(r.Context())
	if err != nil {
		http.Error(w, "could not list tournaments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp := make([]dto.TournamentResponse, 0, len(ts))
	for _, t := range ts {
		resp = append(resp, dto.ToTournamentResponse(t))
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *TournamentHandler) GetTournament(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	t, err := h.tournamentService.GetTournament(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrTournamentNotFound) {
			http.Error(w, "tournament not found", http.StatusNotFound)
			return
		}
		http.Error(w, "could not get tournament", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.ToTournamentResponse(t))
}
