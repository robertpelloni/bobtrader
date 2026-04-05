package risk

import (
	"context"
	"fmt"
	"strings"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
)

type SymbolWhitelistGuard struct {
	Allowed map[string]struct{}
}

func NewSymbolWhitelistGuard(symbols []string) SymbolWhitelistGuard {
	allowed := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		symbol = strings.TrimSpace(strings.ToUpper(symbol))
		if symbol != "" {
			allowed[symbol] = struct{}{}
		}
	}
	return SymbolWhitelistGuard{Allowed: allowed}
}

func (g SymbolWhitelistGuard) Name() string { return "symbol-whitelist" }

func (g SymbolWhitelistGuard) Check(_ context.Context, _ account.Account, intent OrderIntent) error {
	if len(g.Allowed) == 0 {
		return nil
	}
	symbol := strings.ToUpper(strings.TrimSpace(intent.Symbol))
	if _, ok := g.Allowed[symbol]; !ok {
		return fmt.Errorf("symbol %q is not whitelisted", symbol)
	}
	return nil
}
