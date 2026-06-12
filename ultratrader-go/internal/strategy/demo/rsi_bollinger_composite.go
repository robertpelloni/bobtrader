package demo

import (
	"context"
	"fmt"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/indicator"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// RSIBollingerComposite generates buy signals when RSI drops below oversold AND
// price touches the lower Bollinger Band. It generates sell signals when RSI rises
// above overbought AND price touches the upper Bollinger Band.
type RSIBollingerComposite struct {
	accountID   string
	symbol      string
	quantity    string
	rsiPeriod   int
	oversold    float64
	overbought  float64
	bbPeriod    int
	bbMultiplier float64
	rsi         *indicator.RSI
	bb          *indicator.BollingerBands
	lastSignal  string // "buy" or "sell" — prevents repeated signals
}

func NewRSIBollingerComposite(accountID, symbol, quantity string, rsiPeriod int, oversold, overbought float64, bbPeriod int, bbMultiplier float64) *RSIBollingerComposite {
	if rsiPeriod < 2 {
		rsiPeriod = 14
	}
	if oversold <= 0 {
		oversold = 30
	}
	if overbought <= 0 {
		overbought = 70
	}
	if bbPeriod < 5 {
		bbPeriod = 20
	}
	if bbMultiplier <= 0 {
		bbMultiplier = 2.0
	}
	return &RSIBollingerComposite{
		accountID:    accountID,
		symbol:       symbol,
		quantity:     quantity,
		rsiPeriod:    rsiPeriod,
		oversold:     oversold,
		overbought:   overbought,
		bbPeriod:     bbPeriod,
		bbMultiplier: bbMultiplier,
		rsi:          indicator.NewRSI(rsiPeriod),
		bb:           indicator.NewBollingerBands(bbPeriod, bbMultiplier),
	}
}

func (s *RSIBollingerComposite) Name() string {
	return fmt.Sprintf("rsi-bb-comp-rsi%d-bb%d", s.rsiPeriod, s.bbPeriod)
}

func (s *RSIBollingerComposite) OnTick(_ context.Context) ([]strategy.Signal, error) {
	return nil, nil
}

func (s *RSIBollingerComposite) OnMarketTick(_ context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	if tick.Symbol != s.symbol {
		return nil, nil
	}
	price := utils.ParseFloat(tick.Price)
	if price <= 0 {
		return nil, nil
	}

	rsiVal := s.rsi.Update(price)
	bbResult := s.bb.Update(price)

	// Don't signal until indicators are fully warmed up
	if !s.rsi.Ready() || (bbResult.Upper == 0 && bbResult.Lower == 0) {
		return nil, nil
	}

	// Buy when RSI is oversold AND price is below lower BB
	isRSIOversold := rsiVal <= s.oversold && rsiVal > 0 && rsiVal < 100
	isBBBreakoutLow := price <= bbResult.Lower

	if isRSIOversold && isBBBreakoutLow && s.lastSignal != "buy" {
		s.lastSignal = "buy"
		return []strategy.Signal{{
			AccountID: s.accountID,
			Symbol:    s.symbol,
			Action:    "buy",
			Reason:    fmt.Sprintf("RSI(%.1f) oversold + BB lower touch(%.2f)", rsiVal, bbResult.Lower),
			Quantity:  s.quantity,
			OrderType: "market",
		}}, nil
	}

	// Sell when RSI is overbought AND price is above upper BB
	isRSIOverbought := rsiVal >= s.overbought && rsiVal > 0 && rsiVal < 100
	isBBBreakoutHigh := price >= bbResult.Upper

	if isRSIOverbought && isBBBreakoutHigh && s.lastSignal != "sell" {
		s.lastSignal = "sell"
		return []strategy.Signal{{
			AccountID: s.accountID,
			Symbol:    s.symbol,
			Action:    "sell",
			Reason:    fmt.Sprintf("RSI(%.1f) overbought + BB upper touch(%.2f)", rsiVal, bbResult.Upper),
			Quantity:  s.quantity,
			OrderType: "market",
		}}, nil
	}

	// Reset signal state when both indicators return to neutral zones
	bandWidth := bbResult.Upper - bbResult.Lower
	isBBMiddle := bandWidth > 0 && price > bbResult.Lower+bandWidth*0.25 && price < bbResult.Upper-bandWidth*0.25
	isRSIMiddle := rsiVal >= 40 && rsiVal <= 60

	if isBBMiddle && isRSIMiddle {
		s.lastSignal = ""
	}

	return nil, nil
}
