package backtest_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/backtest"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

type mockHistoricalFeed struct {
	err error
}

func (m *mockHistoricalFeed) LatestTick(ctx context.Context, symbol string) (marketdata.Tick, error) {
	return marketdata.Tick{}, nil
}

func (m *mockHistoricalFeed) LatestCandle(ctx context.Context, symbol, interval string) (marketdata.Candle, error) {
	return marketdata.Candle{}, nil
}

func (m *mockHistoricalFeed) HistoricalCandles(ctx context.Context, symbol string, interval string, limit int) ([]marketdata.Candle, error) {
	if m.err != nil {
		return nil, m.err
	}

	now := time.Now().Truncate(time.Hour)
	var candles []marketdata.Candle

	for i := 0; i < limit; i++ {
		ts := now.Add(time.Duration(i) * time.Hour)
		candles = append(candles, marketdata.Candle{
			Symbol:    symbol,
			Timestamp: ts,
			Close:     "100.0",
		})
	}

	// Intentionally drop a middle candle for ETH to test sparse timeline
	if symbol == "ETH" {
		// remove second candle
		candles = append(candles[:1], candles[2:]...)
	}

	return candles, nil
}

func TestMultiSymbolFeed_Synchronize(t *testing.T) {
	symbols := []string{"BTC", "ETH"}
	feed := backtest.NewMultiSymbolFeed(symbols)

	t1 := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC)
	t3 := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	feed.LoadData("BTC", []marketdata.Candle{
		{Symbol: "BTC", Timestamp: t1},
		{Symbol: "BTC", Timestamp: t2},
		{Symbol: "BTC", Timestamp: t3},
	})

	feed.LoadData("ETH", []marketdata.Candle{
		{Symbol: "ETH", Timestamp: t1},
		// Missing t2 intentionally
		{Symbol: "ETH", Timestamp: t3},
	})

	timeline := feed.Synchronize()

	if len(timeline) != 3 {
		t.Fatalf("Expected 3 sync points on timeline, got %d", len(timeline))
	}

	// T1 check
	if len(timeline[0].Candles) != 2 {
		t.Errorf("Expected 2 candles at T1")
	}

	// T2 check
	if len(timeline[1].Candles) != 1 {
		t.Errorf("Expected 1 candle at T2")
	}
	if _, ok := timeline[1].Candles["BTC"]; !ok {
		t.Errorf("Expected BTC at T2")
	}

	// T3 check
	if len(timeline[2].Candles) != 2 {
		t.Errorf("Expected 2 candles at T3")
	}
}

func TestMultiSymbolFeed_FetchData(t *testing.T) {
	symbols := []string{"BTC", "ETH"}
	multiFeed := backtest.NewMultiSymbolFeed(symbols)

	mockFeed := &mockHistoricalFeed{}

	err := multiFeed.FetchData(context.Background(), mockFeed, "1hour", 5)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	timeline := multiFeed.Synchronize()
	if len(timeline) != 5 {
		t.Fatalf("Expected 5 timeline entries, got %d", len(timeline))
	}

	// The 2nd entry (idx 1) for ETH was explicitly dropped in the mock
	if len(timeline[1].Candles) != 1 {
		t.Errorf("Expected ETH to be missing at idx 1")
	}
}

func TestMultiSymbolFeed_FetchData_Error(t *testing.T) {
	symbols := []string{"BTC"}
	multiFeed := backtest.NewMultiSymbolFeed(symbols)

	mockFeed := &mockHistoricalFeed{err: errors.New("api rate limit")}

	err := multiFeed.FetchData(context.Background(), mockFeed, "1hour", 5)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
}
