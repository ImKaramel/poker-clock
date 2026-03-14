package service

import (
	"context"
	"errors"

	"backend/internal/domain"
	"backend/internal/repository"
)

var (
	ErrTournamentNotFound   = errors.New("tournament not found")
	ErrTournamentNotRunning = errors.New("tournament is not running")
	ErrTournamentNotPaused  = errors.New("tournament is not paused")
)

type TournamentService struct {
	repo repository.TournamentRepository
}

func NewTournamentService(repo repository.TournamentRepository) *TournamentService {
	return &TournamentService{repo: repo}
}

func (s *TournamentService) CreateTournament(ctx context.Context, name, ownerID string) (domain.Tournament, error) {
	t := &domain.Tournament{
		Name:              name,
		OwnerID:           ownerID,
		Levels:            nil,
		CurrentLevelIndex: -1,
		State:             domain.TournamentStateStopped,
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
	t, err := s.repo.Get(ctx, tournamentID)
	if err != nil {
		return domain.Tournament{}, err
	}
	if t == nil {
		return domain.Tournament{}, ErrTournamentNotFound
	}

	if err := s.repo.UpdateCurrentLevelIndex(ctx, tournamentID, 0); err != nil {
		return domain.Tournament{}, err
	}
	if err := s.repo.UpdateState(ctx, tournamentID, domain.TournamentStateRunning); err != nil {
		return domain.Tournament{}, err
	}

	t.CurrentLevelIndex = 0
	t.State = domain.TournamentStateRunning
	return *t, nil
}

func (s *TournamentService) PauseTournament(ctx context.Context, tournamentID string) (domain.Tournament, error) {
	t, err := s.repo.Get(ctx, tournamentID)
	if err != nil {
		return domain.Tournament{}, err
	}
	if t == nil {
		return domain.Tournament{}, ErrTournamentNotFound
	}

	if t.State != domain.TournamentStateRunning {
		return domain.Tournament{}, ErrTournamentNotRunning
	}

	if err := s.repo.UpdateState(ctx, tournamentID, domain.TournamentStatePaused); err != nil {
		return domain.Tournament{}, err
	}

	t.State = domain.TournamentStatePaused
	return *t, nil
}

func (s *TournamentService) ResumeTournament(ctx context.Context, tournamentID string) (domain.Tournament, error) {
	t, err := s.repo.Get(ctx, tournamentID)
	if err != nil {
		return domain.Tournament{}, err
	}
	if t == nil {
		return domain.Tournament{}, ErrTournamentNotFound
	}

	if t.State != domain.TournamentStatePaused {
		return domain.Tournament{}, ErrTournamentNotPaused
	}

	if err := s.repo.UpdateState(ctx, tournamentID, domain.TournamentStateRunning); err != nil {
		return domain.Tournament{}, err
	}

	t.State = domain.TournamentStateRunning
	return *t, nil
}

func (s *TournamentService) NextLevel(ctx context.Context, tournamentID string) (domain.Tournament, error) {
	t, err := s.repo.Get(ctx, tournamentID)
	if err != nil {
		return domain.Tournament{}, err
	}
	if t == nil {
		return domain.Tournament{}, ErrTournamentNotFound
	}

	if len(t.Levels) == 0 {
		t.State = domain.TournamentStateStopped
		_ = s.repo.UpdateState(ctx, tournamentID, domain.TournamentStateStopped)
		return *t, nil
	}

	if t.CurrentLevelIndex+1 >= len(t.Levels) {
		t.State = domain.TournamentStateStopped
		_ = s.repo.UpdateState(ctx, tournamentID, domain.TournamentStateStopped)
		return *t, nil
	}

	nextIndex := t.CurrentLevelIndex + 1
	if err := s.repo.UpdateCurrentLevelIndex(ctx, tournamentID, nextIndex); err != nil {
		return domain.Tournament{}, err
	}

	t.CurrentLevelIndex = nextIndex
	return *t, nil
}
