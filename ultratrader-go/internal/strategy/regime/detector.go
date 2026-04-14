package regime

import (
	"math"
)

// Regime represents the current market state classification.
type Regime int

const (
	RegimeTrending Regime = iota
	RegimeRanging
	RegimeVolatile
	RegimeQuiet
)

func (r Regime) String() string {
	switch r {
	case RegimeTrending:
		return "TRENDING"
	case RegimeRanging:
		return "RANGING"
	case RegimeVolatile:
		return "VOLATILE"
	case RegimeQuiet:
		return "QUIET"
	default:
		return "UNKNOWN"
	}
}

// Detector determines the current market regime from price data.
type Detector interface {
	Name() string
	Detect(candles []CandleData) Regime
}

// CandleData is a minimal candle interface for regime detection.
type CandleData struct {
	High   float64
	Low    float64
	Close  float64
	Open   float64
	Volume float64
}

// VolatilityDetector classifies regime based on ATR relative to price.
// High ATR% = Volatile, Medium ATR% = Trending, Low ATR% = Quiet/Ranging.
type VolatilityDetector struct {
	HighVolThreshold float64 // ATR/Close above this = Volatile (e.g., 0.05 = 5%)
	LowVolThreshold  float64 // ATR/Close below this = Quiet (e.g., 0.01 = 1%)
	ATRPeriod        int     // Period for ATR calculation
}

func NewVolatilityDetector(highVol, lowVol float64, atrPeriod int) *VolatilityDetector {
	return &VolatilityDetector{
		HighVolThreshold: highVol,
		LowVolThreshold:  lowVol,
		ATRPeriod:        atrPeriod,
	}
}

func (v *VolatilityDetector) Name() string { return "volatility" }

func (v *VolatilityDetector) Detect(candles []CandleData) Regime {
	if len(candles) < v.ATRPeriod {
		return RegimeQuiet
	}

	atr := calculateATR(candles, v.ATRPeriod)
	if atr <= 0 {
		return RegimeQuiet
	}

	lastClose := candles[len(candles)-1].Close
	if lastClose <= 0 {
		return RegimeQuiet
	}

	volRatio := atr / lastClose

	switch {
	case volRatio > v.HighVolThreshold:
		return RegimeVolatile
	case volRatio > v.LowVolThreshold:
		// Check if trending or ranging using directional movement
		if isTrending(candles) {
			return RegimeTrending
		}
		return RegimeRanging
	default:
		return RegimeQuiet
	}
}

// TrendDetector uses directional movement to classify trend strength.
// Inspired by ADX (Average Directional Index).
type TrendDetector struct {
	Period          int
	TrendThreshold  float64 // ADX above = trending (e.g., 25)
	StrongThreshold float64 // ADX above = strong trend (e.g., 50)
}

func NewTrendDetector(period int, trendThreshold, strongThreshold float64) *TrendDetector {
	return &TrendDetector{
		Period:          period,
		TrendThreshold:  trendThreshold,
		StrongThreshold: strongThreshold,
	}
}

func (td *TrendDetector) Name() string { return "trend" }

func (td *TrendDetector) Detect(candles []CandleData) Regime {
	if len(candles) < td.Period+1 {
		return RegimeQuiet
	}

	adx := calculateADX(candles, td.Period)

	switch {
	case adx > td.StrongThreshold:
		return RegimeTrending
	case adx > td.TrendThreshold:
		return RegimeTrending
	default:
		// Check volatility for ranging vs quiet
		atr := calculateATR(candles, td.Period)
		lastClose := candles[len(candles)-1].Close
		if lastClose > 0 && atr/lastClose > 0.02 {
			return RegimeRanging
		}
		return RegimeQuiet
	}
}

// BollingerBandwidthDetector uses Bollinger Band width to classify regime.
// Narrow bands = ranging/quiet, Wide bands = volatile/trending.
type BollingerBandwidthDetector struct {
	Period          int
	StdDev          float64
	LowBWThreshold  float64 // Bandwidth below this = ranging/quiet
	HighBWThreshold float64 // Bandwidth above this = volatile
}

func NewBollingerBandwidthDetector(period int, stdDev, lowBW, highBW float64) *BollingerBandwidthDetector {
	return &BollingerBandwidthDetector{
		Period:          period,
		StdDev:          stdDev,
		LowBWThreshold:  lowBW,
		HighBWThreshold: highBW,
	}
}

func (bb *BollingerBandwidthDetector) Name() string { return "bollinger-bandwidth" }

