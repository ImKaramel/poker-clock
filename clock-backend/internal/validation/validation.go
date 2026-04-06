package validation

import (
	"errors"
	"net/mail"
	"strings"
)

var (
	ErrInvalidEmail           = errors.New("invalid email")
	ErrInvalidPassword        = errors.New("invalid password")
	ErrInvalidTournamentName  = errors.New("invalid tournament name")
	ErrInvalidBlinds          = errors.New("invalid blinds")
	ErrInvalidDurationMinutes = errors.New("invalid duration minutes")
)

func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return ErrInvalidEmail
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidEmail
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrInvalidPassword
	}
	return nil
}

func ValidateTournamentName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrInvalidTournamentName
	}
	return nil
}

func ValidateBlinds(small, big int) error {
	if small <= 0 || big <= 0 || big < small {
		return ErrInvalidBlinds
	}
	return nil
}

func ValidateDurationMinutes(d int) error {
	if d <= 0 {
		return ErrInvalidDurationMinutes
	}
	return nil
}
