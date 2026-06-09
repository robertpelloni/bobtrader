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
	cacheTTL time.Duration

	mu        sync.Mutex
	cachedBal float64
	cachedAt  time.Time
	quote     string // usually "USDT"
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

// USDTBalance returns the available USDT balance from the Binance
// account, using a cached value when fresh.
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
		// Return stale cached value on error — better than returning 0
		return r.cachedBal
	}

	for _, b := range balances {
		if b.Asset == r.quote {
			free, _ := strconv.ParseFloat(b.Free, 64)
			r.cachedBal = free
			r.cachedAt = time.Now()
			return free
		}
	}

	return r.cachedBal
}

// SetQuoteAsset changes the quote asset to look up (default: "USDT").
func (r *BinanceBalanceReader) SetQuoteAsset(asset string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.quote = asset
	r.cachedAt = time.Time{} // force refresh on next read
}
