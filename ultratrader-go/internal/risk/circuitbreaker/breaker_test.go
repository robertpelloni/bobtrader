package circuitbreaker

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestBreaker_StartsClosed(t *testing.T) {
	b := New(DefaultConfig())
	if b.State() != StateClosed {
		t.Errorf("expected CLOSED, got %s", b.State())
	}
}

func TestBreaker_Execute_Success(t *testing.T) {
	b := New(DefaultConfig())
	err := b.Execute(context.Background(), func() error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	stats := b.Stats()
	if stats.TotalSuccesses != 1 {
		t.Errorf("expected 1 success, got %d", stats.TotalSuccesses)
	}
}

func TestBreaker_Execute_Failure(t *testing.T) {
	b := New(Config{
		FailureThreshold: 3,
		HalfOpenMax:      1,
		OpenTimeout:      50 * time.Millisecond,
	})

	for i := 0; i < 3; i++ {
		b.Execute(context.Background(), func() error { return errors.New("fail") })
	}

	if b.State() != StateOpen {
		t.Errorf("expected OPEN after 3 failures, got %s", b.State())
	}
}

func TestBreaker_BlocksWhenOpen(t *testing.T) {
	b := New(Config{
		FailureThreshold: 1,
		HalfOpenMax:      1,
		OpenTimeout:      100 * time.Millisecond,
	})

	// Trip the breaker
	b.Execute(context.Background(), func() error { return errors.New("fail") })

	// Should be blocked
	err := b.Execute(context.Background(), func() error { return nil })
	if err == nil {
		t.Error("expected error when circuit is open")
	}

	stats := b.Stats()
	if stats.TotalRejected != 1 {
		t.Errorf("expected 1 rejected, got %d", stats.TotalRejected)
	}
}

func TestBreaker_HalfOpenRecovery(t *testing.T) {
	b := New(Config{
		FailureThreshold: 1,
		HalfOpenMax:      1,
		OpenTimeout:      50 * time.Millisecond,
	})

	// Trip the breaker
	b.Execute(context.Background(), func() error { return errors.New("fail") })
	if b.State() != StateOpen {
		t.Fatalf("expected OPEN, got %s", b.State())
	}

	// Wait for half-open transition
	time.Sleep(80 * time.Millisecond)
	if b.State() != StateHalfOpen {
		t.Fatalf("expected HALF_OPEN, got %s", b.State())
	}

	// Succeed in half-open — should close
	err := b.Execute(context.Background(), func() error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	time.Sleep(10 * time.Millisecond) // Let state change callback finish
	if b.State() != StateClosed {
		t.Errorf("expected CLOSED after recovery, got %s", b.State())
	}
}

func TestBreaker_HalfOpen_FailsBack(t *testing.T) {
	b := New(Config{
		FailureThreshold: 1,
		HalfOpenMax:      1,
		OpenTimeout:      50 * time.Millisecond,
	})

	// Trip, wait for half-open
	b.Execute(context.Background(), func() error { return errors.New("fail") })
	time.Sleep(80 * time.Millisecond)

	if b.State() != StateHalfOpen {
		t.Fatalf("expected HALF_OPEN, got %s", b.State())
	}

	// Fail again in half-open
	b.Execute(context.Background(), func() error { return errors.New("still failing") })
	time.Sleep(10 * time.Millisecond)

	if b.State() != StateOpen {
		t.Errorf("expected back to OPEN, got %s", b.State())
	}
}

func TestBreaker_ContextCancellation(t *testing.T) {
	b := New(DefaultConfig())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := b.Execute(ctx, func() error { return nil })
	if err == nil {
		t.Error("expected error from cancelled context")
	}
}

func TestBreaker_StateChangeCallback(t *testing.T) {
	var stateChanges atomic.Int32
	b := New(Config{
		FailureThreshold: 1,
		HalfOpenMax:      1,
		OpenTimeout:      10 * time.Millisecond,
		OnStateChange: func(from, to State) {
			stateChanges.Add(1)
		},
	})

	b.Execute(context.Background(), func() error { return errors.New("fail") })
	time.Sleep(20 * time.Millisecond)
	b.State() // Trigger half-open check
	time.Sleep(50 * time.Millisecond)

	if stateChanges.Load() < 1 {
		t.Errorf("expected state change callbacks, got %d", stateChanges.Load())
	}
}

func TestBreaker_Stats(t *testing.T) {
	b := New(Config{FailureThreshold: 5, OpenTimeout: time.Second})

	b.Execute(context.Background(), func() error { return nil })
	b.Execute(context.Background(), func() error { return errors.New("fail") })
	b.Execute(context.Background(), func() error { return nil })

	stats := b.Stats()
	if stats.TotalSuccesses != 2 {
		t.Errorf("expected 2 successes, got %d", stats.TotalSuccesses)
	}
	if stats.TotalFailures != 1 {
		t.Errorf("expected 1 failure, got %d", stats.TotalFailures)
	}
	if stats.State != StateClosed {
		t.Errorf("expected CLOSED, got %s", stats.State)
	}
}

func TestBreaker_Reset(t *testing.T) {
	b := New(Config{
		FailureThreshold: 1,
		OpenTimeout:      time.Hour,
	})

	b.Execute(context.Background(), func() error { return errors.New("fail") })
	if b.State() != StateOpen {
		t.Fatalf("expected OPEN")
	}

	b.Reset()
	if b.State() != StateClosed {
		t.Errorf("expected CLOSED after reset, got %s", b.State())
	}
}

func TestBreaker_MultipleFailures(t *testing.T) {
	b := New(Config{
		FailureThreshold: 3,
		OpenTimeout:      time.Second,
	})

	// 2 failures should not trip
	b.Execute(context.Background(), func() error { return errors.New("fail") })
	b.Execute(context.Background(), func() error { return errors.New("fail") })
	if b.State() != StateClosed {
		t.Errorf("expected CLOSED after 2 failures (threshold=3), got %s", b.State())
	}

	// 3rd failure trips
	b.Execute(context.Background(), func() error { return errors.New("fail") })
	if b.State() != StateOpen {
		t.Errorf("expected OPEN after 3 failures, got %s", b.State())
	}
}

func TestState_String(t *testing.T) {
	tests := []struct {
		state State
		str   string
	}{
		{StateClosed, "CLOSED"},
		{StateOpen, "OPEN"},
		{StateHalfOpen, "HALF_OPEN"},
		{State(99), "UNKNOWN"},
	}
	for _, tt := range tests {
		if got := tt.state.String(); got != tt.str {
			t.Errorf("State(%d).String() = %q, want %q", tt.state, got, tt.str)
		}
	}
}
