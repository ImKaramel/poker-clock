package timer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"backend/internal/domain"
)

type persistFunc func(state *domain.TimerState) error

type tournamentTimer struct {
	ctx    context.Context
	id     string
	levels []domain.Level

	mu              sync.Mutex
	currentLevelIdx int
	remaining       int
	timerState      string
	levelStartedAt  time.Time

	publish TimerStatePublisher
	persist persistFunc
}

type TimerStatePublisher func(ViewState)

func newTournamentTimer(
	ctx context.Context,
	id string,
	levels []domain.Level,
	currentLevelIdx int,
	remaining int,
	state string,
	levelStartedAt time.Time,
	publish TimerStatePublisher,
	persist persistFunc,
) *tournamentTimer {
	return &tournamentTimer{
		ctx:             ctx,
		id:              id,
		levels:          levels,
		currentLevelIdx: currentLevelIdx,
		remaining:       remaining,
		timerState:      state,
		levelStartedAt:  levelStartedAt,
		publish:         publish,
		persist:         persist,
	}
}

func (t *tournamentTimer) run() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-t.ctx.Done():
			return
		case <-ticker.C:
			t.tick()
		}
	}
}

func (t *tournamentTimer) tick() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.timerState != "running" {
		return
	}

	if t.remaining > 0 {
		t.remaining--
	}

	_ = t.persist(&domain.TimerState{
		TournamentID:      t.id,
		CurrentLevelIndex: t.currentLevelIdx,
		RemainingSeconds:  t.remaining,
		LevelStartedAt:    t.levelStartedAt,
		State:             t.timerState,
	})

	state := t.currentStateLocked()
	t.publish(state)

	if t.remaining <= 0 {
		_ = t.advanceLocked()
	}
}

func (t *tournamentTimer) currentStateLocked() ViewState {
	levelNumber := t.currentLevelIdx + 1
	var sb, bb int
	if t.currentLevelIdx >= 0 && t.currentLevelIdx < len(t.levels) {
		l := t.levels[t.currentLevelIdx]
		sb = l.SmallBlind
		bb = l.BigBlind
	}

	return ViewState{
		LevelNumber:      levelNumber,
		SmallBlind:       sb,
		BigBlind:         bb,
		RemainingSeconds: t.remaining,
	}
}

func (t *tournamentTimer) pause() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.timerState != "running" {
		return fmt.Errorf("cannot pause timer: timer is not running (current state: %s)", t.timerState)
	}
	t.timerState = "paused"
	return t.persist(&domain.TimerState{
		TournamentID:      t.id,
		CurrentLevelIndex: t.currentLevelIdx,
		RemainingSeconds:  t.remaining,
		LevelStartedAt:    t.levelStartedAt,
		State:             t.timerState,
	})
}

func (t *tournamentTimer) resume() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.timerState != "paused" {
		return fmt.Errorf("cannot resume timer: timer is not paused (current state: %s)", t.timerState)
	}
	t.timerState = "running"
	return t.persist(&domain.TimerState{
		TournamentID:      t.id,
		CurrentLevelIndex: t.currentLevelIdx,
		RemainingSeconds:  t.remaining,
		LevelStartedAt:    t.levelStartedAt,
		State:             t.timerState,
	})
}

func (t *tournamentTimer) nextLevel() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.advanceLocked()
}

func (t *tournamentTimer) advanceLocked() error {
	if t.currentLevelIdx+1 >= len(t.levels) {
		t.timerState = "stopped"
		_ = t.persist(&domain.TimerState{
			TournamentID:      t.id,
			CurrentLevelIndex: t.currentLevelIdx,
			RemainingSeconds:  t.remaining,
			LevelStartedAt:    t.levelStartedAt,
			State:             t.timerState,
		})
		return nil
	}

	t.currentLevelIdx++
	t.levelStartedAt = time.Now()
	t.remaining = t.levels[t.currentLevelIdx].DurationMinutes * 60
	t.timerState = "running"

	_ = t.persist(&domain.TimerState{
		TournamentID:      t.id,
		CurrentLevelIndex: t.currentLevelIdx,
		RemainingSeconds:  t.remaining,
		LevelStartedAt:    t.levelStartedAt,
		State:             t.timerState,
	})

	state := t.currentStateLocked()
	t.publish(state)

	return nil
}

func (t *tournamentTimer) state() ViewState {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.currentStateLocked()
}

func (t *tournamentTimer) stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.timerState = "stopped"
	_ = t.persist(&domain.TimerState{
		TournamentID:      t.id,
		CurrentLevelIndex: t.currentLevelIdx,
		RemainingSeconds:  t.remaining,
		LevelStartedAt:    t.levelStartedAt,
		State:             t.timerState,
	})
}
