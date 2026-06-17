package multiexchange_test

import (
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/multiexchange"
)

type mockExchangeManager struct{}

func (m *mockExchangeManager) GetTicker(coin, exchange string) (multiexchange.Ticker, error) {
	if coin == "BTC" {
		switch exchange {
		case "binance":
			return multiexchange.Ticker{Bid: 65000, Ask: 65010}, nil
		case "kucoin":
			return multiexchange.Ticker{Bid: 65020, Ask: 65030}, nil
		case "coinbase":
			return multiexchange.Ticker{Bid: 64900, Ask: 65050}, nil
		}
	}
	return multiexchange.Ticker{}, nil
}

func (m *mockExchangeManager) GetExchanges() []string {
	return []string{"binance", "kucoin", "coinbase"}
}
