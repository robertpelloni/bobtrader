package arbitrage

import (
	"context"
	"fmt"
	"strconv"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// Triangle represents a 3-point arbitrage route.
type Triangle struct {
	Asset1 string // e.g., USDT
	Asset2 string // e.g., BTC
	Asset3 string // e.g., ETH
	Symbol1 string // e.g., BTCUSDT
	Symbol2 string // e.g., ETHBTC
	Symbol3 string // e.g., ETHUSDT
}

// TriangularScanner detects arbitrage opportunities within a single exchange.
type TriangularScanner struct {
	feed         marketdata.Feed
	accountID    string
	triangles    []Triangle
	minProfitPct float64
}

// NewTriangularScanner creates a new triangular arbitrage scanner.
func NewTriangularScanner(feed marketdata.Feed, accountID string, minProfitPct float64) *TriangularScanner {
	if minProfitPct <= 0 {
		minProfitPct = 0.1 // Default 0.1%
	}
	return &TriangularScanner{
		feed:         feed,
		accountID:    accountID,
		minProfitPct: minProfitPct,
		// Example triangles for Binance/Common exchanges
		triangles: []Triangle{
			{Asset1: "USDT", Asset2: "BTC", Asset3: "ETH", Symbol1: "BTCUSDT", Symbol2: "ETHBTC", Symbol3: "ETHUSDT"},
			{Asset1: "USDT", Asset2: "BTC", Asset3: "BNB", Symbol1: "BTCUSDT", Symbol2: "BNBBTC", Symbol3: "BNBUSDT"},
			{Asset1: "USDT", Asset2: "ETH", Asset3: "BNB", Symbol1: "ETHUSDT", Symbol2: "BNBETH", Symbol3: "BNBUSDT"},
		},
	}
}

func (s *TriangularScanner) Name() string { return "triangular-arbitrage" }

// OnTick scans for triangular opportunities.
func (s *TriangularScanner) OnTick(ctx context.Context) ([]strategy.Signal, error) {
	var signals []strategy.Signal

	for _, t := range s.triangles {
		// Direction 1: USDT -> A2 -> A3 -> USDT
		// 1. Buy A2 with USDT (Symbol1)
		// 2. Buy A3 with A2 (Symbol2)
		// 3. Sell A3 for USDT (Symbol3)

		p1, err := s.getPrice(ctx, t.Symbol1)
		if err != nil { continue }
		p2, err := s.getPrice(ctx, t.Symbol2)
		if err != nil { continue }
		p3, err := s.getPrice(ctx, t.Symbol3)
		if err != nil { continue }

		if p1 == 0 || p2 == 0 || p3 == 0 { continue }

		// Path 1: USDT -> Asset2 -> Asset3 -> USDT
		// Start with 1 USDT
		qtyA2 := 1.0 / p1 // Buy A2
		qtyA3 := qtyA2 / p2 // Buy A3 with A2
		finalUSDT := qtyA3 * p3 // Sell A3 for USDT

		profitPct := (finalUSDT - 1.0) * 100.0
		if profitPct > s.minProfitPct {
			signals = append(signals, strategy.Signal{
				Symbol:   t.Symbol1,
				Action:   "buy",
				Quantity: "0.01", // Placeholder
				Reason:   fmt.Sprintf("Triangular Arb 1: %s -> %s -> %s -> %s (Profit: %.4f%%)", t.Asset1, t.Asset2, t.Asset3, t.Asset1, profitPct),
			})
		}

		// Path 2: USDT -> Asset3 -> Asset2 -> USDT
		// 1. Buy A3 with USDT (Symbol3)
		// 2. Sell A3 for A2 (Symbol2)
		// 3. Sell A2 for USDT (Symbol1)
		qtyA3_2 := 1.0 / p3
		qtyA2_2 := qtyA3_2 * p2
		finalUSDT_2 := qtyA2_2 * p1

		profitPct2 := (finalUSDT_2 - 1.0) * 100.0
		if profitPct2 > s.minProfitPct {
			signals = append(signals, strategy.Signal{
				Symbol:   t.Symbol3,
				Action:   "buy",
				Quantity: "0.01", // Placeholder
				Reason:   fmt.Sprintf("Triangular Arb 2: %s -> %s -> %s -> %s (Profit: %.4f%%)", t.Asset1, t.Asset3, t.Asset2, t.Asset1, profitPct2),
			})
		}
	}

	return signals, nil
}

func (s *TriangularScanner) getPrice(ctx context.Context, symbol string) (float64, error) {
	tick, err := s.feed.LatestTick(ctx, symbol)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(tick.Price, 64)
}
