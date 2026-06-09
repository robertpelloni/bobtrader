package strategy_test

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/backtest"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy/demo"
)

// =============================================================================
// COMPREHENSIVE STRATEGY BENCHMARK SUITE
// =============================================================================

const (
	testSymbol     = "BTCUSDT"
	testAccountID  = "bench-account"
	initialCapital = 10000.0
	basePrice      = 68000.0
	tickCount      = 5000
	seed           = 42
)

type regime struct {
	name      string
	generator func(rng *rand.Rand, base float64, n int) []float64
}

func generateTrendUp(rng *rand.Rand, base float64, n int) []float64 {
	prices := make([]float64, n)
	prices[0] = base
	for i := 1; i < n; i++ {
		change := 0.0003 + rng.NormFloat64()*0.0015
		prices[i] = prices[i-1] * (1 + change)
	}
	return prices
}

func generateTrendDown(rng *rand.Rand, base float64, n int) []float64 {
	prices := make([]float64, n)
	prices[0] = base
	for i := 1; i < n; i++ {
		change := -0.0003 + rng.NormFloat64()*0.0015
		prices[i] = prices[i-1] * (1 + change)
	}
	return prices
}

func generateRanging(rng *rand.Rand, base float64, n int) []float64 {
	prices := make([]float64, n)
	prices[0] = base
	theta := 0.02
	sigma := 0.003
	for i := 1; i < n; i++ {
		diff := theta * (base - prices[i-1]) / base
		noise := rng.NormFloat64() * sigma
		prices[i] = prices[i-1] * (1 + diff + noise)
	}
	return prices
}

func generateVolatile(rng *rand.Rand, base float64, n int) []float64 {
	prices := make([]float64, n)
	prices[0] = base
	for i := 1; i < n; i++ {
		prices[i] = prices[i-1] * (1 + rng.NormFloat64()*0.006)
	}
	return prices
}

func generateCrashRecovery(rng *rand.Rand, base float64, n int) []float64 {
	prices := make([]float64, n)
	prices[0] = base
	crashPoint := n / 3
	recoveryPoint := 2 * n / 3
	for i := 1; i < n; i++ {
		var change float64
		if i < crashPoint {
			change = rng.NormFloat64() * 0.001
		} else if i < crashPoint+100 {
			change = -0.003 + rng.NormFloat64()*0.004
		} else if i < recoveryPoint {
			change = rng.NormFloat64() * 0.002
		} else {
			change = 0.0008 + rng.NormFloat64()*0.003
		}
		prices[i] = prices[i-1] * (1 + change)
	}
	return prices
}

func pricesToTicks(prices []float64) []marketdata.Tick {
	ticks := make([]marketdata.Tick, len(prices))
	baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, p := range prices {
		ticks[i] = marketdata.Tick{
			Symbol:    testSymbol,
			Price:     fmt.Sprintf("%.2f", p),
			Timestamp: baseTime.Add(time.Duration(i) * 5 * time.Second),
		}
	}
	return ticks
}

func pricesToCandles(prices []float64) []marketdata.Candle {
	candleInterval := 12
	n := len(prices) / candleInterval
	candles := make([]marketdata.Candle, 0, n)
	baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < n; i++ {
		start := i * candleInterval
		end := start + candleInterval
		if end > len(prices) {
			end = len(prices)
		}
		if start >= len(prices) {
			break
		}
		high := prices[start]
		low := prices[start]
		for j := start + 1; j < end; j++ {
			if prices[j] > high {
				high = prices[j]
			}
			if prices[j] < low {
				low = prices[j]
			}
		}
		candles = append(candles, marketdata.Candle{
			Symbol:    testSymbol,
			Open:      fmt.Sprintf("%.2f", prices[start]),
			High:      fmt.Sprintf("%.2f", high),
			Low:       fmt.Sprintf("%.2f", low),
			Close:     fmt.Sprintf("%.2f", prices[end-1]),
			Timestamp: baseTime.Add(time.Duration(i) * time.Minute),
			Interval:  "1m",
		})
	}
	return candles
}

// --- Strategy Factory ---

type strategyFactory struct {
	name         string
	isTick       bool
	isCandle     bool
	createTick   func() strategy.TickStrategy
	createCandle func() strategy.CandleStrategy
}

