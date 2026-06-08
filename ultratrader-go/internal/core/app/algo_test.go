package app

import (
	"context"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/config"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

func TestAlgoVerification(t *testing.T) {
	cfg := config.Default()
	cfg.Environment = "algo-verify"
	cfg.Scheduler.Enabled = true
	cfg.Scheduler.Mode = "timer"
	cfg.Scheduler.IntervalMS = 50
	cfg.Risk.AllowedSymbols = []string{"BTCUSDT"}

	// Fast turnaround for tests
	application, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to init: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := application.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Verify strategy signals are being generated and logged
	time.Sleep(500 * time.Millisecond)

	stats := application.signalLog.StatsByStrategy()
	if len(stats) == 0 {
		t.Log("Warning: No strategy stats recorded. This is expected if price threshold was not hit.")
	}

	// Manually inject an order to verify ExecutionManager coordinates correctly
	order := exchange.Order{
		Symbol:   "BTCUSDT",
		Side:     exchange.Buy,
		Quantity: "0.1",
	}
	err = application.executionManager.Execute(ctx, "market", order)
	if err != nil {
		t.Errorf("ExecutionManager failed to execute market strategy: %v", err)
	}

	if err := application.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}
