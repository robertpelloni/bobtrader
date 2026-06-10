package demo

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// CrossExchangeArbitrage detects price differences between exchanges.
// When Binance has BTC at $61,000 and Coinbase has it at $61,200,
// we buy on Binance and sell on Coinbase for risk-free profit.
type CrossExchangeArbitrage struct {
	accountID       string
	symbol          string
	quantity        string
	minSpreadPct    float64 // Minimum spread to trigger (e.g., 0.1%)
	maxSpreadPct    float64 // Maximum spread (beyond this, probably bad data)
	exchangePrices  map[string]float64
	mu              sync.RWMutex
	lastSignalTime  time.Time
	cooldown        time.Duration
}

func NewCrossExchangeArbitrage(
	accountID, symbol, quantity string,
	minSpreadPct, maxSpreadPct float64,
) *CrossExchangeArbitrage {
	if minSpreadPct <= 0 {
		minSpreadPct = 0.05 // 0.05% minimum spread
	}
	if maxSpreadPct <= 0 {
		maxSpreadPct = 2.0 // 2% max spread (beyond this = bad data)
	}
	return &CrossExchangeArbitrage{
		accountID:      accountID,
		symbol:         symbol,
		quantity:       quantity,
		minSpreadPct:   minSpreadPct,
		maxSpreadPct:   maxSpreadPct,
		exchangePrices: make(map[string]float64),
		cooldown:       30 * time.Second,
	}
}

func (s *CrossExchangeArbitrage) Name() string {
	return fmt.Sprintf("cross-exchange-arb-%s", s.symbol)
}

func (s *CrossExchangeArbitrage) OnTick(_ context.Context) ([]strategy.Signal, error) {
	return nil, nil
}

// UpdatePrice updates the price from a specific exchange
func (s *CrossExchangeArbitrage) UpdatePrice(exchange string, price float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.exchangePrices[exchange] = price
}

func (s *CrossExchangeArbitrage) OnMarketTick(_ context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	if tick.Symbol != s.symbol {
		return nil, nil
	}

	price := utils.ParseFloat(tick.Price)
	if price <= 0 {
		return nil, nil
	}

	s.mu.Lock()
	s.exchangePrices[tick.Source] = price
	s.mu.Unlock()

	// Need at least 2 exchanges to arbitrage
	s.mu.RLock()
	exchanges := make([]string, 0, len(s.exchangePrices))
	prices := make(map[string]float64)
	for ex, p := range s.exchangePrices {
		exchanges = append(exchanges, ex)
		prices[ex] = p
	}
	s.mu.RUnlock()

	if len(exchanges) < 2 {
		return nil, nil
	}

	// Cooldown check
	if time.Since(s.lastSignalTime) < s.cooldown {
		return nil, nil
	}

	var signals []strategy.Signal

	// Find the lowest and highest priced exchanges
	var lowestEx, highestEx string
	lowestPrice := math.MaxFloat64
	highestPrice := 0.0

	for ex, p := range prices {
		if p < lowestPrice {
			lowestPrice = p
			lowestEx = ex
		}
		if p > highestPrice {
			highestPrice = p
			highestEx = ex
		}
	}

	// Calculate spread
	spreadPct := (highestPrice - lowestPrice) / lowestPrice * 100

	if spreadPct >= s.minSpreadPct && spreadPct <= s.maxSpreadPct {
		// Arbitrage opportunity detected!
		signals = append(signals, strategy.Signal{
			StrategyName: s.Name(),
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "buy",
			Quantity:     s.quantity,
			Reason: fmt.Sprintf("ARBITRAGE: buy on %s at $%.2f, sell on %s at $%.2f (spread: %.3f%%)",
				lowestEx, lowestPrice, highestEx, highestPrice, spreadPct),
		})
		s.lastSignalTime = time.Now()
	}

	return signals, nil
}

// GetSpreadInfo returns current spread information for diagnostics
func (s *CrossExchangeArbitrage) GetSpreadInfo() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info := map[string]interface{}{
		"exchanges": len(s.exchangePrices),
		"prices":    s.exchangePrices,
	}

	if len(s.exchangePrices) >= 2 {
		lowest := math.MaxFloat64
		highest := 0.0
		for _, p := range s.exchangePrices {
			if p < lowest {
				lowest = p
			}
			if p > highest {
				highest = p
			}
		}
		info["spread_pct"] = (highest - lowest) / lowest * 100
		info["lowest"] = lowest
		info["highest"] = highest
	}

	return info
}
