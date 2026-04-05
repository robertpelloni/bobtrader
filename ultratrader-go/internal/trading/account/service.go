package account

import (
	"fmt"
	"sort"
	"strings"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/config"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

type Service struct {
	accounts map[string]Account
}

func NewService(cfg []config.AccountConfig) (*Service, error) {
	accounts := make(map[string]Account, len(cfg))
	for _, item := range cfg {
		if strings.TrimSpace(item.ID) == "" {
			return nil, fmt.Errorf("account id is required")
		}

		caps := make([]exchange.Capability, 0, len(item.Capabilities))
		for _, capability := range item.Capabilities {
			caps = append(caps, exchange.Capability(capability))
		}

		accounts[item.ID] = Account{
			ID:           item.ID,
			Name:         item.Name,
			Enabled:      item.Enabled,
			ExchangeName: item.Exchange,
			Capabilities: caps,
		}
	}

	return &Service{accounts: accounts}, nil
}

func (s *Service) Get(id string) (Account, bool) {
	acct, ok := s.accounts[id]
	return acct, ok
}

func (s *Service) List() []Account {
	keys := make([]string, 0, len(s.accounts))
	for id := range s.accounts {
		keys = append(keys, id)
	}
	sort.Strings(keys)

	out := make([]Account, 0, len(s.accounts))
	for _, id := range keys {
		out = append(out, s.accounts[id])
	}
	return out
}
