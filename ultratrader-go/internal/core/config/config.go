package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Environment string          `json:"environment"`
	EventLog    EventLogConfig  `json:"event_log"`
	Snapshots   SnapshotConfig  `json:"snapshots"`
	Orders      OrderConfig     `json:"orders"`
	Reports     ReportConfig    `json:"reports"`
	Logging     LoggingConfig   `json:"logging"`
	Server      ServerConfig    `json:"server"`
	Scheduler   SchedulerConfig `json:"scheduler"`
	Risk        RiskConfig      `json:"risk"`
	Strategy    StrategyConfig  `json:"strategy"`
	MarketData  MarketDataConfig `json:"market_data"`
	Accounts    []AccountConfig `json:"accounts"`
}

type EventLogConfig struct {
	Path string `json:"path"`
}

type SnapshotConfig struct {
	Path string `json:"path"`
}

type OrderConfig struct {
	Path string `json:"path"`
}

type ReportConfig struct {
	Path string `json:"path"`
}

type LoggingConfig struct {
	Path   string `json:"path"`
	Stdout bool   `json:"stdout"`
}

type ServerConfig struct {
	Enabled bool   `json:"enabled"`
	Address string `json:"address"`
}

type SchedulerConfig struct {
	Enabled    bool   `json:"enabled"`
	Mode       string `json:"mode"`
	IntervalMS int    `json:"interval_ms"`
}

type RiskConfig struct {
	MaxNotional            float64  `json:"max_notional"`
	MaxNotionalPerSymbol   float64  `json:"max_notional_per_symbol"`
	AllowedSymbols         []string `json:"allowed_symbols"`
	CooldownMS             int      `json:"cooldown_ms"`
	DuplicateWindowMS      int      `json:"duplicate_window_ms"`
	DuplicateSideWindowMS  int      `json:"duplicate_side_window_ms"`
	MaxOpenPositions       int      `json:"max_open_positions"`
	MaxConcentrationPct    float64  `json:"max_concentration_pct"`
}

type StrategyConfig struct {
	RiskPct                float64 `json:"risk_pct"`
	MaxNotional            float64 `json:"max_notional"`
	TrailingActivatePct    float64 `json:"trailing_activate_pct"`
	TrailingGapPct         float64 `json:"trailing_gap_pct"`
	TrailingStopLossPct    float64 `json:"trailing_stop_loss_pct"`
	TrailingMaxHoldMinutes int     `json:"trailing_max_hold_minutes"`
	BollingerPeriod        int     `json:"bollinger_period"`
	BollingerStdDev        float64 `json:"bollinger_std_dev"`
	RSIPeriod              int     `json:"rsi_period"`
	RSIOversold            float64 `json:"rsi_oversold"`
	RSIOverbought          float64 `json:"rsi_overbought"`
	EMAFast                int     `json:"ema_fast"`
	EMASlow                int     `json:"ema_slow"`
}

type MarketDataConfig struct {
	Source         string  `json:"source"` // "rest" or "websocket"
	InitialBalance float64 `json:"initial_balance"`
}

type AccountConfig struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Enabled      bool     `json:"enabled"`
	Exchange     string   `json:"exchange"`
	Capabilities []string `json:"capabilities"`
	APIKey       string   `json:"api_key"`
	SecretKey    string   `json:"secret_key"`
	Testnet      bool     `json:"testnet"`
	SecretsFile   string   `json:"secrets_file"` // path to JSON file with api_key/secret_key
}

type secretsFile struct {
	APIKey    string `json:"api_key"`
	SecretKey string `json:"secret_key"`
}

func Default() Config {
	return Config{
		Environment: "development",
		EventLog:    EventLogConfig{Path: filepath.Join("data", "eventlog", "events.jsonl")},
		Snapshots:   SnapshotConfig{Path: filepath.Join("data", "snapshots", "accounts.jsonl")},
		Orders:      OrderConfig{Path: filepath.Join("data", "orders", "orders.jsonl")},
		Reports:     ReportConfig{Path: filepath.Join("data", "reports", "runtime.jsonl")},
		Logging:     LoggingConfig{Path: filepath.Join("data", "logs", "app.jsonl"), Stdout: true},
		Server:      ServerConfig{Enabled: true, Address: "127.0.0.1:0"},
		Scheduler:   SchedulerConfig{Enabled: false, Mode: "timer", IntervalMS: 1000},
		Risk: RiskConfig{
			MaxNotional:          1000,
			MaxNotionalPerSymbol: 0,
			AllowedSymbols:       []string{"BTCUSDT", "ETHUSDT"},
			CooldownMS:           0,
			DuplicateWindowMS:    0,
			DuplicateSideWindowMS: 0,
			MaxOpenPositions:     0,
			MaxConcentrationPct:  0,
		},
		Strategy: StrategyConfig{
			RiskPct:                2.0,
			MaxNotional:            1000,
			TrailingActivatePct:    1.0,
			TrailingGapPct:         0.3,
			TrailingStopLossPct:    3.0,
			TrailingMaxHoldMinutes: 5,
			BollingerPeriod:        20,
			BollingerStdDev:        2.0,
			RSIPeriod:              14,
			RSIOversold:            35,
			RSIOverbought:          65,
			EMAFast:                9,
			EMASlow:                21,
		},
		MarketData: MarketDataConfig{
			Source:         "rest",
			InitialBalance: 10000,
		},
		Accounts: []AccountConfig{{
			ID:           "paper-main",
			Name:         "Paper Main",
			Enabled:      true,
			Exchange:     "paper",
			Capabilities: []string{"spot", "paper", "candles", "balances", "orders"},
		}},
	}
}

