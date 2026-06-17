package rebalancer

import (
	"testing"
	"time"
)

func TestRebalancer_WashSale(t *testing.T) {
	alloc := []Allocation{{Symbol: "BTC", Weight: 0.5}, {Symbol: "USDT", Weight: 0.5}}
	r := New(alloc, 0.01)
	r.SetWashSaleWindow(time.Hour)

	holdings := []Holding{{Symbol: "BTC", Quantity: 1.0, Value: 60.0}, {Symbol: "USDT", Quantity: 40.0, Value: 40.0}}

	// Drift: BTC is 60% (target 50%), USDT is 40% (target 50%).
	// Rebalance should want to sell BTC.
	res := r.Compute(holdings)
	if !res.NeedsRebalance {
		t.Fatal("expected rebalance needed")
	}

	// Record a trade for BTC
	r.RecordTrade("BTC")

	// Now recompute — BTC should be skipped due to wash-sale window
	res2 := r.Compute(holdings)
	foundBTC := false
	for _, o := range res2.Orders {
		if o.Symbol == "BTC" {
			foundBTC = true
		}
	}

	if foundBTC {
		t.Errorf("expected BTC to be skipped due to wash-sale window")
	}
}
