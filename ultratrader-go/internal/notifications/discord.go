package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type DiscordProvider struct {
	webhookURL string
	client     *http.Client
}

func NewDiscordProvider(webhookURL string) *DiscordProvider {
	return &DiscordProvider{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 5 * time.Second},
	}
}

func (d *DiscordProvider) Name() string { return "discord" }

func (d *DiscordProvider) Send(ctx context.Context, n Notification) error {
	if d.webhookURL == "" {
		return nil
	}

	color := 0x448aff // Blue (info)
	switch n.Level {
	case Trade:
		color = 0x00e676 // Green
	case Alert:
		color = 0xffab40 // Orange
	case Critical:
		color = 0xff5252 // Red
	}

	payload := map[string]any{
		"embeds": []map[string]any{
			{
				"title":       fmt.Sprintf("[%s] %s", n.Level, n.Source),
				"description": n.Message,
				"color":       color,
				"timestamp":   time.Now().Format(time.RFC3339),
			},
		},
	}

	b, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", d.webhookURL, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("discord status: %d", resp.StatusCode)
	}

	return nil
}
