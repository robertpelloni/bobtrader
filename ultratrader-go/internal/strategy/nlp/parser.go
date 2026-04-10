package nlp

import (
	"regexp"
	"strconv"
	"strings"
)

// Condition represents a single entry or exit condition parsed from text.
type Condition struct {
	Indicator string      `json:"indicator"`
	Operator  string      `json:"operator"`
	Value     interface{} `json:"value"`
}

// Config represents a parsed strategy configuration.
type Config struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Coins           []string               `json:"coins"`
	Timeframe       string                 `json:"timeframe"`
	EntryConditions []Condition            `json:"entry_conditions"`
	ExitConditions  []Condition            `json:"exit_conditions"`
	RiskManagement  map[string]interface{} `json:"risk_management"`
}

// Parser parses natural language strategy descriptions into executable configs.
type Parser struct {
	coinPattern       *regexp.Regexp
	timeframePattern  *regexp.Regexp
	indicatorPatterns map[string]*regexp.Regexp
}

// NewParser creates a new NLP Strategy Parser.
func NewParser() *Parser {
	return &Parser{
		coinPattern:      regexp.MustCompile(`\b[A-Z]{2,5}\b`),
		timeframePattern: regexp.MustCompile(`(\d+)\s*(min|minute|hour|day|week)s?`),
		indicatorPatterns: map[string]*regexp.Regexp{
			"rsi":             regexp.MustCompile(`(?i)rsi\s*(?:is|drops|goes|breaks|crosses)?\s*(below|above|under|over|<|>)\s*(\d+)`),
			"sma_cross":       regexp.MustCompile(`(?i)(?:price|it)\s*(?:breaks|crosses)\s*(above|below|over|under)\s*sma\s*(\d+)`),
			"macd":            regexp.MustCompile(`(?i)macd\s*(?:crosses|breaks)\s*(above|below|over|under)\s*(?:signal|zero|0)`),
			"volume_spike":    regexp.MustCompile(`(?i)volume\s*(?:is|spikes)?\s*(above|>)\s*(\d+(?:\.\d+)?)\s*(?:x|times)\s*(?:normal|average)`),
			"stop_loss_pct":   regexp.MustCompile(`(?i)stop\s*loss\s*(?:at|of)?\s*(\d+(?:\.\d+)?)%?`),
			"take_profit_pct": regexp.MustCompile(`(?i)take\s*profit\s*(?:at|of)?\s*(\d+(?:\.\d+)?)%?`),
			"atr_stop":        regexp.MustCompile(`(?i)atr\s*stop\s*(\d+(?:\.\d+)?)x?`),
		},
	}
}

// Parse extracts a structured configuration from a natural language description.
func (p *Parser) Parse(text string) Config {
	config := Config{
		Name:            p.generateName(text),
		Description:     text,
		Coins:           p.extractCoins(text),
		Timeframe:       p.extractTimeframe(text),
		EntryConditions: []Condition{},
		ExitConditions:  []Condition{},
		RiskManagement:  make(map[string]interface{}),
	}

	// Split roughly into sentences or clauses
	clauses := regexp.MustCompile(`(?i)\b(and|when|if|\.|,)\b`).Split(text, -1)

	currentAction := "buy"
	for _, clause := range clauses {
		lowerClause := strings.ToLower(strings.TrimSpace(clause))
		if lowerClause == "" {
			continue
		}

		if strings.Contains(lowerClause, "sell") || strings.Contains(lowerClause, "short") || strings.Contains(lowerClause, "exit") {
			currentAction = "sell"
		} else if strings.Contains(lowerClause, "buy") || strings.Contains(lowerClause, "long") || strings.Contains(lowerClause, "enter") {
			currentAction = "buy"
		}

		for indicator, pattern := range p.indicatorPatterns {
			matches := pattern.FindStringSubmatch(lowerClause)
			if len(matches) > 0 {
				condition := Condition{
					Indicator: indicator,
				}

				if len(matches) > 1 {
					op := strings.ToLower(matches[1])
					if op == "below" || op == "under" || op == "<" {
						condition.Operator = "<"
					} else if op == "above" || op == "over" || op == ">" {
						condition.Operator = ">"
					}
				} else {
					condition.Operator = "active"
				}

				// Check for value
				if len(matches) > 2 {
					if val, err := strconv.ParseFloat(matches[2], 64); err == nil {
						condition.Value = val
					} else {
						condition.Value = matches[2]
					}
				} else if len(matches) > 1 && condition.Operator == "" {
					// Sometimes the value is in group 1
					if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
						condition.Value = val
					}
				}

				if strings.HasPrefix(indicator, "stop_loss") || strings.HasPrefix(indicator, "take_profit") || strings.HasPrefix(indicator, "atr_stop") {
					config.RiskManagement[indicator] = condition.Value
				} else if currentAction == "buy" {
					config.EntryConditions = append(config.EntryConditions, condition)
				} else {
					config.ExitConditions = append(config.ExitConditions, condition)
				}
			}
		}
	}

	// Default risk management
	if _, exists := config.RiskManagement["stop_loss_pct"]; !exists {
		config.RiskManagement["stop_loss_pct"] = 5.0
	}
	if _, exists := config.RiskManagement["take_profit_pct"]; !exists {
		config.RiskManagement["take_profit_pct"] = 10.0
	}

	return config
}

func (p *Parser) extractCoins(text string) []string {
	matches := p.coinPattern.FindAllString(strings.ToUpper(text), -1)
	if len(matches) == 0 {
		return []string{"BTC", "ETH"} // Default
	}

	// Deduplicate
	unique := make(map[string]bool)
	var coins []string
	for _, m := range matches {
		if !unique[m] && len(m) > 1 && m != "SMA" && m != "RSI" && m != "MACD" && m != "ATR" {
			unique[m] = true
			coins = append(coins, m)
		}
	}

	if len(coins) == 0 {
		return []string{"BTC", "ETH"} // Default
	}

	return coins
}

func (p *Parser) extractTimeframe(text string) string {
	match := p.timeframePattern.FindStringSubmatch(text)
	if len(match) > 2 {
		n := match[1]
		unit := match[2]
		if strings.Contains(unit, "min") {
			return n + "min"
		}
		if strings.Contains(unit, "hour") {
			return n + "hour"
		}
		if strings.Contains(unit, "day") {
			return n + "day"
		}
		if strings.Contains(unit, "week") {
			return n + "week"
		}
	}
	return "1hour"
}

func (p *Parser) generateName(description string) string {
	words := strings.Fields(strings.ToLower(description))
	var keywords []string

	stopWords := map[string]bool{
		"the": true, "and": true, "when": true, "with": true, "a": true, "an": true, "to": true, "of": true, "at": true, "or": true, "is": true,
	}

	for _, w := range words {
		cleanWord := regexp.MustCompile(`[^a-z]`).ReplaceAllString(w, "")
		if len(cleanWord) > 2 && !stopWords[cleanWord] {
			keywords = append(keywords, cleanWord)
		}
		if len(keywords) >= 3 {
			break
		}
	}

	if len(keywords) > 0 {
		return strings.Join(keywords, "_")
	}
	return "custom_strategy"
}
