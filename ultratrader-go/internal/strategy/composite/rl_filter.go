package composite

import (
	"context"
	"fmt"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/analytics/features"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/analytics/rl"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// RLFilter wraps a strategy and uses reinforcement learning to "vote"
// on whether to execute a signal. It learns from realized PnL.
type RLFilter struct {
	mu        sync.Mutex
	inner     strategy.Strategy
	agent     *rl.QLearningAgent
	extractor *features.Extractor
	lastState rl.State
	lastAct   rl.Action
	lastSig   *strategy.Signal
}

func NewRLFilter(inner strategy.Strategy, agent *rl.QLearningAgent) *RLFilter {
	return &RLFilter{
		inner:     inner,
		agent:     agent,
		extractor: features.NewExtractor(14, 14),
	}
}

func (f *RLFilter) Name() string {
	return fmt.Sprintf("rl-filter(%s)", f.inner.Name())
}

func (f *RLFilter) OnTick(ctx context.Context) ([]strategy.Signal, error) {
	// RL filter mainly operates on events where features can be extracted
	return f.inner.OnTick(ctx)
}

func (f *RLFilter) OnMarketTick(ctx context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	ts, ok := f.inner.(strategy.TickStrategy)
	if !ok {
		return nil, nil
	}

	signals, err := ts.OnMarketTick(ctx, tick)
	if err != nil {
		return nil, err
	}

	return f.filter(ctx, signals, tick), nil
}

func (f *RLFilter) OnMarketCandle(ctx context.Context, candle marketdata.Candle) ([]strategy.Signal, error) {
	cs, ok := f.inner.(strategy.CandleStrategy)
	if !ok {
		return nil, nil
	}

	// Update features
	fMap := f.extractor.Update(features.CandleData{
		Open:   utils.ParseFloat(candle.Open),
		High:   utils.ParseFloat(candle.High),
		Low:    utils.ParseFloat(candle.Low),
		Close:  utils.ParseFloat(candle.Close),
		Volume: utils.ParseFloat(candle.Volume),
	})

	// Discretize state for Q-Table
	names := f.extractor.Names()
	vals := make([]float64, len(names))
	for i, n := range names {
		vals[i] = fMap[n]
	}
	newState := rl.Discretize(vals, 5)

	f.mu.Lock()
	f.lastState = newState
	f.mu.Unlock()

	signals, err := cs.OnMarketCandle(ctx, candle)
	if err != nil {
		return nil, err
	}

	// We pass 0 as dummy price for filter since we only use tick price for real filtering
	return f.filter(ctx, signals, marketdata.Tick{Price: candle.Close}), nil
}

func (f *RLFilter) filter(ctx context.Context, signals []strategy.Signal, tick marketdata.Tick) []strategy.Signal {
	if len(signals) == 0 {
		return nil
	}

	f.mu.Lock()
	state := f.lastState
	f.mu.Unlock()

	// RL Decision: Should we allow this strategy to trade in this state?
	// Action 0 = Hold/Block, Action 1 = Buy/Allow, Action 2 = Sell/Allow
	decision := f.agent.ChooseAction(state)

	var out []strategy.Signal
	for _, s := range signals {
		allowed := false
		if s.Action == "buy" && decision == rl.Buy {
			allowed = true
		} else if s.Action == "sell" && (decision == rl.Sell || decision == rl.Buy) {
			// Sells often allowed to exit positions even if RL is uncertain
			allowed = true
		}

		if allowed {
			s.Reason = fmt.Sprintf("%s [RL: Confirmed]", s.Reason)
			out = append(out, s)
		}
	}

	return out
}

// UpdateReward should be called by a higher-level manager (like Siphoning or Execution)
// to provide feedback to the RL agent.
func (f *RLFilter) UpdateReward(symbol string, pnl float64) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Simple reward logic: PnL is the reward
	// In a real system, we'd need to map which state/action led to this PnL.
	// For this demo, we assume the last action taken by this filter is being rewarded.
	f.agent.Update(f.lastState, f.lastAct, pnl, f.lastState)
}

var _ strategy.Strategy = (*RLFilter)(nil)
var _ strategy.TickStrategy = (*RLFilter)(nil)
var _ strategy.CandleStrategy = (*RLFilter)(nil)
