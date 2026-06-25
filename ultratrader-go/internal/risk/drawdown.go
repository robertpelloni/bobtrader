package risk

import (
	"context"
	"fmt"
	"sync"
)

// DrawdownMonitor tracks portfolio peak value and detects excessive drawdowns.
type DrawdownMonitor struct {
	mu             sync.Mutex
	maxDrawdownPct float64
	peakValue      float64
	currentValue   float64
	triggered      bool
	onTrigger      func(reason string)
}

// NewDrawdownMonitor creates a new monitor. maxDrawdownPct should be a positive fraction (e.g., 0.20 for 20%).
// The onTrigger callback is executed when the drawdown exceeds the threshold.
func NewDrawdownMonitor(maxDrawdownPct float64, onTrigger func(reason string)) *DrawdownMonitor {
	if maxDrawdownPct <= 0 {
		maxDrawdownPct = 1.0 // Disable essentially, or set to 100%
	}
	return &DrawdownMonitor{
		maxDrawdownPct: maxDrawdownPct,
		onTrigger:      onTrigger,
	}
}

// Update processes a new portfolio value, updates the peak, and checks for breaches.
func (d *DrawdownMonitor) Update(ctx context.Context, portfolioValue float64) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	d.mu.Lock()

	if d.triggered {
		d.mu.Unlock()
		return fmt.Errorf("drawdown monitor previously triggered, trading halted")
	}

	d.currentValue = portfolioValue
	if portfolioValue > d.peakValue {
		d.peakValue = portfolioValue
	}

	if d.peakValue <= 0 {
		d.mu.Unlock()
		return nil
	}

	currentDrawdown := (d.peakValue - d.currentValue) / d.peakValue

	if currentDrawdown >= d.maxDrawdownPct {
		d.triggered = true
		reason := fmt.Sprintf("Max drawdown threshold (%.2f%%) exceeded! Peak: $%.2f, Current: $%.2f, Drawdown: %.2f%%",
			d.maxDrawdownPct*100, d.peakValue, d.currentValue, currentDrawdown*100)
		d.mu.Unlock()

		if d.onTrigger != nil {
			d.onTrigger(reason)
		}
		return fmt.Errorf("drawdown triggered: %s", reason)
	}

	d.mu.Unlock()
	return nil
}

// Stats returns the current state of the monitor.
func (d *DrawdownMonitor) Stats() map[string]interface{} {
	d.mu.Lock()
	defer d.mu.Unlock()

	var currentDrawdown float64
	if d.peakValue > 0 {
		currentDrawdown = (d.peakValue - d.currentValue) / d.peakValue
	}

	return map[string]interface{}{
		"max_drawdown_pct_allowed": d.maxDrawdownPct,
		"peak_value":               d.peakValue,
		"current_value":            d.currentValue,
		"current_drawdown_pct":     currentDrawdown,
		"triggered":                d.triggered,
	}
}

// Reset resets the monitor's state.
func (d *DrawdownMonitor) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.peakValue = 0
	d.currentValue = 0
	d.triggered = false
}
