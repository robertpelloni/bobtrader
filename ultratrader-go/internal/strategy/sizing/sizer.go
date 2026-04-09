package sizing

import (
	"math"
)

// SizingInput contains the data needed to calculate position size.
type SizingInput struct {
	PortfolioValue float64 // Total portfolio value
	Price          float64 // Current asset price
	RiskPercent    float64 // Risk as percentage of portfolio (0.01 = 1%)
	ATR            float64 // Average True Range (for volatility-based sizing)
	StopDistance    float64 // Stop loss distance from entry price (0 = use ATR)
}

// PositionSizer calculates the number of units to trade.
type PositionSizer interface {
	Name() string
	Size(input SizingInput) float64
}

// FixedSizer returns a fixed lot size regardless of portfolio.
type FixedSizer struct {
	Quantity float64
}

func NewFixedSizer(qty float64) *FixedSizer {
	return &FixedSizer{Quantity: qty}
}

func (f *FixedSizer) Name() string { return "fixed" }
func (f *FixedSizer) Size(_ SizingInput) float64 {
	return f.Quantity
}

// PercentRiskSizer sizes positions so that a stop-out risks a fixed
// percentage of the portfolio. This is the most common professional sizing method.
// position_size = (portfolio_value * risk_pct) / stop_distance
type PercentRiskSizer struct {
	DefaultRisk float64 // Default risk percentage if not specified in input
}

func NewPercentRiskSizer(defaultRisk float64) *PercentRiskSizer {
	return &PercentRiskSizer{DefaultRisk: defaultRisk}
}

func (p *PercentRiskSizer) Name() string { return "percent-risk" }

func (p *PercentRiskSizer) Size(input SizingInput) float64 {
	risk := input.RiskPercent
	if risk <= 0 {
		risk = p.DefaultRisk
	}
	if risk <= 0 {
		risk = 0.01 // 1% default
	}

	stopDist := input.StopDistance
	if stopDist <= 0 && input.ATR > 0 {
		stopDist = input.ATR // Use ATR as stop distance fallback
	}
	if stopDist <= 0 || input.Price <= 0 {
		return 0
	}

	dollarRisk := input.PortfolioValue * risk
	units := dollarRisk / stopDist
	return math.Max(0, units)
}

// KellySizer uses the Kelly Criterion for optimal position sizing.
// kelly_fraction = win_rate - (loss_rate / avg_win_loss_ratio)
// position_size = kelly_fraction * portfolio_value / price
type KellySizer struct {
	WinRate       float64 // Historical win rate (0.0 to 1.0)
	AvgWinLossRatio float64 // Average win / average loss
	Fraction     float64 // Kelly fraction (0.5 = half Kelly for safety)
}

func NewKellySizer(winRate, avgWinLossRatio, fraction float64) *KellySizer {
	return &KellySizer{
		WinRate:         winRate,
		AvgWinLossRatio: avgWinLossRatio,
		Fraction:        fraction,
	}
}

func (k *KellySizer) Name() string { return "kelly" }

func (k *KellySizer) Size(input SizingInput) float64 {
	if k.AvgWinLossRatio <= 0 || input.Price <= 0 || input.PortfolioValue <= 0 {
		return 0
	}

	// Kelly formula: f* = p - (1-p) / b
	// where p = win rate, b = win/loss ratio
	lossRate := 1 - k.WinRate
	kelly := k.WinRate - (lossRate / k.AvgWinLossRatio)

	// Apply fractional Kelly for safety
	fractionalKelly := kelly * k.Fraction
	if fractionalKelly <= 0 {
		return 0
	}

	// Cap at 25% of portfolio for safety
	if fractionalKelly > 0.25 {
		fractionalKelly = 0.25
	}

	positionValue := input.PortfolioValue * fractionalKelly
	units := positionValue / input.Price
	return math.Max(0, units)
}

// VolatilityTargetSizer adjusts position size to target a constant portfolio
// volatility contribution from each position.
// position_size = (target_vol * portfolio_value) / (price * ATR_normalized)
type VolatilityTargetSizer struct {
	TargetVolatility float64 // Target annual volatility (e.g., 0.15 = 15%)
	ATRPeriods       int     // Number of periods used for ATR calculation
}

func NewVolatilityTargetSizer(targetVol float64, atrPeriods int) *VolatilityTargetSizer {
	return &VolatilityTargetSizer{
		TargetVolatility: targetVol,
		ATRPeriods:       atrPeriods,
	}
}

func (v *VolatilityTargetSizer) Name() string { return "volatility-target" }

func (v *VolatilityTargetSizer) Size(input SizingInput) float64 {
	if input.ATR <= 0 || input.Price <= 0 || input.PortfolioValue <= 0 {
		return 0
	}

	// Normalize ATR to percentage of price
	atrNormalized := input.ATR / input.Price

	// Annualize if we know the period (assume daily for 252 trading days)
	if atrNormalized > 0 {
		atrNormalized = atrNormalized * math.Sqrt(252)
	}

	if atrNormalized <= 0 {
		return 0
	}

	// Position units = target_vol * portfolio / (price * annualized_atr)
	units := (v.TargetVolatility * input.PortfolioValue) / (input.Price * atrNormalized)
	return math.Max(0, units)
}

// EqualWeightSizer divides portfolio equally among N positions.
type EqualWeightSizer struct {
	NumPositions int
}

func NewEqualWeightSizer(numPositions int) *EqualWeightSizer {
	return &EqualWeightSizer{NumPositions: numPositions}
}

func (e *EqualWeightSizer) Name() string { return "equal-weight" }

func (e *EqualWeightSizer) Size(input SizingInput) float64 {
	if e.NumPositions <= 0 || input.Price <= 0 || input.PortfolioValue <= 0 {
		return 0
	}
	allocation := input.PortfolioValue / float64(e.NumPositions)
	return allocation / input.Price
}
