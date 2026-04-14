package nlp

import (
	"regexp"
	"strconv"
	"strings"
)

// Condition defines an entry or exit condition extracted from NLP text.
type Condition struct {
	Indicator string  `json:"indicator"`
	Operator  string  `json:"operator"`
	Value     float64 `json:"value"`
}

// StrategyConfig holds the parsed trading strategy settings.
type StrategyConfig struct {
	Name            string             `json:"name"`
	Coins           []string           `json:"coins"`
	Timeframe       string             `json:"timeframe"`
	EntryConditions []Condition        `json:"entry_conditions"`
	ExitConditions  []Condition        `json:"exit_conditions"`
	RiskManagement  map[string]float64 `json:"risk_management"`
}

// Parser parses natural language text into a structured StrategyConfig.
type Parser struct {
	coinPattern       *regexp.Regexp
	timeframePattern  *regexp.Regexp
	actionPattern     *regexp.Regexp
	indicatorPattern  *regexp.Regexp
	stopProfitPattern *regexp.Regexp
}

// NewParser initializes an NLP parser with predefined regex patterns.
func NewParser() *Parser {
	return &Parser{
		coinPattern:       regexp.MustCompile(`\b(?:BTC|ETH|SOL|ADA|DOT|XRP|LTC|LINK|DOGE|BCH)\b`),
		timeframePattern:  regexp.MustCompile(`(?i)\b(\d+)\s*(min|minute|hour|hr|day|week)s?\b`),
		actionPattern:     regexp.MustCompile(`(?i)\b(buy|sell|long|short)\b`),
		indicatorPattern:  regexp.MustCompile(`(?i)(rsi|sma|ema|macd|bollinger|volume)\s*(?:is)?\s*(above|below|crosses above|crosses below|<|>|over|under|drops below)\s*(\d+(?:\.\d+)?)`),
		stopProfitPattern: regexp.MustCompile(`(?i)(stop loss|take profit|atr stop)\s*(?:at)?\s*(\d+(?:\.\d+)?)%?`),
	}
}

// Parse converts a natural language string into a structured StrategyConfig.
func (p *Parser) Parse(text string) StrategyConfig {
	config := StrategyConfig{
		Name:            p.generateName(text),
		Coins:           p.extractCoins(text),
		Timeframe:       p.extractTimeframe(text),
		EntryConditions: make([]Condition, 0),
		ExitConditions:  make([]Condition, 0),
		RiskManagement:  make(map[string]float64),
	}

	// Simple state tracking
	currentAction := "buy"
	words := strings.Fields(text)

	for _, w := range words {
		lower := strings.ToLower(w)
		if lower == "buy" || lower == "long" {
			currentAction = "buy"
		} else if lower == "sell" || lower == "short" {
			currentAction = "sell"
		}
	}

	// We'll iterate through indicator matches
	matches := p.indicatorPattern.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		if len(match) < 4 {
			continue
		}
		indicator := strings.ToLower(match[1])
		opRaw := strings.ToLower(match[2])
		valStr := match[3]

		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			continue
		}

		operator := "<"
		if strings.Contains(opRaw, "above") || strings.Contains(opRaw, ">") || strings.Contains(opRaw, "over") {
			operator = ">"
		}

		cond := Condition{
			Indicator: indicator,
			Operator:  operator,
			Value:     val,
		}

		// Simple heuristic to split entries vs exits
		// Normally we'd do deeper sentence parsing, but keeping logic parallel to Python version
		if currentAction == "buy" {
			config.EntryConditions = append(config.EntryConditions, cond)
		} else {
			config.ExitConditions = append(config.ExitConditions, cond)
		}
	}

	// Extract Risk Params
	riskMatches := p.stopProfitPattern.FindAllStringSubmatch(text, -1)
	for _, match := range riskMatches {
		if len(match) < 3 {
			continue
		}
		riskTypeRaw := strings.ToLower(match[1])
		valStr := match[2]

		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			continue
		}

		key := "stop_loss_pct"
		if strings.Contains(riskTypeRaw, "take profit") {
			key = "take_profit_pct"
		} else if strings.Contains(riskTypeRaw, "atr stop") {
			key = "atr_stop"
		}

		config.RiskManagement[key] = val
	}

	// Defaults
	if _, ok := config.RiskManagement["stop_loss_pct"]; !ok {
		config.RiskManagement["stop_loss_pct"] = 5.0
	}
	if _, ok := config.RiskManagement["take_profit_pct"]; !ok {
		config.RiskManagement["take_profit_pct"] = 10.0
	}

	return config
}

func (p *Parser) extractCoins(text string) []string {
	matches := p.coinPattern.FindAllString(strings.ToUpper(text), -1)
	if len(matches) == 0 {
		return []string{"BTC", "ETH"}
	}
	// Deduplicate
	seen := make(map[string]bool)
	var coins []string
	for _, m := range matches {
		if !seen[m] {
			seen[m] = true
			coins = append(coins, m)
		}
	}
	return coins
}

func (p *Parser) extractTimeframe(text string) string {
	match := p.timeframePattern.FindStringSubmatch(text)
	if len(match) >= 3 {
		n := match[1]
		unit := strings.ToLower(match[2])
		if strings.HasPrefix(unit, "min") {
			return n + "min"
		}
		if strings.HasPrefix(unit, "hour") || strings.HasPrefix(unit, "hr") {
			return n + "hour"
		}
		if strings.HasPrefix(unit, "day") {
			return n + "day"
		}
		if strings.HasPrefix(unit, "week") {
			return n + "week"
		}
	}
	return "1hour"
}

func (p *Parser) generateName(text string) string {
	words := strings.Fields(strings.ToLower(text))
	var keywords []string
	ignored := map[string]bool{"the": true, "and": true, "when": true, "with": true}
	for _, w := range words {
		if len(w) > 2 && !ignored[w] {
			keywords = append(keywords, w)
		}
		if len(keywords) == 3 {
			break
		}
	}
	if len(keywords) == 0 {
		return "custom_strategy"
	}
	return strings.Join(keywords, "_")
}
