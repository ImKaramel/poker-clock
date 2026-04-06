package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pridecrm/app-backend/internal/domain"
)

type SupportRepo struct {
	pool *pgxpool.Pool
}

func NewSupportRepo(pool *pgxpool.Pool) *SupportRepo {
	return &SupportRepo{pool: pool}
}

func scanTicket(row pgx.Row) (*domain.SupportTicket, error) {
	var t domain.SupportTicket
	err := row.Scan(&t.ID, &t.UserID, &t.Subject, &t.Message, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *SupportRepo) Create(ctx context.Context, t *domain.SupportTicket) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO support_tickets (user_id, subject, message, status)
		VALUES ($1,$2,$3,$4) RETURNING id, created_at, updated_at`,
		t.UserID, t.Subject, t.Message, t.Status,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *SupportRepo) GetByID(ctx context.Context, id int64) (*domain.SupportTicket, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, subject, message, status, created_at, updated_at
		FROM support_tickets WHERE id = $1`, id)
	t, err := scanTicket(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return t, err
}

func (r *SupportRepo) Update(ctx context.Context, t *domain.SupportTicket) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE support_tickets SET subject = $2, message = $3, status = $4, updated_at = NOW()
		WHERE id = $1`,
		t.ID, t.Subject, t.Message, t.Status,
	)
	return err
}

func (r *SupportRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM support_tickets WHERE id = $1`, id)
	return err
}

func (r *SupportRepo) ListByUser(ctx context.Context, userID string) ([]domain.SupportTicket, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, subject, message, status, created_at, updated_at
		FROM support_tickets WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.SupportTicket
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

func (r *SupportRepo) ListAll(ctx context.Context) ([]domain.SupportTicket, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, subject, message, status, created_at, updated_at
		FROM support_tickets ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.SupportTicket
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

func (r *SupportRepo) CountOpen(ctx context.Context) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM support_tickets WHERE status = 'open'`).Scan(&n)
	return n, err
}
