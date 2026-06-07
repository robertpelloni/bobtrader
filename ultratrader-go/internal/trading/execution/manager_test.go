package execution

import (
	"context"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	exchangepaper "github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange/paper"
	"github.com/stretchr/testify/assert"
)

func TestExecutionManager(t *testing.T) {
	ctx := context.Background()
	manager := NewManager()
	paperAdapter := exchangepaper.New()

	marketStrategy := NewMarketStrategy(paperAdapter)
	manager.Register(marketStrategy)

	t.Run("ListStrategies", func(t *testing.T) {
		strategies := manager.ListStrategies()
		assert.Contains(t, strategies, "market")
	})

	t.Run("ExecuteMarketOrder", func(t *testing.T) {
		order := exchange.Order{
			Symbol:   "BTCUSDT",
			Side:     exchange.Buy,
			Quantity: "1.0",
		}
		err := manager.Execute(ctx, "market", order)
		assert.NoError(t, err)
	})

	t.Run("ExecuteUnknownStrategy", func(t *testing.T) {
		order := exchange.Order{Symbol: "BTCUSDT"}
		err := manager.Execute(ctx, "unknown", order)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "execution strategy not found")
	})
}