func Load(path string) (Config, error) {
	if path == "" {
		return Default(), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file: %w", err)
	}
	cfg := Default()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}
	defaults := Default()
	if cfg.EventLog.Path == "" {
		cfg.EventLog.Path = defaults.EventLog.Path
	}
	if cfg.Snapshots.Path == "" {
		cfg.Snapshots.Path = defaults.Snapshots.Path
	}
	if cfg.Orders.Path == "" {
		cfg.Orders.Path = defaults.Orders.Path
	}
	if cfg.Reports.Path == "" {
		cfg.Reports.Path = defaults.Reports.Path
	}
	if cfg.Logging.Path == "" {
		cfg.Logging.Path = defaults.Logging.Path
	}
	if cfg.Server.Address == "" {
		cfg.Server.Address = defaults.Server.Address
	}
	if cfg.Scheduler.Mode == "" {
		cfg.Scheduler.Mode = defaults.Scheduler.Mode
	}
	if cfg.Scheduler.IntervalMS <= 0 {
		cfg.Scheduler.IntervalMS = defaults.Scheduler.IntervalMS
	}
	if cfg.Risk.MaxNotional == 0 {
		cfg.Risk.MaxNotional = defaults.Risk.MaxNotional
	}
	if cfg.Risk.MaxNotionalPerSymbol == 0 {
		cfg.Risk.MaxNotionalPerSymbol = defaults.Risk.MaxNotionalPerSymbol
	}
	if len(cfg.Risk.AllowedSymbols) == 0 {
		cfg.Risk.AllowedSymbols = defaults.Risk.AllowedSymbols
	}
	if len(cfg.Accounts) == 0 {
		cfg.Accounts = defaults.Accounts
	}
	// Strategy defaults
	if cfg.Strategy.RiskPct == 0 {
		cfg.Strategy.RiskPct = defaults.Strategy.RiskPct
	}
	if cfg.Strategy.MaxNotional == 0 {
		cfg.Strategy.MaxNotional = defaults.Strategy.MaxNotional
	}
	if cfg.Strategy.TrailingActivatePct == 0 {
		cfg.Strategy.TrailingActivatePct = defaults.Strategy.TrailingActivatePct
	}
	if cfg.Strategy.TrailingGapPct == 0 {
		cfg.Strategy.TrailingGapPct = defaults.Strategy.TrailingGapPct
	}
	if cfg.Strategy.TrailingStopLossPct == 0 {
		cfg.Strategy.TrailingStopLossPct = defaults.Strategy.TrailingStopLossPct
	}
	if cfg.Strategy.TrailingMaxHoldMinutes == 0 {
		cfg.Strategy.TrailingMaxHoldMinutes = defaults.Strategy.TrailingMaxHoldMinutes
	}
	if cfg.Strategy.BollingerPeriod == 0 {
		cfg.Strategy.BollingerPeriod = defaults.Strategy.BollingerPeriod
	}
	if cfg.Strategy.BollingerStdDev == 0 {
		cfg.Strategy.BollingerStdDev = defaults.Strategy.BollingerStdDev
	}
	if cfg.Strategy.RSIPeriod == 0 {
		cfg.Strategy.RSIPeriod = defaults.Strategy.RSIPeriod
	}
	if cfg.Strategy.RSIOversold == 0 {
		cfg.Strategy.RSIOversold = defaults.Strategy.RSIOversold
	}
	if cfg.Strategy.RSIOverbought == 0 {
		cfg.Strategy.RSIOverbought = defaults.Strategy.RSIOverbought
	}
	if cfg.Strategy.EMAFast == 0 {
		cfg.Strategy.EMAFast = defaults.Strategy.EMAFast
	}
	if cfg.Strategy.EMASlow == 0 {
		cfg.Strategy.EMASlow = defaults.Strategy.EMASlow
	}
	// MarketData defaults
	if cfg.MarketData.Source == "" {
		cfg.MarketData.Source = defaults.MarketData.Source
	}
	if cfg.MarketData.InitialBalance == 0 {
		cfg.MarketData.InitialBalance = defaults.MarketData.InitialBalance
	}
	// Load secrets from separate files if specified
	for i := range cfg.Accounts {
		acct := &cfg.Accounts[i]
		if acct.SecretsFile != "" {
			secData, err := os.ReadFile(acct.SecretsFile)
			if err != nil {
				return Config{}, fmt.Errorf("read secrets file %s: %w", acct.SecretsFile, err)
			}
			var sec secretsFile
			if err := json.Unmarshal(secData, &sec); err != nil {
				return Config{}, fmt.Errorf("unmarshal secrets file %s: %w", acct.SecretsFile, err)
			}
			if sec.APIKey != "" {
				acct.APIKey = sec.APIKey
			}
			if sec.SecretKey != "" {
				acct.SecretKey = sec.SecretKey
			}
		}
	}

	return cfg, nil
}
