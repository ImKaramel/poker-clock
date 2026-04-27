package timer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"backend/internal/domain"
	"backend/internal/repository"
)

type LevelView struct {
	Type            string `json:"type"` // "level" | "break"
	Name            string `json:"name"`
	SmallBlind      int    `json:"small_blind"`
	BigBlind        int    `json:"big_blind"`
	DurationMinutes int    `json:"duration_minutes"`
}

type ViewState struct {
	LevelNumber      int        `json:"level"`
	SmallBlind       int        `json:"small_blind"`
	BigBlind         int        `json:"big_blind"`
	RemainingSeconds int        `json:"remaining_seconds"`
	CurrentType      string     `json:"current_type"`
	CurrentName      string     `json:"current_name"`
	NextLevel        *LevelView `json:"next_level"`

	PlayersCount int `json:"players_count"`
	TotalChips   int `json:"total_chips"`
}

type TournamentRuntimeData struct {
	PlayersCount int
	TotalChips   int
}

type Manager interface {
	StartTournamentTimer(ctx context.Context, tournamentID string) error
	PauseTournamentTimer(tournamentID string) error
	ResumeTournamentTimer(tournamentID string) error
	NextLevel(tournamentID string) error
	GetState(ctx context.Context, tournamentID string) (ViewState, error)
	Subscribe(tournamentID string) (<-chan ViewState, func(), error)
	UpdateStats(tournamentID string, players int, chips int)
	CleanupTimer(tournamentID string)
	Stop()
}

type manager struct {
	tournamentRepo repository.TournamentRepository
	timerRepo      repository.TimerRepository

	mu          sync.RWMutex
	timers      map[string]*tournamentTimer
	runtimeData map[string]TournamentRuntimeData

	subsMu      sync.RWMutex
	subscribers map[string]map[chan ViewState]struct{}

	ctx    context.Context
	cancel context.CancelFunc
}

func NewManager(parent context.Context, tournaments repository.TournamentRepository, timers repository.TimerRepository) Manager {
	ctx, cancel := context.WithCancel(parent)
	m := &manager{
		tournamentRepo: tournaments,
		timerRepo:      timers,
		timers:         make(map[string]*tournamentTimer),
		runtimeData:    make(map[string]TournamentRuntimeData),
		subscribers:    make(map[string]map[chan ViewState]struct{}),
		ctx:            ctx,
		cancel:         cancel,
	}

	// Restore active timers on startup
	go m.RestoreTimers(ctx)

	return m
}

func (m *manager) UpdateStats(tournamentID string, players int, chips int) {
	m.mu.Lock()
	m.runtimeData[tournamentID] = TournamentRuntimeData{
		PlayersCount: players,
		TotalChips:   chips,
	}
	m.mu.Unlock()

	state, err := m.GetState(context.Background(), tournamentID)
	if err == nil {
		m.publish(tournamentID, state)
	}

}

func (m *manager) GetStats(tournamentID string) TournamentRuntimeData {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.runtimeData[tournamentID]
}

