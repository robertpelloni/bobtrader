package scheduler

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/risk"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/execution"
)

// PositionChecker knows whether we hold a position in a symbol and its quantity.
type PositionChecker interface {
	HasOpenPosition(symbol string) bool
	PositionQuantity(symbol string) float64
}

// SmartDispatcher wraps a base Scheduler and adds:
//   - Position awareness (skip buy if already long, skip sell if flat)
//   - Real notional calculation using market data prices
//   - Signal logging for every signal + outcome
type SmartDispatcher struct {
	inner     *Scheduler
	portfolio PositionChecker
	feed      marketdata.Feed
	signalLog *strategy.SignalLog
}

// NewSmartDispatcher creates a dispatcher with position awareness and signal logging.
func NewSmartDispatcher(
	inner *Scheduler,
	portfolio PositionChecker,
	feed marketdata.Feed,
	signalLog *strategy.SignalLog,
) *SmartDispatcher {
	return &SmartDispatcher{
		inner:     inner,
		portfolio: portfolio,
		feed:      feed,
		signalLog: signalLog,
	}
}

func (d *SmartDispatcher) RunOnce(ctx context.Context) error {
	return d.inner.RunOnce(ctx)
}

func (d *SmartDispatcher) RunTick(ctx context.Context, tick marketdata.Tick) error {
	return d.inner.RunTick(ctx, tick)
}

func (d *SmartDispatcher) RunCandle(ctx context.Context, candle marketdata.Candle) error {
	return d.inner.RunCandle(ctx, candle)
}

// smartToOrder builds an order with real market-price-based notional.
func smartToOrder(signal strategy.Signal, marketPrice float64) (exchange.OrderRequest, risk.OrderIntent, error) {
	side := exchange.Buy
	switch strings.ToLower(signal.Action) {
	case "buy":
		side = exchange.Buy
	case "sell":
		side = exchange.Sell
	default:
		return exchange.OrderRequest{}, risk.OrderIntent{}, fmt.Errorf("unsupported action %q", signal.Action)
	}

	orderType := exchange.MarketOrder
	if strings.EqualFold(signal.OrderType, "limit") {
		orderType = exchange.LimitOrder
	}

	request := exchange.OrderRequest{
		Symbol:   signal.Symbol,
		Side:     side,
		Type:     orderType,
		Quantity: signal.Quantity,
	}

	// Calculate real notional using market price
	notional := 1.0
	qty := utils.ParseFloat(signal.Quantity)
	if marketPrice > 0 && qty > 0 {
		notional = marketPrice * qty
	}

	intent := risk.OrderIntent{
		AccountID: signal.AccountID,
		Symbol:    signal.Symbol,
		Side:      riskSide(signal.Action),
		Notional:  notional,
	}

	return request, intent, nil
}

// extractStrategyName tries to pull a strategy name from the signal reason.
func extractStrategyName(reason string) string {
	if reason == "" {
		return "unknown"
	}
	parts := strings.SplitN(reason, ":", 2)
	name := strings.TrimSpace(parts[0])
	if name == "" {
		name = "unknown"
	}
	return name
}

func formatDispQuantity(qty float64) string {
	if qty >= 1 {
		return fmt.Sprintf("%.4f", qty)
	}
	if qty >= 0.01 {
		return fmt.Sprintf("%.6f", qty)
	}
	return fmt.Sprintf("%.8f", qty)
}

func riskSide(action string) risk.OrderSide {
	switch strings.ToLower(action) {
	case "buy":
		return risk.BuySide
	case "sell":
		return risk.SellSide
	default:
		return risk.BuySide
	}
}

// ExecuteSignals evaluates and dispatches signals with full position awareness
// and signal logging. This replaces the scheduler's internal executeSignals.
func ExecuteSignals(ctx context.Context, signals []strategy.Signal, execService *execution.Service, portfolio PositionChecker, feed marketdata.Feed, signalLog *strategy.SignalLog) {
	for _, signal := range signals {
		// Position awareness: skip if already in position
		if portfolio != nil {
			hasPosition := portfolio.HasOpenPosition(signal.Symbol)
			if strings.EqualFold(signal.Action, "buy") && hasPosition {
				recordSignal(signal, strategy.OutcomeSkipped, "already-in-position", "", "", feed, signalLog)
				continue
			}
			if strings.EqualFold(signal.Action, "sell") && !hasPosition {
				recordSignal(signal, strategy.OutcomeSkipped, "no-position-to-sell", "", "", feed, signalLog)
				continue
			}
			// For sell signals, override quantity with full position if needed
			if strings.EqualFold(signal.Action, "sell") && portfolio != nil {
				heldQty := portfolio.PositionQuantity(signal.Symbol)
				if heldQty > 0 {
					signal.Quantity = formatDispQuantity(heldQty)
				}
			}
		}

		// Get market price for notional calculation
		var price float64
		if feed != nil {
			tick, err := feed.LatestTick(ctx, signal.Symbol)
			if err == nil && tick.Price != "" {
				price = utils.ParseFloat(tick.Price)
			}
		}

		// Build order request with real price
		request, intent, err := smartToOrder(signal, price)
		if err != nil {
			recordSignal(signal, strategy.OutcomeBlocked, err.Error(), "", "", feed, signalLog)
			continue
		}

		order, execErr := execService.Execute(ctx, signal.AccountID, request, intent)
		if execErr != nil {
			blockedBy := "unknown"
			var guardErr risk.GuardError
			if errors.As(execErr, &guardErr) {
				blockedBy = guardErr.GuardName
			}
			recordSignal(signal, strategy.OutcomeBlocked, blockedBy, "", "", feed, signalLog)
			continue
		}

		recordSignal(signal, strategy.OutcomeExecuted, "", order.Price, order.ID, feed, signalLog)
	}
}

func recordSignal(signal strategy.Signal, outcome strategy.SignalOutcome, blockedBy, fillPrice, orderID string, feed marketdata.Feed, signalLog *strategy.SignalLog) {
	if signalLog == nil {
		return
	}
	price := ""
	if feed != nil {
		tick, err := feed.LatestTick(context.Background(), signal.Symbol)
		if err == nil {
			price = tick.Price
		}
	}
	signalLog.Record(strategy.LoggedSignal{
		Strategy:  signal.StrategyName,
		Symbol:    signal.Symbol,
		Action:    signal.Action,
		Quantity:  signal.Quantity,
		Price:     price,
		Reason:    signal.Reason,
		Outcome:   outcome,
		BlockedBy: blockedBy,
		FillPrice: fillPrice,
		OrderID:   orderID,
	})
}
