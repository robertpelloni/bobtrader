package strategy

import "context"

type Signal struct {
	AccountID string
	Symbol    string
	Action    string
	Reason    string
	Quantity  string
	OrderType string
}

type Strategy interface {
	Name() string
	OnTick(ctx context.Context) ([]Signal, error)
}

type Runtime struct {
	strategies []Strategy
}

func NewRuntime(strategies ...Strategy) *Runtime {
	return &Runtime{strategies: strategies}
}

func (r *Runtime) Tick(ctx context.Context) ([]Signal, error) {
	var out []Signal
	for _, strategy := range r.strategies {
		signals, err := strategy.OnTick(ctx)
		if err != nil {
			return nil, err
		}
		out = append(out, signals...)
	}
	return out, nil
}
