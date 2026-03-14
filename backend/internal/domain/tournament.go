package domain

type Tournament struct {
	ID                string
	Name              string
	OwnerID           string
	Levels            []Level
	CurrentLevelIndex int
	State             string
}

const (
	TournamentStateStopped = "stopped"
	TournamentStateRunning = "running"
	TournamentStatePaused  = "paused"
)
