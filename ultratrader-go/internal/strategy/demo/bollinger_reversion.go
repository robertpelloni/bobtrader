package demo

import (
	"context"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/indicator"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// BollingerReversion is a candle-based mean-reversion strategy that buys when
// price touches the lower band (oversold) and sells when price touches the
// upper band (overbought).
type BollingerReversion struct {
	accountID string
	symbol    string
	quantity  string
	bb        *indicator.BollingerBands
	warmup    int
	period    int
}

func NewBollingerReversion(accountID, symbol, quantity string, period int, multiplier float64) *BollingerReversion {
	return &BollingerReversion{
		accountID: accountID,
		symbol:    symbol,
		quantity:  quantity,
		bb:        indicator.NewBollingerBands(period, multiplier),
		period:    period,
	}
}

func (s *BollingerReversion) Name() string { return "BollingerReversion" }

func (s *BollingerReversion) CandleEvent(ctx context.Context, candle marketdata.Candle) ([]strategy.Signal, error) {
	closePrice := utils.ParseFloat(candle.Close)

	s.warmup++
	result := s.bb.Update(closePrice)

	// Not enough data yet
	if s.warmup < s.period {
		return nil, nil
	}

	var signals []strategy.Signal

	// Buy when price touches or drops below the lower band (oversold)
	if closePrice <= result.Lower {
		signals = append(signals, strategy.Signal{
			AccountID: s.accountID,
			Symbol:    s.symbol,
			Action:    "buy",
			Quantity:  s.quantity,
			OrderType: "market",
		})
	}

	// Sell when price touches or exceeds the upper band (overbought)
	if closePrice >= result.Upper {
		signals = append(signals, strategy.Signal{
			AccountID: s.accountID,
			Symbol:    s.symbol,
			Action:    "sell",
			Quantity:  s.quantity,
			OrderType: "market",
		})
	}

	return signals, nil
}
