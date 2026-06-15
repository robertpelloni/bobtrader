package demo

import (
	"context"
	"fmt"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy/sizing"
)

// AdaptiveKellySizer wraps a strategy and uses the Kelly Criterion
// to size positions based on the strategy's live performance stats.
type AdaptiveKellySizer struct {
	inner       strategy.Strategy
	log         *strategy.SignalLog
	balance     BalanceReader
	feed        marketdata.Feed
	maxNotional float64
	fraction    float64 // Kelly fraction (e.g. 0.5 for half-Kelly)
}

func NewAdaptiveKellySizer(
	inner strategy.Strategy,
	log *strategy.SignalLog,
	balance BalanceReader,
	feed marketdata.Feed,
	maxNotional float64,
	fraction float64,
) *AdaptiveKellySizer {
	if fraction <= 0 {
		fraction = 0.5
	}
	return &AdaptiveKellySizer{
		inner:       inner,
		log:         log,
		balance:     balance,
		feed:        feed,
		maxNotional: maxNotional,
		fraction:    fraction,
	}
}

func (s *AdaptiveKellySizer) Name() string {
	return s.inner.Name()
}

func (s *AdaptiveKellySizer) OnTick(ctx context.Context) ([]strategy.Signal, error) {
	signals, err := s.inner.OnTick(ctx)
	if err != nil {
		return nil, err
	}
	return s.resize(ctx, signals), nil
}

func (s *AdaptiveKellySizer) OnMarketTick(ctx context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	ts, ok := s.inner.(strategy.TickStrategy)
	if !ok {
		return nil, nil
	}
	signals, err := ts.OnMarketTick(ctx, tick)
	if err != nil {
		return nil, err
	}
	return s.resize(ctx, signals), nil
}

func (s *AdaptiveKellySizer) OnMarketCandle(ctx context.Context, candle marketdata.Candle) ([]strategy.Signal, error) {
	cs, ok := s.inner.(strategy.CandleStrategy)
	if !ok {
		return nil, nil
	}
	signals, err := cs.OnMarketCandle(ctx, candle)
	if err != nil {
		return nil, err
	}
	return s.resize(ctx, signals), nil
}

func (s *AdaptiveKellySizer) resize(ctx context.Context, signals []strategy.Signal) []strategy.Signal {
	if len(signals) == 0 {
		return nil
	}

	stats := s.log.StatsByStrategy()[s.inner.Name()]

	// Default to conservative sizing if not enough trade history
	winRate := stats.WinRate
	if stats.WinTrades + stats.LossTrades < 10 {
		winRate = 0.5 // Start neutral
	}

	// Assume 2:1 reward/risk for Kelly if not tracked (could be improved)
	rewardRisk := 2.0

	ks := sizing.NewKellySizer(winRate, rewardRisk, s.fraction)

	balance := 1000.0
	if s.balance != nil {
		balance = s.balance.USDTBalance()
	}

	out := make([]strategy.Signal, 0, len(signals))
	for _, sig := range signals {
		if sig.Action == "buy" {
			price := 0.0
			tick, err := s.feed.LatestTick(ctx, sig.Symbol)
			if err == nil {
				price = utils.ParseFloat(tick.Price)
			}

			if price > 0 {
				qty := ks.Size(sizing.SizingInput{
					PortfolioValue: balance,
					Price:          price,
				})

				// Clamp to max notional
				notional := qty * price
				if notional > s.maxNotional {
					qty = s.maxNotional / price
				}

				if qty > 0 {
					sig.Quantity = fmt.Sprintf("%.6f", qty)
					sig.Reason = fmt.Sprintf("%s [Kelly Size: WR %.1f%%]", sig.Reason, winRate*100)
				}
			}
		}
		out = append(out, sig)
	}
	return out
}

var _ strategy.Strategy = (*AdaptiveKellySizer)(nil)
var _ strategy.TickStrategy = (*AdaptiveKellySizer)(nil)
var _ strategy.CandleStrategy = (*AdaptiveKellySizer)(nil)