func (bb *BollingerBandwidthDetector) Detect(candles []CandleData) Regime {
	if len(candles) < bb.Period {
		return RegimeQuiet
	}

	bw := bollingerBandwidth(candles, bb.Period, bb.StdDev)
	if bw <= 0 {
		return RegimeQuiet
	}

	switch {
	case bw > bb.HighBWThreshold:
		return RegimeVolatile
	case bw > bb.LowBWThreshold:
		if isTrending(candles) {
			return RegimeTrending
		}
		return RegimeRanging
	default:
		return RegimeQuiet
	}
}

// CompositeDetector combines multiple detectors with majority voting.
type CompositeDetector struct {
	detectors []Detector
}

func NewCompositeDetector(detectors ...Detector) *CompositeDetector {
	return &CompositeDetector{detectors: detectors}
}

func (c *CompositeDetector) Name() string { return "composite" }

func (c *CompositeDetector) Detect(candles []CandleData) Regime {
	if len(c.detectors) == 0 {
		return RegimeQuiet
	}

	votes := make(map[Regime]int)
	for _, d := range c.detectors {
		r := d.Detect(candles)
		votes[r]++
	}

	// Find majority
	var maxVotes int
	var winner Regime
	for r, v := range votes {
		if v > maxVotes {
			maxVotes = v
			winner = r
		}
	}
	return winner
}

// Helper functions

func calculateATR(candles []CandleData, period int) float64 {
	if len(candles) < 2 {
		return 0
	}

	var sum float64
	n := min(len(candles)-1, period)

	for i := len(candles) - n; i < len(candles); i++ {
		tr := trueRange(candles[i-1], candles[i])
		sum += tr
	}

	if n == 0 {
		return 0
	}
	return sum / float64(n)
}

func trueRange(prev, curr CandleData) float64 {
	hl := math.Abs(curr.High - curr.Low)
	hc := math.Abs(curr.High - prev.Close)
	lc := math.Abs(curr.Low - prev.Close)
	return math.Max(hl, math.Max(hc, lc))
}

func isTrending(candles []CandleData) bool {
	if len(candles) < 5 {
		return false
	}
	// Simple check: compare first half average to second half average
	n := len(candles)
	firstHalf := avgClose(candles[:n/2])
	secondHalf := avgClose(candles[n/2:])
	overallAvg := (firstHalf + secondHalf) / 2

	if overallAvg == 0 {
		return false
	}
	change := math.Abs(secondHalf-firstHalf) / overallAvg
	return change > 0.01 // More than 1% directional movement
}

func avgClose(candles []CandleData) float64 {
	var sum float64
	for _, c := range candles {
		sum += c.Close
	}
	if len(candles) == 0 {
		return 0
	}
	return sum / float64(len(candles))
}

func calculateADX(candles []CandleData, period int) float64 {
	if len(candles) < period+1 {
		return 0
	}

	// Simplified ADX: average directional movement
	var sumDX float64
	count := 0

	for i := period; i < len(candles); i++ {
		upMove := candles[i].High - candles[i-1].High
		downMove := candles[i-1].Low - candles[i].Low

		var plusDM, minusDM float64
		if upMove > downMove && upMove > 0 {
			plusDM = upMove
		}
		if downMove > upMove && downMove > 0 {
			minusDM = downMove
		}

		atr := calculateATR(candles[i-period:i+1], period)
		if atr <= 0 {
			continue
		}

		plusDI := 100 * plusDM / atr
		minusDI := 100 * minusDM / atr

		diSum := plusDI + minusDI
		if diSum == 0 {
			continue
		}

		dx := 100 * math.Abs(plusDI-minusDI) / diSum
		sumDX += dx
		count++
	}

	if count == 0 {
		return 0
	}
	return sumDX / float64(count)
}

func bollingerBandwidth(candles []CandleData, period int, stdDev float64) float64 {
	if len(candles) < period {
		return 0
	}

	// Calculate SMA
	var sum float64
	for i := len(candles) - period; i < len(candles); i++ {
		sum += candles[i].Close
	}
	sma := sum / float64(period)

	// Calculate standard deviation
	var sumSqDiff float64
	for i := len(candles) - period; i < len(candles); i++ {
		diff := candles[i].Close - sma
		sumSqDiff += diff * diff
	}
	std := sqrt(sumSqDiff / float64(period))

	upper := sma + stdDev*std
	lower := sma - stdDev*std

	if sma == 0 {
		return 0
	}
	return (upper - lower) / sma
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 20; i++ {
		z = (z + x/z) / 2
	}
	return z
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
