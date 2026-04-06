package domain

import "time"

type TournamentHistory struct {
	ID                int64
	GameID            int64
	Date              time.Time
	Time              *time.Time
	TournamentName    string
	Location          string
	Buyin             int
	ReentryBuyin      *int
	TotalRevenue      int
	ParticipantsCount int
	CompletedAt       time.Time
	Participants      []TournamentParticipant
}

type TournamentParticipant struct {
	ID                  int64
	TournamentHistoryID int64
	UserID              string
	Username            string
	FirstName           string
	LastName            string
	Entries             int
	Rebuys              int
	Addons              int
	TotalSpent          int
	PaymentMethod       *string
}
