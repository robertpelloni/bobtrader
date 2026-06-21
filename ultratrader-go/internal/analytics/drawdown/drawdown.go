package drawdown

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// DrawdownEvent represents a single drawdown occurrence.
type DrawdownEvent struct {
	Timestamp    time.Time
	PeakValue    float64
	TroughValue  float64
	DrawdownPct  float64
	Duration     time.Duration
	Recovered    bool
	RecoveryTime time.Duration
}

// DrawdownStats contains statistical analysis of drawdowns.
type DrawdownStats struct {
	MaxDrawdown       float64
	AvgDrawdown       float64
	MedianDrawdown    float64
	MaxDuration       time.Duration
	AvgDuration       time.Duration
	RecoveryRate      float64
	AvgRecoveryTime   time.Duration
	DrawdownFrequency float64
	TotalDrawdowns    int
}

// Tracker monitors real-time drawdown with configurable thresholds.
type Tracker struct {
	peakValue    float64
	currentValue float64
	maxDrawdown  float64
	currentDD    float64
	history      []DrawdownEvent
	currentEvent *DrawdownEvent
	mu           sync.RWMutex
}

func NewTracker(initialValue float64) *Tracker {
	return &Tracker{
		peakValue:    initialValue,
		currentValue: initialValue,
		maxDrawdown:  0,
		currentDD:    0,
		history:      make([]DrawdownEvent, 0),
	}
}

// Update updates the current value and returns any new drawdown event.
func (t *Tracker) Update(value float64) *DrawdownEvent {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.currentValue = value

	// Update peak if new high
	if value > t.peakValue {
		// If we were in a drawdown, mark it as recovered
		if t.currentEvent != nil {
			now := time.Now()
			t.currentEvent.Recovered = true
			t.currentEvent.RecoveryTime = now.Sub(t.currentEvent.Timestamp)
			t.history = append(t.history, *t.currentEvent)
			t.currentEvent = nil
		}

		t.peakValue = value
		t.currentDD = 0
		return nil
	}

	// Calculate current drawdown
	t.currentDD = (t.peakValue - value) / t.peakValue * 100

	// Update max drawdown
	if t.currentDD > t.maxDrawdown {
		t.maxDrawdown = t.currentDD
	}

	// Track drawdown event
	if t.currentDD > 0.1 { // More than 0.1% drawdown
		if t.currentEvent == nil {
			t.currentEvent = &DrawdownEvent{
				Timestamp:   time.Now(),
				PeakValue:   t.peakValue,
				TroughValue: value,
				DrawdownPct: t.currentDD,
			}
		} else {
			// Update existing event
			if value < t.currentEvent.TroughValue {
				t.currentEvent.TroughValue = value
				t.currentEvent.DrawdownPct = t.currentDD
			}
			t.currentEvent.Duration = time.Since(t.currentEvent.Timestamp)
		}
	}

	return t.currentEvent
}

// GetMaxDrawdown returns the maximum drawdown experienced.
func (t *Tracker) GetMaxDrawdown() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.maxDrawdown
}

// GetCurrentDrawdown returns the current drawdown percentage.
func (t *Tracker) GetCurrentDrawdown() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.currentDD
}

// IsInDrawdown returns whether currently in a drawdown.
func (t *Tracker) IsInDrawdown() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.currentDD > 0
}

// GetDrawdownHistory returns all completed drawdown events.
func (t *Tracker) GetDrawdownHistory() []DrawdownEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]DrawdownEvent, len(t.history))
	copy(result, t.history)
	return result
}

// GetStatus returns current drawdown status.
func (t *Tracker) GetStatus() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return map[string]interface{}{
		"peak_value":    t.peakValue,
		"current_value": t.currentValue,
		"current_dd":    t.currentDD,
		"max_dd":        t.maxDrawdown,
		"in_drawdown":   t.currentDD > 0,
		"history_count": len(t.history),
	}
}

