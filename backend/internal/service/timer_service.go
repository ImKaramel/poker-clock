package service

import (
	"context"
	"errors"

	"backend/internal/repository"
	"backend/internal/timer"
)

var (
	ErrTournamentNotRunning = errors.New("tournament is not running")
	ErrTournamentNotPaused  = errors.New("tournament is not paused")
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
		return errors.New("tournament has no levels")
	}
	return s.manager.StartTournamentTimer(ctx, tournamentID)
}

func (s *TimerService) Pause(tournamentID string) error {
	return s.manager.PauseTournamentTimer(tournamentID)
}

func (s *TimerService) Resume(tournamentID string) error {
	return s.manager.ResumeTournamentTimer(tournamentID)
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
