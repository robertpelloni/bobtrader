package reporting

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// TimePeriod represents a time range.
type TimePeriod struct {
	Start time.Time
	End   time.Time
}

// TradeRecord represents a single trade.
type TradeRecord struct {
	EntryTime  time.Time
	ExitTime   time.Time
	Side       string
	EntryPrice float64
	ExitPrice  float64
	Quantity   float64
	PnL        float64
	PnLPct     float64
	Fees       float64
	Duration   time.Duration
}

// EquityPoint represents a point on the equity curve.
type EquityPoint struct {
	Timestamp time.Time
	Value     float64
	Drawdown  float64
}

// BacktestReport contains comprehensive backtest results.
type BacktestReport struct {
	ID             string
	Strategy       string
	Symbol         string
	Period         TimePeriod
	InitialBalance float64
	FinalBalance   float64

	// Performance Metrics
	TotalReturn  float64
	AnnualReturn float64
	SharpeRatio  float64
	SortinoRatio float64
	CalmarRatio  float64
	MaxDrawdown  float64

	// Trade Statistics
	TotalTrades   int
	WinningTrades int
	LosingTrades  int
	WinRate       float64
	ProfitFactor  float64
	AvgWin        float64
	AvgLoss       float64
	LargestWin    float64
	LargestLoss   float64
	AvgHoldTime   time.Duration

	// Risk Metrics
	Volatility       float64
	DownsideDev      float64
	Beta             float64
	Alpha            float64
	InformationRatio float64

	// Trade Details
	Trades         []TradeRecord
	EquityCurve    []EquityPoint
	DrawdownCurve  []EquityPoint
	MonthlyReturns map[string]float64

	// Metadata
	GeneratedAt time.Time
	Config      map[string]interface{}
}

// ReportGenerator generates backtest reports.
type ReportGenerator struct{}

func NewReportGenerator() *ReportGenerator {
	return &ReportGenerator{}
}

// Generate creates a report from backtest data.
func (g *ReportGenerator) Generate(
	strategy, symbol string,
	initialBalance float64,
	trades []TradeRecord,
	equityCurve []EquityPoint,
	config map[string]interface{},
) *BacktestReport {
	report := &BacktestReport{
		ID:             fmt.Sprintf("backtest-%s-%s-%d", strategy, symbol, time.Now().Unix()),
		Strategy:       strategy,
		Symbol:         symbol,
		InitialBalance: initialBalance,
		Trades:         trades,
		EquityCurve:    equityCurve,
		MonthlyReturns: make(map[string]float64),
		GeneratedAt:    time.Now(),
		Config:         config,
	}

	if len(equityCurve) > 0 {
		report.Period.Start = equityCurve[0].Timestamp
		report.Period.End = equityCurve[len(equityCurve)-1].Timestamp
		report.FinalBalance = equityCurve[len(equityCurve)-1].Value
	}

	// Calculate metrics
	report.calculateTradeStats()
	report.calculatePerformanceMetrics()
	report.calculateRiskMetrics()
	report.calculateMonthlyReturns()
	report.calculateDrawdownCurve()

	return report
}

func (r *BacktestReport) calculateTradeStats() {
	if len(r.Trades) == 0 {
		return
	}

	r.TotalTrades = len(r.Trades)
	totalWin := 0.0
	totalLoss := 0.0
	totalHoldTime := time.Duration(0)

	for _, t := range r.Trades {
		r.TotalTrades++
		if t.PnL > 0 {
			r.WinningTrades++
			totalWin += t.PnL
			if t.PnL > r.LargestWin {
				r.LargestWin = t.PnL
			}
		} else {
			r.LosingTrades++
			totalLoss += math.Abs(t.PnL)
			if math.Abs(t.PnL) > r.LargestLoss {
				r.LargestLoss = math.Abs(t.PnL)
			}
		}
		totalHoldTime += t.Duration
	}

	r.WinRate = float64(r.WinningTrades) / float64(r.TotalTrades) * 100

	if r.WinningTrades > 0 {
		r.AvgWin = totalWin / float64(r.WinningTrades)
	}
	if r.LosingTrades > 0 {
		r.AvgLoss = totalLoss / float64(r.LosingTrades)
	}
	if totalLoss > 0 {
		r.ProfitFactor = totalWin / totalLoss
	}
	if r.TotalTrades > 0 {
		r.AvgHoldTime = totalHoldTime / time.Duration(r.TotalTrades)
	}
}

