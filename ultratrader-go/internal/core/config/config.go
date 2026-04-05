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
	Server      ServerConfig    `json:"server"`
	Accounts    []AccountConfig `json:"accounts"`
}

type EventLogConfig struct {
	Path string `json:"path"`
}

type SnapshotConfig struct {
	Path string `json:"path"`
}

type ServerConfig struct {
	Enabled bool   `json:"enabled"`
	Address string `json:"address"`
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
		Snapshots: SnapshotConfig{
			Path: filepath.Join("data", "snapshots", "accounts.jsonl"),
		},
		Server: ServerConfig{
			Enabled: true,
			Address: "127.0.0.1:8080",
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

	defaults := Default()
	if cfg.EventLog.Path == "" {
		cfg.EventLog.Path = defaults.EventLog.Path
	}
	if cfg.Snapshots.Path == "" {
		cfg.Snapshots.Path = defaults.Snapshots.Path
	}
	if cfg.Server.Address == "" {
		cfg.Server.Address = defaults.Server.Address
	}
	if len(cfg.Accounts) == 0 {
		cfg.Accounts = defaults.Accounts
	}

	return cfg, nil
}
