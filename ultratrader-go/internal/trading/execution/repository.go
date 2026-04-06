package execution

import (
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

type StoredOrder struct {
	Order   exchange.Order `json:"order"`
	SavedAt time.Time      `json:"saved_at"`
}

type Summary struct {
	TotalOrders    int            `json:"total_orders"`
	OrdersBySymbol map[string]int `json:"orders_by_symbol"`
	UniqueSymbols  int            `json:"unique_symbols"`
	TopSymbol      string         `json:"top_symbol,omitempty"`
	TopSymbolCount int            `json:"top_symbol_count,omitempty"`
	LastOrderID    string         `json:"last_order_id,omitempty"`
	LastSymbol     string         `json:"last_symbol,omitempty"`
}

type Repository struct {
	mu     sync.Mutex
	orders map[string]StoredOrder
}

func NewRepository() *Repository { return &Repository{orders: make(map[string]StoredOrder)} }

func (r *Repository) Save(order exchange.Order) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[order.ID] = StoredOrder{Order: order, SavedAt: time.Now().UTC()}
}

func (r *Repository) List() []exchange.Order {
	stored := r.ListStored()
	out := make([]exchange.Order, 0, len(stored))
	for _, item := range stored {
		out = append(out, item.Order)
	}
	return out
}

func (r *Repository) ListStored() []StoredOrder {
	r.mu.Lock()
	defer r.mu.Unlock()
	keys := make([]string, 0, len(r.orders))
	for id := range r.orders {
		keys = append(keys, id)
	}
	sort.Strings(keys)
	out := make([]StoredOrder, 0, len(keys))
	for _, id := range keys {
		out = append(out, r.orders[id])
	}
	return out
}

func (r *Repository) HasRecentSymbol(symbol string, within time.Duration) bool {
	if within <= 0 {
		return false
	}
	now := time.Now().UTC()
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	for _, item := range r.ListStored() {
		if strings.ToUpper(strings.TrimSpace(item.Order.Symbol)) == symbol && now.Sub(item.SavedAt) <= within {
			return true
		}
	}
	return false
}

func (r *Repository) Summary() Summary {
	stored := r.ListStored()
	summary := Summary{OrdersBySymbol: map[string]int{}, TotalOrders: len(stored)}
	for _, item := range stored {
		symbol := strings.ToUpper(strings.TrimSpace(item.Order.Symbol))
		summary.OrdersBySymbol[symbol]++
		if summary.OrdersBySymbol[symbol] > summary.TopSymbolCount {
			summary.TopSymbol = symbol
			summary.TopSymbolCount = summary.OrdersBySymbol[symbol]
		}
		summary.LastOrderID = item.Order.ID
		summary.LastSymbol = symbol
	}
	summary.UniqueSymbols = len(summary.OrdersBySymbol)
	return summary
}
