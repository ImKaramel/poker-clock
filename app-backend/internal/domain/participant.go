package domain

import "time"

type Participant struct {
	ID          int64
	UserID      string
	GameID      int64
	Entries     int
	Rebuys      int
	Addons      int
	FinalPoints int //
	Position    *int
	JoinedAt    time.Time
	Arrived     bool
}
