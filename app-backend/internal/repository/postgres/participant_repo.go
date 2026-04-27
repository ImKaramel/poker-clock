package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pridecrm/app-backend/internal/domain"
)

type ParticipantRepo struct {
	pool *pgxpool.Pool
}

func NewParticipantRepo(pool *pgxpool.Pool) *ParticipantRepo {
	return &ParticipantRepo{pool: pool}
}

func scanParticipant(row pgx.Row) (*domain.Participant, error) {
	var p domain.Participant
	var pos *int
	err := row.Scan(&p.ID, &p.UserID, &p.GameID, &p.Entries, &p.Rebuys, &p.Addons, &p.FinalPoints, &pos, &p.Arrived, &p.IsOut, &p.JoinedAt)
	p.Position = pos
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ParticipantRepo) Create(ctx context.Context, p *domain.Participant) error {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO participants (user_id, game_id, entries, rebuys, addons, final_points, position, arrived, is_out)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id, joined_at`,
		p.UserID, p.GameID, p.Entries, p.Rebuys, p.Addons, p.FinalPoints, p.Position, p.Arrived, p.IsOut,
	).Scan(&p.ID, &p.JoinedAt)
	return err
}

func (r *ParticipantRepo) GetByID(ctx context.Context, id int64) (*domain.Participant, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, game_id, entries, rebuys, addons, final_points, position, arrived, is_out, joined_at
		FROM participants WHERE id = $1`, id)
	p, err := scanParticipant(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return p, err
}

func (r *ParticipantRepo) GetByUserAndGame(ctx context.Context, userID string, gameID int64) (*domain.Participant, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, game_id, entries, rebuys, addons, final_points, position, arrived, is_out, joined_at
		FROM participants WHERE user_id = $1 AND game_id = $2`, userID, gameID)
	p, err := scanParticipant(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return p, err
}

func (r *ParticipantRepo) Update(ctx context.Context, p *domain.Participant, rebuyDelta int) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE participants
		SET
			entries = $2,
			rebuys = rebuys + $3,
			addons = $4,
			final_points = $5,
			position = $6,
			arrived = $7,
			is_out = $8
		WHERE id = $1`,
		p.ID,
		p.Entries,
		rebuyDelta,
		p.Addons,
		p.FinalPoints,
		p.Position,
		p.Arrived,
		p.IsOut,
	)
	return err
}

func (r *ParticipantRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM participants WHERE id = $1`, id)
	return err
}

func (r *ParticipantRepo) DeleteByUserAndGame(ctx context.Context, userID string, gameID int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM participants WHERE user_id = $1 AND game_id = $2`, userID, gameID)
	return err
}

func (r *ParticipantRepo) ListByGame(ctx context.Context, gameID int64) ([]domain.Participant, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, game_id, entries, rebuys, addons, final_points, position, arrived, is_out, joined_at
		FROM participants WHERE game_id = $1 ORDER BY joined_at`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanParticipantRows(rows)
}

func (r *ParticipantRepo) ListByUser(ctx context.Context, userID string) ([]domain.Participant, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, game_id, entries, rebuys, addons, final_points, position, arrived, is_out, joined_at
		FROM participants WHERE user_id = $1 ORDER BY joined_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanParticipantRows(rows)
}

func (r *ParticipantRepo) ListAll(ctx context.Context) ([]domain.Participant, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, game_id, entries, rebuys, addons, final_points, position, arrived, is_out, joined_at
		FROM participants ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanParticipantRows(rows)
}

func (r *ParticipantRepo) ListUpcomingForUser(ctx context.Context, userID string, fromDate time.Time) ([]domain.Participant, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.user_id, p.game_id, p.entries, p.rebuys, p.addons, p.final_points, p.position, p.arrived, p.is_out, p.joined_at
		FROM participants p
		INNER JOIN games g ON g.game_id = p.game_id
		WHERE p.user_id = $1 AND g.date >= $2::date
		ORDER BY g.date, g.time
		LIMIT 5`, userID, fromDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanParticipantRows(rows)
}

func (r *ParticipantRepo) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM participants`).Scan(&n)
	return n, err
}

func (r *ParticipantRepo) CountByGame(ctx context.Context, gameID int64) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM participants WHERE game_id = $1`, gameID).Scan(&n)
	return n, err
}

func scanParticipantRows(rows pgx.Rows) ([]domain.Participant, error) {
	var out []domain.Participant
	for rows.Next() {
		p, err := scanParticipant(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

func (r *ParticipantRepo) SetArrived(ctx context.Context, participantID int64, arrived bool) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE participants SET arrived = $2
		WHERE id = $1`,
		participantID, arrived,
	)
	return err
}