func (r *BacktestReport) calculatePerformanceMetrics() {
	if r.InitialBalance <= 0 {
		return
	}

	r.TotalReturn = (r.FinalBalance - r.InitialBalance) / r.InitialBalance * 100

	// Annualized return
	if r.Period.End.After(r.Period.Start) {
		years := r.Period.End.Sub(r.Period.Start).Hours() / (365.25 * 24)
		if years > 0 {
			r.AnnualReturn = (math.Pow(1+r.TotalReturn/100, 1/years) - 1) * 100
		}
	}

	// Sharpe Ratio (assuming 0% risk-free rate)
	if len(r.EquityCurve) > 1 {
		returns := make([]float64, 0, len(r.EquityCurve)-1)
		for i := 1; i < len(r.EquityCurve); i++ {
			if r.EquityCurve[i-1].Value > 0 {
				ret := (r.EquityCurve[i].Value - r.EquityCurve[i-1].Value) / r.EquityCurve[i-1].Value
				returns = append(returns, ret)
			}
		}

		if len(returns) > 0 {
			mean := mean(returns)
			std := stddev(returns)
			if std > 0 {
				r.SharpeRatio = mean / std * math.Sqrt(252) // Annualized
			}

			// Sortino Ratio (downside deviation)
			downsideReturns := make([]float64, 0)
			for _, ret := range returns {
				if ret < 0 {
					downsideReturns = append(downsideReturns, ret)
				}
			}
			if len(downsideReturns) > 0 {
				r.DownsideDev = stddev(downsideReturns)
				if r.DownsideDev > 0 {
					r.SortinoRatio = mean / r.DownsideDev * math.Sqrt(252)
				}
			}
		}
	}

	// Calmar Ratio
	if r.MaxDrawdown > 0 {
		r.CalmarRatio = r.AnnualReturn / r.MaxDrawdown
	}
}

func (r *BacktestReport) calculateRiskMetrics() {
	if len(r.EquityCurve) < 2 {
		return
	}

	returns := make([]float64, 0, len(r.EquityCurve)-1)
	for i := 1; i < len(r.EquityCurve); i++ {
		if r.EquityCurve[i-1].Value > 0 {
			ret := (r.EquityCurve[i].Value - r.EquityCurve[i-1].Value) / r.EquityCurve[i-1].Value
			returns = append(returns, ret)
		}
	}

	if len(returns) > 0 {
		r.Volatility = stddev(returns) * math.Sqrt(252) // Annualized
	}
}

func (r *BacktestReport) calculateDrawdownCurve() {
	if len(r.EquityCurve) == 0 {
		return
	}

	peak := r.EquityCurve[0].Value
	r.DrawdownCurve = make([]EquityPoint, len(r.EquityCurve))

	for i, point := range r.EquityCurve {
		if point.Value > peak {
			peak = point.Value
		}
		drawdown := 0.0
		if peak > 0 {
			drawdown = (peak - point.Value) / peak * 100
		}
		if drawdown > r.MaxDrawdown {
			r.MaxDrawdown = drawdown
		}
		r.DrawdownCurve[i] = EquityPoint{
			Timestamp: point.Timestamp,
			Value:     point.Value,
			Drawdown:  drawdown,
		}
	}
}

func (r *BacktestReport) calculateMonthlyReturns() {
	if len(r.EquityCurve) < 2 {
		return
	}

	// Group by month
	monthStart := make(map[string]float64)
	monthEnd := make(map[string]float64)

	for _, point := range r.EquityCurve {
		month := point.Timestamp.Format("2006-01")
		if _, ok := monthStart[month]; !ok {
			monthStart[month] = point.Value
		}
		monthEnd[month] = point.Value
	}

	for month, startVal := range monthStart {
		endVal := monthEnd[month]
		if startVal > 0 {
			r.MonthlyReturns[month] = (endVal - startVal) / startVal * 100
		}
	}
}

// GenerateJSON generates a JSON report.
func (g *ReportGenerator) GenerateJSON(report *BacktestReport) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}

// GenerateCSV generates a CSV of trades.
func (g *ReportGenerator) GenerateCSV(report *BacktestReport) ([]byte, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// Header
	writer.Write([]string{
		"Entry Time", "Exit Time", "Side", "Entry Price", "Exit Price",
		"Quantity", "PnL", "PnL %", "Fees", "Duration",
	})

	// Trades
	for _, t := range report.Trades {
		writer.Write([]string{
			t.EntryTime.Format(time.RFC3339),
			t.ExitTime.Format(time.RFC3339),
			t.Side,
			fmt.Sprintf("%.8f", t.EntryPrice),
			fmt.Sprintf("%.8f", t.ExitPrice),
			fmt.Sprintf("%.8f", t.Quantity),
			fmt.Sprintf("%.2f", t.PnL),
			fmt.Sprintf("%.2f", t.PnLPct),
			fmt.Sprintf("%.2f", t.Fees),
			t.Duration.String(),
		})
	}

	writer.Flush()
	return []byte(buf.String()), nil
}

