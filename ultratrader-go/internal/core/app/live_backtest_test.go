package app

import (
	"context"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/backtest"
	marketdatabinance "github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata/binance"
	strategydemo "github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy/demo"
)

func TestLiveRecentBacktest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live backtest in short mode")
	}

	ctx := context.Background()
	symbol := "BTCUSDT"

	// 1. Setup Live Data Provider (Real Binance API)
	// Empty API key, production endpoint
	adapter := marketdatabinance.NewAdapter("", false)
	provider := backtest.NewLiveHistoryProvider(adapter)

	// 2. Fetch 1000 recent 1-hour candles
	t.Logf("Fetching 1000 recent 1h candles for %s...", symbol)
	_, err := provider.FetchCandles(ctx, symbol, "1h", 1000)
	if err != nil {
		t.Fatalf("Failed to fetch live history: %v", err)
	}

	// 3. Setup Strategy and Engine
	// Using more aggressive parameters to ensure trades are triggered
	strat := strategydemo.NewDoubleEMATrendStrategy("live-backtest", symbol, "0.01", 5, 10, 50)
	engine := backtest.NewEngine(strat, 10000.0)

	// 4. Run Simulation
	t.Log("Running live backtest simulation...")
	result, err := engine.RunCandles(ctx, provider)
	if err != nil {
		t.Fatalf("Backtest failed: %v", err)
	}

	// 5. Output Results
	t.Logf("Live Backtest Results for %s:", symbol)
	t.Logf("  Total Trades: %d", result.TotalTrades)
	t.Logf("  Realized PnL: %.2f", result.RealizedPnL)
	t.Logf("  Unrealized PnL: %.2f", result.UnrealizedPnL)
	t.Logf("  Final Portfolio Value: %.2f", result.FinalPortfolioValue)

	if result.TotalTrades == 0 {
		t.Log("Warning: No trades executed on real data. Strategy may be too restrictive for current market.")
	}
}
