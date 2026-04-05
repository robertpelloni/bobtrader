package portfolio

import (
	"sort"
	"strings"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

type Position struct {
	Symbol   string  `json:"symbol"`
	Quantity float64 `json:"quantity"`
}

type Tracker struct {
	mu        sync.Mutex
	positions map[string]float64
}

func NewTracker() *Tracker {
	return &Tracker{positions: make(map[string]float64)}
}

func (t *Tracker) Apply(order exchange.Order) {
	qty := parseFloat(order.Quantity)
	if qty == 0 {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	symbol := strings.ToUpper(strings.TrimSpace(order.Symbol))
	if order.Side == exchange.Sell {
		t.positions[symbol] -= qty
	} else {
		t.positions[symbol] += qty
	}
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
		out = append(out, Position{Symbol: symbol, Quantity: t.positions[symbol]})
	}
	return out
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
