package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr        string
	DatabaseURL string
	JWTSecret   string
	JWTTTL      time.Duration

	TelegramBotToken    string
	FrontendURL         string
	TelegramBotUsername string
	MiniAppURL          string
	AdminTelegramIDs    []string
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

	adminIDsStr := os.Getenv("ADMIN_TELEGRAM_IDS")
	var adminIDs []string
	if adminIDsStr != "" {
		parts := strings.Split(adminIDsStr, ",")
		for _, part := range parts {
			id := strings.TrimSpace(part)
			if id != "" {
				adminIDs = append(adminIDs, id)
			}
		}
	}

	return Config{
		Addr:        ":" + port,
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		JWTTTL:      ttl,

		TelegramBotToken:    os.Getenv("TELEGRAM_BOT_TOKEN"),
		FrontendURL:         os.Getenv("FRONTEND_URL"),
		TelegramBotUsername: os.Getenv("TELEGRAM_BOT_USERNAME"),
		MiniAppURL:          os.Getenv("MINI_APP_URL"),
		AdminTelegramIDs:    adminIDs,
	}
}
