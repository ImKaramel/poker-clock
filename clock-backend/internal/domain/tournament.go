package domain

import "time"

type Tournament struct {
	ID        string
	Name      string
	Levels    []Level
	CreatedAt time.Time
}
