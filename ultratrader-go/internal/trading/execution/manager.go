package execution

import (
	"context"
	"fmt"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

// Strategy represents a specific execution logic (e.g., TWAP, VWAP, Snipe).
type Strategy interface {
	Name() string
	Execute(ctx context.Context, order exchange.Order) error
}

// Manager coordinates multiple execution strategies, inspired by OpenAlice's ToolCenter.
type Manager struct {
	mu         sync.RWMutex
	strategies map[string]Strategy
}

// NewManager creates a new execution manager.
func NewManager() *Manager {
	return &Manager{
		strategies: make(map[string]Strategy),
	}
}

// Register adds an execution strategy to the manager.
func (m *Manager) Register(s Strategy) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.strategies[s.Name()] = s
}

// Execute dispatches an order to the specified strategy.
func (m *Manager) Execute(ctx context.Context, strategyName string, order exchange.Order) error {
	m.mu.RLock()
	s, ok := m.strategies[strategyName]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("execution strategy not found: %s", strategyName)
	}

	return s.Execute(ctx, order)
}

// ListStrategies returns the names of all registered strategies.
func (m *Manager) ListStrategies() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.strategies))
	for name := range m.strategies {
		names = append(names, name)
	}
	return names
}
