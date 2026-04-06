package analysis

import (
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/persistence/reports"
)

func TestBuildRuntimeTrends(t *testing.T) {
	metricReports := []reports.Report{
		{Timestamp: time.Now(), Type: "metrics-snapshot", Payload: map[string]any{"metrics": map[string]any{"success_rate": 0.5, "blocked_rate": 0.5, "block_reasons": map[string]any{"cooldown": 1}}}},
		{Timestamp: time.Now(), Type: "metrics-snapshot", Payload: map[string]any{"metrics": map[string]any{"success_rate": 0.75, "blocked_rate": 0.25, "block_reasons": map[string]any{"cooldown": 2}}}},
	}
	valuationReports := []reports.Report{
		{Timestamp: time.Now(), Type: "portfolio-valuation", Payload: map[string]any{"portfolio_value": 100.0, "realized_pnl": 5.0, "unrealized_pnl": 10.0, "concentration": map[string]any{"BTCUSDT": 1.0}}},
		{Timestamp: time.Now(), Type: "portfolio-valuation", Payload: map[string]any{"portfolio_value": 120.0, "realized_pnl": 6.0, "unrealized_pnl": 11.0, "concentration": map[string]any{"BTCUSDT": 0.8}}},
	}
	executionReports := []reports.Report{
		{Timestamp: time.Now(), Type: "execution-summary", Payload: map[string]any{"summary": map[string]any{"top_symbol": "BTCUSDT", "top_symbol_count": 3}}},
	}

	trends := BuildRuntimeTrends(metricReports, valuationReports, executionReports)
	if trends.SuccessRate.Latest != 0.75 || trends.SuccessRate.Delta != 0.25 {
		t.Fatalf("unexpected success rate trend: %+v", trends.SuccessRate)
	}
	if trends.PortfolioValue.Latest != 120 || trends.PortfolioValue.Delta != 20 {
		t.Fatalf("unexpected portfolio value trend: %+v", trends.PortfolioValue)
	}
	if trends.LatestTopSymbol != "BTCUSDT" || trends.LatestTopSymbolCount != 3 {
		t.Fatalf("unexpected execution trend metadata: %+v", trends)
	}
	if trends.LatestDominantBlockReason != "cooldown" || trends.LatestDominantBlockCount != 2 {
		t.Fatalf("unexpected dominant block reason metadata: %+v", trends)
	}
	if trends.LatestTopConcentration != "BTCUSDT" || trends.LatestTopConcentrationPct != 0.8 {
		t.Fatalf("unexpected concentration trend metadata: %+v", trends)
	}
}
