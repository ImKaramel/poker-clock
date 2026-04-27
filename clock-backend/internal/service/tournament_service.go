package service

import (
	"context"
	"errors"

	"backend/internal/domain"
	"backend/internal/repository"
	"backend/internal/timer"
)

var (
	ErrTournamentNotFound = errors.New("tournament not found")
	ErrLevelNotFound      = errors.New("level not found")
)

type TournamentService struct {
	repo  repository.TournamentRepository
	timer timer.Manager
}

func NewTournamentService(repo repository.TournamentRepository, timerManager timer.Manager) *TournamentService {
	return &TournamentService{
		repo:  repo,
		timer: timerManager,
	}
}

func (s *TournamentService) CreateTournament(ctx context.Context, name string) (domain.Tournament, error) {
	t := &domain.Tournament{
		Name:   name,
		Levels: nil,
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return domain.Tournament{}, err
	}

	return *t, nil
}

func (s *TournamentService) ListTournaments(ctx context.Context) ([]domain.Tournament, error) {
	return s.repo.List(ctx)
}

func (s *TournamentService) GetTournament(ctx context.Context, id string) (domain.Tournament, error) {
	t, err := s.repo.Get(ctx, id)
	if err != nil {
		return domain.Tournament{}, err
	}
	if t == nil {
		return domain.Tournament{}, ErrTournamentNotFound
	}

	return *t, nil
}

func (s *TournamentService) AddLevel(ctx context.Context, tournamentID string, level domain.Level) (domain.Tournament, error) {
	t, err := s.repo.Get(ctx, tournamentID)
	if err != nil {
		return domain.Tournament{}, err
	}
	if t == nil {
		return domain.Tournament{}, ErrTournamentNotFound
	}

	level.Order = len(t.Levels) + 1
	if err := s.repo.AddLevel(ctx, tournamentID, &level); err != nil {
		return domain.Tournament{}, err
	}

	t.Levels = append(t.Levels, level)
	return *t, nil
}

func (s *TournamentService) ListLevels(ctx context.Context, tournamentID string) ([]domain.Level, error) {
	t, err := s.repo.Get(ctx, tournamentID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, ErrTournamentNotFound
	}

	return s.repo.ListLevels(ctx, tournamentID)
}

func (s *TournamentService) StartTournament(ctx context.Context, tournamentID string) (domain.Tournament, error) {
	return domain.Tournament{}, errors.New("StartTournament is now handled by TimerService")
}

func (s *TournamentService) PauseTournament(ctx context.Context, tournamentID string) (domain.Tournament, error) {
	return domain.Tournament{}, errors.New("PauseTournament is now handled by TimerService")
}

func (s *TournamentService) ResumeTournament(ctx context.Context, tournamentID string) (domain.Tournament, error) {
	return domain.Tournament{}, errors.New("ResumeTournament is now handled by TimerService")
}

func (s *TournamentService) NextLevel(ctx context.Context, tournamentID string) (domain.Tournament, error) {
	return domain.Tournament{}, errors.New("NextLevel is now handled by TimerService")
}

func (s *TournamentService) DeleteLevel(ctx context.Context, tournamentID, levelID string) error {
	t, err := s.repo.Get(ctx, tournamentID)
	if err != nil {
		return err
	}
	if t == nil {
		return ErrTournamentNotFound
	}
	levels, err := s.repo.ListLevels(ctx, tournamentID)
	if err != nil {
		return err
	}

	levelExists := false
	for _, level := range levels {
		if level.ID == levelID {
			levelExists = true
			break
		}
	}

	if !levelExists {
		return ErrLevelNotFound
	}
	return s.repo.DeleteLevel(ctx, tournamentID, levelID)
}

func (s *TournamentService) DeleteTournament(ctx context.Context, id string) error {
	t, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if t == nil {
		return ErrTournamentNotFound
	}
	if s.timer != nil {
		s.timer.CleanupTimer(id)
	}

	return s.repo.Delete(ctx, id)
}
