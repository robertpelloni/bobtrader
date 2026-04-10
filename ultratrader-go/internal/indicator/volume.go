package indicator

import "math"

// CandleInput provides OHLCV data for volume-based indicators.
type CandleInput struct {
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

// VWAP (Volume-Weighted Average Price) tracks the average price weighted by volume.
// Typically reset daily. Provides a benchmark for institutional order execution.
type VWAP struct {
	cumVolume float64
	cumTPVol  float64 // cumulative (typical price * volume)
}

// NewVWAP creates a new VWAP calculator.
func NewVWAP() *VWAP {
	return &VWAP{}
}

// Update adds a new candle and returns the current VWAP.
func (v *VWAP) Update(candle CandleInput) float64 {
	// Typical price = (High + Low + Close) / 3
	tp := (candle.High + candle.Low + candle.Close) / 3
	v.cumVolume += candle.Volume
	v.cumTPVol += tp * candle.Volume

	if v.cumVolume == 0 {
		return 0
	}
	return v.cumTPVol / v.cumVolume
}

// Last returns the current VWAP value.
func (v *VWAP) Last() float64 {
	if v.cumVolume == 0 {
		return 0
	}
	return v.cumTPVol / v.cumVolume
}

// Reset clears the VWAP for a new session (e.g., daily reset).
func (v *VWAP) Reset() {
	v.cumVolume = 0
	v.cumTPVol = 0
}

// OBV (On Balance Volume) tracks cumulative volume flow.
// Rising OBV = buying pressure, falling OBV = selling pressure.
type OBV struct {
	obv     float64
	lastClose float64
	started bool
}

// NewOBV creates a new OBV calculator.
func NewOBV() *OBV {
	return &OBV{}
}

// Update adds a new candle and returns the current OBV.
func (o *OBV) Update(candle CandleInput) float64 {
	if !o.started {
		o.obv = candle.Volume
		o.lastClose = candle.Close
		o.started = true
		return o.obv
	}

	if candle.Close > o.lastClose {
		o.obv += candle.Volume
	} else if candle.Close < o.lastClose {
		o.obv -= candle.Volume
	}
	// If close == lastClose, OBV unchanged

	o.lastClose = candle.Close
	return o.obv
}

// Last returns the current OBV value.
func (o *OBV) Last() float64 {
	return o.obv
}

// VolumeSMA tracks the simple moving average of volume.
type VolumeSMA struct {
	period  int
	volumes []float64
}

// NewVolumeSMA creates a volume SMA.
func NewVolumeSMA(period int) *VolumeSMA {
	return &VolumeSMA{period: period}
}

// Update adds a volume reading and returns the average.
func (v *VolumeSMA) Update(volume float64) float64 {
	v.volumes = append(v.volumes, volume)
	if len(v.volumes) > v.period {
		v.volumes = v.volumes[len(v.volumes)-v.period:]
	}
	if len(v.volumes) < v.period {
		return 0
	}
	var sum float64
	for _, vol := range v.volumes {
		sum += vol
	}
	return sum / float64(v.period)
}

func (v *VolumeSMA) Last() float64 {
	if len(v.volumes) < v.period {
		return 0
	}
	var sum float64
	for _, vol := range v.volumes {
		sum += vol
	}
	return sum / float64(len(v.volumes))
}

// VolumeRatio computes the ratio of current volume to average volume.
// Values > 1 indicate above-average volume, < 1 indicate below-average.
type VolumeRatio struct {
	sma *VolumeSMA
}

// NewVolumeRatio creates a volume ratio calculator.
func NewVolumeRatio(period int) *VolumeRatio {
	return &VolumeRatio{sma: NewVolumeSMA(period)}
}

// Update returns the ratio of current volume to the moving average.
func (vr *VolumeRatio) Update(volume float64) float64 {
	avg := vr.sma.Update(volume)
	if avg == 0 {
		return 1.0
	}
	return volume / avg
}

// MFI (Money Flow Index) is a volume-weighted RSI.
// Range: 0 to 100. Above 80 = overbought, below 20 = oversold.
type MFI struct {
	period    int
	posMF     []float64
	negMF     []float64
	lastTP    float64
	started   bool
}

// NewMFI creates a Money Flow Index calculator.
func NewMFI(period int) *MFI {
	return &MFI{period: period}
}

// Update adds a candle and returns the current MFI.
func (m *MFI) Update(candle CandleInput) float64 {
	tp := (candle.High + candle.Low + candle.Close) / 3
	mf := tp * candle.Volume // raw money flow

	if !m.started {
		m.lastTP = tp
		m.started = true
		return 50.0 // Neutral until enough data
	}

	if tp > m.lastTP {
		m.posMF = append(m.posMF, mf)
		m.negMF = append(m.negMF, 0)
	} else if tp < m.lastTP {
		m.posMF = append(m.posMF, 0)
		m.negMF = append(m.negMF, mf)
	} else {
		m.posMF = append(m.posMF, 0)
		m.negMF = append(m.negMF, 0)
	}

	m.lastTP = tp

	// Trim to period
	if len(m.posMF) > m.period {
		m.posMF = m.posMF[len(m.posMF)-m.period:]
		m.negMF = m.negMF[len(m.negMF)-m.period:]
	}

	if len(m.posMF) < m.period {
		return 50.0
	}

	var posSum, negSum float64
	for _, v := range m.posMF {
		posSum += v
	}
	for _, v := range m.negMF {
		negSum += v
	}

	if negSum == 0 {
		return 100.0
	}

	mfr := posSum / negSum
	return 100.0 - (100.0 / (1.0 + mfr))
}

// Last returns the current MFI value.
func (m *MFI) Last() float64 {
	if len(m.posMF) < m.period {
		return 50.0
	}
	var posSum, negSum float64
	for _, v := range m.posMF {
		posSum += v
	}
	for _, v := range m.negMF {
		negSum += v
	}
	if negSum == 0 {
		return 100.0
	}
	mfr := posSum / negSum
	return 100.0 - (100.0 / (1.0 + mfr))
}

// ChaikinMoneyFlow measures accumulation/distribution over a period.
// Positive CMF = buying pressure, negative = selling pressure.
type ChaikinMoneyFlow struct {
	period   int
	mfvs     []float64 // Money Flow Volume = MF Multiplier * Volume
	volumes  []float64
}

// NewChaikinMoneyFlow creates a CMF calculator.
func NewChaikinMoneyFlow(period int) *ChaikinMoneyFlow {
	return &ChaikinMoneyFlow{period: period}
}

// Update adds a candle and returns the current CMF.
func (c *ChaikinMoneyFlow) Update(candle CandleInput) float64 {
	// Money Flow Multiplier = ((Close - Low) - (High - Close)) / (High - Low)
	var mfm float64
	hl := candle.High - candle.Low
	if hl != 0 {
		mfm = ((candle.Close - candle.Low) - (candle.High - candle.Close)) / hl
	}

	mfv := mfm * candle.Volume
	c.mfvs = append(c.mfvs, mfv)
	c.volumes = append(c.volumes, candle.Volume)

	if len(c.mfvs) > c.period {
		c.mfvs = c.mfvs[len(c.mfvs)-c.period:]
		c.volumes = c.volumes[len(c.volumes)-c.period:]
	}

	if len(c.mfvs) < c.period {
		return 0
	}

	var sumMFV, sumVol float64
	for i := range c.mfvs {
		sumMFV += c.mfvs[i]
		sumVol += c.volumes[i]
	}

	if sumVol == 0 {
		return 0
	}
	return sumMFV / sumVol
}

// Last returns the current CMF value.
func (c *ChaikinMoneyFlow) Last() float64 {
	if len(c.mfvs) < c.period {
		return 0
	}
	var sumMFV, sumVol float64
	for i := range c.mfvs {
		sumMFV += c.mfvs[i]
		sumVol += c.volumes[i]
	}
	if sumVol == 0 {
		return 0
	}
	return sumMFV / sumVol
}

// sqrt helper for volume indicators
func vsqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 20; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// Suppress unused import warning
var _ = math.Pi
var _ = vsqrt
