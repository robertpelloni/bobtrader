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

	if cfg.EventLog.Path == "" {
		t.Fatal("expected default event log path")
	}
	if cfg.Snapshots.Path == "" {
		t.Fatal("expected default snapshot path")
	}
	if cfg.Orders.Path == "" {
		t.Fatal("expected default order path")
	}
	if cfg.Server.Address == "" {
		t.Fatal("expected default server address")
	}
	if len(cfg.Accounts) == 0 {
		t.Fatal("expected default account")
	}
}

func TestLoadFileOverridesDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{"environment":"test","event_log":{"path":"tmp/events.jsonl"},"snapshots":{"path":"tmp/snapshots.jsonl"},"orders":{"path":"tmp/orders.jsonl"},"server":{"enabled":false,"address":"127.0.0.1:9191"},"accounts":[{"id":"acct-1","name":"Test","enabled":true,"exchange":"paper","capabilities":["spot"]}]}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load file returned error: %v", err)
	}

	if cfg.Environment != "test" {
		t.Fatalf("expected environment test, got %q", cfg.Environment)
	}
	if cfg.EventLog.Path != "tmp/events.jsonl" {
		t.Fatalf("unexpected event log path: %q", cfg.EventLog.Path)
	}
	if cfg.Snapshots.Path != "tmp/snapshots.jsonl" {
		t.Fatalf("unexpected snapshot path: %q", cfg.Snapshots.Path)
	}
	if cfg.Orders.Path != "tmp/orders.jsonl" {
		t.Fatalf("unexpected orders path: %q", cfg.Orders.Path)
	}
	if cfg.Server.Address != "127.0.0.1:9191" {
		t.Fatalf("unexpected server address: %q", cfg.Server.Address)
	}
	if len(cfg.Accounts) != 1 || cfg.Accounts[0].ID != "acct-1" {
		t.Fatalf("unexpected accounts: %+v", cfg.Accounts)
	}
}
