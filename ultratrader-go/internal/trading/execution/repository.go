package execution

import (
	"sort"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

type Repository struct {
	mu     sync.Mutex
	orders map[string]exchange.Order
}

func NewRepository() *Repository {
	return &Repository{orders: make(map[string]exchange.Order)}
}

func (r *Repository) Save(order exchange.Order) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[order.ID] = order
}

func (r *Repository) List() []exchange.Order {
	r.mu.Lock()
	defer r.mu.Unlock()

	keys := make([]string, 0, len(r.orders))
	for id := range r.orders {
		keys = append(keys, id)
	}
	sort.Strings(keys)

	out := make([]exchange.Order, 0, len(keys))
	for _, id := range keys {
		out = append(out, r.orders[id])
	}
	return out
}
