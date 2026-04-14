package correlation

import (
	"math"
	"testing"
)

const eps = 1e-6

func TestPearson_PerfectPositive(t *testing.T) {
	x := []float64{1, 2, 3, 4, 5}
	y := []float64{2, 4, 6, 8, 10}
	corr := Pearson(x, y)
	if math.Abs(corr-1.0) > eps {
		t.Errorf("expected perfect positive correlation 1.0, got %f", corr)
	}
}

func TestPearson_PerfectNegative(t *testing.T) {
	x := []float64{1, 2, 3, 4, 5}
	y := []float64{10, 8, 6, 4, 2}
	corr := Pearson(x, y)
	if math.Abs(corr-(-1.0)) > eps {
		t.Errorf("expected perfect negative correlation -1.0, got %f", corr)
	}
}

func TestPearson_NoCorrelation(t *testing.T) {
	x := []float64{1, 2, 3, 4, 5}
	y := []float64{3, 1, 4, 1, 5}
	corr := Pearson(x, y)
	if corr > 0.5 || corr < -0.5 {
		t.Errorf("expected near-zero correlation, got %f", corr)
	}
}

func TestPearson_InsufficientData(t *testing.T) {
	corr := Pearson([]float64{1}, []float64{2})
	if corr != 0 {
		t.Errorf("expected 0 for insufficient data, got %f", corr)
	}
}

func TestPearson_DifferentLengths(t *testing.T) {
	x := []float64{1, 2, 3, 4, 5}
	y := []float64{2, 4, 6}
	corr := Pearson(x, y)
	if math.Abs(corr-1.0) > eps {
		t.Errorf("expected 1.0 for overlapping portion, got %f", corr)
	}
}

func TestCorrelationMatrix_Compute(t *testing.T) {
	cm := NewCorrelationMatrix(10)

	// Add correlated prices
	pricesA := []float64{100, 101, 102, 103, 104, 105}
	pricesB := []float64{200, 202, 204, 206, 208, 210} // perfectly correlated

	for _, p := range pricesA {
		cm.AddPrice("A", p)
	}
	for _, p := range pricesB {
		cm.AddPrice("B", p)
	}

	correlations := cm.Compute()
	if len(correlations) != 1 {
		t.Fatalf("expected 1 correlation pair, got %d", len(correlations))
	}

	corr := correlations["A:B"]
	if corr < 0.99 {
		t.Errorf("expected near-perfect correlation, got %f", corr)
	}
}

func TestCorrelationMatrix_DiversificationScore(t *testing.T) {
	cm := NewCorrelationMatrix(10)

	// Add uncorrelated data
	pricesA := []float64{100, 101, 99, 102, 98, 103}
	pricesB := []float64{50, 49, 51, 48, 52, 47} // somewhat inverse

	for _, p := range pricesA {
		cm.AddPrice("A", p)
	}
	for _, p := range pricesB {
		cm.AddPrice("B", p)
	}

	score := cm.DiversificationScore()
	if score < 0 || score > 1 {
		t.Errorf("diversification score should be 0-1, got %f", score)
	}
}

func TestCorrelationMatrix_PerfectCorrelation_LowDiversification(t *testing.T) {
	cm := NewCorrelationMatrix(10)

	for i := 0; i < 5; i++ {
		price := float64(100 + i)
		cm.AddPrice("A", price)
		cm.AddPrice("B", price*2) // Perfect correlation
	}

	score := cm.DiversificationScore()
	if score > 0.1 {
		t.Errorf("expected low diversification for perfectly correlated assets, got %f", score)
	}
}

func TestCorrelationMatrix_SingleAsset(t *testing.T) {
	cm := NewCorrelationMatrix(10)
	cm.AddPrice("A", 100)
	cm.AddPrice("A", 101)

	// Single asset should return max diversification (no correlation to measure)
	score := cm.DiversificationScore()
	if score != 1.0 {
		t.Errorf("expected 1.0 for single asset, got %f", score)
	}
}

func TestCorrelationMatrix_HeatmapData(t *testing.T) {
	cm := NewCorrelationMatrix(10)

	for i := 0; i < 5; i++ {
		cm.AddPrice("BTC", float64(100+i))
		cm.AddPrice("ETH", float64(50+i))
		cm.AddPrice("DOGE", float64(10-i*2))
	}

	cells := cm.HeatmapData()
	if len(cells) != 3 { // 3 choose 2 = 3 pairs
		t.Errorf("expected 3 heatmap cells, got %d", len(cells))
	}

	for _, cell := range cells {
		if cell.Symbol1 == "" || cell.Symbol2 == "" {
			t.Errorf("empty symbol in cell: %+v", cell)
		}
	}
}

func TestCorrelationMatrix_Symbols(t *testing.T) {
	cm := NewCorrelationMatrix(10)
	cm.AddPrice("BTC", 100)
	cm.AddPrice("ETH", 50)

	symbols := cm.Symbols()
	if len(symbols) != 2 {
		t.Errorf("expected 2 symbols, got %d", len(symbols))
	}
}

func TestCorrelationMatrix_WindowTrimming(t *testing.T) {
	cm := NewCorrelationMatrix(5)

	for i := 0; i < 10; i++ {
		cm.AddPrice("A", float64(100+i))
	}

	if len(cm.history["A"]) != 5 {
		t.Errorf("expected 5 prices after trimming, got %d", len(cm.history["A"]))
	}
}

func TestReturns(t *testing.T) {
	prices := []float64{100, 110, 99, 105}
	ret := returns(prices)

	if len(ret) != 3 {
		t.Fatalf("expected 3 returns, got %d", len(ret))
	}
	if math.Abs(ret[0]-0.1) > eps {
		t.Errorf("first return: expected 0.1, got %f", ret[0])
	}
	if math.Abs(ret[1]-(-0.1)) > eps {
		t.Errorf("second return: expected -0.1, got %f", ret[1])
	}
}
