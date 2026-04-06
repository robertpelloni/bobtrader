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
	MaxNotional         float64  `json:"max_notional"`
	AllowedSymbols      []string `json:"allowed_symbols"`
	CooldownMS          int      `json:"cooldown_ms"`
	DuplicateWindowMS   int      `json:"duplicate_window_ms"`
	MaxOpenPositions    int      `json:"max_open_positions"`
	MaxConcentrationPct float64  `json:"max_concentration_pct"`
}

type AccountConfig struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Enabled      bool     `json:"enabled"`
	Exchange     string   `json:"exchange"`
	Capabilities []string `json:"capabilities"`
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
		Risk:        RiskConfig{MaxNotional: 1000, AllowedSymbols: []string{"BTCUSDT", "ETHUSDT"}, CooldownMS: 0, DuplicateWindowMS: 0, MaxOpenPositions: 0, MaxConcentrationPct: 0},
		Accounts:    []AccountConfig{{ID: "paper-main", Name: "Paper Main", Enabled: true, Exchange: "paper", Capabilities: []string{"spot", "paper", "candles", "balances", "orders"}}},
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
	if len(cfg.Risk.AllowedSymbols) == 0 {
		cfg.Risk.AllowedSymbols = defaults.Risk.AllowedSymbols
	}
	if len(cfg.Accounts) == 0 {
		cfg.Accounts = defaults.Accounts
	}
	return cfg, nil
}
