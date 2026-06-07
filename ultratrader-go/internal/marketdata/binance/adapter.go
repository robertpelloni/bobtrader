package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

// Adapter implements a robust Binance market data adapter, inspired by bbgo.
type Adapter struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

// NewAdapter creates a new Binance adapter.
func NewAdapter(apiKey string, isTestnet bool) *Adapter {
	baseURL := "https://api.binance.com"
	if isTestnet {
		baseURL = "https://testnet.binance.vision"
	}

	return &Adapter{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

type tickerPriceResponse struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

// GetPrice fetches the current price for a symbol with robust error handling.
func (a *Adapter) GetPrice(ctx context.Context, symbol string) (string, error) {
	url := fmt.Sprintf("%s/api/v3/ticker/price?symbol=%s", a.baseURL, symbol)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	if a.apiKey != "" {
		req.Header.Set("X-MBX-APIKEY", a.apiKey)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var ticker tickerPriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&ticker); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return ticker.Price, nil
}

// StreamTicks implements marketdata.StreamFeed (partial)
func (a *Adapter) StreamTicks(ctx context.Context, symbol string) (<-chan marketdata.Tick, error) {
	// Robust WebSocket implementation inspired by bbgo would go here.
	return nil, fmt.Errorf("not implemented")
}
