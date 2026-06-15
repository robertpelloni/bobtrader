package demo

import (
	"context"
	"fmt"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/analytics"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// StatisticalArbitrage trades the price divergence between two correlated assets.
type StatisticalArbitrage struct {
	mu           sync.Mutex
	accountID    string
	symbolA      string
	symbolB      string
	quantityA    string
	quantityB    string
	detector     *analytics.ArbitrageDetector
	pricesA      []float64
	pricesB      []float64
	entryZScore  float64
	exitZScore   float64
	lastAction   string
}

func NewStatisticalArbitrage(
	accountID, symbolA, symbolB, qtyA, qtyB string,
	lookback int,
	entryZ, exitZ float64,
) *StatisticalArbitrage {
	return &StatisticalArbitrage{
		accountID:   accountID,
		symbolA:     symbolA,
		symbolB:     symbolB,
		quantityA:   qtyA,
		quantityB:   qtyB,
		detector:    analytics.NewArbitrageDetector(lookback),
		entryZScore: entryZ,
		exitZScore:  exitZ,
		pricesA:     make([]float64, 0, 100),
		pricesB:     make([]float64, 0, 100),
	}
}

func (s *StatisticalArbitrage) Name() string {
	return fmt.Sprintf("stat-arb-%s-%s", s.symbolA, s.symbolB)
}

func (s *StatisticalArbitrage) OnTick(_ context.Context) ([]strategy.Signal, error) { return nil, nil }

func (s *StatisticalArbitrage) OnMarketTick(ctx context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Update price history for the relevant symbol
	if tick.Symbol == s.symbolA {
		s.pricesA = append(s.pricesA, utils.ParseFloat(tick.Price))
	} else if tick.Symbol == s.symbolB {
		s.pricesB = append(s.pricesB, utils.ParseFloat(tick.Price))
	} else {
		return nil, nil
	}

	// Ensure synchronized price series
	n := len(s.pricesA)
	if len(s.pricesB) < n {
		n = len(s.pricesB)
	}
	if n < 20 {
		return nil, nil
	}

	// Align lengths
	pricesA := s.pricesA[len(s.pricesA)-n:]
	pricesB := s.pricesB[len(s.pricesB)-n:]

	stats, err := s.detector.AnalyzePair(s.symbolA, s.symbolB, pricesA, pricesB)
	if err != nil {
		return nil, nil
	}

	// Only trade if assets are sufficiently correlated
	if stats.Correlation < 0.7 {
		return nil, nil
	}

	action := s.detector.CheckSignal(stats, s.entryZScore, s.exitZScore)

	var signals []strategy.Signal
	switch action {
	case "SHORT_A_LONG_B":
		if s.lastAction != "SHORT_A_LONG_B" {
			signals = append(signals,
				strategy.Signal{AccountID: s.accountID, Symbol: s.symbolA, Action: "sell", Quantity: s.quantityA, Reason: fmt.Sprintf("StatArb: A overvalued (Z=%.2f)", stats.ZScore)},
				strategy.Signal{AccountID: s.accountID, Symbol: s.symbolB, Action: "buy", Quantity: s.quantityB, Reason: fmt.Sprintf("StatArb: B undervalued (Z=%.2f)", stats.ZScore)},
			)
			s.lastAction = "SHORT_A_LONG_B"
		}
	case "LONG_A_SHORT_B":
		if s.lastAction != "LONG_A_SHORT_B" {
			signals = append(signals,
				strategy.Signal{AccountID: s.accountID, Symbol: s.symbolA, Action: "buy", Quantity: s.quantityA, Reason: fmt.Sprintf("StatArb: A undervalued (Z=%.2f)", stats.ZScore)},
				strategy.Signal{AccountID: s.accountID, Symbol: s.symbolB, Action: "sell", Quantity: s.quantityB, Reason: fmt.Sprintf("StatArb: B overvalued (Z=%.2f)", stats.ZScore)},
			)
			s.lastAction = "LONG_A_SHORT_B"
		}
	case "FLAT":
		if s.lastAction != "" && s.lastAction != "FLAT" {
			// Exit both
			signals = append(signals,
				strategy.Signal{AccountID: s.accountID, Symbol: s.symbolA, Action: "sell", Quantity: s.quantityA, Reason: "StatArb: Mean reversion complete"},
				strategy.Signal{AccountID: s.accountID, Symbol: s.symbolB, Action: "buy", Quantity: s.quantityB, Reason: "StatArb: Mean reversion complete"},
			)
			s.lastAction = "FLAT"
		}
	}

	return signals, nil
}
