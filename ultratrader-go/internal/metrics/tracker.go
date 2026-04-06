package metrics

import "sync"

type Snapshot struct {
	ExecutionAttempts int            `json:"execution_attempts"`
	ExecutionSuccess  int            `json:"execution_success"`
	ExecutionBlocked  int            `json:"execution_blocked"`
	BlockReasons      map[string]int `json:"block_reasons,omitempty"`
}

type Tracker struct {
	mu                sync.Mutex
	executionAttempts int
	executionSuccess  int
	executionBlocked  int
	blockReasons      map[string]int
}

func NewTracker() *Tracker { return &Tracker{blockReasons: map[string]int{}} }

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

func (t *Tracker) RecordBlocked(reason string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.executionBlocked++
	if reason == "" {
		reason = "unknown"
	}
	t.blockReasons[reason]++
}

func (t *Tracker) Snapshot() Snapshot {
	t.mu.Lock()
	defer t.mu.Unlock()
	reasons := make(map[string]int, len(t.blockReasons))
	for k, v := range t.blockReasons {
		reasons[k] = v
	}
	return Snapshot{ExecutionAttempts: t.executionAttempts, ExecutionSuccess: t.executionSuccess, ExecutionBlocked: t.executionBlocked, BlockReasons: reasons}
}
