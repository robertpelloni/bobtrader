package risk

import (
	"context"
	"strings"
	"testing"
)

func TestDrawdownMonitor(t *testing.T) {
	ctx := context.Background()
	triggered := false
	var triggerReason string

	callback := func(reason string) {
		triggered = true
		triggerReason = reason
	}

	monitor := NewDrawdownMonitor(0.20, callback) // 20% max drawdown

	// Initial value
	err := monitor.Update(ctx, 10000)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Value goes up, peak increases
	err = monitor.Update(ctx, 11000)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	stats := monitor.Stats()
	if stats["peak_value"].(float64) != 11000 {
		t.Errorf("expected peak to be 11000, got %v", stats["peak_value"])
	}

	// Value goes down slightly (less than 20% from 11000, which is 2200, so min is 8800)
	err = monitor.Update(ctx, 9000)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if triggered {
		t.Errorf("should not be triggered yet")
	}

	// Value drops below 20% drawdown (below 8800)
	err = monitor.Update(ctx, 8500)
	if err == nil {
		t.Errorf("expected error on drawdown breach")
	}

	if !triggered {
		t.Errorf("expected callback to be triggered")
	}

	if !strings.Contains(triggerReason, "Max drawdown threshold (20.00%) exceeded") {
		t.Errorf("unexpected trigger reason: %s", triggerReason)
	}

	// Subsequent updates should be blocked
	err = monitor.Update(ctx, 12000)
	if err == nil || !strings.Contains(err.Error(), "previously triggered") {
		t.Errorf("expected previously triggered error, got %v", err)
	}

	// Test Reset
	monitor.Reset()
	triggered = false
	err = monitor.Update(ctx, 100)
	if err != nil {
		t.Errorf("unexpected error after reset: %v", err)
	}
}
