package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pridecrm/app-backend/internal/domain"
)

type TournamentRepo struct {
	pool *pgxpool.Pool
}

func NewTournamentRepo(pool *pgxpool.Pool) *TournamentRepo {
	return &TournamentRepo{pool: pool}
}

func (r *TournamentRepo) CreateHistory(ctx context.Context, h *domain.TournamentHistory) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO tournament_history (game_id, date, time, tournament_name, location, buyin, reentry_buyin,
			total_revenue, participants_count)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, completed_at`,
		h.GameID, h.Date, h.Time, h.TournamentName, h.Location, h.Buyin, h.ReentryBuyin,
		h.TotalRevenue, h.ParticipantsCount,
	).Scan(&h.ID, &h.CompletedAt)
}

func (r *TournamentRepo) GetHistoryByID(ctx context.Context, id int64) (*domain.TournamentHistory, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, game_id, date, time, tournament_name, location, buyin, reentry_buyin,
			total_revenue, participants_count, completed_at
		FROM tournament_history WHERE id = $1`, id)
	h, err := scanHistory(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	parts, err := r.ListTournamentParticipants(ctx, h.ID)
	if err != nil {
		return nil, err
	}
	h.Participants = parts
	return h, nil
}

func scanHistory(row pgx.Row) (*domain.TournamentHistory, error) {
	var h domain.TournamentHistory
	var t *time.Time
	var re *int
	err := row.Scan(
		&h.ID, &h.GameID, &h.Date, &t, &h.TournamentName, &h.Location,
		&h.Buyin, &re, &h.TotalRevenue, &h.ParticipantsCount, &h.CompletedAt,
	)
	h.Time = t
	h.ReentryBuyin = re
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *TournamentRepo) ListHistory(ctx context.Context) ([]domain.TournamentHistory, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, game_id, date, time, tournament_name, location, buyin, reentry_buyin,
			total_revenue, participants_count, completed_at
		FROM tournament_history ORDER BY date DESC, time DESC NULLS LAST`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.TournamentHistory
	for rows.Next() {
		h, err := scanHistory(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *h)
	}
	return out, rows.Err()
}

func (r *TournamentRepo) ListHistoryByUser(ctx context.Context, userID string) ([]domain.TournamentHistory, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT h.id, h.game_id, h.date, h.time, h.tournament_name, h.location, h.buyin, h.reentry_buyin,
			h.total_revenue, h.participants_count, h.completed_at
		FROM tournament_history h
		INNER JOIN tournament_participants tp ON tp.tournament_history_id = h.id
		WHERE tp.user_id = $1
		ORDER BY h.date DESC, h.time DESC NULLS LAST`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.TournamentHistory
	for rows.Next() {
		h, err := scanHistory(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *h)
	}
	return out, rows.Err()
}

func (r *TournamentRepo) UpdateHistory(ctx context.Context, h *domain.TournamentHistory) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE tournament_history SET
			date = $2, time = $3, tournament_name = $4, location = $5, buyin = $6, reentry_buyin = $7,
			total_revenue = $8, participants_count = $9
		WHERE id = $1`,
		h.ID, h.Date, h.Time, h.TournamentName, h.Location, h.Buyin, h.ReentryBuyin,
		h.TotalRevenue, h.ParticipantsCount,
	)
	return err
}

func (r *TournamentRepo) DeleteHistory(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM tournament_history WHERE id = $1`, id)
	return err
}

func (r *TournamentRepo) AddTournamentParticipant(ctx context.Context, p *domain.TournamentParticipant) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO tournament_participants (tournament_history_id, user_id, username, first_name, last_name,
			entries, rebuys, addons, total_spent, payment_method, position, final_points)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING id`,
		p.TournamentHistoryID, p.UserID, p.Username, p.FirstName, p.LastName,
		p.Entries, p.Rebuys, p.Addons, p.TotalSpent, p.PaymentMethod, p.Position, p.FinalPoints,
	).Scan(&p.ID)
}

func (r *TournamentRepo) ListTournamentParticipants(ctx context.Context, historyID int64) ([]domain.TournamentParticipant, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tournament_history_id, user_id, username, first_name, last_name,
			entries, rebuys, addons, total_spent, payment_method, position, final_points
		FROM tournament_participants WHERE tournament_history_id = $1`, historyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.TournamentParticipant
	for rows.Next() {
		var p domain.TournamentParticipant
		var pm *string
		var pos *int
		err := rows.Scan(
			&p.ID, &p.TournamentHistoryID, &p.UserID, &p.Username, &p.FirstName, &p.LastName,
			&p.Entries, &p.Rebuys, &p.Addons, &p.TotalSpent, &pm, &pos, &p.FinalPoints,
		)
		p.PaymentMethod = pm
		p.Position = pos
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
