package demo

import (
	"context"
	"fmt"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// TriangularArbitrage exploits price inefficiencies between three pairs.
// Example: BTC/USDT, ETH/BTC, ETH/USDT
type TriangularArbitrage struct {
	mu           sync.Mutex
	accountID    string
	baseAsset    string // e.g., USDT
	assetA       string // e.g., BTC
	assetB       string // e.g., ETH
	pairBaseA    string // e.g., BTCUSDT
	pairAB       string // e.g., ETHBTC
	pairBaseB    string // e.g., ETHUSDT
	minProfitPct float64
	prices       map[string]float64
	quantities   map[string]float64
}

func NewTriangularArbitrage(
	accountID, base, a, b string,
	minProfit float64,
) *TriangularArbitrage {
	return &TriangularArbitrage{
		accountID:    accountID,
		baseAsset:    base,
		assetA:       a,
		assetB:       b,
		pairBaseA:    a + base,
		pairAB:       b + a,
		pairBaseB:    b + base,
		minProfitPct: minProfit,
		prices:       make(map[string]float64),
		quantities:   make(map[string]float64),
	}
}

func (s *TriangularArbitrage) Name() string {
	return fmt.Sprintf("tri-arb-%s-%s-%s", s.assetA, s.assetB, s.baseAsset)
}

func (s *TriangularArbitrage) OnTick(_ context.Context) ([]strategy.Signal, error) { return nil, nil }

func (s *TriangularArbitrage) OnMarketTick(ctx context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.prices[tick.Symbol] = utils.ParseFloat(tick.Price)
	s.quantities[tick.Symbol] = utils.ParseFloat(tick.Quantity)

	// Need prices for all 3 pairs
	p1, ok1 := s.prices[s.pairBaseA]
	p2, ok2 := s.prices[s.pairAB]
	p3, ok3 := s.prices[s.pairBaseB]

	if !ok1 || !ok2 || !ok3 {
		return nil, nil
	}

	// Logic 1: Buy A with Base, Buy B with A, Sell B for Base
	// (Base -> A -> B -> Base)
	// Example: USDT -> BTC -> ETH -> USDT
	// Price1 (BTCUSDT), Price2 (ETHBTC), Price3 (ETHUSDT)
	// AmountA = 100 / Price1
	// AmountB = AmountA / Price2
	// AmountBase = AmountB * Price3

	profit1 := (1.0 / p1 / p2 * p3) - 1.0

	// Logic 2: Buy B with Base, Buy A with B, Sell A for Base
	// (Base -> B -> A -> Base)
	// Example: USDT -> ETH -> BTC -> USDT
	// AmountB = 100 / Price3
	// AmountA = AmountB * Price2
	// AmountBase = AmountA * Price1

	profit2 := (1.0 / p3 * p2 * p1) - 1.0

	var signals []strategy.Signal

	if profit1*100.0 > s.minProfitPct {
		signals = append(signals, strategy.Signal{
			AccountID: s.accountID,
			Symbol:    s.pairBaseA,
			Action:    "buy",
			Quantity:  "0.01", // Should be dynamic
			Reason:    fmt.Sprintf("TriArb path 1 profit: %.4f%%", profit1*100),
		})
	} else if profit2*100.0 > s.minProfitPct {
		signals = append(signals, strategy.Signal{
			AccountID: s.accountID,
			Symbol:    s.pairBaseB,
			Action:    "buy",
			Quantity:  "0.1", // Should be dynamic
			Reason:    fmt.Sprintf("TriArb path 2 profit: %.4f%%", profit2*100),
		})
	}

	return signals, nil
}
