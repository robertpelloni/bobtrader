package account

import "github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"

type Account struct {
	ID           string
	Name         string
	Enabled      bool
	ExchangeName string
	Capabilities []exchange.Capability
}

func (a Account) Supports(capability exchange.Capability) bool {
	for _, c := range a.Capabilities {
		if c == capability {
			return true
		}
	}
	return false
}
