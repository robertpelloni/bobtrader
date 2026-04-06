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

type GuardError struct {
	GuardName string
	Cause     error
}

func (e GuardError) Error() string {
	if e.Cause == nil {
		return fmt.Sprintf("guard %s failed", e.GuardName)
	}
	return fmt.Sprintf("guard %s failed: %v", e.GuardName, e.Cause)
}

func (e GuardError) Unwrap() error { return e.Cause }

type Pipeline struct {
	guards []Guard
}

func NewPipeline(guards ...Guard) *Pipeline {
	return &Pipeline{guards: guards}
}

func (p *Pipeline) Run(ctx context.Context, acct account.Account, intent OrderIntent) error {
	for _, guard := range p.guards {
		if err := guard.Check(ctx, acct, intent); err != nil {
			return GuardError{GuardName: guard.Name(), Cause: err}
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
