package metrics

import "sync"

type Snapshot struct {
	ExecutionAttempts int            `json:"execution_attempts"`
	ExecutionSuccess  int            `json:"execution_success"`
	ExecutionBlocked  int            `json:"execution_blocked"`
	SuccessRate       float64        `json:"success_rate"`
	BlockedRate       float64        `json:"blocked_rate"`
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
	var successRate float64
	var blockedRate float64
	if t.executionAttempts > 0 {
		successRate = float64(t.executionSuccess) / float64(t.executionAttempts)
		blockedRate = float64(t.executionBlocked) / float64(t.executionAttempts)
	}
	return Snapshot{ExecutionAttempts: t.executionAttempts, ExecutionSuccess: t.executionSuccess, ExecutionBlocked: t.executionBlocked, SuccessRate: successRate, BlockedRate: blockedRate, BlockReasons: reasons}
}
