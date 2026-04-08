package demo

import (
	"context"
	"fmt"
	"math"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/indicator"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// ATRSizing is a candle-based strategy that combines SMA crossover for signal
// generation with ATR-based dynamic position sizing. It scales the order
// quantity inversely with volatility — smaller positions in volatile markets,
// larger positions in calm markets.
type ATRSizing struct {
	accountID    string
	symbol       string
	baseQuantity string
	riskPerTrade float64 // fraction of capital to risk per trade
	fastSMA      *indicator.SMA
	slowSMA      *indicator.SMA
	atr          *indicator.ATR
	fastPeriod   int
	slowPeriod   int
	warmup       int
	prevFast     float64
	prevSlow     float64
}

func NewATRSizing(accountID, symbol, baseQuantity string, riskPerTrade float64, fastPeriod, slowPeriod, atrPeriod int) *ATRSizing {
	return &ATRSizing{
		accountID:    accountID,
		symbol:       symbol,
		baseQuantity: baseQuantity,
		riskPerTrade: riskPerTrade,
		fastSMA:      indicator.NewSMA(fastPeriod),
		slowSMA:      indicator.NewSMA(slowPeriod),
		atr:          indicator.NewATR(atrPeriod),
		fastPeriod:   fastPeriod,
		slowPeriod:   slowPeriod,
	}
}

func (s *ATRSizing) Name() string { return "ATRSizing" }

func (s *ATRSizing) CandleEvent(ctx context.Context, candle marketdata.Candle) ([]strategy.Signal, error) {
	closePrice := utils.ParseFloat(candle.Close)
	high := utils.ParseFloat(candle.High)
	low := utils.ParseFloat(candle.Low)

	s.warmup++

	fastVal := s.fastSMA.Update(closePrice)
	slowVal := s.slowSMA.Update(closePrice)
	_ = s.atr.Update(high, low, closePrice)

	// Need enough data for both SMAs
	if s.warmup < s.slowPeriod {
		s.prevFast = fastVal
		s.prevSlow = slowVal
		return nil, nil
	}

	var signals []strategy.Signal

	// ATR-based position sizing: scale quantity inversely with volatility
	// Higher ATR → smaller position, lower ATR → larger position
	atrVal := s.atr.Last()
	baseQty := utils.ParseFloat(s.baseQuantity)
	quantity := baseQty
	if atrVal > 0 {
		// Normalize: if ATR is large, reduce size; if small, increase
		// Use ratio of baseQuantity risk to ATR as a scaling factor
		quantity = (s.riskPerTrade * closePrice) / atrVal
		// Floor at a minimum
		quantity = math.Max(quantity, 0.001)
	}

	qtyStr := fmt.Sprintf("%.6f", quantity)

	// Bullish crossover: fast SMA crosses above slow SMA
	if s.prevFast <= s.prevSlow && fastVal > slowVal {
		signals = append(signals, strategy.Signal{
			AccountID: s.accountID,
			Symbol:    s.symbol,
			Action:    "buy",
			Quantity:  qtyStr,
			OrderType: "market",
		})
	}

	// Bearish crossover: fast SMA crosses below slow SMA
	if s.prevFast >= s.prevSlow && fastVal < slowVal {
		signals = append(signals, strategy.Signal{
			AccountID: s.accountID,
			Symbol:    s.symbol,
			Action:    "sell",
			Quantity:  qtyStr,
			OrderType: "market",
		})
	}

	s.prevFast = fastVal
	s.prevSlow = slowVal

	return signals, nil
}
