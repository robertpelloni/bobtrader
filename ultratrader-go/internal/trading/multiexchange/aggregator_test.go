package multiexchange_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/multiexchange"
)

type mockExchangeManagerV2 struct {
	mockExchangeManager
}

func (m *mockExchangeManagerV2) GetOrderBook(symbol, exchange string, depth int) (multiexchange.OrderBook, error) {
	if symbol == "BTC" {
		switch exchange {
		case "binance":
			return multiexchange.OrderBook{
				Bids: [][2]float64{{65000, 1.0}, {64990, 2.0}},
				Asks: [][2]float64{{65010, 1.5}, {65020, 3.0}},
			}, nil
		case "coinbase":
			return multiexchange.OrderBook{
				Bids: [][2]float64{{64900, 5.0}},
				Asks: [][2]float64{{65050, 2.0}},
			}, nil
		}
	}
	return multiexchange.OrderBook{}, fmt.Errorf("unknown")
}

func (m *mockExchangeManagerV2) GetExchanges() []string {
	return []string{"binance", "coinbase"}
}

func TestLiquidityAggregator_SplitOrder(t *testing.T) {
	manager := &mockExchangeManagerV2{}
	fees := map[string]float64{
		"binance":  0.1,
		"coinbase": 0.5,
	}

	agg := multiexchange.NewLiquidityAggregator(manager, 0.1, fees)

	legs, err := agg.SplitOrder("BTC", "buy", 2.0)
	assert.NoError(t, err)
	assert.Len(t, legs, 2)

	// Binance ask sum: (65010*1.5)+(65020*3.0) = 97515 + 195060 = 292575
	// Coinbase ask sum: 65050*2.0 = 130100
	// Total liq = 422675
	// Binance share = 292575 / 422675 = ~0.692
	// Coinbase share = 130100 / 422675 = ~0.308

	for _, leg := range legs {
		if leg.Exchange == "binance" {
			assert.InDelta(t, 2.0 * 0.692, leg.Quantity, 0.01)
		} else if leg.Exchange == "coinbase" {
			assert.InDelta(t, 2.0 * 0.308, leg.Quantity, 0.01)
		} else {
			t.Errorf("unexpected exchange %s", leg.Exchange)
		}
	}
}
