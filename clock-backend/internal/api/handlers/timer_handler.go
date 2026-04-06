package handlers

import (
	"encoding/json"
	_ "encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"

	_ "backend/internal/api/dto"
	"backend/internal/service"
)

type TimerHandler struct {
	timerService *service.TimerService
	upgrader     websocket.Upgrader
}

func NewTimerHandler(timerService *service.TimerService) *TimerHandler {
	return &TimerHandler{
		timerService: timerService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (h *TimerHandler) StartTournament(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	if err := h.timerService.Start(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrTournamentNotFound) {
			http.Error(w, "tournament not found", http.StatusNotFound)
			return
		}
		http.Error(w, "could not start tournament", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TimerHandler) PauseTournament(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	if err := h.timerService.Pause(id); err != nil {
		if errors.Is(err, service.ErrTournamentNotFound) {
			http.Error(w, "tournament not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrTournamentNotRunning) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "could not pause tournament", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TimerHandler) ResumeTournament(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	if err := h.timerService.Resume(id); err != nil {
		if errors.Is(err, service.ErrTournamentNotFound) {
			http.Error(w, "tournament not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrTournamentNotPaused) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "could not resume tournament", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TimerHandler) NextLevel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	if err := h.timerService.NextLevel(id); err != nil {
		if errors.Is(err, service.ErrTournamentNotFound) {
			http.Error(w, "tournament not found", http.StatusNotFound)
			return
		}
		http.Error(w, "could not advance level", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TimerHandler) TimerWS(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	state, err := h.timerService.GetState(r.Context(), id)
	if err != nil {
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "could not get timer state"))
		return
	}

	initial := map[string]any{
		"level":             state.LevelNumber,
		"small_blind":       state.SmallBlind,
		"big_blind":         state.BigBlind,
		"remaining_seconds": state.RemainingSeconds,
	}
	if err := conn.WriteJSON(initial); err != nil {
		return
	}

	updates, unsubscribe, err := h.timerService.Subscribe(id)
	if err != nil {
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "could not subscribe to timer"))
		return
	}
	defer unsubscribe()

	for {
		select {
		case <-r.Context().Done():
			return
		case s, ok := <-updates:
			if !ok {
				return
			}
			msg := map[string]any{
				"level":             s.LevelNumber,
				"small_blind":       s.SmallBlind,
				"big_blind":         s.BigBlind,
				"remaining_seconds": s.RemainingSeconds,
			}
			if err := conn.WriteJSON(msg); err != nil {
				return
			}
		}
	}
}

func (h *TimerHandler) UpdateStats(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req struct {
		PlayersCount int `json:"players_count"`
		TotalChips   int `json:"total_chips"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	h.timerService.UpdateStats(id, req.PlayersCount, req.TotalChips)

	w.WriteHeader(http.StatusOK)
}
