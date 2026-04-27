package domain

type Level struct {
	ID              string
	Type            string // "level" | "break"
	Name            string
	SmallBlind      int
	BigBlind        int
	DurationMinutes int
	Order           int
}
