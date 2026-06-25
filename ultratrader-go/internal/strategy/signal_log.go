package strategy

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SignalOutcome describes what happened to a signal after it was generated.
type SignalOutcome string

const (
	OutcomePending  SignalOutcome = "pending"
	OutcomeExecuted SignalOutcome = "executed"
	OutcomeBlocked  SignalOutcome = "blocked"
	OutcomeSkipped  SignalOutcome = "skipped" // e.g. already in position
)

// LoggedSignal records a strategy signal and its resolution.
type LoggedSignal struct {
	Strategy   string        `json:"strategy"`
	Symbol     string        `json:"symbol"`
	Action     string        `json:"action"`
	Quantity   string        `json:"quantity"`
	Price      string        `json:"price,omitempty"`
	Reason     string        `json:"reason,omitempty"`
	Outcome    SignalOutcome `json:"outcome"`
	BlockedBy  string        `json:"blocked_by,omitempty"`
	FillPrice  string        `json:"fill_price,omitempty"`
	OrderID    string        `json:"order_id,omitempty"`
	PnL        float64       `json:"pnl,omitempty"`         // realized PnL for sell trades
	EntryPrice float64       `json:"entry_price,omitempty"` // average entry price for sell trades
	Timestamp  time.Time     `json:"timestamp"`
}

// StrategyStats tracks per-strategy performance.
type StrategyStats struct {
	Name        string  `json:"name"`
	SignalsTotal int    `json:"signals_total"`
	Executed    int     `json:"executed"`
	Blocked     int     `json:"blocked"`
	Skipped     int     `json:"skipped"`
	WinTrades   int     `json:"win_trades"`
	LossTrades  int     `json:"loss_trades"`
	TotalPnL    float64 `json:"total_pnl"`
	WinRate     float64 `json:"win_rate"`
	SuccessRate float64 `json:"success_rate"` // executed / total signals
	SharpeRatio float64 `json:"sharpe_ratio"`
}

// SignalLog records all strategy signals and their outcomes.
type SignalLog struct {
	mu       sync.Mutex
	signals  []LoggedSignal
	maxSize  int
	persistPath string
	persistMu   sync.Mutex
	lastPersist int // index of last persisted signal
}

// NewSignalLog creates a signal log with bounded history.
func NewSignalLog(maxSize int) *SignalLog {
	if maxSize <= 0 {
		maxSize = 10000
	}
	return &SignalLog{
		signals: make([]LoggedSignal, 0, maxSize),
		maxSize: maxSize,
	}
}

// EnablePersistence enables JSONL file persistence for the signal log.
// Signals are appended to the file on each Flush call.
func (l *SignalLog) EnablePersistence(path string) error {
	l.persistMu.Lock()
	defer l.persistMu.Unlock()

	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	l.persistPath = path
	l.lastPersist = 0

	// Create or append to the file
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	f.Close()

	return nil
}

// Flush writes all un-persisted signals to the JSONL file.
func (l *SignalLog) Flush() error {
	l.persistMu.Lock()
	if l.persistPath == "" {
		l.persistMu.Unlock()
		return nil
	}
	path := l.persistPath
	l.persistMu.Unlock()

	l.mu.Lock()
	start := l.lastPersist
	if start < 0 {
		start = 0
	}
	toFlush := make([]LoggedSignal, len(l.signals)-start)
	copy(toFlush, l.signals[start:])
	l.lastPersist = len(l.signals)
	l.mu.Unlock()

	if len(toFlush) == 0 {
		return nil
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for _, s := range toFlush {
		if err := enc.Encode(s); err != nil {
			return err
		}
	}
	return nil
}

// StartAutoFlush starts a goroutine that periodically flushes signals to disk.
// Returns a stop function.
func (l *SignalLog) StartAutoFlush(interval time.Duration) (stop func()) {
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				l.Flush()
			case <-done:
				l.Flush() // final flush on stop
				return
			}
		}
	}()
	return func() { close(done) }
}

// Record adds a signal to the log.
func (l *SignalLog) Record(s LoggedSignal) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if s.Timestamp.IsZero() {
		s.Timestamp = time.Now().UTC()
	}
	l.signals = append(l.signals, s)
	// Evict oldest if over capacity
	if len(l.signals) > l.maxSize {
		evicted := len(l.signals) - l.maxSize
		l.signals = l.signals[evicted:]
		l.persistMu.Lock()
		l.lastPersist -= evicted
		if l.lastPersist < 0 {
			l.lastPersist = 0
		}
		l.persistMu.Unlock()
	}
}

// Recent returns the last n signals.
func (l *SignalLog) Recent(n int) []LoggedSignal {
	l.mu.Lock()
	defer l.mu.Unlock()
	if n <= 0 || n > len(l.signals) {
		n = len(l.signals)
	}
	result := make([]LoggedSignal, n)
	copy(result, l.signals[len(l.signals)-n:])
	return result
}

// StatsByStrategy computes per-strategy statistics.
func (l *SignalLog) StatsByStrategy() map[string]StrategyStats {
	l.mu.Lock()
	defer l.mu.Unlock()
	stats := make(map[string]StrategyStats)
	for _, s := range l.signals {
		st := stats[s.Strategy]
		st.Name = s.Strategy
		st.SignalsTotal++
		switch s.Outcome {
		case OutcomeExecuted:
			st.Executed++
			if s.Action == "sell" {
				st.TotalPnL += s.PnL
				if s.PnL >= 0 {
					st.WinTrades++
				} else {
					st.LossTrades++
				}
			}
		case OutcomeBlocked:
			st.Blocked++
		case OutcomeSkipped:
			st.Skipped++
		}
		stats[s.Strategy] = st
	}
	// Compute derived rates
	for name, st := range stats {
		// Calculate Sharpe Ratio (simplified daily-proxy)
		// Sharpe = (Avg PnL) / StdDev(PnL)
		var pnls []float64
		for _, s := range l.signals {
			if s.Strategy == name && s.Action == "sell" && s.Outcome == OutcomeExecuted {
				pnls = append(pnls, s.PnL)
			}
		}
		if len(pnls) > 2 {
			sum := 0.0
			for _, p := range pnls { sum += p }
			mean := sum / float64(len(pnls))

			sqDiffSum := 0.0
			for _, p := range pnls { sqDiffSum += (p-mean)*(p-mean) }
			stdDev := 0.0
			if len(pnls) > 1 {
				stdDev = math.Sqrt(sqDiffSum / float64(len(pnls)-1))
			}
			if stdDev > 0 {
				st.SharpeRatio = mean / stdDev
			}
		}

		if st.SignalsTotal > 0 {
			st.SuccessRate = float64(st.Executed) / float64(st.SignalsTotal)
		}
		if st.WinTrades+st.LossTrades > 0 {
			st.WinRate = float64(st.WinTrades) / float64(st.WinTrades+st.LossTrades)
		}
		stats[name] = st
	}
	return stats
}

// UpdateOutcome finds a pending signal and updates its outcome.
func (l *SignalLog) UpdateOutcome(strategy, symbol, action string, outcome SignalOutcome, blockedBy, fillPrice, orderID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	// Find most recent matching pending signal
	for i := len(l.signals) - 1; i >= 0; i-- {
		s := &l.signals[i]
		if s.Strategy == strategy && s.Symbol == symbol && s.Action == action && s.Outcome == OutcomePending {
			s.Outcome = outcome
			s.BlockedBy = blockedBy
			s.FillPrice = fillPrice
			s.OrderID = orderID
			return
		}
	}
}

// Count returns total signals recorded.
func (l *SignalLog) Count() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.signals)
}
