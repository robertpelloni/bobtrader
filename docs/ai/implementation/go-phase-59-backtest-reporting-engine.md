# Go Phase 59: Backtest Reporting Engine

## Summary
Comprehensive backtest reporting with HTML/PDF output, performance metrics, trade analysis, and visual charts.

## Current State
- `internal/backtest/engine.go` — Basic backtest execution
- `internal/backtest/optimizer/` — Walk-forward, grid search, Monte Carlo
- No dedicated reporting module

## New Module: `internal/backtest/reporting/`

### 1. BacktestReport
Comprehensive report structure.

```go
type BacktestReport struct {
    ID            string
    Strategy      string
    Symbol        string
    Period        TimePeriod
    InitialBalance float64
    FinalBalance   float64
    
    // Performance Metrics
    TotalReturn    float64
    AnnualReturn   float64
    SharpeRatio    float64
    SortinoRatio   float64
    CalmarRatio    float64
    MaxDrawdown    float64
    
    // Trade Statistics
    TotalTrades    int
    WinningTrades  int
    LosingTrades   int
    WinRate        float64
    ProfitFactor   float64
    AvgWin         float64
    AvgLoss        float64
    LargestWin     float64
    LargestLoss    float64
    AvgHoldTime    time.Duration
    
    // Risk Metrics
    Volatility     float64
    DownsideDev    float64
    Beta           float64
    Alpha          float64
    InformationRatio float64
    
    // Trade Details
    Trades         []TradeRecord
    EquityCurve    []EquityPoint
    DrawdownCurve  []DrawdownPoint
    MonthlyReturns map[string]float64
    
    // Metadata
    GeneratedAt    time.Time
    Config         map[string]interface{}
}

type TradeRecord struct {
    EntryTime   time.Time
    ExitTime    time.Time
    Side        string
    EntryPrice  float64
    ExitPrice   float64
    Quantity    float64
    PnL         float64
    PnLPct      float64
    Fees        float64
    Duration    time.Duration
}

type EquityPoint struct {
    Timestamp time.Time
    Value     float64
    Drawdown  float64
}
```

### 2. ReportGenerator
Generates reports from backtest results.

```go
type ReportGenerator struct {
    templateDir string
}

func (g *ReportGenerator) Generate(engine *BacktestEngine, config ReportConfig) (*BacktestReport, error)
func (g *ReportGenerator) GenerateHTML(report *BacktestReport) ([]byte, error)
func (g *ReportGenerator) GenerateJSON(report *BacktestReport) ([]byte, error)
func (g *ReportGenerator) GenerateCSV(report *BacktestReport) ([]byte, error)
func (g *ReportGenerator) SaveReport(report *BacktestReport, format string, path string) error
```

### 3. PerformanceComparator
Compare multiple backtest runs.

```go
type PerformanceComparator struct {
    reports []*BacktestReport
}

type ComparisonResult struct {
    BestStrategy    string
    WorstStrategy   string
    Metrics         map[string]map[string]float64 // strategy -> metric -> value
    Rankings        map[string][]string            // metric -> ranked strategies
}

func (c *PerformanceComparator) Compare() *ComparisonResult
func (c *PerformanceComparator) GetBestByMetric(metric string) *BacktestReport
func (c *PerformanceComparator) GetRiskAdjustedRanking() []string
```

### 4. WalkForwardReport
Specialized reporting for walk-forward optimization.

```go
type WalkForwardReport struct {
    InSampleResults  []*BacktestReport
    OutOfSampleResults []*BacktestReport
    ParameterStability map[string][]float64
    OverfittingScore   float64
    RobustnessScore    float64
}

func (r *WalkForwardReport) GenerateSummary() string
func (r *WalkForwardReport) GetOptimalParameters() map[string]interface{}
func (r *WalkForwardReport) IsOverfit() bool
```

## Tests
- `TestReportGenerator_Generate`
- `TestReportGenerator_HTML`
- `TestReportGenerator_JSON`
- `TestReportGenerator_CSV`
- `TestPerformanceComparator_Compare`
- `TestPerformanceComparator_Rankings`
- `TestWalkForwardReport_Summary`
- `TestWalkForwardReport_Overfitting`
