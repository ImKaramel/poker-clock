package repository

import (
	"context"

	"backend/internal/domain"
)

type TimerRepository interface {
	GetTimerState(ctx context.Context, tournamentID string) (*domain.TimerState, error)
	UpdateTimerState(ctx context.Context, state *domain.TimerState) error
}
