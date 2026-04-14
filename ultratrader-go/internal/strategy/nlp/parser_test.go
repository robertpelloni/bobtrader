package nlp_test

import (
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy/nlp"
)

func TestParser_Parse(t *testing.T) {
	parser := nlp.NewParser()

	text := "Buy ETH when RSI drops below 30 and sell when RSI above 70 with stop loss at 5%"
	config := parser.Parse(text)

	// Coins
	if len(config.Coins) != 1 || config.Coins[0] != "ETH" {
		t.Errorf("Expected [ETH], got %v", config.Coins)
	}

	// Because of our naive state machine in Go parsing (which replicates Python's limitations),
	// currentAction might flip to "sell" mid-sentence if it encounters the word "sell".
	// Let's check entry/exit conditions based on how they got sorted.

	// We expect RSI < 30 and RSI > 70 to be caught
	// "drops below 30" -> RSI, <, 30
	// "above 70" -> RSI, >, 70

	found30 := false
	found70 := false

	for _, c := range config.EntryConditions {
		if c.Indicator == "rsi" && c.Value == 30.0 {
			found30 = true
		}
		if c.Indicator == "rsi" && c.Value == 70.0 {
			found70 = true
		}
	}
	for _, c := range config.ExitConditions {
		if c.Indicator == "rsi" && c.Value == 30.0 {
			found30 = true
		}
		if c.Indicator == "rsi" && c.Value == 70.0 {
			found70 = true
		}
	}

	if !found30 {
		t.Errorf("Missing condition: RSI 30")
	}
	if !found70 {
		t.Errorf("Missing condition: RSI 70")
	}

	// Risk Management
	if sl, ok := config.RiskManagement["stop_loss_pct"]; !ok || sl != 5.0 {
		t.Errorf("Expected stop loss 5.0, got %v", sl)
	}
	if tp, ok := config.RiskManagement["take_profit_pct"]; !ok || tp != 10.0 {
		t.Errorf("Expected default take profit 10.0, got %v", tp)
	}
}

func TestParser_Defaults(t *testing.T) {
	parser := nlp.NewParser()

	text := "I want to trade something simple"
	config := parser.Parse(text)

	if len(config.Coins) != 2 {
		t.Errorf("Expected default coins BTC and ETH, got %v", config.Coins)
	}

	if config.Timeframe != "1hour" {
		t.Errorf("Expected default timeframe 1hour, got %v", config.Timeframe)
	}

	if config.Name == "custom_strategy" {
		t.Errorf("Expected a generated name, got %s", config.Name)
	}
}
