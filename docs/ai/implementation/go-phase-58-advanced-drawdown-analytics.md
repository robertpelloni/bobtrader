# Go Phase 58: Advanced Drawdown Analytics

## Summary
Dedicated drawdown tracking, analysis, and risk management module with real-time monitoring and historical analysis.

## Current State
- `internal/analytics/journal/journal.go` — Basic MaxDrawdown in PerformanceSnapshot
- `internal/backtest/optimizer/montecarlo.go` — Drawdown in Monte Carlo simulation

## New Module: `internal/analytics/drawdown/`

### 1. DrawdownTracker
Real-time drawdown monitoring with configurable thresholds.

```go
type DrawdownTracker struct {
    peakValue    float64
    currentValue float64
    maxDrawdown  float64
    currentDD    float64
    history      []DrawdownEvent
    mu           sync.RWMutex
}

type DrawdownEvent struct {
    Timestamp    time.Time
    PeakValue    float64
    TroughValue  float64
    DrawdownPct  float64
    Duration     time.Duration
    Recovered    bool
    RecoveryTime time.Duration
}

func (t *DrawdownTracker) Update(value float64) DrawdownEvent
func (t *DrawdownTracker) GetMaxDrawdown() DrawdownEvent
func (t *DrawdownTracker) GetCurrentDrawdown() float64
func (t *DrawdownTracker) GetDrawdownHistory() []DrawdownEvent
func (t *DrawdownTracker) IsInDrawdown() bool
func (t *DrawdownTracker) GetRecoveryTime() time.Duration
```

### 2. DrawdownAnalyzer
Statistical analysis of drawdown patterns.

```go
type DrawdownAnalyzer struct {
    events []DrawdownEvent
}

type DrawdownStats struct {
    MaxDrawdown      float64
    AvgDrawdown      float64
    MedianDrawdown   float64
    MaxDuration      time.Duration
    AvgDuration      time.Duration
    RecoveryRate     float64 // % of drawdowns that recovered
    AvgRecoveryTime  time.Duration
    DrawdownFrequency float64 // drawdowns per day
}

func (a *DrawdownAnalyzer) CalculateStats() DrawdownStats
func (a *DrawdownAnalyzer) GetWorstDrawdowns(n int) []DrawdownEvent
func (a *DrawdownAnalyzer) GetDrawdownDistribution() map[string]int
func (a *DrawdownAnalyzer) PredictRecoveryTime(currentDD float64) time.Duration
```

### 3. DrawdownGuard
Risk guard that blocks trading during excessive drawdown.

```go
type DrawdownGuard struct {
    maxDrawdownPct   float64 // e.g., 10% max drawdown
    cooldownDuration time.Duration
    tracker          *DrawdownTracker
}

func (g *DrawdownGuard) Check(intent risk.OrderIntent) error
func (g *DrawdownGuard) IsTradingAllowed() bool
func (g *DrawdownGuard) GetStatus() map[string]interface{}
```

## Tests
- `TestDrawdownTracker_Update`
- `TestDrawdownTracker_MaxDrawdown`
- `TestDrawdownTracker_Recovery`
- `TestDrawdownAnalyzer_Stats`
- `TestDrawdownAnalyzer_WorstDrawdowns`
- `TestDrawdownGuard_Check`
- `TestDrawdownGuard_Cooldown`
