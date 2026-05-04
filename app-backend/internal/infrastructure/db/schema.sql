

CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(64) PRIMARY KEY,
    password VARCHAR(128) NOT NULL DEFAULT '',
    last_login TIMESTAMPTZ,
    is_superuser BOOLEAN NOT NULL DEFAULT FALSE,
    username VARCHAR(100) NOT NULL,
    nick_name VARCHAR(100),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone_number VARCHAR(20),
    email VARCHAR(254),
    date_of_birth DATE,
    points INT NOT NULL DEFAULT 0,
    total_games_played INT NOT NULL DEFAULT 0,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    is_staff BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_banned BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE users ADD COLUMN IF NOT EXISTS photo_url TEXT;
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM (
            SELECT LOWER(username)
            FROM users
            GROUP BY LOWER(username)
            HAVING COUNT(*) > 1
        ) duplicate_usernames
    ) THEN
        CREATE UNIQUE INDEX IF NOT EXISTS idx_users_telegram_username_unique_ci
            ON users (LOWER(username));
    END IF;
END $$;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_nick_name_unique_ci
    ON users (LOWER(nick_name))
    WHERE nick_name IS NOT NULL;

CREATE TABLE IF NOT EXISTS games (
    game_id SERIAL PRIMARY KEY,
    date DATE NOT NULL,
    time TIME NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    buyin NUMERIC(10, 2) NOT NULL DEFAULT 0,
    reentry_buyin NUMERIC(10, 2) NOT NULL DEFAULT 0,
    location VARCHAR(255) NOT NULL DEFAULT '',
    photo VARCHAR(512),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    base_points INT NOT NULL DEFAULT 100,
    points_per_extra_player INT NOT NULL DEFAULT 10,
    min_players_for_extra_points INT NOT NULL DEFAULT 10
);
ALTER TABLE games ADD COLUMN IF NOT EXISTS name VARCHAR(255) NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS participants (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    game_id INT NOT NULL REFERENCES games(game_id) ON DELETE CASCADE,
    entries INT NOT NULL DEFAULT 1,
    rebuys INT NOT NULL DEFAULT 0,
    addons INT NOT NULL DEFAULT 0,
    final_points INT NOT NULL DEFAULT 0,
    position INT,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, game_id)
);
ALTER TABLE participants
    ADD COLUMN IF NOT EXISTS arrived BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE participants
    ADD COLUMN IF NOT EXISTS is_out BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS support_tickets (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    subject VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tournament_history (
    id BIGSERIAL PRIMARY KEY,
    game_id INT NOT NULL UNIQUE REFERENCES games(game_id) ON DELETE CASCADE,
    date DATE NOT NULL,
    time TIME,
    tournament_name VARCHAR(255) NOT NULL,
    location VARCHAR(255) NOT NULL,
    buyin INT NOT NULL DEFAULT 0,
    reentry_buyin INT,
    total_revenue INT NOT NULL DEFAULT 0,
    participants_count INT NOT NULL DEFAULT 0,
    completed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tournament_participants (
    id BIGSERIAL PRIMARY KEY,
    tournament_history_id BIGINT NOT NULL REFERENCES tournament_history(id) ON DELETE CASCADE,
    user_id VARCHAR(100) NOT NULL,
    username VARCHAR(100) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL DEFAULT '',
    entries INT NOT NULL DEFAULT 1,
    rebuys INT NOT NULL DEFAULT 0,
    addons INT NOT NULL DEFAULT 0,
    total_spent INT NOT NULL DEFAULT 0,
    payment_method VARCHAR(20)
);
ALTER TABLE tournament_participants
    ADD COLUMN IF NOT EXISTS position INT;
ALTER TABLE tournament_participants
    ADD COLUMN IF NOT EXISTS final_points INT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_participants_game ON participants(game_id);
CREATE INDEX IF NOT EXISTS idx_participants_user ON participants(user_id);
CREATE INDEX IF NOT EXISTS idx_support_tickets_user ON support_tickets(user_id);
CREATE INDEX IF NOT EXISTS idx_tournament_participants_th ON tournament_participants(tournament_history_id);
