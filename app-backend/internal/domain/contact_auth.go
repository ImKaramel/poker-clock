package domain

import "time"

type ContactAuthChallenge struct {
	Token          string
	Status         string
	TelegramUserID *string
	PhoneNumber    *string
	CreatedAt      time.Time
	ExpiresAt      time.Time
	CompletedAt    *time.Time
	ConsumedAt     *time.Time
}
