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
	if len(cfg.Accounts) == 0 {
		t.Fatal("expected default account")
	}
}

func TestLoadFileOverridesDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{"environment":"test","event_log":{"path":"tmp/events.jsonl"},"accounts":[{"id":"acct-1","name":"Test","enabled":true,"exchange":"paper","capabilities":["spot"]}]}`
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
	if len(cfg.Accounts) != 1 || cfg.Accounts[0].ID != "acct-1" {
		t.Fatalf("unexpected accounts: %+v", cfg.Accounts)
	}
}
