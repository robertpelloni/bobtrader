package demo

import (
	"context"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/indicator"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// MACDCrossover is a candle-based strategy that generates buy signals when the
// MACD line crosses above the signal line (bullish crossover) and sell signals
// when the MACD line crosses below the signal line (bearish crossover).
type MACDCrossover struct {
	accountID   string
	symbol      string
	quantity    string
	macd        *indicator.MACD
	prevHist    float64
	prevMACD    float64
	prevSig     float64
	initialized bool
}

func NewMACDCrossover(accountID, symbol, quantity string, fastPeriod, slowPeriod, signalPeriod int) *MACDCrossover {
	return &MACDCrossover{
		accountID: accountID,
		symbol:    symbol,
		quantity:  quantity,
		macd:      indicator.NewMACD(fastPeriod, slowPeriod, signalPeriod),
	}
}

func (s *MACDCrossover) Name() string { return "MACDCrossover" }

func (s *MACDCrossover) CandleEvent(ctx context.Context, candle marketdata.Candle) ([]strategy.Signal, error) {
	closePrice := utils.ParseFloat(candle.Close)

	result := s.macd.Update(closePrice)

	var signals []strategy.Signal

	if s.initialized {
		// Bullish crossover: histogram crosses from negative to positive
		if s.prevHist <= 0 && result.Histogram > 0 {
			signals = append(signals, strategy.Signal{
				AccountID: s.accountID,
				Symbol:    s.symbol,
				Action:    "buy",
				Quantity:  s.quantity,
				OrderType: "market",
			})
		}
		// Bearish crossover: histogram crosses from positive to negative
		if s.prevHist >= 0 && result.Histogram < 0 {
			signals = append(signals, strategy.Signal{
				AccountID: s.accountID,
				Symbol:    s.symbol,
				Action:    "sell",
				Quantity:  s.quantity,
				OrderType: "market",
			})
		}
	}

	s.prevHist = result.Histogram
	s.prevMACD = result.MACD
	s.prevSig = result.Signal
	s.initialized = true

	return signals, nil
}
