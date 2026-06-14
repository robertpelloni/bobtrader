package demo

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

// BinanceBalanceReader implements BalanceReader by querying the real
// Binance spot account via the exchange adapter. Results are cached
// for cacheTTL to avoid hitting rate limits on every tick.
type BinanceBalanceReader struct {
	adapter  exchange.Adapter
	feed     PriceQuerier
	cacheTTL time.Duration

	mu        sync.Mutex
	cachedBal float64
	cachedAt  time.Time
	quote     string // usually "USDT"
}

// PriceQuerier can fetch a ticker price for a symbol.
type PriceQuerier interface {
	GetTickerPrice(ctx context.Context, symbol string) (string, error)
}

// NewBinanceBalanceReader creates a balance reader backed by a real
// Binance exchange adapter. It caches the result for the given TTL
// (recommended: 30–60s to stay within rate limits).
func NewBinanceBalanceReader(adapter exchange.Adapter, cacheTTL time.Duration) *BinanceBalanceReader {
	if cacheTTL <= 0 {
		cacheTTL = 30 * time.Second
	}
	return &BinanceBalanceReader{
		adapter:  adapter,
		cacheTTL: cacheTTL,
		quote:    "USDT",
	}
}

// SetPriceQuerier sets the price source for non-USDT asset valuation.
func (r *BinanceBalanceReader) SetPriceQuerier(q PriceQuerier) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.feed = q
	r.cachedAt = time.Time{} // force refresh
}

// USDTBalance returns the available USDT cash balance.
func (r *BinanceBalanceReader) USDTBalance() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	if time.Since(r.cachedAt) < r.cacheTTL {
		return r.cachedBal
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	balances, err := r.adapter.Balances(ctx)
	if err != nil {
		return r.cachedBal
	}

	totalUSDT := 0.0
	for _, b := range balances {
		free, _ := strconv.ParseFloat(b.Free, 64)
		if free <= 0 {
			continue
		}
		if b.Asset == "USDT" || b.Asset == "BUSD" || b.Asset == "USD" {
			totalUSDT += free
		}
	}

	r.cachedBal = totalUSDT
	r.cachedAt = time.Now()
	return totalUSDT
}

// SetQuoteAsset changes the quote asset to look up (default: "USDT").
func (r *BinanceBalanceReader) SetQuoteAsset(asset string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.quote = asset
	r.cachedAt = time.Time{} // force refresh on next read
}