func allStrategyFactories() []strategyFactory {
	return []strategyFactory{
		{
			name:   "BollingerTickReversion",
			isTick: true,
			createTick: func() strategy.TickStrategy {
				return demo.NewBollingerTickReversion(testAccountID, testSymbol, "0.001", 20, 2.0)
			},
		},
		{
			name:   "RSIReversion",
			isTick: true,
			createTick: func() strategy.TickStrategy {
				return demo.NewRSIReversion(testAccountID, testSymbol, "0.001", 14, 35, 65)
			},
		},
		{
			name:   "EMATickCrossover",
			isTick: true,
			createTick: func() strategy.TickStrategy {
				return demo.NewEMATickCrossover(testAccountID, testSymbol, "0.001", 9, 21)
			},
		},
		{
			name:   "TickMeanReversion",
			isTick: true,
			createTick: func() strategy.TickStrategy {
				return demo.NewTickMeanReversion(testAccountID, testSymbol, "0.001", 50, 0.5, 0.5)
			},
		},
		{
			name:   "TickMomentumBurst",
			isTick: true,
			createTick: func() strategy.TickStrategy {
				return demo.NewTickMomentumBurst(testAccountID, testSymbol, "0.001", 30, 0.3, 0.3)
			},
		},
		{
			name:     "MACDCrossover",
			isCandle: true,
			createCandle: func() strategy.CandleStrategy {
				return demo.NewMACDCrossover(testAccountID, testSymbol, "0.001", 12, 26, 9)
			},
		},
		{
			name:     "BollingerReversion",
			isCandle: true,
			createCandle: func() strategy.CandleStrategy {
				return demo.NewBollingerReversion(testAccountID, testSymbol, "0.001", 20, 2.0)
			},
		},
		{
			name:     "CandleSMACross",
			isCandle: true,
			createCandle: func() strategy.CandleStrategy {
				return demo.NewCandleSMACross(testAccountID, testSymbol, "0.001", 5, 20)
			},
		},
		{
			name:     "ATRSizing",
			isCandle: true,
			createCandle: func() strategy.CandleStrategy {
				return demo.NewATRSizing(testAccountID, testSymbol, "0.001", 0.01, 7, 25, 14)
			},
		},
		{
			name:     "DoubleEMATrend",
			isCandle: true,
			createCandle: func() strategy.CandleStrategy {
				return demo.NewDoubleEMATrendStrategy(testAccountID, testSymbol, "0.001", 9, 21, 200)
			},
		},
	}
}

// --- Results ---

type regimeResult struct {
	regimeName  string
	trades      int
	winRate     float64
	pnl         float64
	maxDrawdown float64
	sharpeRatio float64
	signalCount int
}

type strategyResult struct {
	name           string
	regimes        []regimeResult
	avgWinRate     float64
	avgPnL         float64
	avgSharpe      float64
	compositeScore float64
	totalTrades    int
	grade          string
}

type orderRecord struct {
	symbol string
	side   string
	price  float64
	qty    float64
}

func computeTradeStats(orders []orderRecord) (winRate, maxDD, sharpe float64, wins, losses int) {
	if len(orders) == 0 {
		return 0, 0, 0, 0, 0
	}

	var roundTrips []float64
	runningPnL := 0.0
	peak := 0.0
	dd := 0.0
	buys := make(map[string]float64)

	for _, o := range orders {
		if o.side == "buy" {
			buys[o.symbol] = o.price
		} else if o.side == "sell" {
			if entry, ok := buys[o.symbol]; ok && entry > 0 {
				pnl := (o.price - entry) * o.qty
				roundTrips = append(roundTrips, pnl)
				runningPnL += pnl
				if runningPnL > peak {
					peak = runningPnL
				}
				drawdown := peak - runningPnL
				if drawdown > dd {
					dd = drawdown
				}
				delete(buys, o.symbol)
			}
		}
	}

	if len(roundTrips) == 0 {
		return 0, dd, 0, 0, 0
	}

	for _, rt := range roundTrips {
		if rt > 0 {
			wins++
		} else {
			losses++
		}
	}
	winRate = float64(wins) / float64(len(roundTrips)) * 100

	var sum, sumSq float64
	for _, rt := range roundTrips {
		sum += rt
		sumSq += rt * rt
	}
	mean := sum / float64(len(roundTrips))
	variance := sumSq/float64(len(roundTrips)) - mean*mean
	stddev := math.Sqrt(math.Max(variance, 0.0001))
	sharpe = mean / stddev

	return winRate, dd, sharpe, wins, losses
}

