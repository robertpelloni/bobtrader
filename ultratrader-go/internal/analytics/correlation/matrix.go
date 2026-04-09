package correlation

import (
	"math"
)

// CorrelationMatrix computes rolling Pearson correlations between symbol pairs.
type CorrelationMatrix struct {
	windowSize int
	history    map[string][]float64 // symbol -> price history
}

// NewCorrelationMatrix creates a correlation matrix with the given lookback window.
func NewCorrelationMatrix(windowSize int) *CorrelationMatrix {
	if windowSize < 2 {
		windowSize = 2
	}
	return &CorrelationMatrix{
		windowSize: windowSize,
		history:    make(map[string][]float64),
	}
}

// AddPrice adds a price observation for a symbol.
func (cm *CorrelationMatrix) AddPrice(symbol string, price float64) {
	cm.history[symbol] = append(cm.history[symbol], price)
	// Trim to window size
	if len(cm.history[symbol]) > cm.windowSize {
		cm.history[symbol] = cm.history[symbol][len(cm.history[symbol])-cm.windowSize:]
	}
}

// Pearson computes the Pearson correlation coefficient between two series.
// Returns 0 if insufficient data.
func Pearson(x, y []float64) float64 {
	n := min(len(x), len(y))
	if n < 2 {
		return 0
	}

	// Use the overlapping portion
	x = x[:n]
	y = y[:n]

	var sumX, sumY, sumXY, sumX2, sumY2 float64
	for i := 0; i < n; i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
		sumY2 += y[i] * y[i]
	}

	nf := float64(n)
	numerator := nf*sumXY - sumX*sumY
	denominator := math.Sqrt((nf*sumX2 - sumX*sumX) * (nf*sumY2 - sumY*sumY))

	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}

// Compute returns a map of symbol pair correlations.
// Keys are formatted as "SYMBOL1:SYMBOL2" (alphabetically sorted).
func (cm *CorrelationMatrix) Compute() map[string]float64 {
	result := make(map[string]float64)

	symbols := make([]string, 0, len(cm.history))
	for s := range cm.history {
		if len(cm.history[s]) >= 2 {
			symbols = append(symbols, s)
		}
	}

	for i := 0; i < len(symbols); i++ {
		for j := i + 1; j < len(symbols); j++ {
			// Compute returns instead of raw prices for meaningful correlation
			retI := returns(cm.history[symbols[i]])
			retJ := returns(cm.history[symbols[j]])

			n := min(len(retI), len(retJ))
			if n < 2 {
				continue
			}

			// Use the most recent overlapping returns
			corr := Pearson(retI[len(retI)-n:], retJ[len(retJ)-n:])

			key := symbols[i] + ":" + symbols[j]
			result[key] = corr
		}
	}

	return result
}

// DiversificationScore computes a portfolio diversification score from 0 to 1.
// 1 = perfectly diversified (uncorrelated), 0 = all perfectly correlated.
func (cm *CorrelationMatrix) DiversificationScore() float64 {
	correlations := cm.Compute()
	if len(correlations) == 0 {
		return 1.0 // Single asset = no correlation = max diversification by default
	}

	var totalAbsCorr float64
	for _, corr := range correlations {
		totalAbsCorr += math.Abs(corr)
	}

	avgAbsCorr := totalAbsCorr / float64(len(correlations))
	return 1 - avgAbsCorr
}

// Returns converts a price series to percentage returns.
func returns(prices []float64) []float64 {
	if len(prices) < 2 {
		return nil
	}
	ret := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		if prices[i-1] != 0 {
			ret[i-1] = (prices[i] - prices[i-1]) / prices[i-1]
		}
	}
	return ret
}

// HeatmapData generates data suitable for a correlation heatmap visualization.
type HeatmapCell struct {
	Symbol1 string  `json:"symbol1"`
	Symbol2 string  `json:"symbol2"`
	Correlation float64 `json:"correlation"`
}

// HeatmapData returns all correlation pairs as a flat list for visualization.
func (cm *CorrelationMatrix) HeatmapData() []HeatmapCell {
	correlations := cm.Compute()
	cells := make([]HeatmapCell, 0, len(correlations))

	for key, corr := range correlations {
		// Parse key
		var s1, s2 string
		for i, c := range key {
			if c == ':' {
				s1 = key[:i]
				s2 = key[i+1:]
				break
			}
		}
		cells = append(cells, HeatmapCell{
			Symbol1:     s1,
			Symbol2:     s2,
			Correlation: math.Round(corr*1000) / 1000, // 3 decimal places
		})
	}
	return cells
}

// Symbols returns the list of symbols currently tracked.
func (cm *CorrelationMatrix) Symbols() []string {
	symbols := make([]string, 0, len(cm.history))
	for s := range cm.history {
		symbols = append(symbols, s)
	}
	return symbols
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
