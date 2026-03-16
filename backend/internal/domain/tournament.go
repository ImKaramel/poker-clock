package domain

import "time"

type Tournament struct {
	ID        string
	Name      string
	OwnerID   string
	Levels    []Level
	CreatedAt time.Time
}
