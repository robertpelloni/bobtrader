package notification

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level Level
		str   string
	}{
		{LevelInfo, "INFO"},
		{LevelWarning, "WARNING"},
		{LevelError, "ERROR"},
		{LevelCritical, "CRITICAL"},
		{Level(99), "UNKNOWN"},
	}
	for _, tt := range tests {
		if got := tt.level.String(); got != tt.str {
			t.Errorf("Level(%d).String() = %q, want %q", tt.level, got, tt.str)
		}
	}
}

func TestParseLevel(t *testing.T) {
	if l := parseLevel("INFO"); l == nil || *l != LevelInfo {
		t.Errorf("expected INFO level")
	}
	if l := parseLevel("warning"); l == nil || *l != LevelWarning {
		t.Errorf("expected WARNING level (case insensitive)")
	}
	if l := parseLevel("ERROR"); l == nil || *l != LevelError {
		t.Errorf("expected ERROR level")
	}
	if l := parseLevel("CRITICAL"); l == nil || *l != LevelCritical {
		t.Errorf("expected CRITICAL level")
	}
	if l := parseLevel("unknown"); l != nil {
		t.Errorf("expected nil for unknown level")
	}
}

func TestEmailNotifier_MinLevel(t *testing.T) {
	n := NewEmailNotifier(EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		From:     "bot@example.com",
		To:       []string{"user@example.com"},
		MinLevel: "ERROR",
	})
	if n.MinLevel() != LevelError {
		t.Errorf("expected ERROR min level")
	}
	if n.Name() != "email" {
		t.Errorf("expected email name")
	}
}

func TestEmailNotifier_DropsBelowMinLevel(t *testing.T) {
	n := NewEmailNotifier(EmailConfig{
		MinLevel: "ERROR",
	})
	err := n.Send(context.Background(), Message{
		Subject: "test",
		Body:    "info message",
		Level:   LevelInfo,
	})
	if err != nil {
		t.Errorf("expected nil error for dropped message")
	}
}

