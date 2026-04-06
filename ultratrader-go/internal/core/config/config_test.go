package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefault(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load default returned error: %v", err)
	}
	if cfg.EventLog.Path == "" || cfg.Snapshots.Path == "" || cfg.Orders.Path == "" || cfg.Reports.Path == "" || cfg.Logging.Path == "" {
		t.Fatal("expected default persistence/logging paths")
	}
	if cfg.Server.Address == "" {
		t.Fatal("expected default server address")
	}
	if cfg.Scheduler.IntervalMS <= 0 {
		t.Fatal("expected default scheduler interval")
	}
	if cfg.Risk.MaxNotional <= 0 || len(cfg.Risk.AllowedSymbols) == 0 {
		t.Fatal("expected default risk config")
	}
	if len(cfg.Accounts) == 0 {
		t.Fatal("expected default account")
	}
}

func TestLoadFileOverridesDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{"environment":"test","event_log":{"path":"tmp/events.jsonl"},"snapshots":{"path":"tmp/snapshots.jsonl"},"orders":{"path":"tmp/orders.jsonl"},"reports":{"path":"tmp/runtime.jsonl"},"logging":{"path":"tmp/app.jsonl","stdout":false},"server":{"enabled":false,"address":"127.0.0.1:9191"},"scheduler":{"enabled":true,"mode":"stream","interval_ms":500},"risk":{"max_notional":250,"max_notional_per_symbol":120,"allowed_symbols":["BTCUSDT"],"cooldown_ms":1000,"duplicate_window_ms":2000,"duplicate_side_window_ms":1500,"max_open_positions":3,"max_concentration_pct":40},"accounts":[{"id":"acct-1","name":"Test","enabled":true,"exchange":"paper","capabilities":["spot"]}]}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load file returned error: %v", err)
	}
	if cfg.Environment != "test" || cfg.EventLog.Path != "tmp/events.jsonl" || cfg.Snapshots.Path != "tmp/snapshots.jsonl" || cfg.Orders.Path != "tmp/orders.jsonl" || cfg.Reports.Path != "tmp/runtime.jsonl" || cfg.Logging.Path != "tmp/app.jsonl" {
		t.Fatalf("unexpected config paths: %+v", cfg)
	}
	if cfg.Server.Address != "127.0.0.1:9191" {
		t.Fatalf("unexpected server address: %q", cfg.Server.Address)
	}
	if !cfg.Scheduler.Enabled || cfg.Scheduler.Mode != "stream" || cfg.Scheduler.IntervalMS != 500 {
		t.Fatalf("unexpected scheduler config: %+v", cfg.Scheduler)
	}
	if cfg.Risk.MaxNotional != 250 || cfg.Risk.MaxNotionalPerSymbol != 120 || cfg.Risk.CooldownMS != 1000 || cfg.Risk.DuplicateWindowMS != 2000 || cfg.Risk.DuplicateSideWindowMS != 1500 || cfg.Risk.MaxOpenPositions != 3 || cfg.Risk.MaxConcentrationPct != 40 || len(cfg.Risk.AllowedSymbols) != 1 || cfg.Risk.AllowedSymbols[0] != "BTCUSDT" {
		t.Fatalf("unexpected risk config: %+v", cfg.Risk)
	}
	if len(cfg.Accounts) != 1 || cfg.Accounts[0].ID != "acct-1" {
		t.Fatalf("unexpected accounts: %+v", cfg.Accounts)
	}
}
