package app

import (
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/config"
)

func TestLiveTradingInitialization(t *testing.T) {
	cfg := config.Default()
	cfg.Accounts = []config.AccountConfig{
		{
			ID: "live-test",
			Name: "Live Test",
			Exchange: "binance",
			Enabled: true,
			APIKey: "test-key",
			SecretKey: "test-secret",
			Testnet: true,
		},
	}

	application, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to init App with live config: %v", err)
	}

	if application.executionService == nil {
		t.Fatal("ExecutionService not initialized")
	}

	// Verify that the registry can create an adapter for this account
	adapter, err := application.exchangeRegistry.CreateForAccount("binance", "test-key", "test-secret", true)
	if err != nil {
		t.Fatalf("Failed to create Binance adapter for account: %v", err)
	}
	if adapter.Name() != "binance" {
		t.Errorf("Expected binance adapter, got %s", adapter.Name())
	}
}
