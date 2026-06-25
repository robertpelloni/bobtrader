package sizing

import (
	"math"
)

// ATRSizer implements volatility-based position sizing using ATR.
// It targets a specific risk amount per trade based on the ATR-derived stop distance.
type ATRSizer struct {
	RiskPerTrade float64 // Risk amount in quote currency (e.g., $100)
	ATRExpansion float64 // Multiplier for ATR to set stop distance (e.g., 2.0)
}

func NewATRSizer(risk float64, expansion float64) *ATRSizer {
	return &ATRSizer{
		RiskPerTrade: risk,
		ATRExpansion: expansion,
	}
}

func (a *ATRSizer) Name() string { return "atr-volatility" }

func (a *ATRSizer) Size(input SizingInput) float64 {
	if input.ATR <= 0 || input.Price <= 0 {
		return 0
	}

	// stop_distance = ATR * expansion
	stopDist := input.ATR * a.ATRExpansion
	if stopDist <= 0 {
		return 0
	}

	// RiskPerTrade = units * stop_distance
	// units = RiskPerTrade / stop_distance
	units := a.RiskPerTrade / stopDist

	// Safety cap: never more than 50% of portfolio
	if input.PortfolioValue > 0 {
		maxUnits := (input.PortfolioValue * 0.5) / input.Price
		if units > maxUnits {
			units = maxUnits
		}
	}

	return math.Max(0, units)
}
