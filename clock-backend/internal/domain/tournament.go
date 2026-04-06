package domain

import "time"

type Tournament struct {
	ID        string
	Name      string
	Levels    []Level
	CreatedAt time.Time

	// Stats from App Backend
	PlayersCount int
	TotalChips   int
}
