package backtest

import (
	"context"
	"fmt"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/portfolio"
)

// HistoryProvider provides a sequence of historical market data ticks.
type HistoryProvider interface {
	Ticks() []marketdata.Tick
}

// CandleHistoryProvider provides a sequence of historical market data candles.
type CandleHistoryProvider interface {
	Candles() []marketdata.Candle
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

// EmulatorOptions defines the simulated trading friction for the backtesting engine.
type EmulatorOptions struct {
	MakerFeeRate float64 // e.g., 0.001 for 0.1% maker fee
	TakerFeeRate float64 // e.g., 0.001 for 0.1% taker fee
	SlippageRate float64 // e.g., 0.0005 for 0.05% slippage on market orders
}

// DefaultEmulatorOptions provides a baseline setup with 0.1% fees and zero slippage.
func DefaultEmulatorOptions() EmulatorOptions {
	return EmulatorOptions{
		MakerFeeRate: 0.001,
		TakerFeeRate: 0.001,
		SlippageRate: 0.0,
	}
}

// Engine orchestrates the historical simulation.
type Engine struct {
	strategy strategy.Strategy
	tracker  *portfolio.Tracker
	orders   []exchange.Order
	opts     EmulatorOptions
}

// NewEngine creates a new backtesting engine with default emulator options.
func NewEngine(s strategy.Strategy, initialCapital float64) *Engine {
	return NewEngineWithOptions(s, initialCapital, DefaultEmulatorOptions())
}

// NewEngineWithOptions creates a new backtesting engine with specific emulation friction.
func NewEngineWithOptions(s strategy.Strategy, initialCapital float64, opts EmulatorOptions) *Engine {
	tracker := portfolio.NewTracker()

	return &Engine{
		strategy: s,
		tracker:  tracker,
		orders:   make([]exchange.Order, 0),
		opts:     opts,
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

// RunTicks executes the simulation over the provided tick history.
func (e *Engine) RunTicks(ctx context.Context, h HistoryProvider) (Result, error) {
	tickStrat, ok := e.strategy.(strategy.TickStrategy)
	if !ok {
		return Result{}, fmt.Errorf("strategy does not implement TickStrategy")
	}

	ticks := h.Ticks()
	if len(ticks) == 0 {
		return Result{}, fmt.Errorf("no historical tick data provided")
	}

	var lastPrice string
	var startTime time.Time = ticks[0].Timestamp
	var endTime time.Time = ticks[len(ticks)-1].Timestamp

	for _, tick := range ticks {
		lastPrice = tick.Price

		// Pass the tick to the strategy.
		signals, err := tickStrat.OnMarketTick(ctx, tick)
		if err != nil {
			return Result{}, fmt.Errorf("strategy error on tick %v: %w", tick, err)
		}

		e.processSignals(signals, tick.Price)
	}

	return e.buildResult(ctx, lastPrice, startTime, endTime), nil
}

// RunCandles executes the simulation over the provided candle history.
func (e *Engine) RunCandles(ctx context.Context, h CandleHistoryProvider) (Result, error) {
	candleStrat, ok := e.strategy.(strategy.CandleStrategy)
	if !ok {
		return Result{}, fmt.Errorf("strategy does not implement CandleStrategy")
	}

	candles := h.Candles()
	if len(candles) == 0 {
		return Result{}, fmt.Errorf("no historical candle data provided")
	}

	var lastPrice string
	var startTime time.Time = candles[0].Timestamp
	var endTime time.Time = candles[len(candles)-1].Timestamp

	for _, candle := range candles {
		lastPrice = candle.Close

		// Pass the candle to the strategy.
		signals, err := candleStrat.OnMarketCandle(ctx, candle)
		if err != nil {
			return Result{}, fmt.Errorf("strategy error on candle %v: %w", candle, err)
		}

		e.processSignals(signals, candle.Close)
	}

	return e.buildResult(ctx, lastPrice, startTime, endTime), nil
}

func (e *Engine) processSignals(signals []strategy.Signal, rawPrice string) {
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

		priceVal := utils.ParseFloat(rawPrice)

		// Apply simulated slippage and fees. All demo strategy orders are treated as Market orders.
		// A more complex emulator would check if sig.OrderType is "limit" and apply MakerFeeRate.
		if side == exchange.Buy {
			priceVal = priceVal * (1.0 + e.opts.SlippageRate) * (1.0 + e.opts.TakerFeeRate)
		} else {
			priceVal = priceVal * (1.0 - e.opts.SlippageRate) * (1.0 - e.opts.TakerFeeRate)
		}

		simulatedPriceStr := fmt.Sprintf("%f", priceVal)

		order := exchange.Order{
			ID:       fmt.Sprintf("bt-ord-%d", len(e.orders)+1),
			Symbol:   sig.Symbol,
			Side:     side,
			Type:     exchange.MarketOrder, // Treating signals as market execution by default
			Status:   "filled",
			Quantity: sig.Quantity,
			Price:    simulatedPriceStr,
		}
		e.orders = append(e.orders, order)

		// Update the simulated portfolio.
		e.tracker.Apply(order)
	}
}

func (e *Engine) buildResult(ctx context.Context, lastPrice string, startTime, endTime time.Time) Result {
	feed := &simpleFeed{currentPrice: lastPrice}
	totalValue := e.tracker.TotalMarketValue(ctx, feed)
	realized := e.tracker.TotalRealizedPnL()
	unrealized := e.tracker.TotalUnrealizedPnL(ctx, feed)

	return Result{
		TotalTrades:         len(e.orders),
		FinalPortfolioValue: totalValue,
		RealizedPnL:         realized,
		UnrealizedPnL:       unrealized,
		StartTime:           startTime,
		EndTime:             endTime,
		Orders:              e.orders,
	}
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

// MemoryCandleHistory is a simple history provider backed by an in-memory slice of candles.
type MemoryCandleHistory struct {
	data []marketdata.Candle
}

func NewMemoryCandleHistory(data []marketdata.Candle) *MemoryCandleHistory {
	return &MemoryCandleHistory{data: data}
}

func (h *MemoryCandleHistory) Candles() []marketdata.Candle {
	return h.data
}