// Analyzer provides statistical analysis of drawdowns.
type Analyzer struct {
	events []DrawdownEvent
}

func NewAnalyzer(events []DrawdownEvent) *Analyzer {
	return &Analyzer{events: events}
}

// CalculateStats calculates comprehensive drawdown statistics.
func (a *Analyzer) CalculateStats() DrawdownStats {
	if len(a.events) == 0 {
		return DrawdownStats{}
	}

	stats := DrawdownStats{
		TotalDrawdowns: len(a.events),
	}

	// Sort by drawdown percentage
	drawdowns := make([]float64, len(a.events))
	durations := make([]time.Duration, len(a.events))
	recovered := 0
	totalRecoveryTime := time.Duration(0)

	for i, e := range a.events {
		drawdowns[i] = e.DrawdownPct
		durations[i] = e.Duration
		if e.Recovered {
			recovered++
			totalRecoveryTime += e.RecoveryTime
		}
	}

	sort.Float64s(drawdowns)
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })

	// Max drawdown
	stats.MaxDrawdown = drawdowns[len(drawdowns)-1]

	// Average drawdown
	sum := 0.0
	for _, dd := range drawdowns {
		sum += dd
	}
	stats.AvgDrawdown = sum / float64(len(drawdowns))

	// Median drawdown
	mid := len(drawdowns) / 2
	if len(drawdowns)%2 == 0 {
		stats.MedianDrawdown = (drawdowns[mid-1] + drawdowns[mid]) / 2
	} else {
		stats.MedianDrawdown = drawdowns[mid]
	}

	// Duration stats
	stats.MaxDuration = durations[len(durations)-1]
	totalDuration := time.Duration(0)
	for _, d := range durations {
		totalDuration += d
	}
	stats.AvgDuration = totalDuration / time.Duration(len(durations))

	// Recovery stats
	stats.RecoveryRate = float64(recovered) / float64(len(a.events)) * 100
	if recovered > 0 {
		stats.AvgRecoveryTime = totalRecoveryTime / time.Duration(recovered)
	}

	// Frequency (drawdowns per day)
	if len(a.events) >= 2 {
		timeSpan := a.events[len(a.events)-1].Timestamp.Sub(a.events[0].Timestamp)
		if timeSpan > 0 {
			stats.DrawdownFrequency = float64(len(a.events)) / timeSpan.Hours() * 24
		}
	}

	return stats
}

// GetWorstDrawdowns returns the N worst drawdowns.
func (a *Analyzer) GetWorstDrawdowns(n int) []DrawdownEvent {
	if n <= 0 {
		n = 5
	}

	// Sort by drawdown percentage (descending)
	sorted := make([]DrawdownEvent, len(a.events))
	copy(sorted, a.events)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].DrawdownPct > sorted[j].DrawdownPct
	})

	if n > len(sorted) {
		n = len(sorted)
	}
	return sorted[:n]
}

// GetDrawdownDistribution returns drawdown distribution by range.
func (a *Analyzer) GetDrawdownDistribution() map[string]int {
	dist := map[string]int{
		"0-1%":   0,
		"1-2%":   0,
		"2-5%":   0,
		"5-10%":  0,
		"10-20%": 0,
		"20%+":   0,
	}

	for _, e := range a.events {
		switch {
		case e.DrawdownPct < 1:
			dist["0-1%"]++
		case e.DrawdownPct < 2:
			dist["1-2%"]++
		case e.DrawdownPct < 5:
			dist["2-5%"]++
		case e.DrawdownPct < 10:
			dist["5-10%"]++
		case e.DrawdownPct < 20:
			dist["10-20%"]++
		default:
			dist["20%+"]++
		}
	}

	return dist
}

