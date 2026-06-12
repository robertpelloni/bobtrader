package demo

import (
	"context"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// PortfolioSizer wraps another TickStrategy and adjusts signal quantities
// based on portfolio balance and risk parameters. Instead of fixed quantities,
// it calculates position size from:
//   - Available USDT balance (via BalanceReader)
//   - Risk per trade as % of portfolio
//   - Current market price
//
// For sell signals, it passes through the original quantity.
type PortfolioSizer struct {
	inner       strategy.TickStrategy
	symbol      string
	balance     BalanceReader
	feed        marketdata.Feed
	riskPct     float64 // % of balance to risk per trade
	maxNotional float64 // absolute max notional per trade
}

// BalanceReader reads the available USDT balance.
type BalanceReader interface {
	USDTBalance() float64
}

func NewPortfolioSizer(
	inner strategy.TickStrategy,
	symbol string,
	balance BalanceReader,
	feed marketdata.Feed,
	riskPct float64,
	maxNotional float64,
) *PortfolioSizer {
	if riskPct <= 0 {
		riskPct = 2.0 // default 2% of balance per trade
	}
	if maxNotional <= 0 {
		maxNotional = 1000 // default max
	}
	return &PortfolioSizer{
		inner:       inner,
		symbol:      symbol,
		balance:     balance,
		feed:        feed,
		riskPct:     riskPct,
		maxNotional: maxNotional,
	}
}

func (s *PortfolioSizer) Name() string {
	return s.inner.Name()
}

func (s *PortfolioSizer) OnTick(ctx context.Context) ([]strategy.Signal, error) {
	return nil, nil
}

func (s *PortfolioSizer) OnMarketTick(ctx context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	signals, err := s.inner.OnMarketTick(ctx, tick)
	if err != nil {
		return nil, err
	}
	if len(signals) == 0 {
		return nil, nil
	}

	// Resize each signal
	out := make([]strategy.Signal, 0, len(signals))
	for _, sig := range signals {
		if sig.Symbol != s.symbol {
			out = append(out, sig)
			continue
		}

		if sig.Action == "buy" {
			sig.Quantity = s.sizeBuy(ctx, tick)
		}
		// Sell quantity passes through (set by TrailingTakeProfit to position qty)

		out = append(out, sig)
	}
	return out, nil
}

func (s *PortfolioSizer) sizeBuy(ctx context.Context, tick marketdata.Tick) string {
	// Get current price
	price := utils.ParseFloat(tick.Price)
	if price <= 0 {
		if s.feed != nil {
			t, err := s.feed.LatestTick(ctx, s.symbol)
			if err == nil {
				price = utils.ParseFloat(t.Price)
			}
		}
	}
	if price <= 0 {
		return "0.001" // absolute minimum fallback
	}

	// Get available balance
	balance := 1000.0 // default fallback
	if s.balance != nil {
		balance = s.balance.USDTBalance()
	}
	if balance <= 0 {
		return "0" // no balance, no trade
	}

	// Calculate notional = balance * riskPct / 100
	notional := balance * s.riskPct / 100.0

	// Clamp to max notional
	if notional > s.maxNotional {
		notional = s.maxNotional
	}

	// Don't exceed available balance (leave 5% buffer for fees)
	if notional > balance*0.95 {
		notional = balance * 0.95
	}

	// Calculate quantity = notional / price
	quantity := notional / price

	if quantity <= 0 {
		return "0"
	}

	return formatQuantity(s.symbol, quantity)
}
