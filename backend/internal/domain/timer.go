package domain

import "time"

type TimerState struct {
	TournamentID      string
	CurrentLevelIndex int
	RemainingSeconds  int
	LevelStartedAt    time.Time
	State             string
}
