package postgres

import (
	"context"
	"fmt"
)

func RunMigrations(ctx context.Context, db *DB) error {
	queries := []string{
		// tournaments table
		`CREATE TABLE IF NOT EXISTS tournaments (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			state TEXT NOT NULL DEFAULT 'stopped',
			current_level_index INT NOT NULL DEFAULT -1,
			level_started_at TIMESTAMP,
			remaining_seconds INT,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		// levels table
		`CREATE TABLE IF NOT EXISTS levels (
			id UUID PRIMARY KEY,
			tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
			small_blind INT NOT NULL,
			big_blind INT NOT NULL,
			duration_minutes INT NOT NULL,
			level_order INT NOT NULL
		)`,
		// indexes
		`CREATE INDEX IF NOT EXISTS idx_levels_tournament_id ON levels(tournament_id)`,
	}

	for _, q := range queries {
		if _, err := db.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("run migration %q: %w", q, err)
		}
	}

	return nil
}
