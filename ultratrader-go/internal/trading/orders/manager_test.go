package orders

import (
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

func TestConditionalOrder_StopLossBuy(t *testing.T) {
	co := &ConditionalOrder{
		Side:         exchange.Buy,
		Type:         StopLoss,
		TriggerPrice: 9500,
		Active:       true,
	}

	// Price above stop -> no trigger
	if co.ShouldTrigger(10000) {
		t.Error("should not trigger above stop loss")
	}

	// Price drops to stop -> trigger
	if !co.ShouldTrigger(9500) {
		t.Error("should trigger at stop loss price")
	}
}

func TestConditionalOrder_StopLossSell(t *testing.T) {
	co := &ConditionalOrder{
		Side:         exchange.Sell,
		Type:         StopLoss,
		TriggerPrice: 10500,
		Active:       true,
	}

	// Price below stop -> no trigger
	if co.ShouldTrigger(10000) {
		t.Error("should not trigger below stop for sell")
	}

	// Price rises to stop -> trigger
	if !co.ShouldTrigger(10500) {
		t.Error("should trigger at stop loss for sell")
	}
}

func TestConditionalOrder_TakeProfit(t *testing.T) {
	co := &ConditionalOrder{
		Side:         exchange.Buy,
		Type:         TakeProfit,
		TriggerPrice: 11000,
		Active:       true,
	}

	if co.ShouldTrigger(10000) {
		t.Error("should not trigger below TP")
	}
	if !co.ShouldTrigger(11000) {
		t.Error("should trigger at TP")
	}
	if !co.ShouldTrigger(12000) {
		t.Error("should trigger above TP")
	}
}

func TestConditionalOrder_TrailingStopBuy(t *testing.T) {
	co := &ConditionalOrder{
		Side:         exchange.Buy,
		Type:         TrailingStop,
		TrailPercent: 5.0, // 5% trail
		HighestPrice: 10000,
		Active:       true,
	}

	// Price rises -> update highest
	if co.ShouldTrigger(11000) {
		t.Error("should not trigger while rising")
	}
	if co.HighestPrice != 11000 {
		t.Errorf("expected highest 11000, got %f", co.HighestPrice)
	}

	// Price drops 5% from high -> trigger
	trailTrigger := 11000 * 0.95 // 10450
	if !co.ShouldTrigger(trailTrigger) {
		t.Error("should trigger at trail level")
	}
}

func TestConditionalOrder_TrailingStopSell(t *testing.T) {
	co := &ConditionalOrder{
		Side:         exchange.Sell,
		Type:         TrailingStop,
		TrailPercent: 3.0,
		LowestPrice:  10000,
		Active:       true,
	}

	// Price drops -> update lowest
	if co.ShouldTrigger(9000) {
		t.Error("should not trigger while falling (sell trailing)")
	}
	if co.LowestPrice != 9000 {
		t.Errorf("expected lowest 9000, got %f", co.LowestPrice)
	}

	// Price rises 3% from low -> trigger
	trailTrigger := 9000 * 1.03 // 9270
	if !co.ShouldTrigger(trailTrigger) {
		t.Error("should trigger at trail level for sell")
	}
}

func TestConditionalOrder_AlreadyTriggered(t *testing.T) {
	co := &ConditionalOrder{
		Side:         exchange.Buy,
		Type:         StopLoss,
		TriggerPrice: 9500,
		Active:       false,
		Triggered:    true,
	}

	if co.ShouldTrigger(9000) {
		t.Error("should not trigger already triggered order")
	}
}

func TestManager_PlaceConditional(t *testing.T) {
	m := NewManager()
	id := m.PlaceConditional(ConditionalOrder{
		Symbol:       "BTCUSDT",
		Side:         exchange.Sell,
		Type:         StopLoss,
		Quantity:     "1.0",
		TriggerPrice: 9500,
	})

	if id == "" {
		t.Error("expected non-empty ID")
	}

	order, ok := m.Get(id)
	if !ok {
		t.Fatal("order not found")
	}
	if !order.Active {
		t.Error("expected active")
	}
	if order.Symbol != "BTCUSDT" {
		t.Errorf("expected BTCUSDT, got %s", order.Symbol)
	}
}

func TestManager_CheckPrice(t *testing.T) {
	m := NewManager()
	// Buy position: stop loss triggers when price drops below trigger
	m.PlaceConditional(ConditionalOrder{
		Symbol:       "BTCUSDT",
		Side:         exchange.Buy, // long position protection
		Type:         StopLoss,
		Quantity:     "1.0",
		TriggerPrice: 9500,
	})

	// Price above stop -> no triggers
	triggered := m.CheckPrice("BTCUSDT", 10000)
	if len(triggered) != 0 {
		t.Errorf("expected 0 triggers, got %d", len(triggered))
	}

	// Price drops below stop -> trigger
	triggered = m.CheckPrice("BTCUSDT", 9400)
	if len(triggered) != 1 {
		t.Fatalf("expected 1 trigger, got %d", len(triggered))
	}
	if triggered[0].Type != StopLoss {
		t.Errorf("expected stop loss type, got %s", triggered[0].Type)
	}
}

func TestManager_BracketOrder(t *testing.T) {
	m := NewManager()
	stopID, tpID := m.PlaceBracketOrder("BTCUSDT", exchange.Buy, "1.0", 9500, 11000)

	if stopID == "" || tpID == "" {
		t.Fatal("expected non-empty IDs")
	}

	stop, ok := m.Get(stopID)
	if !ok {
		t.Fatal("stop order not found")
	}
	if stop.Type != StopLoss {
		t.Errorf("expected StopLoss, got %s", stop.Type)
	}

	tp, ok := m.Get(tpID)
	if !ok {
		t.Fatal("TP order not found")
	}
	if tp.Type != TakeProfit {
		t.Errorf("expected TakeProfit, got %s", tp.Type)
	}

	// Both should share the same group
	if stop.GroupID != tp.GroupID {
		t.Error("expected same group ID for bracket orders")
	}
}

func TestManager_BracketOCO(t *testing.T) {
	m := NewManager()
	stopID, tpID := m.PlaceBracketOrder("BTCUSDT", exchange.Buy, "1.0", 9500, 11000)

	// Hit take profit
	triggered := m.CheckPrice("BTCUSDT", 11500)
	if len(triggered) != 1 {
		t.Fatalf("expected 1 trigger, got %d", len(triggered))
	}

	// Stop loss should be cancelled (OCO)
	stop, _ := m.Get(stopID)
	if stop.Active {
		t.Error("stop should be cancelled after TP triggers (OCO)")
	}

	tp, _ := m.Get(tpID)
	if tp.Active {
		t.Error("TP should be inactive after triggering")
	}
}

func TestManager_TrailingStop(t *testing.T) {
	m := NewManager()
	id := m.PlaceTrailingStop("BTCUSDT", exchange.Sell, "1.0", 5.0, 10000)

	order, _ := m.Get(id)
	if order.Type != TrailingStop {
		t.Errorf("expected TrailingStop, got %s", order.Type)
	}
	if order.TrailPercent != 5.0 {
		t.Errorf("expected 5%% trail, got %f%%", order.TrailPercent)
	}
}

func TestManager_Cancel(t *testing.T) {
	m := NewManager()
	id := m.PlaceConditional(ConditionalOrder{
		Symbol:       "BTCUSDT",
		Side:         exchange.Sell,
		Type:         StopLoss,
		TriggerPrice: 9500,
	})

	if !m.Cancel(id) {
		t.Error("expected successful cancel")
	}

	order, _ := m.Get(id)
	if order.Active {
		t.Error("expected inactive after cancel")
	}

	// Cancel again should fail
	if m.Cancel(id) {
		t.Error("expected failed cancel for already cancelled order")
	}
}

func TestManager_ActiveOrders(t *testing.T) {
	m := NewManager()
	m.PlaceConditional(ConditionalOrder{Symbol: "A", Type: StopLoss, TriggerPrice: 100, Active: true})
	m.PlaceConditional(ConditionalOrder{Symbol: "B", Type: TakeProfit, TriggerPrice: 200, Active: true})

	active := m.ActiveOrders()
	if len(active) != 2 {
		t.Errorf("expected 2 active, got %d", len(active))
	}
}

func TestConditionalOrder_Trigger(t *testing.T) {
	co := &ConditionalOrder{Active: true}
	co.Trigger()

	if co.Active {
		t.Error("expected inactive after trigger")
	}
	if !co.Triggered {
		t.Error("expected triggered flag")
	}
	if co.TriggeredAt == nil {
		t.Error("expected triggered timestamp")
	}
}
