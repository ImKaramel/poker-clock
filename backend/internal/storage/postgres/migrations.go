package postgres

import (
	"context"
)

func RunMigrations(ctx context.Context, db *DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS tournaments (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			owner_id UUID NOT NULL,
			current_level_index INT NOT NULL DEFAULT 0,
			state TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS levels (
			id UUID PRIMARY KEY,
			tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
			small_blind INT NOT NULL,
			big_blind INT NOT NULL,
			duration_minutes INT NOT NULL,
			level_order INT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_levels_tournament_id ON levels(tournament_id)`,
	}

	for _, q := range queries {
		if _, err := db.ExecContext(ctx, q); err != nil {
			return err
		}
	}

	return nil
}