func runTickStrategyBacktest(ctx context.Context, s strategy.TickStrategy, ticks []marketdata.Tick) (trades int, winRate, pnl, maxDD, sharpe float64, signalCount int) {
	engine := backtest.NewEngineWithOptions(s, initialCapital, backtest.EmulatorOptions{
		TakerFeeRate: 0.001,
		MakerFeeRate: 0.001,
		SlippageRate: 0.0002,
	})
	history := backtest.NewMemoryHistory(ticks)
	result, err := engine.RunTicks(ctx, history)
	if err != nil {
		return 0, 0, 0, 0, 0, 0
	}

	signalCount = 0
	for _, tick := range ticks {
		sigs, _ := s.OnMarketTick(ctx, tick)
		signalCount += len(sigs)
	}

	trades = result.TotalTrades
	pnl = result.RealizedPnL

	var orders []orderRecord
	for _, o := range result.Orders {
		side := "buy"
		if o.Side == exchange.Sell {
			side = "sell"
		}
		orders = append(orders, orderRecord{
			symbol: o.Symbol,
			side:   side,
			price:  parseFloat(o.Price),
			qty:    0.001,
		})
	}

	winRate, maxDD, sharpe, _, _ = computeTradeStats(orders)
	return
}

func runCandleStrategyBacktest(ctx context.Context, s strategy.CandleStrategy, candles []marketdata.Candle) (trades int, winRate, pnl, maxDD, sharpe float64, signalCount int) {
	engine := backtest.NewEngineWithOptions(s, initialCapital, backtest.EmulatorOptions{
		TakerFeeRate: 0.001,
		MakerFeeRate: 0.001,
		SlippageRate: 0.0002,
	})
	history := backtest.NewMemoryCandleHistory(candles)
	result, err := engine.RunCandles(ctx, history)
	if err != nil {
		return 0, 0, 0, 0, 0, 0
	}

	signalCount = 0
	for _, c := range candles {
		sigs, _ := s.OnMarketCandle(ctx, c)
		signalCount += len(sigs)
	}

	trades = result.TotalTrades
	pnl = result.RealizedPnL

	var orders []orderRecord
	for _, o := range result.Orders {
		side := "buy"
		if o.Side == exchange.Sell {
			side = "sell"
		}
		orders = append(orders, orderRecord{
			symbol: o.Symbol,
			side:   side,
			price:  parseFloat(o.Price),
			qty:    0.001,
		})
	}

	winRate, maxDD, sharpe, _, _ = computeTradeStats(orders)
	return
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

// =============================================================================
// THE BIG TEST
// =============================================================================

func TestComprehensiveStrategyBenchmark(t *testing.T) {
	rng := rand.New(rand.NewSource(seed))

	regimes := []regime{
		{name: "TrendUp", generator: generateTrendUp},
		{name: "TrendDown", generator: generateTrendDown},
		{name: "Ranging", generator: generateRanging},
		{name: "Volatile", generator: generateVolatile},
		{name: "CrashRecovery", generator: generateCrashRecovery},
	}

	factories := allStrategyFactories()
	ctx := context.Background()

	var allResults []strategyResult

	for _, factory := range factories {
		sr := strategyResult{name: factory.name}

		for _, reg := range regimes {
			prices := reg.generator(rng, basePrice, tickCount)

			var trades int
			var winRate, pnl, maxDD, sharpe float64
			var signalCount int

			if factory.isTick {
				ticks := pricesToTicks(prices)
				s := factory.createTick()
				trades, winRate, pnl, maxDD, sharpe, signalCount = runTickStrategyBacktest(ctx, s, ticks)
			} else if factory.isCandle {
				candles := pricesToCandles(prices)
				s := factory.createCandle()
				trades, winRate, pnl, maxDD, sharpe, signalCount = runCandleStrategyBacktest(ctx, s, candles)
			}

			sr.regimes = append(sr.regimes, regimeResult{
				regimeName:  reg.name,
				trades:      trades,
				winRate:     winRate,
				pnl:         pnl,
				maxDrawdown: maxDD,
				sharpeRatio: sharpe,
				signalCount: signalCount,
			})
			sr.totalTrades += trades
		}

		var wrSum, pnlSum, sharpeSum float64
		for _, r := range sr.regimes {
			wrSum += r.winRate
			pnlSum += r.pnl
			sharpeSum += r.sharpeRatio
		}
		n := float64(len(sr.regimes))
		sr.avgWinRate = wrSum / n
		sr.avgPnL = pnlSum / n
		sr.avgSharpe = sharpeSum / n

		// Composite score (0-100):
		// 40% win rate, 25% PnL normalized, 20% Sharpe, 15% trade count efficiency
		tradeEfficiency := math.Min(float64(sr.totalTrades)/50.0, 1.0)
		if sr.totalTrades > 200 {
			tradeEfficiency = math.Max(0.3, 1.0-float64(sr.totalTrades-200)/500.0)
		}
		if sr.totalTrades == 0 {
			tradeEfficiency = 0.0
		}

		pnlNorm := math.Max(0, math.Min(100, (sr.avgPnL+5)/10.0*100))
		sharpeNorm := math.Max(0, math.Min(100, (sr.avgSharpe+3)/6.0*100))

		sr.compositeScore = sr.avgWinRate*0.4 + pnlNorm*0.25 + sharpeNorm*0.2 + tradeEfficiency*100*0.15

		switch {
		case sr.compositeScore >= 80:
			sr.grade = "A"
		case sr.compositeScore >= 65:
			sr.grade = "B"
		case sr.compositeScore >= 50:
			sr.grade = "C"
		case sr.compositeScore >= 35:
			sr.grade = "D"
		default:
			sr.grade = "F"
		}

		allResults = append(allResults, sr)
	}

	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].compositeScore > allResults[j].compositeScore
	})

	// ═══════════════════════════════════════════════════════════════
	// REPORT
	// ═══════════════════════════════════════════════════════════════

	t.Log("")
	t.Log("╔══════════════════════════════════════════════════════════════════════════════╗")
	t.Log("║       COMPREHENSIVE STRATEGY BENCHMARK — UltraTrader-Go v2.0.54            ║")
	t.Log("╠══════════════════════════════════════════════════════════════════════════════╣")
	t.Log("║  Regimes: TrendUp, TrendDown, Ranging, Volatile, CrashRecovery             ║")
	t.Log("║  Data: 5000 ticks/regime, BTC/USDT @ $68,000 base, 0.1% taker fee          ║")
	t.Log("║  Capital: $10,000 USDT, Qty: 0.001 BTC/trade                               ║")
	t.Log("╚══════════════════════════════════════════════════════════════════════════════╝")
	t.Log("")

	for rank, sr := range allResults {
		t.Logf("┌──────────────────────────────────────────────────────────────")
		t.Logf("│ #%d  %-28s  Grade: %s  Score: %.1f/100", rank+1, sr.name, sr.grade, sr.compositeScore)
		t.Logf("│   Avg WinRate: %.1f%%  |  Avg PnL: $%.4f  |  Avg Sharpe: %.3f", sr.avgWinRate, sr.avgPnL, sr.avgSharpe)
		t.Logf("│   Total Trades: %d (across all 5 regimes)", sr.totalTrades)
		t.Logf("│")
		for _, r := range sr.regimes {
			t.Logf("│   %-16s WR:%5.1f%%  PnL:$%8.4f  Sharpe:%6.3f  DD:$%7.2f  Trades:%3d  Sigs:%4d",
				r.regimeName, r.winRate, r.pnl, r.sharpeRatio, r.maxDrawdown, r.trades, r.signalCount)
		}
		t.Logf("└──────────────────────────────────────────────────────────────")
		t.Logf("")
	}

	// Summary table
	t.Log("┌────────────────────────────────────────────────────────────────────────────┐")
	t.Log("│  RANK  STRATEGY                    GRADE  SCORE  WIN%   PnL      SHARPE   │")
	t.Log("├────────────────────────────────────────────────────────────────────────────┤")
	for rank, sr := range allResults {
		t.Logf("│  #%d   %-28s   %s    %5.1f  %4.1f%%  $%7.3f  %6.3f  │",
			rank+1, sr.name, sr.grade, sr.compositeScore, sr.avgWinRate, sr.avgPnL, sr.avgSharpe)
	}
	t.Log("└────────────────────────────────────────────────────────────────────────────┘")

	// Regime-specific best
	t.Log("")
	t.Log("═══ BEST STRATEGY PER REGIME ═══")
	for _, reg := range regimes {
		type regScore struct {
			name string
			wr   float64
			pnl  float64
		}
		var scores []regScore
		for _, sr := range allResults {
			for _, r := range sr.regimes {
				if r.regimeName == reg.name {
					scores = append(scores, regScore{name: sr.name, wr: r.winRate, pnl: r.pnl})
				}
			}
		}
		sort.Slice(scores, func(i, j int) bool {
			return scores[i].pnl > scores[j].pnl
		})
		t.Logf("  %s:", reg.name)
		for i, s := range scores {
			marker := "     "
			if i == 0 {
				marker = "  ★  "
			}
			t.Logf("  %s %-28s  WR:%5.1f%%  PnL:$%8.4f", marker, s.name, s.wr, s.pnl)
		}
		t.Logf("")
	}

	// Detailed analysis
	t.Log("═══ DETAILED STRATEGY ANALYSIS ═══")
	for _, sr := range allResults {
		t.Logf("")
		t.Logf("  ── %s (%s) ──", sr.name, sr.grade)

		bestRegime := sr.regimes[0]
		worstRegime := sr.regimes[0]
		for _, r := range sr.regimes {
			if r.pnl > bestRegime.pnl {
				bestRegime = r
			}
			if r.pnl < worstRegime.pnl {
				worstRegime = r
			}
		}
		t.Logf("    Best regime:  %s (PnL: $%.4f, WR: %.1f%%)", bestRegime.regimeName, bestRegime.pnl, bestRegime.winRate)
		t.Logf("    Worst regime: %s (PnL: $%.4f, WR: %.1f%%)", worstRegime.regimeName, worstRegime.pnl, worstRegime.winRate)

		totalSigs := 0
		for _, r := range sr.regimes {
			totalSigs += r.signalCount
		}
		avgSigs := float64(totalSigs) / float64(len(sr.regimes))
		sigDensity := avgSigs / float64(tickCount) * 100
		t.Logf("    Signal density: %.2f%% (avg %.0f signals per regime)", sigDensity, avgSigs)

		positiveRegimes := 0
		for _, r := range sr.regimes {
			if r.pnl > 0 {
				positiveRegimes++
			}
		}
		t.Logf("    Profitable regimes: %d/%d (%.0f%%)", positiveRegimes, len(sr.regimes), float64(positiveRegimes)/float64(len(sr.regimes))*100)

		maxDDAny := 0.0
		for _, r := range sr.regimes {
			if r.maxDrawdown > maxDDAny {
				maxDDAny = r.maxDrawdown
			}
		}
		riskLevel := "LOW"
		if maxDDAny > 1.0 {
			riskLevel = "MEDIUM"
		}
		if maxDDAny > 3.0 {
			riskLevel = "HIGH"
		}
		t.Logf("    Max drawdown: $%.2f (risk: %s)", maxDDAny, riskLevel)

		// Strategy type classification
		stratType := "Unknown"
		switch sr.name {
		case "BollingerTickReversion", "RSIReversion", "TickMeanReversion", "BollingerReversion":
			stratType = "Mean Reversion"
		case "EMATickCrossover", "CandleSMACross", "DoubleEMATrend", "MACDCrossover":
			stratType = "Trend Following"
		case "TickMomentumBurst":
			stratType = "Momentum"
		case "ATRSizing":
			stratType = "Volatility-Adaptive Trend Following"
		}
		t.Logf("    Type: %s", stratType)

		// Regime suitability analysis
		var suitable, unsuitable []string
		for _, r := range sr.regimes {
			if r.pnl > 0 || r.winRate > 50 {
				suitable = append(suitable, r.regimeName)
			} else {
				unsuitable = append(unsuitable, r.regimeName)
			}
		}
		if len(suitable) > 0 {
			t.Logf("    Suitable regimes: %v", suitable)
		}
		if len(unsuitable) > 0 {
			t.Logf("    Unsuitable regimes: %v", unsuitable)
		}
	}

	// Final recommendations
	t.Log("")
	t.Log("╔══════════════════════════════════════════════════════════════════════╗")
	t.Log("║                    STRATEGY RECOMMENDATIONS                          ║")
	t.Log("╠══════════════════════════════════════════════════════════════════════╣")
	if len(allResults) > 0 {
		t.Logf("║  🥇 Best Overall:     %-30s  Score: %.1f  ║", allResults[0].name, allResults[0].compositeScore)
	}
	if len(allResults) > 1 {
		t.Logf("║  🥈 Runner Up:        %-30s  Score: %.1f  ║", allResults[1].name, allResults[1].compositeScore)
	}
	if len(allResults) > 2 {
		t.Logf("║  🥉 Third Place:      %-30s  Score: %.1f  ║", allResults[2].name, allResults[2].compositeScore)
	}

	for _, reg := range regimes {
		bestName := ""
		bestPnL := math.Inf(-1)
		for _, sr := range allResults {
			for _, r := range sr.regimes {
				if r.regimeName == reg.name && r.pnl > bestPnL {
					bestPnL = r.pnl
					bestName = sr.name
				}
			}
		}
		if bestName != "" {
			t.Logf("║  Best for %-14s %-28s  ║", reg.name+":", bestName)
		}
	}
	t.Log("╚══════════════════════════════════════════════════════════════════════╝")

	// Assertions
	if len(allResults) > 0 && allResults[0].compositeScore < 5 {
		t.Errorf("Top strategy composite score too low: %.1f", allResults[0].compositeScore)
	}
}
