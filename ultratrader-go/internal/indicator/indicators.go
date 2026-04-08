package indicator

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
