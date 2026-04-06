package portfolio

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

type Position struct {
	Symbol            string  `json:"symbol"`
	Quantity          float64 `json:"quantity"`
	AverageEntryPrice float64 `json:"average_entry_price,omitempty"`
	CostBasis         float64 `json:"cost_basis,omitempty"`
	MarketPrice       float64 `json:"market_price,omitempty"`
	MarketValue       float64 `json:"market_value,omitempty"`
	UnrealizedPnL     float64 `json:"unrealized_pnl,omitempty"`
	RealizedPnL       float64 `json:"realized_pnl,omitempty"`
}

type state struct {
	quantity    float64
	avgEntry    float64
	realizedPnL float64
}

type Tracker struct {
	mu        sync.Mutex
	positions map[string]state
}

func NewTracker() *Tracker { return &Tracker{positions: make(map[string]state)} }

func (t *Tracker) Apply(order exchange.Order) {
	qty := parseFloat(order.Quantity)
	price := parseFloat(order.Price)
	if qty == 0 {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	symbol := strings.ToUpper(strings.TrimSpace(order.Symbol))
	st := t.positions[symbol]
	if order.Side == exchange.Sell {
		if price > 0 {
			st.realizedPnL += (price - st.avgEntry) * qty
		}
		st.quantity -= qty
		if st.quantity <= 0 {
			st.quantity = 0
			st.avgEntry = 0
		}
	} else {
		if price > 0 {
			totalCost := (st.avgEntry * st.quantity) + (price * qty)
			st.quantity += qty
			if st.quantity > 0 {
				st.avgEntry = totalCost / st.quantity
			}
		} else {
			st.quantity += qty
		}
	}
	t.positions[symbol] = st
}

func (t *Tracker) Positions() []Position {
	t.mu.Lock()
	defer t.mu.Unlock()
	keys := make([]string, 0, len(t.positions))
	for symbol := range t.positions {
		keys = append(keys, symbol)
	}
	sort.Strings(keys)
	out := make([]Position, 0, len(keys))
	for _, symbol := range keys {
		st := t.positions[symbol]
		out = append(out, Position{Symbol: symbol, Quantity: st.quantity, AverageEntryPrice: st.avgEntry, CostBasis: st.avgEntry * st.quantity, RealizedPnL: st.realizedPnL})
	}
	return out
}

func (t *Tracker) HasOpenPosition(symbol string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	st, ok := t.positions[strings.ToUpper(strings.TrimSpace(symbol))]
	return ok && st.quantity > 0
}

func (t *Tracker) OpenPositionCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	count := 0
	for _, st := range t.positions {
		if st.quantity > 0 {
			count++
		}
	}
	return count
}

func (t *Tracker) CurrentValue(symbol string) float64 {
	for _, position := range t.Positions() {
		if strings.EqualFold(position.Symbol, symbol) {
			if position.MarketValue > 0 {
				return position.MarketValue
			}
			return position.CostBasis
		}
	}
	return 0
}

func (t *Tracker) TotalValue() float64 {
	var total float64
	for _, position := range t.Positions() {
		total += position.CostBasis
	}
	return total
}

func (t *Tracker) ValuedPositions(ctx context.Context, feed marketdata.Feed) []Position {
	positions := t.Positions()
	if feed == nil {
		return positions
	}
	for i := range positions {
		tick, err := feed.LatestTick(ctx, positions[i].Symbol)
		if err != nil {
			continue
		}
		price := parseFloat(tick.Price)
		positions[i].MarketPrice = price
		positions[i].MarketValue = price * positions[i].Quantity
		positions[i].UnrealizedPnL = (price - positions[i].AverageEntryPrice) * positions[i].Quantity
	}
	return positions
}

func (t *Tracker) TotalMarketValue(ctx context.Context, feed marketdata.Feed) float64 {
	var total float64
	for _, position := range t.ValuedPositions(ctx, feed) {
		total += position.MarketValue
	}
	return total
}

func (t *Tracker) Concentration(ctx context.Context, feed marketdata.Feed) map[string]float64 {
	positions := t.ValuedPositions(ctx, feed)
	var total float64
	for _, position := range positions {
		total += position.MarketValue
	}
	out := make(map[string]float64, len(positions))
	if total <= 0 {
		return out
	}
	for _, position := range positions {
		out[position.Symbol] = position.MarketValue / total
	}
	return out
}

func (t *Tracker) TotalUnrealizedPnL(ctx context.Context, feed marketdata.Feed) float64 {
	var total float64
	for _, position := range t.ValuedPositions(ctx, feed) {
		total += position.UnrealizedPnL
	}
	return total
}

func (t *Tracker) TotalRealizedPnL() float64 {
	var total float64
	for _, position := range t.Positions() {
		total += position.RealizedPnL
	}
	return total
}

func parseFloat(value string) float64 {
	var whole, frac float64
	var fracDiv float64 = 1
	seenDot := false
	for _, ch := range strings.TrimSpace(value) {
		switch {
		case ch == '.':
			if seenDot {
				return 0
			}
			seenDot = true
		case ch >= '0' && ch <= '9':
			digit := float64(ch - '0')
			if !seenDot {
				whole = whole*10 + digit
			} else {
				fracDiv *= 10
				frac += digit / fracDiv
			}
		default:
			return 0
		}
	}
	return whole + frac
}
