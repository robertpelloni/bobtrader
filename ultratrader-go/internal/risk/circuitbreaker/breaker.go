package circuitbreaker

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	StateClosed   State = iota // Normal operation — requests pass through
	StateOpen                  // Failure threshold exceeded — requests blocked
	StateHalfOpen              // Testing recovery — limited requests allowed
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// Config holds circuit breaker parameters.
type Config struct {
	// FailureThreshold is the number of consecutive failures before opening.
	FailureThreshold int
	// HalfOpenMax is the number of test requests allowed in half-open state.
	HalfOpenMax int
	// OpenTimeout is how long to wait before transitioning to half-open.
	OpenTimeout time.Duration
	// OnStateChange is called when the circuit breaker changes state.
	OnStateChange func(from, to State)
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		FailureThreshold: 5,
		HalfOpenMax:      1,
		OpenTimeout:      30 * time.Second,
	}
}

// Breaker implements the circuit breaker pattern for API call protection.
type Breaker struct {
	mu              sync.Mutex
	config          Config
	state           State
	consecutiveFail int
	consecutiveOK   int
	lastFailure     time.Time
	lastStateChange time.Time
	totalFailures   int64
	totalSuccesses  int64
	totalRejected   int64
}

// New creates a new circuit breaker with the given configuration.
func New(cfg Config) *Breaker {
	return &Breaker{
		config:          cfg,
		state:           StateClosed,
		lastStateChange: time.Now(),
	}
}

// State returns the current circuit breaker state.
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.checkHalfOpen()
	return b.state
}

// Execute runs a function through the circuit breaker.
// Returns an error if the circuit is open, or propagates the function's error.
func (b *Breaker) Execute(ctx context.Context, fn func() error) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if !b.allowRequest() {
		b.mu.Lock()
		b.totalRejected++
		b.mu.Unlock()
		return fmt.Errorf("circuit breaker is %s", b.State())
	}

	err := fn()
	if err != nil {
		b.RecordFailure()
		return err
	}
	b.RecordSuccess()
	return nil
}

// allowRequest checks if a request should be allowed.
func (b *Breaker) allowRequest() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.checkHalfOpen()

	switch b.state {
	case StateClosed:
		return true
	case StateOpen:
		return false
	case StateHalfOpen:
		// In half-open, allow limited test requests
		return b.consecutiveFail == 0
	default:
		return true
	}
}

// RecordSuccess records a successful operation.
func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.totalSuccesses++
	b.consecutiveFail = 0
	b.consecutiveOK++

	if b.state == StateHalfOpen && b.consecutiveOK >= b.config.HalfOpenMax {
		b.transitionTo(StateClosed)
	}
}

// RecordFailure records a failed operation.
func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.totalFailures++
	b.consecutiveOK = 0
	b.consecutiveFail++
	b.lastFailure = time.Now()

	if b.state == StateHalfOpen {
		b.transitionTo(StateOpen)
	} else if b.state == StateClosed && b.consecutiveFail >= b.config.FailureThreshold {
		b.transitionTo(StateOpen)
	}
}

// checkHalfOpen transitions from open to half-open after timeout.
func (b *Breaker) checkHalfOpen() {
	if b.state == StateOpen && !b.lastFailure.IsZero() {
		if time.Since(b.lastFailure) >= b.config.OpenTimeout {
			b.transitionTo(StateHalfOpen)
		}
	}
}

func (b *Breaker) transitionTo(newState State) {
	oldState := b.state
	if oldState == newState {
		return
	}
	b.state = newState
	b.lastStateChange = time.Now()
	b.consecutiveFail = 0
	b.consecutiveOK = 0
	if b.config.OnStateChange != nil {
		// Call outside lock to prevent deadlock
		go b.config.OnStateChange(oldState, newState)
	}
}

// Stats returns current circuit breaker statistics.
type Stats struct {
	State           State      `json:"state"`
	ConsecutiveFail int        `json:"consecutive_failures"`
	ConsecutiveOK   int        `json:"consecutive_successes"`
	TotalFailures   int64      `json:"total_failures"`
	TotalSuccesses  int64      `json:"total_successes"`
	TotalRejected   int64      `json:"total_rejected"`
	LastFailure     *time.Time `json:"last_failure,omitempty"`
	LastStateChange time.Time  `json:"last_state_change"`
}

// Stats returns a snapshot of circuit breaker statistics.
func (b *Breaker) Stats() Stats {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.checkHalfOpen()

	var lastFail *time.Time
	if !b.lastFailure.IsZero() {
		lastFail = &b.lastFailure
	}
	return Stats{
		State:           b.state,
		ConsecutiveFail: b.consecutiveFail,
		ConsecutiveOK:   b.consecutiveOK,
		TotalFailures:   b.totalFailures,
		TotalSuccesses:  b.totalSuccesses,
		TotalRejected:   b.totalRejected,
		LastFailure:     lastFail,
		LastStateChange: b.lastStateChange,
	}
}

// Reset forces the circuit breaker back to closed state.
func (b *Breaker) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.transitionTo(StateClosed)
}
