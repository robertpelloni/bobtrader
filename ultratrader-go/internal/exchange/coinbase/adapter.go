package coinbase

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

const defaultBaseURL = "https://api.coinbase.com/api/v3/brokerage"

type Config struct {
	APIKey    string `json:"api_key"`
	SecretKey string `json:"secret_key"`
	BaseURL   string `json:"base_url"`
}

type Adapter struct {
	config     Config
	httpClient *http.Client
	baseURL    string
}

func New(cfg Config) *Adapter {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Adapter{
		config:     cfg,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
	}
}

func (a *Adapter) Name() string { return "coinbase" }

func (a *Adapter) Capabilities() []exchange.Capability {
	return []exchange.Capability{
		exchange.CapabilitySpot,
		exchange.CapabilityBalances,
		exchange.CapabilityOrders,
		exchange.CapabilityTickers,
	}
}

func (a *Adapter) ListMarkets(ctx context.Context) ([]exchange.Market, error) {
	var resp struct {
		Products []struct {
			ProductID      string `json:"product_id"`
			BaseCurrencyID string `json:"base_currency_id"`
			QuoteCurrencyID string `json:"quote_currency_id"`
			IsDisabled     bool   `json:"is_disabled"`
		} `json:"products"`
	}

	if err := a.doRequest(ctx, "GET", "/products", nil, &resp); err != nil {
		return nil, err
	}

	var markets []exchange.Market
	for _, p := range resp.Products {
		if p.IsDisabled {
			continue
		}
		markets = append(markets, exchange.Market{
			Symbol:     p.ProductID,
			BaseAsset:  p.BaseCurrencyID,
			QuoteAsset: p.QuoteCurrencyID,
		})
	}
	return markets, nil
}

func (a *Adapter) Balances(ctx context.Context) ([]exchange.Balance, error) {
	var resp struct {
		Accounts []struct {
			Currency string `json:"currency"`
			AvailableBalance struct {
				Value string `json:"value"`
			} `json:"available_balance"`
			Hold struct {
				Value string `json:"value"`
			} `json:"hold"`
		} `json:"accounts"`
	}

	if err := a.doRequest(ctx, "GET", "/accounts", nil, &resp); err != nil {
		return nil, err
	}

	var balances []exchange.Balance
	for _, acc := range resp.Accounts {
		val, _ := strconv.ParseFloat(acc.AvailableBalance.Value, 64)
		hold, _ := strconv.ParseFloat(acc.Hold.Value, 64)
		if val > 0 || hold > 0 {
			balances = append(balances, exchange.Balance{
				Asset:  acc.Currency,
				Free:   acc.AvailableBalance.Value,
				Locked: acc.Hold.Value,
			})
		}
	}
	return balances, nil
}

func (a *Adapter) PlaceOrder(ctx context.Context, req exchange.OrderRequest) (exchange.Order, error) {
	orderID := fmt.Sprintf("cb-%d", time.Now().UnixNano())

	params := map[string]interface{}{
		"client_order_id": orderID,
		"product_id":      req.Symbol,
		"side":            strings.ToUpper(string(req.Side)),
		"order_configuration": map[string]interface{}{
			"market_market_ioc": map[string]interface{}{
				"base_size": req.Quantity,
			},
		},
	}

	// This is a simplified market order implementation for the demo
	var resp struct {
		OrderID string `json:"order_id"`
	}

	if err := a.doRequest(ctx, "POST", "/orders", params, &resp); err != nil {
		return exchange.Order{}, err
	}

	return exchange.Order{
		ID:        resp.OrderID,
		Symbol:    req.Symbol,
		Side:      req.Side,
		Type:      req.Type,
		Status:    exchange.StatusOpen,
		Quantity:  req.Quantity,
		Timestamp: time.Now(),
	}, nil
}

func (a *Adapter) GetTickerPrice(ctx context.Context, symbol string) (string, error) {
	var resp struct {
		Price string `json:"price"`
	}
	if err := a.doRequest(ctx, "GET", "/products/"+symbol+"/ticker", nil, &resp); err != nil {
		return "", err
	}
	return resp.Price, nil
}

func (a *Adapter) sign(req *http.Request, method, path, body string) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	message := timestamp + method + path + body

	h := hmac.New(sha256.New, []byte(a.config.SecretKey))
	h.Write([]byte(message))
	signature := hex.EncodeToString(h.Sum(nil))

	req.Header.Set("CB-ACCESS-KEY", a.config.APIKey)
	req.Header.Set("CB-ACCESS-SIGN", signature)
	req.Header.Set("CB-ACCESS-TIMESTAMP", timestamp)
}

func (a *Adapter) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	u, _ := url.Parse(a.baseURL + path)
	var bodyReader io.Reader
	bodyStr := ""

	if body != nil {
		b, _ := json.Marshal(body)
		bodyStr = string(b)
		bodyReader = strings.NewReader(bodyStr)
	}

	req, _ := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	req.Header.Set("Content-Type", "application/json")

	if a.config.APIKey != "" {
		a.sign(req, method, path, bodyStr)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("coinbase error %d: %s", resp.StatusCode, string(b))
	}

	return json.Unmarshal(b, result)
}
