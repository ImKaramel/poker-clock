package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"backend/internal/api/dto"
	"backend/internal/domain"
	"backend/internal/service"
)

type LevelHandler struct {
	tournamentService *service.TournamentService
}

func NewLevelHandler(tournamentService *service.TournamentService) *LevelHandler {
	return &LevelHandler{
		tournamentService: tournamentService,
	}
}

func (h *LevelHandler) AddLevel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	var req CreateLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	level := domain.Level{
		SmallBlind:      req.SmallBlind,
		BigBlind:        req.BigBlind,
		DurationMinutes: req.DurationMinutes,
	}

	t, err := h.tournamentService.AddLevel(id, level)
	if err != nil {
		http.Error(w, "tournament not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.ToTournamentResponse(t))
}

func (h *LevelHandler) ListLevels(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	levels, err := h.tournamentService.ListLevels(id)
	if err != nil {
		http.Error(w, "tournament not found", http.StatusNotFound)
		return
	}

	resp := make([]dto.LevelResponse, 0, len(levels))
	for _, l := range levels {
		resp = append(resp, dto.ToLevelResponse(l))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
