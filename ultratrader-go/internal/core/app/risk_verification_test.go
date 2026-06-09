package app

import (
	"context"
	"strings"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/config"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/risk"
)

func TestRiskControlVerification(t *testing.T) {
	cfg := config.Default()
	cfg.Risk.AllowedSymbols = []string{"BTCUSDT"}
	cfg.Risk.MaxNotional = 100

	application, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to init: %v", err)
	}

	ctx := context.Background()

	t.Run("Whitelist_Violation", func(t *testing.T) {
		req := exchange.OrderRequest{Symbol: "DOGEUSDT", Side: exchange.Buy, Quantity: "1"}
		intent := risk.OrderIntent{Symbol: "DOGEUSDT", Notional: 10}
		_, err := application.executionService.Execute(ctx, "paper-main", req, intent)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !strings.Contains(err.Error(), "not whitelisted") && !strings.Contains(err.Error(), "not in whitelist") {
			t.Errorf("Expected whitelist error, got: %v", err)
		}
	})

	t.Run("MaxNotional_Violation", func(t *testing.T) {
		req := exchange.OrderRequest{Symbol: "BTCUSDT", Side: exchange.Buy, Quantity: "1"}
		intent := risk.OrderIntent{Symbol: "BTCUSDT", Notional: 500}
		_, err := application.executionService.Execute(ctx, "paper-main", req, intent)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !strings.Contains(err.Error(), "exceeds maximum notional") && !strings.Contains(err.Error(), "exceeds limit") {
			t.Errorf("Expected max notional error, got: %v", err)
		}
	})

	t.Run("Valid_Order_Passes", func(t *testing.T) {
		req := exchange.OrderRequest{Symbol: "BTCUSDT", Side: exchange.Buy, Quantity: "0.001"}
		intent := risk.OrderIntent{Symbol: "BTCUSDT", Notional: 65}
		_, err := application.executionService.Execute(ctx, "paper-main", req, intent)
		if err != nil {
			t.Errorf("Valid order should have passed guards, got: %v", err)
		}
	})
}
