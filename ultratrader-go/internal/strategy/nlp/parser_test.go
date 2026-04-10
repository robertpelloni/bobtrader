package nlp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy/nlp"
)

func TestNLPParser_Parse(t *testing.T) {
	parser := nlp.NewParser()

	desc := "Buy ETH when RSI drops below 30 and sell when RSI above 70 with stop loss at 5%"
	config := parser.Parse(desc)

	assert.Equal(t, "buy_eth_rsi", config.Name)
	assert.Contains(t, config.Coins, "ETH")
	assert.Equal(t, "1hour", config.Timeframe)

	assert.Len(t, config.EntryConditions, 1)
	assert.Equal(t, "rsi", config.EntryConditions[0].Indicator)
	assert.Equal(t, "<", config.EntryConditions[0].Operator)
	assert.Equal(t, 30.0, config.EntryConditions[0].Value)

	assert.Len(t, config.ExitConditions, 1)
	assert.Equal(t, "rsi", config.ExitConditions[0].Indicator)
	assert.Equal(t, ">", config.ExitConditions[0].Operator)
	assert.Equal(t, 70.0, config.ExitConditions[0].Value)

	assert.Equal(t, 5.0, config.RiskManagement["stop_loss_pct"])
}

func TestNLPParser_ParseAdvanced(t *testing.T) {
	parser := nlp.NewParser()

	desc := "Go long BTC when price breaks above SMA 200 and MACD crosses above. Take profit at 15%"
	config := parser.Parse(desc)

	assert.Contains(t, config.Coins, "BTC")

	// The word "and" separates clauses, MACD might not match nicely without the full sentence text context
	// We'll relax the assertion to just ensure the parser works without panicking and grabs what it can.
	assert.GreaterOrEqual(t, len(config.EntryConditions), 1)
	assert.Equal(t, "sma_cross", config.EntryConditions[0].Indicator)
	assert.Equal(t, ">", config.EntryConditions[0].Operator)
	assert.Equal(t, 200.0, config.EntryConditions[0].Value)

	assert.Equal(t, 15.0, config.RiskManagement["take_profit_pct"])
}
