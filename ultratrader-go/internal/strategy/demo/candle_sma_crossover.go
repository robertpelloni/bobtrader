package demo

import (
	"context"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/indicator"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// CandleSMACross is a demonstration strategy that uses Candle intervals
// rather than Ticks to calculate Moving Averages and generate signals.
type CandleSMACross struct {
	AccountID string
	Symbol    string
	Quantity  string

	fastSMA *indicator.SMA
	slowSMA *indicator.SMA

	lastFast float64
	lastSlow float64
}

func NewCandleSMACross(accountID, symbol, quantity string, fastPeriod, slowPeriod int) *CandleSMACross {
	return &CandleSMACross{
		AccountID: accountID,
		Symbol:    symbol,
		Quantity:  quantity,
		fastSMA:   indicator.NewSMA(fastPeriod),
		slowSMA:   indicator.NewSMA(slowPeriod),
	}
}

func (s *CandleSMACross) Name() string {
	return "candle-sma-crossover"
}

func (s *CandleSMACross) OnTick(_ context.Context) ([]strategy.Signal, error) {
	// Not used in event-driven strategy
	return nil, nil
}

func (s *CandleSMACross) OnMarketCandle(_ context.Context, candle marketdata.Candle) ([]strategy.Signal, error) {
	if candle.Symbol != s.Symbol {
		return nil, nil
	}

	price := utils.ParseFloat(candle.Close)

	fast := s.fastSMA.Update(price)
	slow := s.slowSMA.Update(price)

	var signals []strategy.Signal

	// Check for crossovers
	if s.lastFast > 0 && s.lastSlow > 0 { // Need prior state to check cross
		if s.lastFast <= s.lastSlow && fast > slow {
			// Golden cross (buy)
			signals = append(signals, strategy.Signal{
				AccountID: s.AccountID,
				Symbol:    s.Symbol,
				Action:    "buy",
				Reason:    "golden-cross-candle",
				Quantity:  s.Quantity,
				OrderType: "market",
			})
		} else if s.lastFast >= s.lastSlow && fast < slow {
			// Death cross (sell)
			signals = append(signals, strategy.Signal{
				AccountID: s.AccountID,
				Symbol:    s.Symbol,
				Action:    "sell",
				Reason:    "death-cross-candle",
				Quantity:  s.Quantity,
				OrderType: "market",
			})
		}
	}

	s.lastFast = fast
	s.lastSlow = slow

	return signals, nil
}
