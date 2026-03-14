package dto

import "backend/internal/domain"

type LevelResponse struct {
	ID              string `json:"id"`
	SmallBlind      int    `json:"small_blind"`
	BigBlind        int    `json:"big_blind"`
	DurationMinutes int    `json:"duration_minutes"`
	Order           int    `json:"order"`
}

type TournamentResponse struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	OwnerID           string          `json:"owner_id"`
	CurrentLevelIndex int             `json:"current_level_index"`
	State             string          `json:"state"`
	Levels            []LevelResponse `json:"levels"`
}

func ToLevelResponse(l domain.Level) LevelResponse {
	return LevelResponse{
		ID:              l.ID,
		SmallBlind:      l.SmallBlind,
		BigBlind:        l.BigBlind,
		DurationMinutes: l.DurationMinutes,
		Order:           l.Order,
	}
}

func ToTournamentResponse(t domain.Tournament) TournamentResponse {
	levels := make([]LevelResponse, 0, len(t.Levels))
	for _, l := range t.Levels {
		levels = append(levels, ToLevelResponse(l))
	}

	return TournamentResponse{
		ID:                t.ID,
		Name:              t.Name,
		OwnerID:           t.OwnerID,
		CurrentLevelIndex: t.CurrentLevelIndex,
		State:             t.State,
		Levels:            levels,
	}
}
