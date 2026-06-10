package demo

import (
	"context"
	"fmt"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// WhaleAlertStrategy trades based on large whale movements.
// Key signals:
// - Large outflow from exchange → Bullish (whales accumulating)
// - Large inflow to exchange → Bearish (whales preparing to sell)
// - Spike in whale activity → High volatility incoming
type WhaleAlertStrategy struct {
	accountID      string
	symbol         string
	quantity       string
	whaleSentiment float64 // Current whale sentiment score
	lastUpdateTime time.Time
	cooldown       time.Duration
	lastSignalTime time.Time
	priceHistory   []float64
	maxHistory     int
}

func NewWhaleAlertStrategy(
	accountID, symbol, quantity string,
) *WhaleAlertStrategy {
	return &WhaleAlertStrategy{
		accountID:    accountID,
		symbol:       symbol,
		quantity:     quantity,
		cooldown:     5 * time.Minute,
		priceHistory: make([]float64, 0, 100),
		maxHistory:   100,
	}
}

func (s *WhaleAlertStrategy) Name() string {
	return fmt.Sprintf("whale-alert-%s", s.symbol)
}

func (s *WhaleAlertStrategy) OnTick(_ context.Context) ([]strategy.Signal, error) {
	return nil, nil
}

// UpdateWhaleSentiment is called by the sentiment engine to update whale data
func (s *WhaleAlertStrategy) UpdateWhaleSentiment(score float64) {
	s.whaleSentiment = score
	s.lastUpdateTime = time.Now()
}

func (s *WhaleAlertStrategy) OnMarketTick(_ context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	if tick.Symbol != s.symbol {
		return nil, nil
	}

	price := utils.ParseFloat(tick.Price)
	if price <= 0 {
		return nil, nil
	}

	// Track price history
	s.priceHistory = append(s.priceHistory, price)
	if len(s.priceHistory) > s.maxHistory {
		s.priceHistory = s.priceHistory[1:]
	}

	// Cooldown check
	if time.Since(s.lastSignalTime) < s.cooldown {
		return nil, nil
	}

	// Need some history and recent whale data
	if len(s.priceHistory) < 20 {
		return nil, nil
	}

	// Whale data should be recent (within 10 minutes)
	if time.Since(s.lastUpdateTime) > 10*time.Minute {
		return nil, nil
	}

	var signals []strategy.Signal

	// Calculate price trend (20-period SMA)
	sma := calculateSMA(s.priceHistory, 20)
	priceVsSMA := (price - sma) / sma * 100

	// ── BUY SIGNAL: Whale accumulation + price dip ──
	// Whales are moving coins OFF exchanges (bullish)
	// AND price is below SMA (buying the dip)
	if s.whaleSentiment > 0.3 && priceVsSMA < -0.2 {
		signals = append(signals, strategy.Signal{
			StrategyName: s.Name(),
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "buy",
			Quantity:     s.quantity,
			Reason: fmt.Sprintf("WHALE ACCUMULATION: sentiment=%.2f (outflows > inflows), price %.2f%% below SMA - whales buying dip",
				s.whaleSentiment, priceVsSMA),
		})
		s.lastSignalTime = time.Now()
	}

	// ── SELL SIGNAL: Whale distribution + price rally ──
	// Whales are moving coins ONTO exchanges (bearish)
	// AND price is above SMA (selling the rally)
	if s.whaleSentiment < -0.3 && priceVsSMA > 0.2 {
		signals = append(signals, strategy.Signal{
			StrategyName: s.Name(),
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "sell",
			Quantity:     s.quantity,
			Reason: fmt.Sprintf("WHALE DISTRIBUTION: sentiment=%.2f (inflows > outflows), price %.2f%% above SMA - whales selling rally",
				s.whaleSentiment, priceVsSMA),
		})
		s.lastSignalTime = time.Now()
	}

	// ── STRONG BUY: Massive whale accumulation ──
	// Very strong outflow signal (>0.6) regardless of price position
	if s.whaleSentiment > 0.6 {
		signals = append(signals, strategy.Signal{
			StrategyName: s.Name(),
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "buy",
			Quantity:     s.quantity,
			Reason: fmt.Sprintf("MASSIVE WHALE OUTFLOW: sentiment=%.2f - major accumulation detected",
				s.whaleSentiment),
		})
		s.lastSignalTime = time.Now()
	}

	// ── STRONG SELL: Massive whale distribution ──
	// Very strong inflow signal (<-0.6) regardless of price position
	if s.whaleSentiment < -0.6 {
		signals = append(signals, strategy.Signal{
			StrategyName: s.Name(),
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "sell",
			Quantity:     s.quantity,
			Reason: fmt.Sprintf("MASSIVE WHALE INFLOW: sentiment=%.2f - major distribution detected",
				s.whaleSentiment),
		})
		s.lastSignalTime = time.Now()
	}

	return signals, nil
}

// GetWhaleStatus returns current whale status for diagnostics
func (s *WhaleAlertStrategy) GetWhaleStatus() map[string]interface{} {
	return map[string]interface{}{
		"symbol":           s.symbol,
		"whale_sentiment":  s.whaleSentiment,
		"last_update":      s.lastUpdateTime.Format(time.RFC3339),
		"data_age_seconds": time.Since(s.lastUpdateTime).Seconds(),
		"interpretation":   interpretWhaleSentiment(s.whaleSentiment),
	}
}

func interpretWhaleSentiment(score float64) string {
	switch {
	case score > 0.6:
		return "MASSIVE ACCUMULATION - whales aggressively buying"
	case score > 0.3:
		return "ACCUMULATION - whales moving coins off exchanges"
	case score > 0.1:
		return "SLIGHT ACCUMULATION - mild outflow trend"
	case score < -0.6:
		return "MASSIVE DISTRIBUTION - whales aggressively selling"
	case score < -0.3:
		return "DISTRIBUTION - whales moving coins to exchanges"
	case score < -0.1:
		return "SLIGHT DISTRIBUTION - mild inflow trend"
	default:
		return "NEUTRAL - no significant whale activity"
	}
}
