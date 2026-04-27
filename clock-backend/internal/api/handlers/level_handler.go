package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"backend/internal/api/dto"
	"backend/internal/domain"
	"backend/internal/service"
	"backend/internal/validation"
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

	if req.Type != "level" && req.Type != "break" {
		http.Error(w, "invalid type: must be 'level' or 'break'", http.StatusBadRequest)
		return
	}

	if req.Type == "level" {
		if err := validation.ValidateBlinds(req.SmallBlind, req.BigBlind); err != nil {
			http.Error(w, "invalid blinds", http.StatusBadRequest)
			return
		}
	} else if req.Type == "break" {
		req.SmallBlind = 0
		req.BigBlind = 0
	}

	if err := validation.ValidateDurationMinutes(req.DurationMinutes); err != nil {
		http.Error(w, "invalid duration_minutes", http.StatusBadRequest)
		return
	}

	var level domain.Level
	level.Type = req.Type
	level.Name = req.Name
	level.SmallBlind = req.SmallBlind
	level.BigBlind = req.BigBlind
	level.DurationMinutes = req.DurationMinutes

	t, err := h.tournamentService.AddLevel(r.Context(), id, level)
	if err != nil {
		if errors.Is(err, service.ErrTournamentNotFound) {
			http.Error(w, "tournament not found", http.StatusNotFound)
			return
		}
		http.Error(w, "could not add level", http.StatusInternalServerError)
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

	levels, err := h.tournamentService.ListLevels(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrTournamentNotFound) {
			http.Error(w, "tournament not found", http.StatusNotFound)
			return
		}
		http.Error(w, "could not list levels", http.StatusInternalServerError)
		return
	}

	resp := make([]dto.LevelResponse, 0, len(levels))
	for _, l := range levels {
		resp = append(resp, dto.ToLevelResponse(l))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *LevelHandler) DeleteLevel(w http.ResponseWriter, r *http.Request) {
	tournamentID := chi.URLParam(r, "id")
	if tournamentID == "" {
		http.Error(w, "missing tournament id", http.StatusBadRequest)
		return
	}

	levelID := chi.URLParam(r, "levelId")
	if levelID == "" {
		http.Error(w, "missing level id", http.StatusBadRequest)
		return
	}

	err := h.tournamentService.DeleteLevel(r.Context(), tournamentID, levelID)
	if err != nil {
		if errors.Is(err, service.ErrTournamentNotFound) {
			http.Error(w, "tournament not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrLevelNotFound) {
			http.Error(w, "level not found", http.StatusNotFound)
			return
		}
		http.Error(w, "could not delete level", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
