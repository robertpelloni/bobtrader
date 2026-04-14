package optimizer

import (
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/backtest"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

func TestSharpeScorer_Positive(t *testing.T) {
	scorer := SharpeScorer()
	result := backtest.Result{TotalTrades: 10, RealizedPnL: 100}
	score := scorer(result)
	if score <= 0 {
		t.Errorf("expected positive Sharpe score for positive PnL, got %v", score)
	}
}

func TestSharpeScorer_Negative(t *testing.T) {
	scorer := SharpeScorer()
	result := backtest.Result{TotalTrades: 10, RealizedPnL: -50}
	score := scorer(result)
	if score >= 0 {
		t.Errorf("expected negative Sharpe score for negative PnL, got %v", score)
	}
}

func TestSharpeScorer_NoTrades(t *testing.T) {
	scorer := SharpeScorer()
	result := backtest.Result{TotalTrades: 0}
	score := scorer(result)
	if score != 0 {
		t.Errorf("expected zero score with no trades, got %v", score)
	}
}

func TestProfitFactorScorer(t *testing.T) {
	scorer := ProfitFactorScorer()

	// Positive PnL
	posResult := backtest.Result{TotalTrades: 5, RealizedPnL: 500}
	posScore := scorer(posResult)
	if posScore <= 1 {
		t.Errorf("expected profit factor > 1 for positive PnL, got %v", posScore)
	}

	// Negative PnL
	negResult := backtest.Result{TotalTrades: 5, RealizedPnL: -200}
	negScore := scorer(negResult)
	if negScore >= 0 {
		t.Errorf("expected negative profit factor for losses, got %v", negScore)
	}
}

func TestWinRateScorer(t *testing.T) {
	scorer := WinRateScorer()

	result := backtest.Result{
		TotalTrades: 4,
		Orders: []exchange.Order{
			{Side: "buy", Price: "100"},
			{Side: "sell", Price: "110"}, // win
			{Side: "buy", Price: "100"},
			{Side: "sell", Price: "95"}, // loss
		},
	}

	score := scorer(result)
	// 1 win out of 2 completed pairs = 0.5
	if score < 0.49 || score > 0.51 {
		t.Errorf("expected win rate ~0.5, got %v", score)
	}
}

func TestWinRateScorer_Perfect(t *testing.T) {
	scorer := WinRateScorer()

	result := backtest.Result{
		TotalTrades: 2,
		Orders: []exchange.Order{
			{Side: "buy", Price: "100"},
			{Side: "sell", Price: "120"}, // win
		},
	}

	score := scorer(result)
	if score != 1.0 {
		t.Errorf("expected perfect win rate 1.0, got %v", score)
	}
}

func TestCompositeScorer(t *testing.T) {
	composite := NewCompositeScorer().
		Add(DefaultScorer, 0.5).
		Add(SharpeScorer(), 0.5)

	result := backtest.Result{TotalTrades: 10, RealizedPnL: 100}
	score := composite.Scorer()(result)

	if score <= 0 {
		t.Errorf("expected positive composite score, got %v", score)
	}
}

func TestSimpleParseFloat(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"100", 100},
		{"100.50", 100.5},
		{"0.001", 0.001},
		{"-50.5", -50.5},
		{"abc", 0},
		{"", 0},
	}
	for _, tt := range tests {
		got := simpleParseFloat(tt.input)
		if got != tt.expected {
			t.Errorf("simpleParseFloat(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}
