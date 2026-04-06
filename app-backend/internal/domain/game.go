package domain

import "time"

type Game struct {
	GameID                   int64
	Date                     time.Time
	Time                     time.Time
	Description              string
	Buyin                    float64
	ReentryBuyin             float64
	Location                 string
	Photo                    *string
	IsActive                 bool
	Completed                bool
	CompletedAt              *time.Time
	CreatedAt                time.Time
	BasePoints               int
	PointsPerExtraPlayer     int
	MinPlayersForExtraPoints int
}
