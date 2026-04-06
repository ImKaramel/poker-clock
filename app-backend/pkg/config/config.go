package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr        string
	DatabaseURL string
	JWTSecret   string
	JWTTTL      time.Duration

	TelegramBotToken    string
	TelegramBotUsername string
	MiniAppURL          string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ttl := 7 * 24 * time.Hour
	if s := os.Getenv("JWT_TTL_HOURS"); s != "" {
		if h, err := strconv.Atoi(s); err == nil && h > 0 {
			ttl = time.Duration(h) * time.Hour
		}
	}

	return Config{
		Addr:        ":" + port,
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		JWTTTL:      ttl,

		TelegramBotToken:    os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramBotUsername: os.Getenv("TELEGRAM_BOT_USERNAME"),
		MiniAppURL:          os.Getenv("MINI_APP_URL"),
	}
}
