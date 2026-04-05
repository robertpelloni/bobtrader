package risk

import (
	"context"
	"fmt"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
)

type OrderIntent struct {
	AccountID string
	Symbol    string
	Notional  float64
}

type Guard interface {
	Name() string
	Check(ctx context.Context, acct account.Account, intent OrderIntent) error
}

type Pipeline struct {
	guards []Guard
}

func NewPipeline(guards ...Guard) *Pipeline {
	return &Pipeline{guards: guards}
}

func (p *Pipeline) Run(ctx context.Context, acct account.Account, intent OrderIntent) error {
	for _, guard := range p.guards {
		if err := guard.Check(ctx, acct, intent); err != nil {
			return fmt.Errorf("guard %s failed: %w", guard.Name(), err)
		}
	}
	return nil
}

func (p *Pipeline) Names() []string {
	out := make([]string, 0, len(p.guards))
	for _, guard := range p.guards {
		out = append(out, guard.Name())
	}
	return out
}
