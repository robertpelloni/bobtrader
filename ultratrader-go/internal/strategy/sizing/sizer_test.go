package sizing

import (
	"math"
	"testing"
)

const eps = 1e-6

func TestFixedSizer(t *testing.T) {
	s := NewFixedSizer(1.5)
	if s.Name() != "fixed" {
		t.Errorf("expected fixed name")
	}
	result := s.Size(SizingInput{PortfolioValue: 100000, Price: 50000})
	if math.Abs(result-1.5) > eps {
		t.Errorf("expected 1.5, got %f", result)
	}
}

func TestPercentRiskSizer_Basic(t *testing.T) {
	s := NewPercentRiskSizer(0.01)
	if s.Name() != "percent-risk" {
		t.Errorf("expected percent-risk name")
	}

	// $100k portfolio, 1% risk = $1000 risk budget
	// Stop distance = $1000
	// Position = $1000 / $1000 = 1 unit
	result := s.Size(SizingInput{
		PortfolioValue: 100000,
		Price:          50000,
		RiskPercent:    0.01,
		StopDistance:   1000,
	})
	if math.Abs(result-1.0) > eps {
		t.Errorf("expected 1.0, got %f", result)
	}
}

func TestPercentRiskSizer_ATRFallback(t *testing.T) {
	s := NewPercentRiskSizer(0.02)

	// No stop distance, should fall back to ATR
	result := s.Size(SizingInput{
		PortfolioValue: 100000,
		Price:          50000,
		RiskPercent:    0.01,
		ATR:            500,
	})
	// Risk budget = 1000, stop = 500 (ATR), units = 2
	if math.Abs(result-2.0) > eps {
		t.Errorf("expected 2.0, got %f", result)
	}
}

func TestPercentRiskSizer_ZeroStop(t *testing.T) {
	s := NewPercentRiskSizer(0.01)
	result := s.Size(SizingInput{
		PortfolioValue: 100000,
		Price:          50000,
	})
	if result != 0 {
		t.Errorf("expected 0 for zero stop distance, got %f", result)
	}
}

func TestPercentRiskSizer_DefaultRisk(t *testing.T) {
	s := NewPercentRiskSizer(0.02) // 2% default risk

	// No risk specified in input, should use 2% default
	result := s.Size(SizingInput{
		PortfolioValue: 100000,
		Price:          50000,
		StopDistance:   1000,
	})
	// $100k * 2% = $2000 / $1000 = 2 units
	if math.Abs(result-2.0) > eps {
		t.Errorf("expected 2.0, got %f", result)
	}
}

func TestKellySizer_Basic(t *testing.T) {
	// 60% win rate, 1.5:1 win/loss ratio, full Kelly
	s := NewKellySizer(0.6, 1.5, 1.0)
	if s.Name() != "kelly" {
		t.Errorf("expected kelly name")
	}

	// Kelly = 0.6 - (0.4 / 1.5) = 0.6 - 0.267 = 0.333
	// Full Kelly at $100k portfolio
	// Position = 0.333 * 100000 / 50000 = 0.666 units
	result := s.Size(SizingInput{
		PortfolioValue: 100000,
		Price:          50000,
	})
	if result <= 0 {
		t.Errorf("expected positive position size, got %f", result)
	}
}

func TestKellySizer_HalfKelly(t *testing.T) {
	// Use parameters where full Kelly won't be capped at 25%
	full := NewKellySizer(0.6, 1.5, 1.0)
	half := NewKellySizer(0.6, 1.5, 0.5)

	input := SizingInput{PortfolioValue: 100000, Price: 50000}
	fullSize := full.Size(input)
	halfSize := half.Size(input)

	// Kelly = 0.6 - 0.4/1.5 = 0.333..., full fraction
	// Half Kelly = 0.333 * 0.5 = 0.167
	// Full Kelly gets capped at 0.25, so full = 0.25 * 100k / 50k = 0.5
	// Half Kelly = 0.167 * 100k / 50k = 0.333
	if fullSize <= halfSize {
		t.Errorf("full Kelly (%f) should be > half Kelly (%f)", fullSize, halfSize)
	}
}

func TestKellySizer_LosingStrategy(t *testing.T) {
	// 30% win rate = negative Kelly
	s := NewKellySizer(0.3, 1.0, 1.0)
	result := s.Size(SizingInput{
		PortfolioValue: 100000,
		Price:          50000,
	})
	if result != 0 {
		t.Errorf("expected 0 for negative Kelly, got %f", result)
	}
}

func TestKellySizer_CappedAt25Percent(t *testing.T) {
	// Very high Kelly should be capped
	s := NewKellySizer(0.9, 5.0, 1.0)
	result := s.Size(SizingInput{
		PortfolioValue: 100000,
		Price:          50000,
	})
	maxUnits := 0.25 * 100000 / 50000 // 25% of portfolio
	if result > maxUnits+eps {
		t.Errorf("expected at most %f units (25%% cap), got %f", maxUnits, result)
	}
}

func TestVolatilityTargetSizer(t *testing.T) {
	s := NewVolatilityTargetSizer(0.15, 14)
	if s.Name() != "volatility-target" {
		t.Errorf("expected volatility-target name")
	}

	result := s.Size(SizingInput{
		PortfolioValue: 100000,
		Price:          50000,
		ATR:            1500, // 3% daily ATR
	})
	if result <= 0 {
		t.Errorf("expected positive size, got %f", result)
	}
}

func TestVolatilityTargetSizer_ZeroATR(t *testing.T) {
	s := NewVolatilityTargetSizer(0.15, 14)
	result := s.Size(SizingInput{
		PortfolioValue: 100000,
		Price:          50000,
	})
	if result != 0 {
		t.Errorf("expected 0 for zero ATR, got %f", result)
	}
}

func TestEqualWeightSizer(t *testing.T) {
	s := NewEqualWeightSizer(4)
	if s.Name() != "equal-weight" {
		t.Errorf("expected equal-weight name")
	}

	// $100k / 4 = $25k per position, at $50k price = 0.5 units
	result := s.Size(SizingInput{
		PortfolioValue: 100000,
		Price:          50000,
	})
	if math.Abs(result-0.5) > eps {
		t.Errorf("expected 0.5, got %f", result)
	}
}

func TestEqualWeightSizer_SinglePosition(t *testing.T) {
	s := NewEqualWeightSizer(1)
	result := s.Size(SizingInput{
		PortfolioValue: 100000,
		Price:          50000,
	})
	if math.Abs(result-2.0) > eps {
		t.Errorf("expected 2.0 (full portfolio), got %f", result)
	}
}

func TestEqualWeightSizer_ZeroPositions(t *testing.T) {
	s := NewEqualWeightSizer(0)
	result := s.Size(SizingInput{
		PortfolioValue: 100000,
		Price:          50000,
	})
	if result != 0 {
		t.Errorf("expected 0 for zero positions, got %f", result)
	}
}
