package marketdata

import "time"

type Tick struct {
	Symbol    string
	Price     string
	Source    string
	Timestamp time.Time
}

type Candle struct {
	Symbol    string
	Interval  string
	Open      string
	High      string
	Low       string
	Close     string
	Volume    string
	Timestamp time.Time
}
