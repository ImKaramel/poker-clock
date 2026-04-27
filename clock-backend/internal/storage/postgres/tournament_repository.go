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
		`INSERT INTO tournaments (id, name, state)
     			VALUES ($1, $2, $3)`,
		t.ID, t.Name, "stopped",
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
		`SELECT id, name, created_at
		 FROM tournaments WHERE id = $1`,
		id,
	).Scan(&t.ID, &t.Name, &t.CreatedAt)
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
			t.created_at,
			l.id,
			l.type,
			l.name,
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
			createdAt time.Time

			levelID         sql.NullString
			levelType       sql.NullString
			levelName       sql.NullString
			smallBlind      sql.NullInt32
			bigBlind        sql.NullInt32
			durationMinutes sql.NullInt32
			levelOrder      sql.NullInt32
		)

		if err := rows.Scan(
			&tID,
			&name,
			&createdAt,
			&levelID,
			&levelType,
			&levelName,
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
				CreatedAt: createdAt,
			}
		}

		if levelID.Valid {
			l := domain.Level{
				ID:              levelID.String,
				Type:            levelType.String,
				Name:            levelName.String,
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
		`INSERT INTO levels (
		id,
		tournament_id,
		type,
		name,
		small_blind,
		big_blind,
		duration_minutes,
		level_order
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		level.ID,
		tournamentID,
		level.Type,
		level.Name,
		level.SmallBlind,
		level.BigBlind,
		level.DurationMinutes,
		level.Order,
	)
	if err != nil {
		return fmt.Errorf("insert level: %w", err)
	}

	return nil
}

func (r *TournamentRepository) ListLevels(ctx context.Context, tournamentID string) ([]domain.Level, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT
					id,
					type,
					name,
					small_blind,
					big_blind,
					duration_minutes,
					level_order
				FROM levels
				WHERE tournament_id = $1
				ORDER BY level_order`,
		tournamentID,
	)
	if err != nil {
		return nil, fmt.Errorf("list levels for tournament %s: %w", tournamentID, err)
	}
	defer rows.Close()

	var levels []domain.Level
	for rows.Next() {
		var l domain.Level
		if err := rows.Scan(
			&l.ID,
			&l.Type,
			&l.Name,
			&l.SmallBlind,
			&l.BigBlind,
			&l.DurationMinutes,
			&l.Order,
		); err != nil {
			return nil, fmt.Errorf("scan level row: %w", err)
		}
		levels = append(levels, l)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate levels: %w", err)
	}

	return levels, nil
}

func (r *TournamentRepository) DeleteLevel(ctx context.Context, tournamentID, levelID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	var deletedOrder int
	err = tx.QueryRowContext(ctx,
		`SELECT level_order FROM levels WHERE id = $1 AND tournament_id = $2`,
		levelID, tournamentID,
	).Scan(&deletedOrder)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("level not found")
		}
		return fmt.Errorf("get level order: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`DELETE FROM levels WHERE id = $1 AND tournament_id = $2`,
		levelID, tournamentID,
	)
	if err != nil {
		return fmt.Errorf("delete level: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE levels SET level_order = level_order - 1 
		 WHERE tournament_id = $1 AND level_order > $2`,
		tournamentID, deletedOrder,
	)
	if err != nil {
		return fmt.Errorf("reindex levels: %w", err)
	}

	return nil
}

func (r *TournamentRepository) UpdateState(ctx context.Context, tournamentID, state string) error {
	return fmt.Errorf("UpdateState is no longer supported; timer state is managed via TimerRepository")
}

func (r *TournamentRepository) UpdateCurrentLevelIndex(ctx context.Context, tournamentID string, index int) error {
	return fmt.Errorf("UpdateCurrentLevelIndex is no longer supported; timer state is managed via TimerRepository")
}

func (r *TournamentRepository) Delete(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	//_, err = tx.ExecContext(ctx, `
	//	UPDATE tournaments
	//	SET state = 'stopped',
	//	    current_level_index = NULL,
	//	    level_started_at = NULL,
	//	    remaining_seconds = NULL
	//	WHERE id = $1
	//`, id)
	//if err != nil {
	//	return fmt.Errorf("clear timer state: %w", err)
	//}

	_, err = tx.ExecContext(ctx, `
		DELETE FROM levels WHERE tournament_id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("delete levels: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		DELETE FROM tournaments WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("delete tournament: %w", err)
	}

	return nil
}

func (r *TournamentRepository) UpdateTimerState() error {
	return fmt.Errorf("UpdateTimerState is no longer supported; use TimerRepository instead")
}
