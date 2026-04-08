package strategy

import (
	"context"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

type Signal struct {
	AccountID string
	Symbol    string
	Action    string
	Reason    string
	Quantity  string
	OrderType string
}

type Strategy interface {
	Name() string
	OnTick(ctx context.Context) ([]Signal, error)
}

type TickStrategy interface {
	Strategy
	OnMarketTick(ctx context.Context, tick marketdata.Tick) ([]Signal, error)
}

type CandleStrategy interface {
	Strategy
	OnMarketCandle(ctx context.Context, candle marketdata.Candle) ([]Signal, error)
}

type Runtime struct {
	strategies []Strategy
}

func NewRuntime(strategies ...Strategy) *Runtime {
	return &Runtime{strategies: strategies}
}

func (r *Runtime) Tick(ctx context.Context) ([]Signal, error) {
	var out []Signal
	for _, strategy := range r.strategies {
		signals, err := strategy.OnTick(ctx)
		if err != nil {
			return nil, err
		}
		out = append(out, signals...)
	}
	return out, nil
}

func (r *Runtime) TickEvent(ctx context.Context, tick marketdata.Tick) ([]Signal, error) {
	var out []Signal
	for _, candidate := range r.strategies {
		strategy, ok := candidate.(TickStrategy)
		if !ok {
			continue
		}
		signals, err := strategy.OnMarketTick(ctx, tick)
		if err != nil {
			return nil, err
		}
		out = append(out, signals...)
	}
	return out, nil
}

func (r *Runtime) CandleEvent(ctx context.Context, candle marketdata.Candle) ([]Signal, error) {
	var out []Signal
	for _, candidate := range r.strategies {
		strategy, ok := candidate.(CandleStrategy)
		if !ok {
			continue
		}
		signals, err := strategy.OnMarketCandle(ctx, candle)
		if err != nil {
			return nil, err
		}
		out = append(out, signals...)
	}
	return out, nil
}
