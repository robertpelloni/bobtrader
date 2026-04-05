package risk

import (
	"context"
	"errors"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
)

type stubGuard struct {
	name string
	err  error
}

func (g stubGuard) Name() string                                                    { return g.name }
func (g stubGuard) Check(_ context.Context, _ account.Account, _ OrderIntent) error { return g.err }

func TestPipelineStopsOnError(t *testing.T) {
	pipeline := NewPipeline(
		stubGuard{name: "ok"},
		stubGuard{name: "fail", err: errors.New("blocked")},
	)

	err := pipeline.Run(context.Background(), account.Account{ID: "acct"}, OrderIntent{AccountID: "acct", Symbol: "BTCUSDT", Notional: 100})
	if err == nil {
		t.Fatal("expected pipeline error")
	}
}
