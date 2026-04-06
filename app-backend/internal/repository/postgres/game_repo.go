package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pridecrm/app-backend/internal/domain"
)

type GameRepo struct {
	pool *pgxpool.Pool
}

func NewGameRepo(pool *pgxpool.Pool) *GameRepo {
	return &GameRepo{pool: pool}
}

func scanGame(row pgx.Row) (*domain.Game, error) {
	var g domain.Game
	var photo *string
	var completedAt *time.Time
	err := row.Scan(
		&g.GameID, &g.Date, &g.Time, &g.Description, &g.Buyin, &g.ReentryBuyin,
		&g.Location, &photo, &g.IsActive, &g.Completed, &completedAt, &g.CreatedAt,
		&g.BasePoints, &g.PointsPerExtraPlayer, &g.MinPlayersForExtraPoints,
	)
	g.Photo = photo
	g.CompletedAt = completedAt
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *GameRepo) Create(ctx context.Context, g *domain.Game) error {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO games (date, time, description, buyin, reentry_buyin, location, photo, is_active,
			completed, completed_at, base_points, points_per_extra_player, min_players_for_extra_points)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING game_id`,
		g.Date, g.Time, g.Description, g.Buyin, g.ReentryBuyin, g.Location, g.Photo, g.IsActive,
		g.Completed, g.CompletedAt, g.BasePoints, g.PointsPerExtraPlayer, g.MinPlayersForExtraPoints,
	).Scan(&g.GameID)
	return err
}

func (r *GameRepo) GetByID(ctx context.Context, id int64) (*domain.Game, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT game_id, date, time, description, buyin, reentry_buyin, location, photo, is_active,
			completed, completed_at, created_at, base_points, points_per_extra_player, min_players_for_extra_points
		FROM games WHERE game_id = $1`, id)
	g, err := scanGame(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return g, err
}

func (r *GameRepo) Update(ctx context.Context, g *domain.Game) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE games SET
			date = $2, time = $3, description = $4, buyin = $5, reentry_buyin = $6, location = $7, photo = $8,
			is_active = $9, completed = $10, completed_at = $11,
			base_points = $12, points_per_extra_player = $13, min_players_for_extra_points = $14
		WHERE game_id = $1`,
		g.GameID, g.Date, g.Time, g.Description, g.Buyin, g.ReentryBuyin, g.Location, g.Photo, g.IsActive,
		g.Completed, g.CompletedAt, g.BasePoints, g.PointsPerExtraPlayer, g.MinPlayersForExtraPoints,
	)
	return err
}

func (r *GameRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM games WHERE game_id = $1`, id)
	return err
}

func (r *GameRepo) ListAll(ctx context.Context) ([]domain.Game, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT game_id, date, time, description, buyin, reentry_buyin, location, photo, is_active,
			completed, completed_at, created_at, base_points, points_per_extra_player, min_players_for_extra_points
		FROM games ORDER BY date, time`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGameRows(rows)
}

func (r *GameRepo) ListActive(ctx context.Context) ([]domain.Game, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT game_id, date, time, description, buyin, reentry_buyin, location, photo, is_active,
			completed, completed_at, created_at, base_points, points_per_extra_player, min_players_for_extra_points
		FROM games WHERE is_active = TRUE ORDER BY date, time`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGameRows(rows)
}

func scanGameRows(rows pgx.Rows) ([]domain.Game, error) {
	var out []domain.Game
	for rows.Next() {
		g, err := scanGame(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *g)
	}
	return out, rows.Err()
}

func (r *GameRepo) ListRecent(ctx context.Context, limit int) ([]domain.Game, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT game_id, date, time, description, buyin, reentry_buyin, location, photo, is_active,
			completed, completed_at, created_at, base_points, points_per_extra_player, min_players_for_extra_points
		FROM games ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGameRows(rows)
}

func (r *GameRepo) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM games`).Scan(&n)
	return n, err
}

func (r *GameRepo) CountActive(ctx context.Context) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM games WHERE is_active = TRUE`).Scan(&n)
	return n, err
}
