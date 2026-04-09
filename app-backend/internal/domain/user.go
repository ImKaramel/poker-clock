package domain

import "time"

type User struct {
	UserID           string
	Password         string
	LastLogin        *time.Time
	IsSuperuser      bool
	Username         string
	NickName         *string
	FirstName        *string
	LastName         *string
	PhoneNumber      *string
	Email            *string
	DateOfBirth      *time.Time
	Points           int
	TotalGamesPlayed int
	IsAdmin          bool
	IsStaff          bool
	IsActive         bool
	IsBanned         bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	PhotoURL         *string
}
