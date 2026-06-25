package demo

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

type NewListingSniper struct {
	accountID    string
	quantity     string
	seenSymbols  map[string]bool
	mu           sync.Mutex
	bootTime     time.Time
	baseCurrency string
	testMode     bool
}

func NewNewListingSniper(accountID, quantity, baseCurrency string) *NewListingSniper {
	return &NewListingSniper{
		accountID:    accountID,
		quantity:     quantity,
		seenSymbols:  make(map[string]bool),
		bootTime:     time.Now(),
		baseCurrency: baseCurrency,
	}
}

func NewTestListingSniper(accountID, quantity, baseCurrency string) *NewListingSniper {
	s := NewNewListingSniper(accountID, quantity, baseCurrency)
	s.testMode = true
	return s
}

func (s *NewListingSniper) Name() string {
	return "NewListingSniper"
}

func (s *NewListingSniper) OnTick(ctx context.Context) ([]strategy.Signal, error) {
	return nil, nil
}

func (s *NewListingSniper) PreWarm(symbols []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, sym := range symbols {
		s.seenSymbols[sym] = true
	}
}

func (s *NewListingSniper) OnMarketTick(ctx context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	if s.baseCurrency != "" && !strings.HasSuffix(tick.Symbol, s.baseCurrency) {
		return nil, nil
	}

	price := utils.ParseFloat(tick.Price)
	if price <= 0 {
		return nil, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.seenSymbols[tick.Symbol] {
		return nil, nil
	}

	s.seenSymbols[tick.Symbol] = true

	if !s.testMode && time.Since(s.bootTime) < 10*time.Second {
		return nil, nil
	}

	return []strategy.Signal{{
		AccountID: s.accountID,
		Symbol:    tick.Symbol,
		Action:    "buy",
		Reason:    fmt.Sprintf("New token listing detected! Sniping at %s", tick.Price),
		Quantity:  s.quantity,
		OrderType: "market",
	}}, nil
}
