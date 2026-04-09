package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/pridecrm/app-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	var nick, fn, ln, phone, email *string
	var dob *time.Time
	var lastLogin *time.Time
	var photoURL *string
	err := row.Scan(
		&u.UserID, &u.Password, &lastLogin, &u.IsSuperuser,
		&u.Username, &nick, &fn, &ln, &phone, &email, &dob,
		&u.Points, &u.TotalGamesPlayed, &u.IsAdmin, &u.IsStaff, &u.IsActive, &u.IsBanned,
		&u.CreatedAt, &u.UpdatedAt, &photoURL,
	)
	u.LastLogin = lastLogin
	u.NickName = nick
	u.FirstName = fn
	u.LastName = ln
	u.PhoneNumber = phone
	u.Email = email
	u.DateOfBirth = dob
	u.PhotoURL = photoURL
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	_, err := r.pool.Exec(ctx,
		`
		INSERT INTO users (
		  user_id, password, last_login, is_superuser, username,
		  nick_name, first_name, last_name,
		  phone_number, email, date_of_birth,
		  points, total_games_played, is_admin, is_staff, is_active, is_banned,
		  photo_url
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
		`,
		u.UserID, u.Password, u.LastLogin, u.IsSuperuser, u.Username, u.NickName, u.FirstName, u.LastName,
		u.PhoneNumber, u.Email, u.DateOfBirth, u.Points, u.TotalGamesPlayed, u.IsAdmin, u.IsStaff, u.IsActive, u.IsBanned,
		u.PhotoURL,
	)
	return err
}

func (r *UserRepo) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT user_id, password, last_login, is_superuser, username, nick_name, first_name, last_name,
			phone_number, email, date_of_birth, points, total_games_played, is_admin, is_staff, is_active, is_banned,
			created_at, updated_at, photo_url
		FROM users WHERE user_id = $1`, userID)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return u, err
}

func (r *UserRepo) Update(ctx context.Context, u *domain.User) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET
			password = $2, last_login = $3, is_superuser = $4, username = $5, nick_name = $6,
			first_name = $7, last_name = $8, phone_number = $9, email = $10, date_of_birth = $11,
			points = $12, total_games_played = $13, is_admin = $14, is_staff = $15, is_active = $16, is_banned = $17,
			updated_at = NOW(), photo_url = $18
		WHERE user_id = $1`,
		u.UserID, u.Password, u.LastLogin, u.IsSuperuser, u.Username, u.NickName, u.FirstName, u.LastName,
		u.PhoneNumber, u.Email, u.DateOfBirth, u.Points, u.TotalGamesPlayed, u.IsAdmin, u.IsStaff, u.IsActive, u.IsBanned,
		u.PhotoURL,
	)
	return err
}

func (r *UserRepo) Delete(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE user_id = $1`, userID)
	return err
}

func (r *UserRepo) List(ctx context.Context) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT user_id, password, last_login, is_superuser, username, nick_name, first_name, last_name,
			phone_number, email, date_of_birth, points, total_games_played, is_admin, is_staff, is_active, is_banned,
			created_at, updated_at, photo_url
		FROM users ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUserRows(rows)
}

func scanUserRows(rows pgx.Rows) ([]domain.User, error) {
	var out []domain.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *u)
	}
	return out, rows.Err()
}

func (r *UserRepo) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}

func (r *UserRepo) CountBanned(ctx context.Context) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE is_banned = TRUE`).Scan(&n)
	return n, err
}

func (r *UserRepo) ListRecent(ctx context.Context, limit int) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT user_id, password, last_login, is_superuser, username, nick_name, first_name, last_name,
			phone_number, email, date_of_birth, points, total_games_played, is_admin, is_staff, is_active, is_banned,
			created_at, updated_at, photo_url
		FROM users ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUserRows(rows)
}

func (r *UserRepo) ListForRating(ctx context.Context) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT user_id, password, last_login, is_superuser, username, nick_name, first_name, last_name,
			phone_number, email, date_of_birth, points, total_games_played, is_admin, is_staff, is_active, is_banned,
			created_at, updated_at, photo_url
		FROM users WHERE is_banned = FALSE ORDER BY points DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUserRows(rows)
}