// GenerateSummary generates a text summary.
func (g *ReportGenerator) GenerateSummary(report *BacktestReport) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("=== Backtest Report: %s ===\n", report.ID))
	sb.WriteString(fmt.Sprintf("Strategy: %s\n", report.Strategy))
	sb.WriteString(fmt.Sprintf("Symbol: %s\n", report.Symbol))
	sb.WriteString(fmt.Sprintf("Period: %s to %s\n", report.Period.Start.Format("2006-01-02"), report.Period.End.Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("\n--- Performance ---\n"))
	sb.WriteString(fmt.Sprintf("Total Return: %.2f%%\n", report.TotalReturn))
	sb.WriteString(fmt.Sprintf("Annual Return: %.2f%%\n", report.AnnualReturn))
	sb.WriteString(fmt.Sprintf("Sharpe Ratio: %.2f\n", report.SharpeRatio))
	sb.WriteString(fmt.Sprintf("Sortino Ratio: %.2f\n", report.SortinoRatio))
	sb.WriteString(fmt.Sprintf("Calmar Ratio: %.2f\n", report.CalmarRatio))
	sb.WriteString(fmt.Sprintf("Max Drawdown: %.2f%%\n", report.MaxDrawdown))
	sb.WriteString(fmt.Sprintf("\n--- Trades ---\n"))
	sb.WriteString(fmt.Sprintf("Total Trades: %d\n", report.TotalTrades))
	sb.WriteString(fmt.Sprintf("Win Rate: %.1f%%\n", report.WinRate))
	sb.WriteString(fmt.Sprintf("Profit Factor: %.2f\n", report.ProfitFactor))
	sb.WriteString(fmt.Sprintf("Avg Win: $%.2f\n", report.AvgWin))
	sb.WriteString(fmt.Sprintf("Avg Loss: $%.2f\n", report.AvgLoss))
	sb.WriteString(fmt.Sprintf("Largest Win: $%.2f\n", report.LargestWin))
	sb.WriteString(fmt.Sprintf("Largest Loss: $%.2f\n", report.LargestLoss))
	sb.WriteString(fmt.Sprintf("Avg Hold Time: %v\n", report.AvgHoldTime))
	sb.WriteString(fmt.Sprintf("\n--- Risk ---\n"))
	sb.WriteString(fmt.Sprintf("Volatility: %.2f%%\n", report.Volatility*100))
	sb.WriteString(fmt.Sprintf("Downside Dev: %.2f%%\n", report.DownsideDev*100))

	return sb.String()
}

// PerformanceComparator compares multiple backtest reports.
type PerformanceComparator struct {
	reports []*BacktestReport
}

func NewPerformanceComparator(reports []*BacktestReport) *PerformanceComparator {
	return &PerformanceComparator{reports: reports}
}

// ComparisonResult contains comparison results.
type ComparisonResult struct {
	BestStrategy  string
	WorstStrategy string
	Metrics       map[string]map[string]float64
	Rankings      map[string][]string
}

// Compare compares all reports.
func (c *PerformanceComparator) Compare() *ComparisonResult {
	if len(c.reports) == 0 {
		return nil
	}

	result := &ComparisonResult{
		Metrics:  make(map[string]map[string]float64),
		Rankings: make(map[string][]string),
	}

	// Collect metrics
	for _, r := range c.reports {
		result.Metrics[r.Strategy] = map[string]float64{
			"total_return":  r.TotalReturn,
			"sharpe_ratio":  r.SharpeRatio,
			"max_drawdown":  r.MaxDrawdown,
			"win_rate":      r.WinRate,
			"profit_factor": r.ProfitFactor,
		}
	}

	// Rank by each metric
	metrics := []string{"total_return", "sharpe_ratio", "win_rate", "profit_factor"}
	for _, metric := range metrics {
		type pair struct {
			strategy string
			value    float64
		}
		pairs := make([]pair, 0, len(c.reports))
		for _, r := range c.reports {
			pairs = append(pairs, pair{r.Strategy, result.Metrics[r.Strategy][metric]})
		}
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].value > pairs[j].value
		})
		ranked := make([]string, len(pairs))
		for i, p := range pairs {
			ranked[i] = p.strategy
		}
		result.Rankings[metric] = ranked
	}

	// Best by Sharpe
	if len(c.reports) > 0 {
		best := c.reports[0]
		worst := c.reports[0]
		for _, r := range c.reports[1:] {
			if r.SharpeRatio > best.SharpeRatio {
				best = r
			}
			if r.SharpeRatio < worst.SharpeRatio {
				worst = r
			}
		}
		result.BestStrategy = best.Strategy
		result.WorstStrategy = worst.Strategy
	}

	return result
}

// Helper functions
func mean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

func stddev(data []float64) float64 {
	if len(data) < 2 {
		return 0
	}
	m := mean(data)
	sum := 0.0
	for _, v := range data {
		diff := v - m
		sum += diff * diff
	}
	return math.Sqrt(sum / float64(len(data)-1))
}
