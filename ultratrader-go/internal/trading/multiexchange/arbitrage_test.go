package multiexchange_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/multiexchange"
)

func TestArbitrageExecutor_Scan(t *testing.T) {
	manager := &mockExchangeManager{}
	fees := map[string]float64{
		"binance":  0.1,
		"coinbase": 0.5,
		"kucoin":   0.1,
	}

	arb := multiexchange.NewArbitrageExecutor(manager, fees, 0.05)

	// Binance ask 65010, Kucoin bid 65020 (diff is 10)
	// Coinbase bid 64900
	// 65020 vs 65010 -> spread is (10/65010) = 0.015%
	// Not enough to cover 0.2% fees, so this will find 0 valid opportunities
	opps := arb.Scan([]string{"BTC"})
	assert.Empty(t, opps)
}
