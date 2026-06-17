package notifications

import (
	"context"
	"fmt"
	"sync"
)

type Level string

const (
	Info     Level = "info"
	Trade    Level = "trade"
	Alert    Level = "alert"
	Critical Level = "critical"
)

type Notification struct {
	Level   Level
	Message string
	Source  string
	Data    map[string]any
}

type Provider interface {
	Name() string
	Send(ctx context.Context, n Notification) error
}

type Manager struct {
	mu        sync.RWMutex
	providers []Provider
}

func NewManager() *Manager {
	return &Manager{
		providers: make([]Provider, 0),
	}
}

func (m *Manager) Register(p Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers = append(m.providers, p)
}

func (m *Manager) Notify(ctx context.Context, n Notification) {
	m.mu.RLock()
	providers := m.providers
	m.mu.RUnlock()

	for _, p := range providers {
		go func(p Provider) {
			if err := p.Send(ctx, n); err != nil {
				fmt.Printf("Notification failed for %s: %v\n", p.Name(), err)
			}
		}(p)
	}
}
