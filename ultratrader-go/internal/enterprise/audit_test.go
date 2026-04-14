package enterprise_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/enterprise"
)

func TestAuditLogger_Log(t *testing.T) {
	// Create a temporary file
	tmpfile, err := ioutil.TempFile("", "audit_test_*.jsonl")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name()) // Clean up after test
	tmpfile.Close()

	// Initialize Logger
	logger, err := enterprise.NewAuditLogger(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to init audit logger: %v", err)
	}

	// 1. Log First Event
	err = logger.Log("user-123", "login", "auth_service", map[string]interface{}{"ip": "192.168.1.1"})
	if err != nil {
		t.Fatalf("Failed to log first event: %v", err)
	}

	// 2. Log Second Event
	err = logger.Log("user-123", "update_settings", "config_api", map[string]interface{}{"setting": "theme", "value": "dark"})
	if err != nil {
		t.Fatalf("Failed to log second event: %v", err)
	}

	logger.Close()

	// Read file contents
	content, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Fatalf("Expected 2 lines, got %d", len(lines))
	}

	var event1, event2 enterprise.AuditEvent
	if err := json.Unmarshal([]byte(lines[0]), &event1); err != nil {
		t.Fatalf("Failed to parse first event: %v", err)
	}
	if err := json.Unmarshal([]byte(lines[1]), &event2); err != nil {
		t.Fatalf("Failed to parse second event: %v", err)
	}

	// Verify chain integrity
	if event1.PrevHash != "0000000000000000000000000000000000000000000000000000000000000000" {
		t.Errorf("First event should have the zeroed genesis hash, got %s", event1.PrevHash)
	}

	if event2.PrevHash != event1.Hash {
		t.Errorf("Second event PrevHash (%s) does not match first event Hash (%s)", event2.PrevHash, event1.Hash)
	}

	if len(event1.Hash) != 64 || len(event2.Hash) != 64 {
		t.Errorf("Hashes must be 64 character hex strings (SHA-256)")
	}
}
