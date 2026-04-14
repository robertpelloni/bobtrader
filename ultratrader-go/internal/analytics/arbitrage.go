package analytics

import (
	"fmt"
	"math"
)

// PairStats holds the statistical relationship between two assets.
type PairStats struct {
	SymbolA      string
	SymbolB      string
	Correlation  float64
	ZScore       float64
	Spread       float64
	SpreadMean   float64
	SpreadStdDev float64
}

// ArbitrageDetector identifies statistical arbitrage opportunities.
type ArbitrageDetector struct {
	lookback int
}

// NewArbitrageDetector creates a detector with a specific lookback window.
func NewArbitrageDetector(lookback int) *ArbitrageDetector {
	if lookback < 2 {
		lookback = 20
	}
	return &ArbitrageDetector{
		lookback: lookback,
	}
}

// AnalyzePair calculates correlation and z-score for two price series.
func (d *ArbitrageDetector) AnalyzePair(symbolA, symbolB string, pricesA, pricesB []float64) (PairStats, error) {
	if len(pricesA) != len(pricesB) {
		return PairStats{}, fmt.Errorf("price series lengths must match (%d != %d)", len(pricesA), len(pricesB))
	}
	if len(pricesA) < d.lookback {
		return PairStats{}, fmt.Errorf("insufficient data: need %d data points, got %d", d.lookback, len(pricesA))
	}

	// Use only the lookback window
	windowA := pricesA[len(pricesA)-d.lookback:]
	windowB := pricesB[len(pricesB)-d.lookback:]

	// 1. Calculate Correlation
	correlation := calculatePearsonCorrelation(windowA, windowB)

	// 2. Calculate Spread (Ratio-based spread is more stable for crypto than simple difference)
	// Spread = log(PriceA) - log(PriceB) ≈ PriceA / PriceB
	var spreads []float64
	var sumSpread float64
	for i := 0; i < d.lookback; i++ {
		// Avoid log(0)
		if windowA[i] <= 0 || windowB[i] <= 0 {
			return PairStats{}, fmt.Errorf("prices must be > 0 for ratio spread calculation")
		}
		spread := math.Log(windowA[i]) - math.Log(windowB[i])
		spreads = append(spreads, spread)
		sumSpread += spread
	}

	meanSpread := sumSpread / float64(d.lookback)

	// 3. Calculate Spread Standard Deviation
	var sumSqDiff float64
	for _, s := range spreads {
		diff := s - meanSpread
		sumSqDiff += diff * diff
	}
	stdDevSpread := math.Sqrt(sumSqDiff / float64(d.lookback))

	// 4. Calculate Z-Score of the current (latest) spread
	currentSpread := spreads[len(spreads)-1]

	var zScore float64
	if stdDevSpread > 0 {
		zScore = (currentSpread - meanSpread) / stdDevSpread
	}

	return PairStats{
		SymbolA:      symbolA,
		SymbolB:      symbolB,
		Correlation:  correlation,
		ZScore:       zScore,
		Spread:       currentSpread,
		SpreadMean:   meanSpread,
		SpreadStdDev: stdDevSpread,
	}, nil
}

// CheckSignal determines if the z-score indicates a trade entry or exit.
// Returns an action string: "LONG_A_SHORT_B", "SHORT_A_LONG_B", or "FLAT".
func (d *ArbitrageDetector) CheckSignal(stats PairStats, entryZScore, exitZScore float64) string {
	// If z-score is highly positive, Spread is wider than normal (A is overvalued relative to B).
	// Action: Short A, Long B.
	if stats.ZScore >= entryZScore {
		return "SHORT_A_LONG_B"
	}

	// If z-score is highly negative, Spread is narrower than normal (A is undervalued relative to B).
	// Action: Long A, Short B.
	if stats.ZScore <= -entryZScore {
		return "LONG_A_SHORT_B"
	}

	// If the z-score reverts to the mean (crosses inside the exit threshold), flatten the position.
	if math.Abs(stats.ZScore) <= exitZScore {
		return "FLAT"
	}

	return "HOLD"
}

// calculatePearsonCorrelation computes the Pearson correlation coefficient between two slices.
func calculatePearsonCorrelation(x, y []float64) float64 {
	var sumX, sumY, sumXY, sumSqX, sumSqY float64
	n := float64(len(x))

	for i := 0; i < len(x); i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumSqX += x[i] * x[i]
		sumSqY += y[i] * y[i]
	}

	numerator := (n * sumXY) - (sumX * sumY)
	denominator := math.Sqrt(((n * sumSqX) - (sumX * sumX)) * ((n * sumSqY) - (sumY * sumY)))

	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}
