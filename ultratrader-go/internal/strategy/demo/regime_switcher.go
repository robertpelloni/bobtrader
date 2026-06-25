package demo

import (
	"context"
	"fmt"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy/regime"
)

// RegimeSwitcher dynamically activates sub-strategies based on market state.
type RegimeSwitcher struct {
	mu            sync.RWMutex
	regimeScanner *MacroRegimeStrategy
	scalper       strategy.Strategy
	trendFollower strategy.Strategy
	symbol        string
}

func NewRegimeSwitcher(symbol string, scanner *MacroRegimeStrategy, scalper, trend strategy.Strategy) *RegimeSwitcher {
	return &RegimeSwitcher{
		regimeScanner: scanner,
		scalper:       scalper,
		trendFollower: trend,
		symbol:        symbol,
	}
}

func (s *RegimeSwitcher) Name() string {
	return fmt.Sprintf("regime-switcher-%s", s.symbol)
}

func (s *RegimeSwitcher) OnTick(ctx context.Context) ([]strategy.Signal, error) {
	active := s.getActiveStrategy()
	if active == nil {
		return nil, nil
	}
	return active.OnTick(ctx)
}

func (s *RegimeSwitcher) OnMarketTick(ctx context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	active := s.getActiveStrategy()
	if ts, ok := active.(strategy.TickStrategy); ok {
		return ts.OnMarketTick(ctx, tick)
	}
	return nil, nil
}

func (s *RegimeSwitcher) OnMarketCandle(ctx context.Context, candle marketdata.Candle) ([]strategy.Signal, error) {
	// First update the scanner
	_, _ = s.regimeScanner.OnMarketCandle(ctx, candle)

	active := s.getActiveStrategy()
	if cs, ok := active.(strategy.CandleStrategy); ok {
		return cs.OnMarketCandle(ctx, candle)
	}
	return nil, nil
}

func (s *RegimeSwitcher) getActiveStrategy() strategy.Strategy {
	current := s.regimeScanner.CurrentRegime()

	switch current {
	case regime.RegimeTrending:
		return s.trendFollower
	case regime.RegimeRanging, regime.RegimeVolatile:
		return s.scalper
	default:
		return nil
	}
}

var _ strategy.Strategy = (*RegimeSwitcher)(nil)
var _ strategy.TickStrategy = (*RegimeSwitcher)(nil)
var _ strategy.CandleStrategy = (*RegimeSwitcher)(nil)
