package timer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"backend/internal/domain"
	"backend/internal/repository"
)

type ViewState struct {
	LevelNumber      int
	SmallBlind       int
	BigBlind         int
	RemainingSeconds int
}

type Manager interface {
	StartTournamentTimer(ctx context.Context, tournamentID string) error
	PauseTournamentTimer(tournamentID string) error
	ResumeTournamentTimer(tournamentID string) error
	NextLevel(tournamentID string) error
	GetState(ctx context.Context, tournamentID string) (ViewState, error)
	Subscribe(tournamentID string) (<-chan ViewState, func(), error)
	Stop()
}

type manager struct {
	tournamentRepo repository.TournamentRepository
	timerRepo      repository.TimerRepository

	mu     sync.RWMutex
	timers map[string]*tournamentTimer

	subsMu      sync.RWMutex
	subscribers map[string]map[chan ViewState]struct{}

	ctx    context.Context
	cancel context.CancelFunc
}

func NewManager(parent context.Context, tournaments repository.TournamentRepository, timers repository.TimerRepository) Manager {
	ctx, cancel := context.WithCancel(parent)
	return &manager{
		tournamentRepo: tournaments,
		timerRepo:      timers,
		timers:         make(map[string]*tournamentTimer),
		subscribers:    make(map[string]map[chan ViewState]struct{}),
		ctx:            ctx,
		cancel:         cancel,
	}
}

func (m *manager) StartTournamentTimer(ctx context.Context, tournamentID string) error {
	m.mu.Lock()
	if _, ok := m.timers[tournamentID]; ok {
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()

	t, err := m.tournamentRepo.Get(ctx, tournamentID)
	if err != nil {
		return fmt.Errorf("load tournament for timer: %w", err)
	}
	if t == nil {
		return fmt.Errorf("tournament %s not found", tournamentID)
	}
	if len(t.Levels) == 0 {
		return fmt.Errorf("tournament %s has no levels", tournamentID)
	}

	now := time.Now()

	timerState, err := m.timerRepo.GetTimerState(ctx, tournamentID)
	if err != nil {
		return fmt.Errorf("get timer state: %w", err)
	}
	if timerState == nil {
		timerState = &domain.TimerState{
			TournamentID:      tournamentID,
			CurrentLevelIndex: 0,
			RemainingSeconds:  t.Levels[0].DurationMinutes * 60,
			LevelStartedAt:    now,
			State:             "running",
		}
	} else {
		if timerState.CurrentLevelIndex < 0 {
			timerState.CurrentLevelIndex = 0
		}
		if timerState.CurrentLevelIndex >= len(t.Levels) {
			timerState.CurrentLevelIndex = len(t.Levels) - 1
		}
		if timerState.RemainingSeconds <= 0 {
			timerState.RemainingSeconds = t.Levels[timerState.CurrentLevelIndex].DurationMinutes * 60
		}
		timerState.State = "running"
		timerState.LevelStartedAt = now
	}

	if err := m.timerRepo.UpdateTimerState(ctx, timerState); err != nil {
		return fmt.Errorf("init timer state: %w", err)
	}

	tt := newTournamentTimer(m.ctx, tournamentID, t.Levels, timerState.CurrentLevelIndex, timerState.RemainingSeconds, timerState.State, timerState.LevelStartedAt,
		func(state ViewState) {
			m.publish(tournamentID, state)
		},
		func(s *domain.TimerState) error {
			return m.timerRepo.UpdateTimerState(context.Background(), s)
		},
	)

	m.mu.Lock()
	m.timers[tournamentID] = tt
	m.mu.Unlock()

	go tt.run()

	return nil
}

func (m *manager) PauseTournamentTimer(tournamentID string) error {
	m.mu.RLock()
	tt, ok := m.timers[tournamentID]
	m.mu.RUnlock()
	if !ok {
		return nil
	}
	tt.pause()
	return nil
}

func (m *manager) ResumeTournamentTimer(tournamentID string) error {
	m.mu.RLock()
	tt, ok := m.timers[tournamentID]
	m.mu.RUnlock()
	if !ok {
		return nil
	}
	tt.resume()
	return nil
}

func (m *manager) NextLevel(tournamentID string) error {
	m.mu.RLock()
	tt, ok := m.timers[tournamentID]
	m.mu.RUnlock()
	if !ok {
		return nil
	}
	return tt.nextLevel()
}

func (m *manager) GetState(ctx context.Context, tournamentID string) (ViewState, error) {
	m.mu.RLock()
	if tt, ok := m.timers[tournamentID]; ok {
		state := tt.state()
		m.mu.RUnlock()
		return state, nil
	}
	m.mu.RUnlock()

	t, err := m.tournamentRepo.Get(ctx, tournamentID)
	if err != nil {
		return ViewState{}, fmt.Errorf("get tournament for state: %w", err)
	}
	if t == nil {
		return ViewState{}, fmt.Errorf("tournament %s not found", tournamentID)
	}

	timerState, err := m.timerRepo.GetTimerState(ctx, tournamentID)
	if err != nil {
		return ViewState{}, fmt.Errorf("get timer state: %w", err)
	}
	if timerState == nil {
		return ViewState{}, nil
	}

	levelIdx := timerState.CurrentLevelIndex
	levelNumber := levelIdx + 1
	var sb, bb int
	if levelIdx >= 0 && levelIdx < len(t.Levels) {
		l := t.Levels[levelIdx]
		sb = l.SmallBlind
		bb = l.BigBlind
	}

	return ViewState{
		LevelNumber:      levelNumber,
		SmallBlind:       sb,
		BigBlind:         bb,
		RemainingSeconds: timerState.RemainingSeconds,
	}, nil
}

func (m *manager) Subscribe(tournamentID string) (<-chan ViewState, func(), error) {
	ch := make(chan ViewState, 16)

	m.subsMu.Lock()
	defer m.subsMu.Unlock()

	subs, ok := m.subscribers[tournamentID]
	if !ok {
		subs = make(map[chan ViewState]struct{})
		m.subscribers[tournamentID] = subs
	}
	subs[ch] = struct{}{}

	unsubscribe := func() {
		m.subsMu.Lock()
		defer m.subsMu.Unlock()
		if m.subscribers[tournamentID] == nil {
			return
		}
		delete(m.subscribers[tournamentID], ch)
		close(ch)
		if len(m.subscribers[tournamentID]) == 0 {
			delete(m.subscribers, tournamentID)
		}
	}

	return ch, unsubscribe, nil
}

func (m *manager) publish(tournamentID string, state ViewState) {
	m.subsMu.RLock()
	defer m.subsMu.RUnlock()

	for ch := range m.subscribers[tournamentID] {
		select {
		case ch <- state:
		default:
		}
	}
}

func (m *manager) Stop() {
	m.cancel()

	m.mu.Lock()
	defer m.mu.Unlock()
	for _, tt := range m.timers {
		tt.stop()
	}
	m.timers = make(map[string]*tournamentTimer)
}
