package risk

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
)

// CooldownGuard prevents rapid repeated trades on the same symbol.
// It tracks cooldowns separately for buys and sells:
//   - Buy cooldown: prevents buying the same symbol too quickly
//   - Sell cooldown: much shorter, prevents rapid sell-rebuy cycles
type CooldownGuard struct {
	BuyDuration  time.Duration
	SellDuration time.Duration
	mu           sync.Mutex
	lastBuy      map[string]time.Time
	lastSell     map[string]time.Time
}

func NewCooldownGuard(duration time.Duration) *CooldownGuard {
	return &CooldownGuard{
		BuyDuration:  duration,
		SellDuration: duration / 3, // sells cool down 3x faster than buys
		lastBuy:      map[string]time.Time{},
		lastSell:     map[string]time.Time{},
	}
}

func (g *CooldownGuard) Name() string {
	return "cooldown"
}

func (g *CooldownGuard) Check(_ context.Context, acct account.Account, intent OrderIntent) error {
	if g == nil {
		return nil
	}

	// Exit orders bypass cooldown — we always want to allow closing positions
	if intent.IsExit {
		return nil
	}

	key := acct.ID + ":" + strings.ToUpper(strings.TrimSpace(intent.Symbol))
	now := time.Now().UTC()

	g.mu.Lock()
	defer g.mu.Unlock()

	switch intent.Side {
	case SellSide:
		if g.SellDuration <= 0 {
			return nil
		}
		if last, ok := g.lastSell[key]; ok && now.Sub(last) < g.SellDuration {
			return fmt.Errorf("sell cooldown active for %s", key)
		}
		g.lastSell[key] = now
		// Also record as a general activity for buy cooldown
		g.lastBuy[key] = now
		return nil

	default: // buy side
		if g.BuyDuration <= 0 {
			return nil
		}
		if last, ok := g.lastBuy[key]; ok && now.Sub(last) < g.BuyDuration {
			return fmt.Errorf("cooldown active for %s", key)
		}
		g.lastBuy[key] = now
		return nil
	}
}
