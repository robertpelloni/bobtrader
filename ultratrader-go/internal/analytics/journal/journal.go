package journal

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// Side represents the direction of a trade.
type Side string

const (
	Buy  Side = "buy"
	Sell Side = "sell"
)

// Entry represents a single trade journal entry.
type Entry struct {
	ID        string    `json:"id"`
	Symbol    string    `json:"symbol"`
	Side      Side      `json:"side"`
	Price     float64   `json:"price"`
	Quantity  float64   `json:"quantity"`
	PnL       float64   `json:"pnl,omitempty"`
	Reason    string    `json:"reason,omitempty"`
	Strategy  string    `json:"strategy,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	GroupID   string    `json:"group_id,omitempty"` // Links related trades
}

// TradeGroup represents a completed trade (entry + exit).
type TradeGroup struct {
	GroupID    string        `json:"group_id"`
	Symbol     string        `json:"symbol"`
	EntrySide  Side          `json:"entry_side"`
	EntryPrice float64       `json:"entry_price"`
	ExitPrice  float64       `json:"exit_price"`
	Quantity   float64       `json:"quantity"`
	PnL        float64       `json:"pnl"`
	EntryTime  time.Time     `json:"entry_time"`
	ExitTime   time.Time     `json:"exit_time"`
	Duration   time.Duration `json:"duration"`
	Strategy   string        `json:"strategy,omitempty"`
}

// PerformanceStats contains computed performance metrics.
type PerformanceStats struct {
	TotalTrades    int     `json:"total_trades"`
	WinningTrades  int     `json:"winning_trades"`
	LosingTrades   int     `json:"losing_trades"`
	WinRate        float64 `json:"win_rate"`
	AvgWin         float64 `json:"avg_win"`
	AvgLoss        float64 `json:"avg_loss"`
	ProfitFactor   float64 `json:"profit_factor"`
	TotalPnL       float64 `json:"total_pnl"`
	MaxDrawdown    float64 `json:"max_drawdown"`
	SharpeRatio    float64 `json:"sharpe_ratio"`
	LargestWin     float64 `json:"largest_win"`
	LargestLoss    float64 `json:"largest_loss"`
	AvgHoldingTime string  `json:"avg_holding_time"`
}

// Journal manages trade entries and computes analytics.
type Journal struct {
	mu      sync.RWMutex
	entries []Entry
	groups  []TradeGroup
}

// New creates a new trade journal.
func New() *Journal {
	return &Journal{
		entries: make([]Entry, 0),
		groups:  make([]TradeGroup, 0),
	}
}

// Record adds a trade entry to the journal.
func (j *Journal) Record(entry Entry) Entry {
	j.mu.Lock()
	defer j.mu.Unlock()

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}
	if entry.ID == "" {
		entry.ID = fmt.Sprintf("T-%d", len(j.entries)+1)
	}
	j.entries = append(j.entries, entry)

	// If this entry has a group ID, try to resolve completed trade groups
	if entry.GroupID != "" {
		j.tryResolveGroup(entry.GroupID)
	}

	return entry
}

// tryResolveGroup checks if a complete trade group exists.
func (j *Journal) tryResolveGroup(groupID string) {
	var buys, sells []Entry
	for _, e := range j.entries {
		if e.GroupID != groupID {
			continue
		}
		if e.Side == Buy {
			buys = append(buys, e)
		} else if e.Side == Sell {
			sells = append(sells, e)
		}
	}

	// Match buys with sells (FIFO)
	for len(buys) > 0 && len(sells) > 0 {
		buy := buys[0]
		sell := sells[0]
		buys = buys[1:]
		sells = sells[1:]

		qty := buy.Quantity
		if sell.Quantity < qty {
			qty = sell.Quantity
		}

		pnl := (sell.Price - buy.Price) * qty

		group := TradeGroup{
			GroupID:    groupID,
			Symbol:     buy.Symbol,
			EntrySide:  Buy,
			EntryPrice: buy.Price,
			ExitPrice:  sell.Price,
			Quantity:   qty,
			PnL:        pnl,
			EntryTime:  buy.Timestamp,
			ExitTime:   sell.Timestamp,
			Duration:   sell.Timestamp.Sub(buy.Timestamp),
			Strategy:   buy.Strategy,
		}
		j.groups = append(j.groups, group)
	}
}

// Entries returns all journal entries.
func (j *Journal) Entries() []Entry {
	j.mu.RLock()
	defer j.mu.RUnlock()
	result := make([]Entry, len(j.entries))
	copy(result, j.entries)
	return result
}

// TradeGroups returns all resolved trade groups.
func (j *Journal) TradeGroups() []TradeGroup {
	j.mu.RLock()
	defer j.mu.RUnlock()
	result := make([]TradeGroup, len(j.groups))
	copy(result, j.groups)
	return result
}

// Stats computes performance statistics from resolved trade groups.
func (j *Journal) Stats() PerformanceStats {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return computeStats(j.groups)
}

// StatsSince computes performance statistics since a given time.
func (j *Journal) StatsSince(since time.Time) PerformanceStats {
	j.mu.RLock()
	defer j.mu.RUnlock()

	var filtered []TradeGroup
	for _, g := range j.groups {
		if !g.ExitTime.Before(since) {
			filtered = append(filtered, g)
		}
	}
	return computeStats(filtered)
}

// StatsByStrategy computes performance statistics for a specific strategy.
func (j *Journal) StatsByStrategy(strategy string) PerformanceStats {
	j.mu.RLock()
	defer j.mu.RUnlock()

	var filtered []TradeGroup
	for _, g := range j.groups {
		if g.Strategy == strategy {
			filtered = append(filtered, g)
		}
	}
	return computeStats(filtered)
}

// MaxDrawdown calculates the maximum drawdown from a PnL series.
func MaxDrawdown(pnls []float64) float64 {
	if len(pnls) == 0 {
		return 0
	}

	var peak, maxDD float64
	cumulative := 0.0

	for _, pnl := range pnls {
		cumulative += pnl
		if cumulative > peak {
			peak = cumulative
		}
		dd := peak - cumulative
		if dd > maxDD {
			maxDD = dd
		}
	}
	return maxDD
}

func computeStats(groups []TradeGroup) PerformanceStats {
	if len(groups) == 0 {
		return PerformanceStats{}
	}

	// Sort by exit time for drawdown calculation
	sorted := make([]TradeGroup, len(groups))
	copy(sorted, groups)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ExitTime.Before(sorted[j].ExitTime)
	})

	var totalPnL, totalWins, totalLosses float64
	var wins, losses int
	var largestWin, largestLoss float64
	var totalDuration time.Duration
	pnls := make([]float64, len(sorted))

	for i, g := range sorted {
		pnls[i] = g.PnL
		totalPnL += g.PnL
		totalDuration += g.Duration

		if g.PnL > 0 {
			wins++
			totalWins += g.PnL
			if g.PnL > largestWin {
				largestWin = g.PnL
			}
		} else if g.PnL < 0 {
			losses++
			totalLosses += g.PnL
			if g.PnL < largestLoss {
				largestLoss = g.PnL
			}
		}
	}

	winRate := float64(wins) / float64(len(groups))
	var avgWin, avgLoss, profitFactor float64
	if wins > 0 {
		avgWin = totalWins / float64(wins)
	}
	if losses > 0 {
		avgLoss = totalLosses / float64(losses)
	}
	if totalLosses != 0 {
		profitFactor = totalWins / (-totalLosses)
	} else if totalWins > 0 {
		profitFactor = float64(wins) // No losses = infinite, capped
	}

	// Sharpe approximation from trade PnLs
	var sharpe float64
	if len(pnls) > 1 {
		var sum, sumSq float64
		for _, p := range pnls {
			sum += p
			sumSq += p * p
		}
		mean := sum / float64(len(pnls))
		variance := sumSq/float64(len(pnls)) - mean*mean
		if variance > 0 {
			sharpe = mean / sqrt(variance)
		}
	}

	avgHolding := time.Duration(0)
	if len(groups) > 0 {
		avgHolding = totalDuration / time.Duration(len(groups))
	}

	return PerformanceStats{
		TotalTrades:    len(groups),
		WinningTrades:  wins,
		LosingTrades:   losses,
		WinRate:        winRate,
		AvgWin:         avgWin,
		AvgLoss:        avgLoss,
		ProfitFactor:   profitFactor,
		TotalPnL:       totalPnL,
		MaxDrawdown:    MaxDrawdown(pnls),
		SharpeRatio:    sharpe,
		LargestWin:     largestWin,
		LargestLoss:    largestLoss,
		AvgHoldingTime: avgHolding.String(),
	}
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 20; i++ {
		z = (z + x/z) / 2
	}
	return z
}
