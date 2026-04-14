package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"backend/internal/domain"
	"backend/internal/repository"
)

var _ repository.TimerRepository = (*TimerRepository)(nil)

type TimerRepository struct {
	db *DB
}

func NewTimerRepository(db *DB) *TimerRepository {
	return &TimerRepository{db: db}
}

func (r *TimerRepository) GetTimerState(ctx context.Context, tournamentID string) (*domain.TimerState, error) {
	var (
		currentLevelIndex sql.NullInt32
		state             sql.NullString
		levelStartedAt    sql.NullTime
		remainingSeconds  sql.NullInt32
	)

	err := r.db.QueryRowContext(ctx, `
		SELECT current_level_index, state, level_started_at, remaining_seconds
		FROM tournaments
		WHERE id = $1`,
		tournamentID,
	).Scan(&currentLevelIndex, &state, &levelStartedAt, &remainingSeconds)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get timer state for %s: %w", tournamentID, err)
	}

	if !currentLevelIndex.Valid || !state.Valid {
		return nil, nil
	}

	ts := &domain.TimerState{
		TournamentID:      tournamentID,
		CurrentLevelIndex: int(currentLevelIndex.Int32),
		State:             state.String,
		RemainingSeconds:  0,
	}

	if remainingSeconds.Valid {
		ts.RemainingSeconds = int(remainingSeconds.Int32)
	}
	if levelStartedAt.Valid {
		ts.LevelStartedAt = levelStartedAt.Time
	}

	return ts, nil
}

func (r *TimerRepository) GetAllTimerStates(ctx context.Context) ([]*domain.TimerState, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, current_level_index, state, level_started_at, remaining_seconds
		FROM tournaments
		WHERE state IS NOT NULL AND state != 'stopped'
	`)
	if err != nil {
		return nil, fmt.Errorf("get all timer states: %w", err)
	}
	defer rows.Close()

	var states []*domain.TimerState
	for rows.Next() {
		var (
			tournamentID      string
			currentLevelIndex sql.NullInt32
			state             sql.NullString
			levelStartedAt    sql.NullTime
			remainingSeconds  sql.NullInt32
		)

		err := rows.Scan(&tournamentID, &currentLevelIndex, &state, &levelStartedAt, &remainingSeconds)
		if err != nil {
			return nil, fmt.Errorf("scan timer state: %w", err)
		}

		if !currentLevelIndex.Valid || !state.Valid {
			continue
		}

		ts := &domain.TimerState{
			TournamentID:      tournamentID,
			CurrentLevelIndex: int(currentLevelIndex.Int32),
			State:             state.String,
			RemainingSeconds:  0,
		}

		if remainingSeconds.Valid {
			ts.RemainingSeconds = int(remainingSeconds.Int32)
		}
		if levelStartedAt.Valid {
			ts.LevelStartedAt = levelStartedAt.Time
		}

		states = append(states, ts)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate timer states: %w", err)
	}

	return states, nil
}

func (r *TimerRepository) UpdateTimerState(ctx context.Context, s *domain.TimerState) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE tournaments
		 SET current_level_index = $1,
		     level_started_at = $2,
		     remaining_seconds = $3,
		     state = $4
		 WHERE id = $5`,
		s.CurrentLevelIndex,
		s.LevelStartedAt,
		s.RemainingSeconds,
		s.State,
		s.TournamentID,
	)
	if err != nil {
		return fmt.Errorf("update timer state: %w", err)
	}
	return nil
}
