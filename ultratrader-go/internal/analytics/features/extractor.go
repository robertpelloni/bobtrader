package features

import (
	"math"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/indicator"
)

// FeatureMap is a named map of feature values.
type FeatureMap map[string]float64

// CandleData provides OHLCV data for feature extraction.
type CandleData struct {
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

// Extractor computes features from candle data.
type Extractor struct {
	priceWindow int
	volWindow   int

	// Internal state
	sma     *indicator.SMA
	ema     *indicator.EMA
	rsi     *indicator.RSI
	volSMA  *indicator.VolumeSMA
	closes  []float64
	volumes []float64
}

// NewExtractor creates a feature extractor with given lookback windows.
func NewExtractor(priceWindow, volWindow int) *Extractor {
	return &Extractor{
		priceWindow: priceWindow,
		volWindow:   volWindow,
		sma:        indicator.NewSMA(priceWindow),
		ema:        indicator.NewEMA(priceWindow),
		rsi:        indicator.NewRSI(14),
		volSMA:    indicator.NewVolumeSMA(volWindow),
	}
}

// Update adds a candle and returns the computed features.
func (e *Extractor) Update(candle CandleData) FeatureMap {
	features := make(FeatureMap)

	close := candle.Close
	e.sma.Update(close)
	e.ema.Update(close)
	rsiVal := e.rsi.Update(close)
	e.volSMA.Update(candle.Volume)

	e.closes = append(e.closes, close)
	e.volumes = append(e.volumes, candle.Volume)

	// Trim to prevent unbounded growth
	maxLen := e.priceWindow * 3
	if len(e.closes) > maxLen {
		e.closes = e.closes[len(e.closes)-maxLen:]
		e.volumes = e.volumes[len(e.volumes)-maxLen:]
	}

	// Price features
	features["close"] = close

	if sma := e.sma.Last(); sma > 0 {
		features["sma"] = sma
		features["close_sma_ratio"] = close / sma
	}
	if ema := e.ema.Last(); ema > 0 {
		features["ema"] = ema
		features["close_ema_ratio"] = close / ema
	}

	// RSI
	features["rsi"] = rsiVal

	// Returns
	if len(e.closes) >= 2 {
		prevClose := e.closes[len(e.closes)-2]
		if prevClose > 0 {
			features["return_1"] = (close - prevClose) / prevClose
		}
	}
	if len(e.closes) >= 5 {
		prevClose5 := e.closes[len(e.closes)-5]
		if prevClose5 > 0 {
			features["return_5"] = (close - prevClose5) / prevClose5
		}
	}
	if len(e.closes) >= 10 {
		prevClose10 := e.closes[len(e.closes)-10]
		if prevClose10 > 0 {
			features["return_10"] = (close - prevClose10) / prevClose10
		}
	}

	// Volatility (rolling std of returns)
	if len(e.closes) >= e.priceWindow {
		returns := make([]float64, 0, e.priceWindow-1)
		start := len(e.closes) - e.priceWindow
		for i := start + 1; i < len(e.closes); i++ {
			if e.closes[i-1] > 0 {
				returns = append(returns, (e.closes[i]-e.closes[i-1])/e.closes[i-1])
			}
		}
		features["volatility"] = stdDev(returns)
	}

	// Volume features
	features["volume"] = candle.Volume
	if avgVol := e.volSMA.Last(); avgVol > 0 {
		features["volume_ratio"] = candle.Volume / avgVol
	}

	// Intraday range
	if candle.High > 0 && candle.Low > 0 {
		features["range_pct"] = (candle.High - candle.Low) / candle.Close
		features["body_pct"] = math.Abs(candle.Close-candle.Open) / candle.Close
	}

	// Upper/lower shadow
	if candle.High > candle.Low {
		features["upper_shadow"] = (candle.High - math.Max(candle.Open, candle.Close)) / (candle.High - candle.Low)
		features["lower_shadow"] = (math.Min(candle.Open, candle.Close) - candle.Low) / (candle.High - candle.Low)
	}

	return features
}

// Names returns the list of feature names that this extractor produces.
func (e *Extractor) Names() []string {
	return []string{
		"close", "sma", "ema", "close_sma_ratio", "close_ema_ratio",
		"rsi", "return_1", "return_5", "return_10", "volatility",
		"volume", "volume_ratio", "range_pct", "body_pct",
		"upper_shadow", "lower_shadow",
	}
}

// FeatureMatrix builds a matrix of features from historical candles.
// Each row is a timestamp, each column is a feature.
type FeatureMatrix struct {
	Headers []string
	Rows    [][]float64
}

// BuildMatrix extracts features from a series of candles.
func BuildMatrix(candles []CandleData, priceWindow, volWindow int) *FeatureMatrix {
	ext := NewExtractor(priceWindow, volWindow)
	headers := ext.Names()

	var rows [][]float64
	for _, c := range candles {
		features := ext.Update(c)
		row := make([]float64, len(headers))
		for i, name := range headers {
			row[i] = features[name]
		}
		rows = append(rows, row)
	}

	return &FeatureMatrix{
		Headers: headers,
		Rows:    rows,
	}
}

func stdDev(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	var sumSqDiff float64
	for _, v := range values {
		diff := v - mean
		sumSqDiff += diff * diff
	}
	variance := sumSqDiff / float64(len(values)-1)
	return math.Sqrt(variance)
}

// Ensure utils.ParseFloat is available
var _ = utils.ParseFloat
