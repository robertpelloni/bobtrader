package demo

import (
	"context"
	"fmt"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// USDTStablecoinScalp trades USDT stablecoin fluctuations around the $1.00 peg.
// Strategy: buy when price dips below buyThreshold, sell when price recovers to sellThreshold.
// Designed for the 0.9991-1.0000 range with occasional depeg events.
type USDTStablecoinScalp struct {
	accountID     string
	symbol        string
	quantity      string
	buyThreshold  float64 // e.g., 0.9992 - buy below this
	sellThreshold float64 // e.g., 0.9999 - sell above this
	stopLoss      float64 // e.g., 0.9800 - panic sell if below this
	maxPosition   float64 // max USDT to hold in notional
	lastBuyPrice  float64
	inPosition    bool
	priceHistory  []float64
	maxHistory    int
}

func NewUSDTStablecoinScalp(
	accountID, symbol, quantity string,
	buyThreshold, sellThreshold, stopLoss, maxPosition float64,
) *USDTStablecoinScalp {
	if buyThreshold <= 0 {
		buyThreshold = 0.9992
	}
	if sellThreshold <= 0 {
		sellThreshold = 0.9999
	}
	if stopLoss <= 0 {
		stopLoss = 0.9800
	}
	if maxPosition <= 0 {
		maxPosition = 1000.0
	}
	return &USDTStablecoinScalp{
		accountID:     accountID,
		symbol:        symbol,
		quantity:      quantity,
		buyThreshold:  buyThreshold,
		sellThreshold: sellThreshold,
		stopLoss:      stopLoss,
		maxPosition:   maxPosition,
		priceHistory:  make([]float64, 0, 100),
		maxHistory:    100,
	}
}

func (s *USDTStablecoinScalp) Name() string {
	return fmt.Sprintf("usdt-scalp-%s", s.symbol)
}

func (s *USDTStablecoinScalp) OnTick(_ context.Context) ([]strategy.Signal, error) {
	return nil, nil
}

func (s *USDTStablecoinScalp) OnMarketTick(_ context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	if tick.Symbol != s.symbol {
		return nil, nil
	}

	price := utils.ParseFloat(tick.Price)
	if price <= 0 {
		return nil, nil
	}

	// Track price history for trend analysis
	s.priceHistory = append(s.priceHistory, price)
	if len(s.priceHistory) > s.maxHistory {
		s.priceHistory = s.priceHistory[1:]
	}

	// Need at least a few data points
	if len(s.priceHistory) < 3 {
		return nil, nil
	}

	// Calculate simple moving average of last 5 prices
	sma := s.calculateSMA(5)

	var signals []strategy.Signal

	// STOP LOSS: If USDT depegs severely, panic sell
	if s.inPosition && price < s.stopLoss {
		signals = append(signals, strategy.Signal{
			StrategyName: s.Name(),
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "sell",
			Quantity:     s.quantity,
			Reason:       fmt.Sprintf("DEPEG STOP LOSS: price %.4f below stop %.4f", price, s.stopLoss),
		})
		s.inPosition = false
		s.lastBuyPrice = 0
		return signals, nil
	}

	// SELL: Price recovered to sell threshold (take profit)
	if s.inPosition && price >= s.sellThreshold {
		profitPct := (price - s.lastBuyPrice) / s.lastBuyPrice * 100
		signals = append(signals, strategy.Signal{
			StrategyName: s.Name(),
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "sell",
			Quantity:     s.quantity,
			Reason: fmt.Sprintf("TAKE PROFIT: price %.4f >= threshold %.4f (bought at %.4f, ~%.4f%% profit)",
				price, s.sellThreshold, s.lastBuyPrice, profitPct),
		})
		s.inPosition = false
		s.lastBuyPrice = 0
		return signals, nil
	}

	// BUY: Price dipped below buy threshold AND trending back up (SMA > current)
	if !s.inPosition && price <= s.buyThreshold {
		// Confirm it's bouncing: current price should be near or above recent average
		// This prevents buying into a falling knife
		if price >= sma*0.9995 { // Allow tiny margin below SMA
			signals = append(signals, strategy.Signal{
				StrategyName: s.Name(),
				AccountID:    s.accountID,
				Symbol:       s.symbol,
				Action:       "buy",
				Quantity:     s.quantity,
				Reason: fmt.Sprintf("STABLECOIN DIP: price %.4f <= threshold %.4f (SMA=%.4f)",
					price, s.buyThreshold, sma),
			})
			s.inPosition = true
			s.lastBuyPrice = price
		}
	}

	return signals, nil
}

func (s *USDTStablecoinScalp) calculateSMA(period int) float64 {
	if len(s.priceHistory) == 0 {
		return 0
	}
	start := len(s.priceHistory) - period
	if start < 0 {
		start = 0
	}
	sum := 0.0
	count := 0
	for _, p := range s.priceHistory[start:] {
		sum += p
		count++
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}
