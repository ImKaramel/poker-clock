package services

import "github.com/pridecrm/app-backend/internal/domain"

// Clock is a placeholder integration for hardware/sync clocks (e.g. participant sync to devices).
type Clock struct{}

// SyncParticipants reserved for future clock/device integration; intentionally empty.
func (c *Clock) SyncParticipants(gameID string, participants []domain.Participant) {}
