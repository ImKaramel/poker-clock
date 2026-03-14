package postgres

import (
	"context"
	"database/sql"

	"backend/internal/domain"
	"backend/internal/repository"

	"github.com/google/uuid"
)

var _ repository.TournamentRepository = (*TournamentRepository)(nil)

type TournamentRepository struct {
	db *DB
}

func NewTournamentRepository(db *DB) *TournamentRepository {
	return &TournamentRepository{db: db}
}

func (r *TournamentRepository) Create(ctx context.Context, t *domain.Tournament) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO tournaments (id, name, owner_id, current_level_index, state)
		 VALUES ($1, $2, $3, $4, $5)`,
		t.ID, t.Name, t.OwnerID, t.CurrentLevelIndex, t.State,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *TournamentRepository) Get(ctx context.Context, id string) (*domain.Tournament, error) {
	var t domain.Tournament
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, owner_id, current_level_index, state
		 FROM tournaments WHERE id = $1`,
		id,
	).Scan(&t.ID, &t.Name, &t.OwnerID, &t.CurrentLevelIndex, &t.State)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	levels, err := r.ListLevels(ctx, id)
	if err != nil {
		return nil, err
	}
	t.Levels = levels

	return &t, nil
}

func (r *TournamentRepository) List(ctx context.Context) ([]domain.Tournament, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, owner_id, current_level_index, state FROM tournaments ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.Tournament
	for rows.Next() {
		var t domain.Tournament
		if err := rows.Scan(&t.ID, &t.Name, &t.OwnerID, &t.CurrentLevelIndex, &t.State); err != nil {
			return nil, err
		}
		t.Levels, err = r.ListLevels(ctx, t.ID)
		if err != nil {
			return nil, err
		}
		list = append(list, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *TournamentRepository) AddLevel(ctx context.Context, tournamentID string, level *domain.Level) error {
	if level.ID == "" {
		level.ID = uuid.New().String()
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO levels (id, tournament_id, small_blind, big_blind, duration_minutes, level_order)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		level.ID, tournamentID, level.SmallBlind, level.BigBlind, level.DurationMinutes, level.Order,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *TournamentRepository) ListLevels(ctx context.Context, tournamentID string) ([]domain.Level, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, small_blind, big_blind, duration_minutes, level_order
		 FROM levels WHERE tournament_id = $1 ORDER BY level_order`,
		tournamentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var levels []domain.Level
	for rows.Next() {
		var l domain.Level
		if err := rows.Scan(&l.ID, &l.SmallBlind, &l.BigBlind, &l.DurationMinutes, &l.Order); err != nil {
			return nil, err
		}
		levels = append(levels, l)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return levels, nil
}

func (r *TournamentRepository) UpdateState(ctx context.Context, tournamentID, state string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE tournaments SET state = $1 WHERE id = $2`,
		state, tournamentID,
	)
	return err
}

func (r *TournamentRepository) UpdateCurrentLevelIndex(ctx context.Context, tournamentID string, index int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE tournaments SET current_level_index = $1 WHERE id = $2`,
		index, tournamentID,
	)
	return err
}
