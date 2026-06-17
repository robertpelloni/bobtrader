package influx

import (
	"testing"
)

// This test requires a running InfluxDB instance, so we'll just check compilation and basic structure.
func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:8086", "token", "org", "bucket")
	if client == nil {
		t.Fatal("expected client to be initialized")
	}
}

// We could add a mock-based test here if needed for deeper verification without a real DB.
