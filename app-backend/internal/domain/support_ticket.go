package domain

import "time"

type SupportTicket struct {
	ID        int64
	UserID    string
	Subject   string
	Message   string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}
