package repository

import (
	"context"

	"backend/internal/domain"
)

type TournamentRepository interface {
	Create(ctx context.Context, t *domain.Tournament) error
	Get(ctx context.Context, id string) (*domain.Tournament, error)
	List(ctx context.Context) ([]domain.Tournament, error)
	AddLevel(ctx context.Context, tournamentID string, level *domain.Level) error
	ListLevels(ctx context.Context, tournamentID string) ([]domain.Level, error)
}
