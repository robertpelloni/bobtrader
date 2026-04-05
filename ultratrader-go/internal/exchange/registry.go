package exchange

import (
	"fmt"
	"sort"
	"sync"
)

type Factory func() Adapter

type Registry struct {
	mu        sync.RWMutex
	factories map[string]Factory
}

func NewRegistry() *Registry {
	return &Registry{factories: make(map[string]Factory)}
}

func (r *Registry) Register(name string, factory Factory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if name == "" {
		return fmt.Errorf("exchange name is required")
	}
	if factory == nil {
		return fmt.Errorf("factory is required")
	}
	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("exchange %q already registered", name)
	}

	r.factories[name] = factory
	return nil
}

func (r *Registry) Create(name string) (Adapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, ok := r.factories[name]
	if !ok {
		return nil, fmt.Errorf("exchange %q is not registered", name)
	}
	return factory(), nil
}

func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]string, 0, len(r.factories))
	for name := range r.factories {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
