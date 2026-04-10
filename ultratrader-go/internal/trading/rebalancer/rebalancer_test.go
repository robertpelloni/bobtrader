package rebalancer_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/rebalancer"
)

func TestRebalancer_CheckDrift(t *testing.T) {
	config := rebalancer.Config{
		Enabled:           true,
		DriftThresholdPct: 5.0,
		TargetAllocations: map[string]float64{
			"BTC": 50.0,
			"ETH": 30.0,
			"SOL": 20.0,
		},
	}

	reb := rebalancer.NewRebalancer(config)
	portfolio := map[string]float64{
		"BTC": 7000.0, // 70%
		"ETH": 2000.0, // 20%
		"SOL": 1000.0, // 10%
	}

	adjustments := reb.CheckDrift(portfolio, 10000.0)

	assert.Len(t, adjustments, 3)

	var btcAdj, ethAdj, solAdj rebalancer.Adjustment
	for _, a := range adjustments {
		switch a.Symbol {
		case "BTC":
			btcAdj = a
		case "ETH":
			ethAdj = a
		case "SOL":
			solAdj = a
		}
	}

	assert.Equal(t, "SELL", btcAdj.Action)
	assert.InDelta(t, 2000.0, btcAdj.DiffUSD, 0.01)

	assert.Equal(t, "BUY", ethAdj.Action)
	assert.InDelta(t, 1000.0, ethAdj.DiffUSD, 0.01)

	assert.Equal(t, "BUY", solAdj.Action)
	assert.InDelta(t, 1000.0, solAdj.DiffUSD, 0.01)
}

func TestRebalancer_IsRebalanceDue(t *testing.T) {
	config := rebalancer.Config{
		TriggerMode:            "time",
		RebalanceIntervalHours: 24,
	}

	reb := rebalancer.NewRebalancer(config)

	// Due
	lastDue := time.Now().Add(-25 * time.Hour)
	assert.True(t, reb.IsRebalanceDue(lastDue))

	// Not due
	lastNotDue := time.Now().Add(-23 * time.Hour)
	assert.False(t, reb.IsRebalanceDue(lastNotDue))
}

func TestRebalancer_GenerateOrders_WashSale(t *testing.T) {
	config := rebalancer.Config{
		DriftThresholdPct: 5.0,
		TargetAllocations: map[string]float64{
			"BTC": 50.0,
		},
		AvoidWashSales: true,
	}

	reb := rebalancer.NewRebalancer(config)
	portfolio := map[string]float64{
		"BTC": 7000.0, // 70%, needs to sell
	}
	prices := map[string]float64{"BTC": 65000.0}
	costs := map[string]float64{"BTC": 70000.0} // Loss position
	history := []rebalancer.Trade{
		{Symbol: "BTC", Side: "BUY", Timestamp: time.Now().Add(-5 * 24 * time.Hour)},
	}

	orders := reb.GenerateRebalanceOrders(portfolio, 10000.0, prices, costs, history)
	assert.Empty(t, orders, "Wash sale should prevent SELL order")

	// Disable wash sale prevention
	config.AvoidWashSales = false
	reb2 := rebalancer.NewRebalancer(config)
	orders2 := reb2.GenerateRebalanceOrders(portfolio, 10000.0, prices, costs, history)
	assert.Len(t, orders2, 1)
	assert.Equal(t, "SELL", orders2[0].Side)
}
