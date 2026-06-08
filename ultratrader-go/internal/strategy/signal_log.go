package strategy

import (
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
	Strategy  string        `json:"strategy"`
	Symbol    string        `json:"symbol"`
	Action    string        `json:"action"`
	Quantity  string        `json:"quantity"`
	Price     string        `json:"price,omitempty"`
	Reason    string        `json:"reason,omitempty"`
	Outcome   SignalOutcome `json:"outcome"`
	BlockedBy string        `json:"blocked_by,omitempty"`
	FillPrice string        `json:"fill_price,omitempty"`
	OrderID   string        `json:"order_id,omitempty"`
	PnL       float64       `json:"pnl,omitempty"`        // realized PnL for sell trades
	EntryPrice float64      `json:"entry_price,omitempty"` // average entry price for sell trades
	Timestamp  time.Time    `json:"timestamp"`
}

// StrategyStats tracks per-strategy performance.
type StrategyStats struct {
	Name         string  `json:"name"`
	SignalsTotal int     `json:"signals_total"`
	Executed     int     `json:"executed"`
	Blocked      int     `json:"blocked"`
	Skipped      int     `json:"skipped"`
	WinTrades    int     `json:"win_trades"`
	LossTrades   int     `json:"loss_trades"`
	TotalPnL     float64 `json:"total_pnl"`
	WinRate      float64 `json:"win_rate"`
	SuccessRate  float64 `json:"success_rate"` // executed / total signals
}

// SignalLog records all strategy signals and their outcomes.
type SignalLog struct {
	mu      sync.Mutex
	signals []LoggedSignal
	maxSize int
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
		l.signals = l.signals[len(l.signals)-l.maxSize:]
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
