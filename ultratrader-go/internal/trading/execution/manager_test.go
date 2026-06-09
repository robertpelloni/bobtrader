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

	t.Run("WolfBotBollinger_Logic", func(t *testing.T) {
		strategy := NewWolfBotBollingerStrategy(paperAdapter, 2)

		// Initial state
		assert.Equal(t, "none", strategy.UpdateState(0.5))

		// Reach upper band - should suggest sell (reversion) initially
		assert.Equal(t, "sell-reversion", strategy.UpdateState(1.0))
		assert.Equal(t, "down", strategy.lastTrend)

		// Reach upper band again - still sell-reversion as breakout limit (2) not hit
		assert.Equal(t, "sell-reversion", strategy.UpdateState(1.0))

		// Reach upper band 3rd time - now should be buy-breakout
		assert.Equal(t, "buy-breakout", strategy.UpdateState(1.0))
		assert.Equal(t, "up", strategy.lastTrend)
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
