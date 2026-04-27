package service

import (
	"context"
	"errors"
	_ "fmt"
	"strings"

	"backend/internal/repository"
	"backend/internal/timer"
)

var (
	ErrTournamentNotRunning     = errors.New("tournament is not running")
	ErrTournamentNotPaused      = errors.New("tournament is not paused")
	ErrTournamentAlreadyRunning = errors.New("tournament is already running")
	ErrTournamentNoLevels       = errors.New("tournament has no levels")
)

type TimerService struct {
	tournaments repository.TournamentRepository
	timers      repository.TimerRepository
	manager     timer.Manager
}

func NewTimerService(
	tournaments repository.TournamentRepository,
	timers repository.TimerRepository,
	manager timer.Manager,
) *TimerService {
	return &TimerService{
		tournaments: tournaments,
		timers:      timers,
		manager:     manager,
	}
}

func (s *TimerService) Start(ctx context.Context, tournamentID string) error {
	t, err := s.tournaments.Get(ctx, tournamentID)
	if err != nil {
		return err
	}
	if t == nil {
		return ErrTournamentNotFound
	}
	if len(t.Levels) == 0 {
		return ErrTournamentNoLevels
	}

	err = s.manager.StartTournamentTimer(ctx, tournamentID)
	return s.convertTimerError(err)
}

func (s *TimerService) Pause(tournamentID string) error {
	err := s.manager.PauseTournamentTimer(tournamentID)
	if err != nil {
		return s.convertTimerError(err)
	}
	return nil
}

func (s *TimerService) Resume(tournamentID string) error {
	err := s.manager.ResumeTournamentTimer(tournamentID)
	if err != nil {
		return s.convertTimerError(err)
	}
	return nil
}

func (s *TimerService) NextLevel(tournamentID string) error {
	return s.manager.NextLevel(tournamentID)
}

type TimerViewState = timer.ViewState

func (s *TimerService) GetState(ctx context.Context, tournamentID string) (TimerViewState, error) {
	return s.manager.GetState(ctx, tournamentID)
}

func (s *TimerService) Subscribe(tournamentID string) (<-chan TimerViewState, func(), error) {
	return s.manager.Subscribe(tournamentID)
}

func (s *TimerService) UpdateStats(tournamentID string, players, chips int) {
	s.manager.UpdateStats(tournamentID, players, chips)
}

func (s *TimerService) convertTimerError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Map timer errors to service sentinel errors
	if strings.Contains(errStr, "timer not found") {
		return ErrTournamentNotFound
	}
	if strings.Contains(errStr, "already running") {
		return ErrTournamentAlreadyRunning
	}
	if strings.Contains(errStr, "cannot pause timer") || strings.Contains(errStr, "timer is not running") {
		return ErrTournamentNotRunning
	}
	if strings.Contains(errStr, "cannot resume timer") || strings.Contains(errStr, "timer is not paused") {
		return ErrTournamentNotPaused
	}
	if strings.Contains(errStr, "has no levels") {
		return ErrTournamentNoLevels
	}
	if strings.Contains(errStr, "tournament not found") {
		return ErrTournamentNotFound
	}

	// Return original error if no mapping found
	return err
}
