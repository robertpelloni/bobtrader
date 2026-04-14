package analytics

import (
	"fmt"
)

// OrderFlowData contains aggregated trade volume data.
type OrderFlowData struct {
	BuyVolume  float64
	SellVolume float64
	Delta      float64 // Net difference between buy and sell volume
	CVD        float64 // Cumulative Volume Delta
}

// DivergenceType indicates whether order flow agrees or disagrees with price action.
type DivergenceType string

const (
	BullishDivergence DivergenceType = "BULLISH_DIVERGENCE"
	BearishDivergence DivergenceType = "BEARISH_DIVERGENCE"
	NoDivergence      DivergenceType = "NO_DIVERGENCE"
)

// OrderFlowAnalyzer detects cumulative volume delta (CVD) and divergences.
type OrderFlowAnalyzer struct {
	lookback int
}

// NewOrderFlowAnalyzer creates a new analyzer.
func NewOrderFlowAnalyzer(lookback int) *OrderFlowAnalyzer {
	if lookback < 5 {
		lookback = 5
	}
	return &OrderFlowAnalyzer{
		lookback: lookback,
	}
}

// Analyze calculates the CVD for a sequence of volumes and identifies divergences.
func (a *OrderFlowAnalyzer) Analyze(prices []float64, buyVolumes []float64, sellVolumes []float64) ([]OrderFlowData, DivergenceType, error) {
	if len(prices) != len(buyVolumes) || len(prices) != len(sellVolumes) {
		return nil, NoDivergence, fmt.Errorf("length mismatch between prices and volumes")
	}

	if len(prices) < a.lookback {
		return nil, NoDivergence, fmt.Errorf("insufficient data for order flow analysis")
	}

	var results []OrderFlowData
	var currentCVD float64

	for i := 0; i < len(prices); i++ {
		bv := buyVolumes[i]
		sv := sellVolumes[i]
		delta := bv - sv
		currentCVD += delta

		results = append(results, OrderFlowData{
			BuyVolume:  bv,
			SellVolume: sv,
			Delta:      delta,
			CVD:        currentCVD,
		})
	}

	// Divergence logic
	// Look at the trend over the lookback window.
	startIndex := len(prices) - a.lookback
	endIndex := len(prices) - 1

	startPrice := prices[startIndex]
	endPrice := prices[endIndex]
	priceTrend := endPrice - startPrice

	startCVD := results[startIndex].CVD
	endCVD := results[endIndex].CVD
	cvdTrend := endCVD - startCVD

	// Divergence: Price makes higher high, but CVD makes lower high (Bearish)
	// Simplification for the lookback window: Price went up, but net selling occurred.
	if priceTrend > 0 && cvdTrend < 0 {
		return results, BearishDivergence, nil
	}

	// Divergence: Price makes lower low, but CVD makes higher low (Bullish)
	// Simplification: Price went down, but net buying occurred.
	if priceTrend < 0 && cvdTrend > 0 {
		return results, BullishDivergence, nil
	}

	return results, NoDivergence, nil
}
