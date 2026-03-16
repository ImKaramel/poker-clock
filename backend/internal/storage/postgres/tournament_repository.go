package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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
		`INSERT INTO tournaments (id, name, owner_id)
		 VALUES ($1, $2, $3)`,
		t.ID, t.Name, t.OwnerID,
	)
	if err != nil {
		return fmt.Errorf("insert tournament: %w", err)
	}

	return nil
}

func (r *TournamentRepository) Get(ctx context.Context, id string) (*domain.Tournament, error) {
	var (
		t domain.Tournament
	)
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, owner_id, created_at
		 FROM tournaments WHERE id = $1`,
		id,
	).Scan(&t.ID, &t.Name, &t.OwnerID, &t.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get tournament %s: %w", id, err)
	}

	levels, err := r.ListLevels(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list levels for tournament %s: %w", id, err)
	}
	t.Levels = levels

	return &t, nil
}

func (r *TournamentRepository) List(ctx context.Context) ([]domain.Tournament, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			t.id,
			t.name,
			t.owner_id,
			t.created_at,
			l.id,
			l.small_blind,
			l.big_blind,
			l.duration_minutes,
			l.level_order
		FROM tournaments t
		LEFT JOIN levels l ON l.tournament_id = t.id
		ORDER BY t.id, l.level_order`,
	)
	if err != nil {
		return nil, fmt.Errorf("list tournaments: %w", err)
	}
	defer rows.Close()

	var (
		list      []domain.Tournament
		currentID string
		current   *domain.Tournament
	)

	for rows.Next() {
		var (
			tID       string
			name      string
			ownerID   string
			createdAt time.Time

			levelID         sql.NullString
			smallBlind      sql.NullInt32
			bigBlind        sql.NullInt32
			durationMinutes sql.NullInt32
			levelOrder      sql.NullInt32
		)

		if err := rows.Scan(
			&tID,
			&name,
			&ownerID,
			&createdAt,
			&levelID,
			&smallBlind,
			&bigBlind,
			&durationMinutes,
			&levelOrder,
		); err != nil {
			return nil, fmt.Errorf("scan tournament row: %w", err)
		}

		if current == nil || tID != currentID {
			if current != nil {
				list = append(list, *current)
			}
			currentID = tID
			current = &domain.Tournament{
				ID:        tID,
				Name:      name,
				OwnerID:   ownerID,
				CreatedAt: createdAt,
			}
		}

		if levelID.Valid {
			l := domain.Level{
				ID:              levelID.String,
				SmallBlind:      int(smallBlind.Int32),
				BigBlind:        int(bigBlind.Int32),
				DurationMinutes: int(durationMinutes.Int32),
				Order:           int(levelOrder.Int32),
			}
			current.Levels = append(current.Levels, l)
		}
	}

	if current != nil {
		list = append(list, *current)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tournaments: %w", err)
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
		return fmt.Errorf("insert level: %w", err)
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
		return nil, fmt.Errorf("list levels for tournament %s: %w", tournamentID, err)
	}
	defer rows.Close()

	var levels []domain.Level
	for rows.Next() {
		var l domain.Level
		if err := rows.Scan(&l.ID, &l.SmallBlind, &l.BigBlind, &l.DurationMinutes, &l.Order); err != nil {
			return nil, fmt.Errorf("scan level row: %w", err)
		}
		levels = append(levels, l)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate levels: %w", err)
	}

	return levels, nil
}

func (r *TournamentRepository) UpdateState(ctx context.Context, tournamentID, state string) error {
	return fmt.Errorf("UpdateState is no longer supported; timer state is managed via TimerRepository")
}

func (r *TournamentRepository) UpdateCurrentLevelIndex(ctx context.Context, tournamentID string, index int) error {
	return fmt.Errorf("UpdateCurrentLevelIndex is no longer supported; timer state is managed via TimerRepository")
}

func (r *TournamentRepository) UpdateTimerState(ctx context.Context, tournamentID string, currentLevelIndex int, levelStartedAt time.Time, remainingSeconds int, state string) error {
	return fmt.Errorf("UpdateTimerState is no longer supported; use TimerRepository instead")
}