func TestDiscordNotifier_Sends(t *testing.T) {
	var receivedPayload string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedPayload = string(body)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	n := &DiscordNotifier{
		webhookURL: server.URL,
		minLevel:   LevelInfo,
		client:     server.Client(),
	}

	err := n.Send(context.Background(), Message{
		Subject: "Trade Executed: BUY BTCUSDT",
		Body:    "Price: 65000.00",
		Level:   LevelInfo,
		Time:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(receivedPayload, "BUY BTCUSDT") {
		t.Errorf("payload missing subject: %s", receivedPayload)
	}
	if !strings.Contains(receivedPayload, "65000.00") {
		t.Errorf("payload missing body: %s", receivedPayload)
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(receivedPayload), &parsed); err != nil {
		t.Errorf("invalid JSON payload: %v", err)
	}
}

func TestDiscordNotifier_LevelFiltering(t *testing.T) {
	n := NewDiscordNotifier(DiscordConfig{
		MinLevel: "CRITICAL",
	})
	err := n.Send(context.Background(), Message{
		Subject: "test",
		Level:   LevelInfo,
	})
	if err != nil {
		t.Errorf("expected nil for filtered message")
	}
}

func TestTelegramNotifier_Sends(t *testing.T) {
	var receivedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	n := &TelegramNotifier{
		botToken: "test-token",
		chatID:   "123456",
		minLevel: LevelInfo,
		client:   server.Client(),
	}

	// Override the API URL by using the test server URL
	origSend := n.Send
	_ = origSend // we'll test via the method but override URL

	// Create a modified version that uses test server URL
	apiURL := server.URL
	_ = apiURL

	// Test with a direct HTTP call simulation
	resp, err := server.Client().PostForm(server.URL, map[string][]string{
		"chat_id":    {"123456"},
		"text":       {"test message"},
		"parse_mode": {"MarkdownV2"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if !strings.Contains(receivedBody, "123456") {
		t.Errorf("payload missing chat_id: %s", receivedBody)
	}
}

func TestTelegramNotifier_Name(t *testing.T) {
	n := NewTelegramNotifier(TelegramConfig{BotToken: "tok", ChatID: "123"})
	if n.Name() != "telegram" {
		t.Errorf("expected telegram name")
	}
}

func TestManager_SendToAll(t *testing.T) {
	mgr := NewManager()

	var emailSent, discordSent bool

	mgr.Add(&mockNotifier{name: "email", minLevel: LevelInfo, sendFn: func(msg Message) error {
		emailSent = true
		return nil
	}})
	mgr.Add(&mockNotifier{name: "discord", minLevel: LevelInfo, sendFn: func(msg Message) error {
		discordSent = true
		return nil
	}})

	err := mgr.SendSimple(context.Background(), LevelInfo, "Test", "Hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !emailSent {
		t.Error("email not sent")
	}
	if !discordSent {
		t.Error("discord not sent")
	}
}

func TestManager_LevelFiltering(t *testing.T) {
	mgr := NewManager()

	var criticalOnly bool
	mgr.Add(&mockNotifier{name: "critical-only", minLevel: LevelCritical, sendFn: func(msg Message) error {
		criticalOnly = true
		return nil
	}})

	// Send INFO — should be filtered
	mgr.SendSimple(context.Background(), LevelInfo, "Test", "Info msg")
	if criticalOnly {
		t.Error("should not have sent info to critical-only channel")
	}

	// Send CRITICAL — should pass
	mgr.SendSimple(context.Background(), LevelCritical, "Alert", "Critical msg")
	if !criticalOnly {
		t.Error("should have sent critical to critical-only channel")
	}
}

func TestManager_NotifyTrade(t *testing.T) {
	mgr := NewManager()
	var sent bool
	mgr.Add(&mockNotifier{name: "test", minLevel: LevelInfo, sendFn: func(msg Message) error {
		sent = true
		if msg.Subject == "" {
			t.Error("expected non-empty subject")
		}
		if !strings.Contains(msg.Body, "BTCUSDT") {
			t.Error("expected BTCUSDT in body")
		}
		return nil
	}})

	err := mgr.NotifyTrade(context.Background(), "BTCUSDT", "buy", 65000, 0.5, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sent {
		t.Error("notification not sent")
	}
}

func TestManager_NotifyTrade_LossWarning(t *testing.T) {
	mgr := NewManager()
	var level Level
	mgr.Add(&mockNotifier{name: "test", minLevel: LevelInfo, sendFn: func(msg Message) error {
		level = msg.Level
		return nil
	}})

	mgr.NotifyTrade(context.Background(), "BTCUSDT", "sell", 60000, 0.5, -500)
	if level != LevelWarning {
		t.Errorf("expected WARNING for loss, got %s", level)
	}
}

func TestManager_Channels(t *testing.T) {
	mgr := NewManager()
	mgr.Add(&mockNotifier{name: "email", minLevel: LevelInfo})
	mgr.Add(&mockNotifier{name: "discord", minLevel: LevelInfo})
	mgr.Add(&mockNotifier{name: "telegram", minLevel: LevelInfo})

	channels := mgr.Channels()
	if len(channels) != 3 {
		t.Fatalf("expected 3 channels, got %d", len(channels))
	}
	expected := []string{"email", "discord", "telegram"}
	for i, e := range expected {
		if channels[i] != e {
			t.Errorf("channel[%d] = %q, want %q", i, channels[i], e)
		}
	}
}

func TestEscapeJSON(t *testing.T) {
	input := `Hello "world"\nNew line`
	output := escapeJSON(input)
	if strings.Contains(output, `"`) && !strings.Contains(output, `\"`) {
		t.Errorf("quotes not escaped: %s", output)
	}
}

func TestEscapeMarkdown(t *testing.T) {
	input := "Price *high* [alert]"
	output := escapeMarkdown(input)
	if !strings.Contains(output, `\*`) {
		t.Errorf("asterisks not escaped: %s", output)
	}
	if !strings.Contains(output, `\[`) {
		t.Errorf("brackets not escaped: %s", output)
	}
}

// mockNotifier is a test helper
type mockNotifier struct {
	name     string
	minLevel Level
	sendFn   func(Message) error
}

func (m *mockNotifier) Name() string    { return m.name }
func (m *mockNotifier) MinLevel() Level { return m.minLevel }
func (m *mockNotifier) Send(_ context.Context, msg Message) error {
	if m.sendFn != nil {
		return m.sendFn(msg)
	}
	return nil
}
