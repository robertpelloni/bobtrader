package scheduler

import (
	"context"
	"fmt"
	"strings"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/risk"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/execution"
)

type Scheduler struct {
	runtime   *strategy.Runtime
	execution *execution.Service
}

func New(runtime *strategy.Runtime, execution *execution.Service) *Scheduler {
	return &Scheduler{runtime: runtime, execution: execution}
}

func (s *Scheduler) RunOnce(ctx context.Context) error {
	signals, err := s.runtime.Tick(ctx)
	if err != nil {
		return fmt.Errorf("runtime tick: %w", err)
	}
	return s.executeSignals(ctx, signals)
}

func (s *Scheduler) RunTick(ctx context.Context, tick marketdata.Tick) error {
	signals, err := s.runtime.TickEvent(ctx, tick)
	if err != nil {
		return fmt.Errorf("runtime tick event: %w", err)
	}
	return s.executeSignals(ctx, signals)
}

func (s *Scheduler) executeSignals(ctx context.Context, signals []strategy.Signal) error {
	for _, signal := range signals {
		request, intent, err := toOrder(signal)
		if err != nil {
			return err
		}
		if _, err := s.execution.Execute(ctx, signal.AccountID, request, intent); err != nil {
			return fmt.Errorf("execute signal for %s: %w", signal.Symbol, err)
		}
	}
	return nil
}

func toOrder(signal strategy.Signal) (exchange.OrderRequest, risk.OrderIntent, error) {
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
	intent := risk.OrderIntent{
		AccountID: signal.AccountID,
		Symbol:    signal.Symbol,
		Notional:  1,
	}
	return request, intent, nil
}
