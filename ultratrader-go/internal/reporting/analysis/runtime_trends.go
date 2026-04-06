package analysis

import (
	"encoding/json"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/metrics"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/persistence/reports"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/execution"
)

type NumericTrend struct {
	Samples  int     `json:"samples"`
	Latest   float64 `json:"latest"`
	Previous float64 `json:"previous"`
	Delta    float64 `json:"delta"`
}

type RuntimeTrends struct {
	MetricsSamples            int                `json:"metrics_samples"`
	ValuationSamples          int                `json:"valuation_samples"`
	ExecutionSummarySamples   int                `json:"execution_summary_samples"`
	SuccessRate               NumericTrend       `json:"success_rate"`
	BlockedRate               NumericTrend       `json:"blocked_rate"`
	PortfolioValue            NumericTrend       `json:"portfolio_value"`
	RealizedPnL               NumericTrend       `json:"realized_pnl"`
	UnrealizedPnL             NumericTrend       `json:"unrealized_pnl"`
	DominantBlockCount        NumericTrend       `json:"dominant_block_count"`
	TopConcentrationPct       NumericTrend       `json:"top_concentration_pct"`
	LatestBlockReasons        map[string]int     `json:"latest_block_reasons,omitempty"`
	LatestDominantBlockReason string             `json:"latest_dominant_block_reason,omitempty"`
	LatestDominantBlockCount  int                `json:"latest_dominant_block_count,omitempty"`
	LatestConcentration       map[string]float64 `json:"latest_concentration,omitempty"`
	LatestTopConcentration    string             `json:"latest_top_concentration,omitempty"`
	LatestTopConcentrationPct float64            `json:"latest_top_concentration_pct,omitempty"`
	LatestTopSymbol           string             `json:"latest_top_symbol,omitempty"`
	LatestTopSymbolCount      int                `json:"latest_top_symbol_count,omitempty"`
}

type metricsPayload struct {
	Metrics metrics.Snapshot `json:"metrics"`
}

type valuationPayload struct {
	PortfolioValue float64            `json:"portfolio_value"`
	RealizedPnL    float64            `json:"realized_pnl"`
	UnrealizedPnL  float64            `json:"unrealized_pnl"`
	Concentration  map[string]float64 `json:"concentration,omitempty"`
}

type executionPayload struct {
	Summary execution.Summary `json:"summary"`
}

func BuildRuntimeTrends(metricReports, valuationReports, executionReports []reports.Report) RuntimeTrends {
	trends := RuntimeTrends{
		MetricsSamples:          len(metricReports),
		ValuationSamples:        len(valuationReports),
		ExecutionSummarySamples: len(executionReports),
	}

	metricPayloads := decodeReports[metricsPayload](metricReports)
	valuationPayloads := decodeReports[valuationPayload](valuationReports)
	executionPayloads := decodeReports[executionPayload](executionReports)

	trends.SuccessRate = numericTrendFromMetrics(metricPayloads, func(p metricsPayload) float64 { return p.Metrics.SuccessRate })
	trends.BlockedRate = numericTrendFromMetrics(metricPayloads, func(p metricsPayload) float64 { return p.Metrics.BlockedRate })
	trends.PortfolioValue = numericTrendFromValuation(valuationPayloads, func(p valuationPayload) float64 { return p.PortfolioValue })
	trends.RealizedPnL = numericTrendFromValuation(valuationPayloads, func(p valuationPayload) float64 { return p.RealizedPnL })
	trends.UnrealizedPnL = numericTrendFromValuation(valuationPayloads, func(p valuationPayload) float64 { return p.UnrealizedPnL })
	trends.DominantBlockCount = numericTrendFromMetrics(metricPayloads, func(p metricsPayload) float64 { return float64(maxCount(p.Metrics.BlockReasons)) })
	trends.TopConcentrationPct = numericTrendFromValuation(valuationPayloads, func(p valuationPayload) float64 { return maxConcentration(p.Concentration) })

	if len(metricPayloads) > 0 {
		trends.LatestBlockReasons = metricPayloads[len(metricPayloads)-1].Metrics.BlockReasons
		for reason, count := range trends.LatestBlockReasons {
			if count > trends.LatestDominantBlockCount {
				trends.LatestDominantBlockReason = reason
				trends.LatestDominantBlockCount = count
			}
		}
	}
	if len(valuationPayloads) > 0 {
		trends.LatestConcentration = valuationPayloads[len(valuationPayloads)-1].Concentration
		for symbol, pct := range trends.LatestConcentration {
			if pct > trends.LatestTopConcentrationPct {
				trends.LatestTopConcentration = symbol
				trends.LatestTopConcentrationPct = pct
			}
		}
	}
	if len(executionPayloads) > 0 {
		last := executionPayloads[len(executionPayloads)-1].Summary
		trends.LatestTopSymbol = last.TopSymbol
		trends.LatestTopSymbolCount = last.TopSymbolCount
	}

	return trends
}

func decodeReports[T any](items []reports.Report) []T {
	out := make([]T, 0, len(items))
	for _, item := range items {
		var payload T
		data, err := json.Marshal(item.Payload)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			continue
		}
		out = append(out, payload)
	}
	return out
}

func numericTrendFromMetrics[T any](items []T, selector func(T) float64) NumericTrend {
	return numericTrend(items, selector)
}

func numericTrendFromValuation[T any](items []T, selector func(T) float64) NumericTrend {
	return numericTrend(items, selector)
}

func numericTrend[T any](items []T, selector func(T) float64) NumericTrend {
	trend := NumericTrend{Samples: len(items)}
	if len(items) == 0 {
		return trend
	}
	trend.Latest = selector(items[len(items)-1])
	if len(items) > 1 {
		trend.Previous = selector(items[len(items)-2])
		trend.Delta = trend.Latest - trend.Previous
		return trend
	}
	trend.Previous = trend.Latest
	return trend
}

func maxCount(values map[string]int) int {
	max := 0
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

func maxConcentration(values map[string]float64) float64 {
	max := 0.0
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}
