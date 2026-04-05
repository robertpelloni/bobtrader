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
	Accounts    []AccountConfig `json:"accounts"`
}

type EventLogConfig struct {
	Path string `json:"path"`
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
		EventLog: EventLogConfig{
			Path: filepath.Join("data", "eventlog", "events.jsonl"),
		},
		Accounts: []AccountConfig{
			{
				ID:           "paper-main",
				Name:         "Paper Main",
				Enabled:      true,
				Exchange:     "paper",
				Capabilities: []string{"spot", "paper", "candles", "balances", "orders"},
			},
		},
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

	if cfg.EventLog.Path == "" {
		cfg.EventLog.Path = Default().EventLog.Path
	}

	if len(cfg.Accounts) == 0 {
		cfg.Accounts = Default().Accounts
	}

	return cfg, nil
}
