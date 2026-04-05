package metrics

import "sync"

type Snapshot struct {
	ExecutionAttempts int `json:"execution_attempts"`
	ExecutionSuccess  int `json:"execution_success"`
	ExecutionBlocked  int `json:"execution_blocked"`
}

type Tracker struct {
	mu                sync.Mutex
	executionAttempts int
	executionSuccess  int
	executionBlocked  int
}

func NewTracker() *Tracker { return &Tracker{} }

func (t *Tracker) RecordAttempt() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.executionAttempts++
}

func (t *Tracker) RecordSuccess() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.executionSuccess++
}

func (t *Tracker) RecordBlocked() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.executionBlocked++
}

func (t *Tracker) Snapshot() Snapshot {
	t.mu.Lock()
	defer t.mu.Unlock()
	return Snapshot{ExecutionAttempts: t.executionAttempts, ExecutionSuccess: t.executionSuccess, ExecutionBlocked: t.executionBlocked}
}
