package app

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/backtest"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	strategydemo "github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy/demo"
)

func TestSyntheticBacktest(t *testing.T) {
	ctx := context.Background()
	symbol := "BTCUSDT"

	// 1. Setup Synthetic Data (to avoid region restrictions in sandbox)
	now := time.Now().Round(time.Hour)
	var candles []marketdata.Candle
	for i := 0; i < 300; i++ {
		price := 60000.0 + float64(i)*10.0 // Slow uptrend
		if i > 250 {
			price = 65000.0 - float64(i-250)*100.0 // Sharp drop
		}

		candles = append(candles, marketdata.Candle{
			Symbol:    symbol,
			Close:     fmt.Sprintf("%.2f", price),
			Timestamp: now.Add(time.Duration(i) * time.Hour),
		})
	}
	provider := backtest.NewMemoryCandleHistory(candles)

	// 2. Setup Strategy and Engine
	strat := strategydemo.NewDoubleEMATrendStrategy("backtest-acct", symbol, "0.01", 9, 21, 200)
	engine := backtest.NewEngine(strat, 10000.0)

	// 3. Run Simulation
	t.Log("Running synthetic backtest simulation...")
	result, err := engine.RunCandles(ctx, provider)
	if err != nil {
		t.Fatalf("Backtest failed: %v", err)
	}

	// 4. Validate Results
	t.Logf("Backtest Completed: Trades=%d FinalValue=%.2f PnL=%.2f",
		result.TotalTrades, result.FinalPortfolioValue, result.RealizedPnL)

	if result.TotalTrades == 0 {
		t.Log("Warning: Zero trades executed. Double-check trend filter logic.")
	}
}
