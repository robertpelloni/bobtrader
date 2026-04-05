package risk

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
)

type CooldownGuard struct {
	Duration time.Duration
	mu       sync.Mutex
	lastSeen map[string]time.Time
}

func NewCooldownGuard(duration time.Duration) *CooldownGuard {
	return &CooldownGuard{Duration: duration, lastSeen: map[string]time.Time{}}
}

func (g *CooldownGuard) Name() string { return "cooldown" }

func (g *CooldownGuard) Check(_ context.Context, acct account.Account, intent OrderIntent) error {
	if g == nil || g.Duration <= 0 {
		return nil
	}
	key := acct.ID + ":" + strings.ToUpper(strings.TrimSpace(intent.Symbol))
	now := time.Now().UTC()
	g.mu.Lock()
	defer g.mu.Unlock()
	if last, ok := g.lastSeen[key]; ok && now.Sub(last) < g.Duration {
		return fmt.Errorf("cooldown active for %s", key)
	}
	g.lastSeen[key] = now
	return nil
}
