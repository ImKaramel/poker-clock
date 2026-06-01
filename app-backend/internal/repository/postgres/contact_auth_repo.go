package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pridecrm/app-backend/internal/domain"
)

type ContactAuthRepo struct {
	pool *pgxpool.Pool
}

func NewContactAuthRepo(pool *pgxpool.Pool) *ContactAuthRepo {
	return &ContactAuthRepo{pool: pool}
}

func scanContactAuth(row pgx.Row) (*domain.ContactAuthChallenge, error) {
	var c domain.ContactAuthChallenge
	err := row.Scan(
		&c.Token,
		&c.Status,
		&c.TelegramUserID,
		&c.PhoneNumber,
		&c.CreatedAt,
		&c.ExpiresAt,
		&c.CompletedAt,
		&c.ConsumedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ContactAuthRepo) Create(ctx context.Context, challenge *domain.ContactAuthChallenge) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO contact_auth_challenges (token, status, expires_at)
		VALUES ($1, $2, $3)
		RETURNING created_at`,
		challenge.Token,
		challenge.Status,
		challenge.ExpiresAt,
	).Scan(&challenge.CreatedAt)
}

func (r *ContactAuthRepo) GetByToken(ctx context.Context, token string) (*domain.ContactAuthChallenge, error) {
	return scanContactAuth(r.pool.QueryRow(ctx, `
		SELECT token, status, telegram_user_id, phone_number, created_at, expires_at, completed_at, consumed_at
		FROM contact_auth_challenges
		WHERE token = $1`, token))
}

func (r *ContactAuthRepo) Complete(ctx context.Context, token string, telegramUserID string, phoneNumber string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE contact_auth_challenges
		SET status = 'completed',
			telegram_user_id = $2,
			phone_number = $3,
			completed_at = NOW()
		WHERE token = $1
			AND status = 'pending'
			AND consumed_at IS NULL
			AND expires_at > NOW()`,
		token,
		telegramUserID,
		phoneNumber,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *ContactAuthRepo) Consume(ctx context.Context, token string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE contact_auth_challenges
		SET status = 'consumed',
			consumed_at = NOW()
		WHERE token = $1
			AND status = 'completed'
			AND consumed_at IS NULL`,
		token,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
