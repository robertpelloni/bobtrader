package backtest

import (
	"context"
	"fmt"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/portfolio"
)

// HistoryProvider provides a sequence of historical market data ticks.
type HistoryProvider interface {
	Ticks() []marketdata.Tick
}

// Result holds the final statistics of a backtest run.
type Result struct {
	TotalTrades         int
	FinalPortfolioValue float64
	RealizedPnL         float64
	UnrealizedPnL       float64
	StartTime           time.Time
	EndTime             time.Time
	Orders              []exchange.Order
}

// Engine orchestrates the historical simulation.
type Engine struct {
	strategy strategy.TickStrategy
	history  HistoryProvider
	tracker  *portfolio.Tracker
	orders   []exchange.Order
}

// NewEngine creates a new backtesting engine.
func NewEngine(s strategy.TickStrategy, h HistoryProvider, initialCapital float64) *Engine {
	// Initialize a simulated portfolio tracker.
	tracker := portfolio.NewTracker()
	// Optionally, we could pre-fund the tracker with initialCapital if it tracked cash.
	// For now, our tracker mainly tracks symbol quantities and PnL.

	return &Engine{
		strategy: s,
		history:  h,
		tracker:  tracker,
		orders:   make([]exchange.Order, 0),
	}
}

// simpleFeed implements marketdata.Feed for the backtest context, providing the *current* tick price for valuation.
type simpleFeed struct {
	currentPrice string
}

func (f *simpleFeed) LatestTick(_ context.Context, symbol string) (marketdata.Tick, error) {
	return marketdata.Tick{Symbol: symbol, Price: f.currentPrice, Timestamp: time.Now()}, nil
}
func (f *simpleFeed) LatestCandle(_ context.Context, symbol, interval string) (marketdata.Candle, error) {
	return marketdata.Candle{}, fmt.Errorf("not implemented in backtest")
}

// Run executes the simulation over the provided history.
func (e *Engine) Run(ctx context.Context) (Result, error) {
	ticks := e.history.Ticks()
	if len(ticks) == 0 {
		return Result{}, fmt.Errorf("no historical data provided")
	}

	var lastPrice string
	var startTime time.Time = ticks[0].Timestamp
	var endTime time.Time = ticks[len(ticks)-1].Timestamp

	for _, tick := range ticks {
		lastPrice = tick.Price

		// Pass the tick to the strategy.
		signals, err := e.strategy.OnMarketTick(ctx, tick)
		if err != nil {
			return Result{}, fmt.Errorf("strategy error on tick %v: %w", tick, err)
		}

		// Process generated signals into simulated orders.
		for _, sig := range signals {
			var side exchange.OrderSide
			if sig.Action == "buy" {
				side = exchange.Buy
			} else if sig.Action == "sell" {
				side = exchange.Sell
			} else {
				continue // Ignore unsupported actions
			}

			// Simulate order execution at the current tick price.
			// In a more advanced backtester, slippage, fees, and spread would be modeled here.
			order := exchange.Order{
				ID:       fmt.Sprintf("bt-ord-%d", len(e.orders)+1),
				Symbol:   sig.Symbol,
				Side:     side,
				Type:     exchange.MarketOrder,
				Status:   "filled",
				Quantity: sig.Quantity,
				Price:    tick.Price,
			}
			e.orders = append(e.orders, order)

			// Update the simulated portfolio.
			e.tracker.Apply(order)
		}
	}

	// Final valuation using the last seen price.
	feed := &simpleFeed{currentPrice: lastPrice}
	totalValue := e.tracker.TotalMarketValue(ctx, feed)
	realized := e.tracker.TotalRealizedPnL()
	unrealized := e.tracker.TotalUnrealizedPnL(ctx, feed)

	// Since we aren't tracking a base cash balance initially, the portfolio value
	// is just the value of held assets. A full system would add initialCapital + realized.

	// Assuming initialCapital was held entirely in cash, and we bought assets using it:
	// Final cash = initialCapital - cost_of_assets_bought + revenue_from_assets_sold (which is represented in realized PnL and current asset values)
	// Simplified: Final value = Total Asset Value + (Cash). We need a way to track Cash.
	// For this phase, we'll return the aggregate metrics from the tracker.

	return Result{
		TotalTrades:         len(e.orders),
		FinalPortfolioValue: totalValue,
		RealizedPnL:         realized,
		UnrealizedPnL:       unrealized,
		StartTime:           startTime,
		EndTime:             endTime,
		Orders:              e.orders,
	}, nil
}

// MemoryHistory is a simple history provider backed by an in-memory slice.
type MemoryHistory struct {
	data []marketdata.Tick
}

func NewMemoryHistory(data []marketdata.Tick) *MemoryHistory {
	return &MemoryHistory{data: data}
}

func (h *MemoryHistory) Ticks() []marketdata.Tick {
	return h.data
}
