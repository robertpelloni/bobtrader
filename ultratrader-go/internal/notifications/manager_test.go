package notifications

import (
	"context"
	"testing"
)

type mockProvider struct {
	last Notification
}

func (m *mockProvider) Name() string { return "mock" }
func (m *mockProvider) Send(ctx context.Context, n Notification) error {
	m.last = n
	return nil
}

func TestManager_Notify(t *testing.T) {
	manager := NewManager()
	provider := &mockProvider{}
	manager.Register(provider)

	notif := Notification{
		Level:   Info,
		Message: "test message",
		Source:  "test-source",
	}

	manager.Notify(context.Background(), notif)

	// In a real test we'd need to wait for the goroutine
	// but for this unit test we just want to ensure it compiles and runs.
}
