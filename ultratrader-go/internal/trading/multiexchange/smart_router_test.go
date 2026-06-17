package multiexchange_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/multiexchange"
)

type mockExchangeManager struct{}

func (m *mockExchangeManager) GetTicker(coin, exchange string) (multiexchange.Ticker, error) {
	if coin == "BTC" {
		switch exchange {
		case "binance":
			return multiexchange.Ticker{Bid: 65000, Ask: 65010}, nil
		case "coinbase":
			return multiexchange.Ticker{Bid: 64900, Ask: 65050}, nil
		case "kucoin":
			return multiexchange.Ticker{Bid: 65020, Ask: 65040}, nil
		}
	}
	return multiexchange.Ticker{}, fmt.Errorf("unknown coin/exchange")
}

func (m *mockExchangeManager) GetExchanges() []string {
	return []string{"binance", "coinbase", "kucoin"}
}

func TestSmartRouter_CompareRoutes(t *testing.T) {
	manager := &mockExchangeManager{}
	fees := map[string]float64{
		"binance":  0.1,
		"coinbase": 0.5,
		"kucoin":   0.1,
	}

	router := multiexchange.NewSmartRouter(manager, fees)

	// Test BUY (looks for lowest Ask)
	buyRoutes, err := router.CompareRoutes("BTC", "buy", 1.0)
	assert.NoError(t, err)
	assert.Len(t, buyRoutes, 3)

	// Binance Ask is 65010, fee is 0.1% -> 65010 * 1.001 = 65075.01
	// Kucoin Ask is 65040, fee is 0.1% -> 65040 * 1.001 = 65105.04
	// Coinbase Ask is 65050, fee is 0.5% -> 65050 * 1.005 = 65375.25
	assert.Equal(t, "binance", buyRoutes[0].Exchange)
	assert.Equal(t, "kucoin", buyRoutes[1].Exchange)
	assert.Equal(t, "coinbase", buyRoutes[2].Exchange)

	// Test SELL (looks for highest Bid)
	sellRoutes, err := router.CompareRoutes("BTC", "sell", 1.0)
	assert.NoError(t, err)
	assert.Len(t, sellRoutes, 3)

	// Kucoin Bid is 65020, fee is 0.1% -> 65020 * 0.999 = 64954.98
	// Binance Bid is 65000, fee is 0.1% -> 65000 * 0.999 = 64935.00
	// Coinbase Bid is 64900, fee is 0.5% -> 64900 * 0.995 = 64575.50
	assert.Equal(t, "kucoin", sellRoutes[0].Exchange)
	assert.Equal(t, "binance", sellRoutes[1].Exchange)
	assert.Equal(t, "coinbase", sellRoutes[2].Exchange)
}