func (m *manager) StartTournamentTimer(ctx context.Context, tournamentID string) error {
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

	m.mu.Lock()
	if tt, exists := m.timers[tournamentID]; exists {
		m.mu.Unlock()
		state := tt.state()
		if state.RemainingSeconds > 0 {
			return fmt.Errorf("timer for tournament %s is already running", tournamentID)
		}
		m.mu.Lock()
		delete(m.timers, tournamentID)
		m.mu.Unlock()
	} else {
		m.mu.Unlock()
	}

	now := time.Now()

	timerState, err := m.timerRepo.GetTimerState(ctx, tournamentID)
	if err != nil {
		return fmt.Errorf("get timer state: %w", err)
	}

	timerState = &domain.TimerState{
		TournamentID:      tournamentID,
		CurrentLevelIndex: 0,
		RemainingSeconds:  t.Levels[0].DurationMinutes * 60,
		LevelStartedAt:    now,
		State:             "running",
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
		m.onTournamentCompleted,
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

	fmt.Println("PAUSE DEBUG:", tournamentID, "exists:", ok)

	if !ok {
		return fmt.Errorf("timer for tournament %s not found", tournamentID)
	}

	return tt.pause()
}

func (m *manager) ResumeTournamentTimer(tournamentID string) error {
	m.mu.RLock()
	tt, ok := m.timers[tournamentID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("timer for tournament %s not found", tournamentID)
	}
	return tt.resume()
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
	stats := m.GetStats(tournamentID)
	nextLevel := m.getNextLevel(t.Levels, levelIdx)
	return ViewState{
		LevelNumber:      levelNumber,
		SmallBlind:       sb,
		BigBlind:         bb,
		RemainingSeconds: timerState.RemainingSeconds,
		NextLevel:        nextLevel,
		PlayersCount:     stats.PlayersCount,
		TotalChips:       stats.TotalChips,
	}, nil
}

func (m *manager) getNextLevel(levels []domain.Level, currentLevelIdx int) *LevelView {
	nextLevelIdx := currentLevelIdx + 1
	if nextLevelIdx >= len(levels) {
		return nil
	}

	nextLevel := levels[nextLevelIdx]
	return &LevelView{
		Type:            nextLevel.Type,
		Name:            nextLevel.Name,
		SmallBlind:      nextLevel.SmallBlind,
		BigBlind:        nextLevel.BigBlind,
		DurationMinutes: nextLevel.DurationMinutes,
	}
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
	subs := make([]chan ViewState, 0, len(m.subscribers[tournamentID]))
	for ch := range m.subscribers[tournamentID] {
		subs = append(subs, ch)
	}
	m.subsMu.RUnlock()

	// Fan-out to each subscriber in a separate goroutine to prevent blocking
	for _, ch := range subs {
		go func(subscriber chan ViewState) {
			select {
			case subscriber <- state:
			default:
				// Skip slow subscribers
			}
		}(ch)
	}
}

func (m *manager) RestoreTimers(ctx context.Context) {
	timerStates, err := m.timerRepo.GetAllTimerStates(ctx)
	if err != nil {
		// Log error but don't fail startup
		return
	}

	for _, timerState := range timerStates {
		if timerState.State != "running" && timerState.State != "paused" {
			continue
		}

		t, err := m.tournamentRepo.Get(ctx, timerState.TournamentID)
		if err != nil || t == nil {
			continue // Skip invalid tournaments
		}
		if len(t.Levels) == 0 {
			continue // Skip tournaments without levels
		}
		if timerState.CurrentLevelIndex >= len(t.Levels) {
			resetState := &domain.TimerState{
				TournamentID:      timerState.TournamentID,
				CurrentLevelIndex: 0,
				RemainingSeconds:  t.Levels[0].DurationMinutes * 60,
				LevelStartedAt:    time.Time{},
				State:             "stopped",
			}
			m.timerRepo.UpdateTimerState(ctx, resetState)
			continue
		}

		if timerState.CurrentLevelIndex < 0 {
			timerState.CurrentLevelIndex = 0
		}
		if timerState.RemainingSeconds <= 0 {
			timerState.RemainingSeconds = t.Levels[timerState.CurrentLevelIndex].DurationMinutes * 60
		}

		// Create and start the timer
		tt := newTournamentTimer(m.ctx, timerState.TournamentID, t.Levels, timerState.CurrentLevelIndex, timerState.RemainingSeconds, timerState.State, timerState.LevelStartedAt,
			func(state ViewState) {
				m.publish(timerState.TournamentID, state)
			},
			func(s *domain.TimerState) error {
				return m.timerRepo.UpdateTimerState(context.Background(), s)
			},
			m.onTournamentCompleted,
		)

		m.mu.Lock()
		m.timers[timerState.TournamentID] = tt
		m.mu.Unlock()

		go tt.run()
	}
}

func (m *manager) onTournamentCompleted(tournamentID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.timers, tournamentID)

	go func() {
		ctx := context.Background()
		t, err := m.tournamentRepo.Get(ctx, tournamentID)
		if err != nil || t == nil || len(t.Levels) == 0 {
			return
		}

		resetState := &domain.TimerState{
			TournamentID:      tournamentID,
			CurrentLevelIndex: 0,
			RemainingSeconds:  t.Levels[0].DurationMinutes * 60,
			LevelStartedAt:    time.Time{},
			State:             "stopped",
		}

		m.timerRepo.UpdateTimerState(ctx, resetState)
	}()
}

func (m *manager) CleanupTimer(tournamentID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if tt, exists := m.timers[tournamentID]; exists {
		tt.stop()
		delete(m.timers, tournamentID)
	}
	delete(m.runtimeData, tournamentID)

	m.subsMu.Lock()
	delete(m.subscribers, tournamentID)
	m.subsMu.Unlock()

	go func() {
		ctx := context.Background()
		m.timerRepo.UpdateTimerState(ctx, &domain.TimerState{
			TournamentID:      tournamentID,
			CurrentLevelIndex: 0,
			RemainingSeconds:  0,
			LevelStartedAt:    time.Time{},
			State:             "stopped",
		})
	}()
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