// PredictRecoveryTime estimates recovery time based on historical data.
func (a *Analyzer) PredictRecoveryTime(currentDD float64) time.Duration {
	if len(a.events) == 0 || currentDD <= 0 {
		return 0
	}

	// Find similar drawdowns (within 50% of current)
	var similar []DrawdownEvent
	for _, e := range a.events {
		if e.Recovered && e.DrawdownPct >= currentDD*0.5 && e.DrawdownPct <= currentDD*1.5 {
			similar = append(similar, e)
		}
	}

	if len(similar) == 0 {
		// Use all recovered drawdowns
		for _, e := range a.events {
			if e.Recovered {
				similar = append(similar, e)
			}
		}
	}

	if len(similar) == 0 {
		return 0
	}

	// Average recovery time
	total := time.Duration(0)
	for _, e := range similar {
		total += e.RecoveryTime
	}
	return total / time.Duration(len(similar))
}

// Guard is a risk guard that blocks trading during excessive drawdown.
type Guard struct {
	maxDrawdownPct   float64
	cooldownDuration time.Duration
	tracker          *Tracker
	lastBlockTime    time.Time
	mu               sync.RWMutex
}

func NewGuard(maxDrawdownPct float64, cooldownDuration time.Duration, tracker *Tracker) *Guard {
	if maxDrawdownPct <= 0 {
		maxDrawdownPct = 10.0 // 10% default
	}
	if cooldownDuration <= 0 {
		cooldownDuration = 5 * time.Minute
	}
	return &Guard{
		maxDrawdownPct:   maxDrawdownPct,
		cooldownDuration: cooldownDuration,
		tracker:          tracker,
	}
}

// Check checks if trading should be allowed.
func (g *Guard) Check() error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.tracker == nil {
		return nil
	}

	currentDD := g.tracker.GetCurrentDrawdown()

	// Block if drawdown exceeds threshold
	if currentDD > g.maxDrawdownPct {
		return fmt.Errorf("drawdown guard: current drawdown %.2f%% exceeds max %.2f%%", currentDD, g.maxDrawdownPct)
	}

	// Block if recently hit max drawdown (cooldown)
	if !g.lastBlockTime.IsZero() && time.Since(g.lastBlockTime) < g.cooldownDuration {
		remaining := g.cooldownDuration - time.Since(g.lastBlockTime)
		return fmt.Errorf("drawdown guard: cooldown active for %v", remaining)
	}

	return nil
}

// IsTradingAllowed returns whether trading is currently allowed.
func (g *Guard) IsTradingAllowed() bool {
	return g.Check() == nil
}

// GetStatus returns guard status.
func (g *Guard) GetStatus() map[string]interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()

	status := map[string]interface{}{
		"max_drawdown_pct":  g.maxDrawdownPct,
		"cooldown_duration": g.cooldownDuration.String(),
		"trading_allowed":   g.IsTradingAllowed(),
	}

	if g.tracker != nil {
		status["current_dd"] = g.tracker.GetCurrentDrawdown()
		status["max_dd"] = g.tracker.GetMaxDrawdown()
		status["in_drawdown"] = g.tracker.IsInDrawdown()
	}

	return status
}

// RecordBlock records that trading was blocked due to drawdown.
func (g *Guard) RecordBlock() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.lastBlockTime = time.Now()
}

// String returns a string representation.
func (s DrawdownStats) String() string {
	return fmt.Sprintf(
		"MaxDD=%.2f%% AvgDD=%.2f%% MedDD=%.2f%% Recovery=%.0f%% AvgRecovery=%v",
		s.MaxDrawdown, s.AvgDrawdown, s.MedianDrawdown, s.RecoveryRate, s.AvgRecoveryTime,
	)
}

// String returns a string representation.
func (e DrawdownEvent) String() string {
	return fmt.Sprintf(
		"DD=%.2f%% Peak=%.2f Trough=%.2f Duration=%v Recovered=%v",
		e.DrawdownPct, e.PeakValue, e.TroughValue, e.Duration, e.Recovered,
	)
}
