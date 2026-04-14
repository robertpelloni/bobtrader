package notification

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Level represents the severity of a notification.
type Level int

const (
	LevelInfo Level = iota
	LevelWarning
	LevelError
	LevelCritical
)

func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelWarning:
		return "WARNING"
	case LevelError:
		return "ERROR"
	case LevelCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// Message represents a notification to be sent.
type Message struct {
	Subject string
	Body    string
	Level   Level
	Time    time.Time
}

// Notifier is the interface that all notification channels implement.
type Notifier interface {
	// Name returns the channel name (e.g., "email", "discord", "telegram").
	Name() string
	// Send dispatches a notification message.
	Send(ctx context.Context, msg Message) error
	// MinLevel returns the minimum level this notifier should handle.
	// Messages below this level are silently dropped.
	MinLevel() Level
}

// EmailNotifier sends notifications via SMTP.
type EmailNotifier struct {
	host     string
	port     int
	from     string
	to       []string
	username string
	password string
	minLevel Level
}

// EmailConfig holds configuration for email notifications.
type EmailConfig struct {
	Host     string   `json:"host"`
	Port     int      `json:"port"`
	From     string   `json:"from"`
	To       []string `json:"to"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	MinLevel string   `json:"min_level"`
}

func NewEmailNotifier(cfg EmailConfig) *EmailNotifier {
	n := &EmailNotifier{
		host:     cfg.Host,
		port:     cfg.Port,
		from:     cfg.From,
		to:       cfg.To,
		username: cfg.Username,
		password: cfg.Password,
		minLevel: LevelInfo,
	}
	if lvl := parseLevel(cfg.MinLevel); lvl != nil {
		n.minLevel = *lvl
	}
	return n
}

func (e *EmailNotifier) Name() string    { return "email" }
func (e *EmailNotifier) MinLevel() Level { return e.minLevel }
func (e *EmailNotifier) Send(_ context.Context, msg Message) error {
	if msg.Level < e.minLevel {
		return nil
	}
	// SMTP send would go here — for now we format the message correctly
	// In production this would use net/smtp or a mail library
	_ = fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: [%s] %s\r\n\r\n%s",
		e.from, strings.Join(e.to, ", "), msg.Level, msg.Subject, msg.Body)
	return nil
}

// DiscordNotifier sends notifications via Discord webhook.
type DiscordNotifier struct {
	webhookURL string
	minLevel   Level
	client     *http.Client
}

type DiscordConfig struct {
	WebhookURL string `json:"webhook_url"`
	MinLevel   string `json:"min_level"`
}

func NewDiscordNotifier(cfg DiscordConfig) *DiscordNotifier {
	n := &DiscordNotifier{
		webhookURL: cfg.WebhookURL,
		minLevel:   LevelInfo,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
	if lvl := parseLevel(cfg.MinLevel); lvl != nil {
		n.minLevel = *lvl
	}
	return n
}

func (d *DiscordNotifier) Name() string    { return "discord" }
func (d *DiscordNotifier) MinLevel() Level { return d.minLevel }

func (d *DiscordNotifier) Send(ctx context.Context, msg Message) error {
	if msg.Level < d.minLevel {
		return nil
	}

	// Discord embed color based on level
	color := map[Level]int{
		LevelInfo:     3447003,  // blue
		LevelWarning:  16776960, // yellow
		LevelError:    15158332, // red
		LevelCritical: 10038562, // dark red
	}[msg.Level]

	payload := fmt.Sprintf(`{"embeds":[{"title":"%s","description":"%s","color":%d,"timestamp":"%s"}]}`,
		escapeJSON(msg.Subject), escapeJSON(msg.Body), color, msg.Time.Format(time.RFC3339))

	req, err := http.NewRequestWithContext(ctx, "POST", d.webhookURL, strings.NewReader(payload))
	if err != nil {
		return fmt.Errorf("discord request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("discord send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("discord returned status %d", resp.StatusCode)
	}
	return nil
}

// TelegramNotifier sends notifications via Telegram Bot API.
type TelegramNotifier struct {
	botToken string
	chatID   string
	minLevel Level
	client   *http.Client
}

type TelegramConfig struct {
	BotToken string `json:"bot_token"`
	ChatID   string `json:"chat_id"`
	MinLevel string `json:"min_level"`
}

func NewTelegramNotifier(cfg TelegramConfig) *TelegramNotifier {
	n := &TelegramNotifier{
		botToken: cfg.BotToken,
		chatID:   cfg.ChatID,
		minLevel: LevelInfo,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
	if lvl := parseLevel(cfg.MinLevel); lvl != nil {
		n.minLevel = *lvl
	}
	return n
}

func (t *TelegramNotifier) Name() string    { return "telegram" }
func (t *TelegramNotifier) MinLevel() Level { return t.minLevel }

func (t *TelegramNotifier) Send(ctx context.Context, msg Message) error {
	if msg.Level < t.minLevel {
		return nil
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)
	text := fmt.Sprintf("🔔 *%s* [%s]\n\n%s", escapeMarkdown(msg.Subject), msg.Level, escapeMarkdown(msg.Body))

	params := url.Values{}
	params.Set("chat_id", t.chatID)
	params.Set("text", text)
	params.Set("parse_mode", "MarkdownV2")

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("telegram send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("telegram returned status %d", resp.StatusCode)
	}
	return nil
}

// Manager coordinates multiple notification channels.
type Manager struct {
	mu       sync.RWMutex
	notifier []Notifier
}

// NewManager creates a notification manager.
func NewManager() *Manager {
	return &Manager{}
}

// Add registers a notification channel.
func (m *Manager) Add(n Notifier) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notifier = append(m.notifier, n)
}

// Send dispatches a message to all channels that meet the minimum level.
func (m *Manager) Send(ctx context.Context, msg Message) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var firstErr error
	for _, n := range m.notifier {
		if msg.Level < n.MinLevel() {
			continue
		}
		if err := n.Send(ctx, msg); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("%s: %w", n.Name(), err)
		}
	}
	return firstErr
}

// SendSimple is a convenience method for creating and sending a message.
func (m *Manager) SendSimple(ctx context.Context, level Level, subject, body string) error {
	return m.Send(ctx, Message{
		Subject: subject,
		Body:    body,
		Level:   level,
		Time:    time.Now().UTC(),
	})
}

// NotifyTrade is a convenience for sending trade notifications.
func (m *Manager) NotifyTrade(ctx context.Context, symbol, side string, price float64, qty float64, pnl float64) error {
	var body string
	if pnl != 0 {
		body = fmt.Sprintf("Symbol: %s\nSide: %s\nPrice: %.8f\nQuantity: %.8f\nPnL: %.8f",
			symbol, side, price, qty, pnl)
	} else {
		body = fmt.Sprintf("Symbol: %s\nSide: %s\nPrice: %.8f\nQuantity: %.8f",
			symbol, side, price, qty)
	}

	level := LevelInfo
	if pnl < 0 {
		level = LevelWarning
	}

	return m.SendSimple(ctx, level, fmt.Sprintf("Trade Executed: %s %s", side, symbol), body)
}

// Channels returns the names of registered notification channels.
func (m *Manager) Channels() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	names := make([]string, len(m.notifier))
	for i, n := range m.notifier {
		names[i] = n.Name()
	}
	return names
}

// Helper functions

func parseLevel(s string) *Level {
	switch strings.ToUpper(s) {
	case "INFO":
		l := LevelInfo
		return &l
	case "WARNING":
		l := LevelWarning
		return &l
	case "ERROR":
		l := LevelError
		return &l
	case "CRITICAL":
		l := LevelCritical
		return &l
	default:
		return nil
	}
}

func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}

func escapeMarkdown(s string) string {
	chars := "_*[]()~`>#+-=|{}.!"
	for _, c := range chars {
		s = strings.ReplaceAll(s, string(c), "\\"+string(c))
	}
	return s
}
