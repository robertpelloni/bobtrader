package indicator

import "math"

// SMA (Simple Moving Average) calculates the average of the last n values.
type SMA struct {
	period int
	values []float64
}

func NewSMA(period int) *SMA {
	return &SMA{period: period}
}

func (s *SMA) Update(value float64) float64 {
	s.values = append(s.values, value)
	if len(s.values) > s.period {
		s.values = s.values[len(s.values)-s.period:]
	}
	if len(s.values) < s.period {
		return 0
	}
	var sum float64
	for _, v := range s.values {
		sum += v
	}
	return sum / float64(s.period)
}

func (s *SMA) Last() float64 {
	if len(s.values) < s.period {
		return 0
	}
	var sum float64
	for _, v := range s.values {
		sum += v
	}
	return sum / float64(s.period)
}

// EMA (Exponential Moving Average) calculates the weighted average.
type EMA struct {
	period int
	value  float64
	k      float64
	init   bool
}

func NewEMA(period int) *EMA {
	return &EMA{
		period: period,
		k:      2.0 / float64(period+1),
	}
}

func (e *EMA) Update(value float64) float64 {
	if !e.init {
		e.value = value
		e.init = true
		return e.value
	}
	e.value = (value-e.value)*e.k + e.value
	return e.value
}

func (e *EMA) Last() float64 {
	return e.value
}

// RSI (Relative Strength Index) measures the speed and change of price movements.
type RSI struct {
	period  int
	prev    float64
	avgGain float64
	avgLoss float64
	init    bool
	count   int
}

func NewRSI(period int) *RSI {
	return &RSI{period: period}
}

func (r *RSI) Update(value float64) float64 {
	if !r.init {
		r.prev = value
		r.init = true
		return 50
	}

	change := value - r.prev
	r.prev = value

	var gain, loss float64
	if change > 0 {
		gain = change
	} else {
		loss = -change
	}

	if r.count < r.period {
		r.avgGain += gain
		r.avgLoss += loss
		r.count++
		if r.count == r.period {
			r.avgGain /= float64(r.period)
			r.avgLoss /= float64(r.period)
		}
		return 50
	}

	r.avgGain = (r.avgGain*float64(r.period-1) + gain) / float64(r.period)
	r.avgLoss = (r.avgLoss*float64(r.period-1) + loss) / float64(r.period)

	if r.avgLoss == 0 {
		return 100
	}

	rs := r.avgGain / r.avgLoss
	return 100 - (100 / (1 + rs))
}

func (r *RSI) Last() float64 {
	if r.count < r.period {
		return 50
	}
	if r.avgLoss == 0 {
		return 100
	}
	rs := r.avgGain / r.avgLoss
	return 100 - (100 / (1 + rs))
}

// MACD (Moving Average Convergence Divergence) tracks the relationship between
// two EMAs. It produces three values: the MACD line, the Signal line, and the
// Histogram (MACD - Signal).
type MACD struct {
	fast   *EMA
	slow   *EMA
	signal *EMA
}

type MACDResult struct {
	MACD      float64
	Signal    float64
	Histogram float64
}

func NewMACD(fastPeriod, slowPeriod, signalPeriod int) *MACD {
	return &MACD{
		fast:   NewEMA(fastPeriod),
		slow:   NewEMA(slowPeriod),
		signal: NewEMA(signalPeriod),
	}
}

func (m *MACD) Update(value float64) MACDResult {
	fastVal := m.fast.Update(value)
	slowVal := m.slow.Update(value)
	macdLine := fastVal - slowVal
	signalLine := m.signal.Update(macdLine)
	return MACDResult{
		MACD:      macdLine,
		Signal:    signalLine,
		Histogram: macdLine - signalLine,
	}
}

func (m *MACD) Last() MACDResult {
	macdLine := m.fast.Last() - m.slow.Last()
	signalLine := m.signal.Last()
	return MACDResult{
		MACD:      macdLine,
		Signal:    signalLine,
		Histogram: macdLine - signalLine,
	}
}

// BollingerBands measures volatility by plotting bands above and below a
// simple moving average. Default multiplier is 2.0 standard deviations.
type BollingerBands struct {
	period     int
	multiplier float64
	values     []float64
}

type BollingerBandsResult struct {
	Upper     float64
	Middle    float64
	Lower     float64
	Bandwidth float64 // (Upper - Lower) / Middle
}

func NewBollingerBands(period int, multiplier float64) *BollingerBands {
	return &BollingerBands{period: period, multiplier: multiplier}
}

func (b *BollingerBands) Update(value float64) BollingerBandsResult {
	b.values = append(b.values, value)
	if len(b.values) > b.period {
		b.values = b.values[len(b.values)-b.period:]
	}
	if len(b.values) < b.period {
		return BollingerBandsResult{}
	}

	// Calculate SMA (middle band)
	var sum float64
	for _, v := range b.values {
		sum += v
	}
	middle := sum / float64(b.period)

	// Calculate standard deviation
	var sqSum float64
	for _, v := range b.values {
		diff := v - middle
		sqSum += diff * diff
	}
	stdDev := math.Sqrt(sqSum / float64(b.period))

	upper := middle + b.multiplier*stdDev
	lower := middle - b.multiplier*stdDev

	bandwidth := 0.0
	if middle != 0 {
		bandwidth = (upper - lower) / middle
	}

	return BollingerBandsResult{
		Upper:     upper,
		Middle:    middle,
		Lower:     lower,
		Bandwidth: bandwidth,
	}
}

func (b *BollingerBands) Last() BollingerBandsResult {
	if len(b.values) < b.period {
		return BollingerBandsResult{}
	}

	var sum float64
	for _, v := range b.values {
		sum += v
	}
	middle := sum / float64(b.period)

	var sqSum float64
	for _, v := range b.values {
		diff := v - middle
		sqSum += diff * diff
	}
	stdDev := math.Sqrt(sqSum / float64(b.period))

	upper := middle + b.multiplier*stdDev
	lower := middle - b.multiplier*stdDev

	bandwidth := 0.0
	if middle != 0 {
		bandwidth = (upper - lower) / middle
	}

	return BollingerBandsResult{
		Upper:     upper,
		Middle:    middle,
		Lower:     lower,
		Bandwidth: bandwidth,
	}
}

// ATR (Average True Range) is a volatility indicator that uses the greatest
// of: current high - current low, |current high - previous close|,
// |current low - previous close|.
type ATR struct {
	period int
	prev   float64
	atr    float64
	count  int
	init   bool
}

func NewATR(period int) *ATR {
	return &ATR{period: period}
}

// Update accepts high, low, close values and returns the current ATR.
func (a *ATR) Update(high, low, close float64) float64 {
	if !a.init {
		a.prev = close
		a.init = true
		return 0
	}

	tr1 := high - low
	tr2 := math.Abs(high - a.prev)
	tr3 := math.Abs(low - a.prev)

	trueRange := tr1
	if tr2 > trueRange {
		trueRange = tr2
	}
	if tr3 > trueRange {
		trueRange = tr3
	}

	a.prev = close

	if a.count < a.period {
		a.atr += trueRange
		a.count++
		if a.count == a.period {
			a.atr /= float64(a.period)
			return a.atr
		}
		return a.atr / float64(a.count)
	} else {
		a.atr = (a.atr*float64(a.period-1) + trueRange) / float64(a.period)
		return a.atr
	}
}

func (a *ATR) Last() float64 {
	if a.count < a.period {
		if a.count == 0 {
			return 0
		}
		return a.atr / float64(a.count)
	}
	return a.atr
}
