package kucoin

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
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

const defaultBaseURL = "https://api.kucoin.com"

type Config struct {
	APIKey     string `json:"api_key"`
	SecretKey  string `json:"secret_key"`
	Passphrase string `json:"passphrase"`
	BaseURL    string `json:"base_url"`
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

func (a *Adapter) Name() string { return "kucoin" }

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
		Code string `json:"code"`
		Data []struct {
			Symbol          string `json:"symbol"`
			BaseCurrency    string `json:"baseCurrency"`
			QuoteCurrency   string `json:"quoteCurrency"`
			BaseIncrement   string `json:"baseIncrement"`
			QuoteIncrement  string `json:"quoteIncrement"`
			PriceIncrement  string `json:"priceIncrement"`
			IsMarginEnabled bool   `json:"isMarginEnabled"`
			EnableTrading   bool   `json:"enableTrading"`
		} `json:"data"`
	}

	if err := a.publicGet(ctx, "/api/v1/symbols", nil, &resp); err != nil {
		return nil, err
	}

	var markets []exchange.Market
	for _, s := range resp.Data {
		if !s.EnableTrading {
			continue
		}
		markets = append(markets, exchange.Market{
			Symbol:     s.Symbol,
			BaseAsset:  s.BaseCurrency,
			QuoteAsset: s.QuoteCurrency,
		})
	}
	return markets, nil
}

func (a *Adapter) Balances(ctx context.Context) ([]exchange.Balance, error) {
	var resp struct {
		Code string `json:"code"`
		Data []struct {
			Currency  string `json:"currency"`
			Type      string `json:"type"`
			Balance   string `json:"balance"`
			Available string `json:"available"`
			Holds     string `json:"holds"`
		} `json:"data"`
	}

	if err := a.privateGet(ctx, "/api/v1/accounts", nil, &resp); err != nil {
		return nil, err
	}

	var balances []exchange.Balance
	for _, b := range resp.Data {
		if b.Type != "trade" {
			continue
		}
		val, _ := strconv.ParseFloat(b.Balance, 64)
		if val > 0 {
			balances = append(balances, exchange.Balance{
				Asset:  b.Currency,
				Free:   b.Available,
				Locked: b.Holds,
				Total:  b.Balance,
			})
		}
	}
	return balances, nil
}

func (a *Adapter) GetTickerPrice(ctx context.Context, symbol string) (string, error) {
	var resp struct {
		Code string `json:"code"`
		Data struct {
			Price string `json:"price"`
		} `json:"data"`
	}
	if err := a.publicGet(ctx, "/api/v1/market/orderbook/level1", url.Values{"symbol": {symbol}}, &resp); err != nil {
		return "", err
	}
	return resp.Data.Price, nil
}

func (a *Adapter) PlaceOrder(ctx context.Context, req exchange.OrderRequest) (exchange.Order, error) {
	params := map[string]interface{}{
		"clientOid":   fmt.Sprintf("%d", time.Now().UnixNano()),
		"side":        strings.ToLower(string(req.Side)),
		"symbol":      req.Symbol,
		"type":        strings.ToLower(string(req.Type)),
		"size":        req.Quantity,
	}

	if req.Type == exchange.LimitOrder {
		params["price"] = req.Price
	}

	var resp struct {
		Code string `json:"code"`
		Data struct {
			OrderId string `json:"orderId"`
		} `json:"data"`
	}

	if err := a.privatePost(ctx, "/api/v1/orders", params, &resp); err != nil {
		return exchange.Order{}, err
	}

	return exchange.Order{
		ID:        resp.Data.OrderId,
		Symbol:    req.Symbol,
		Side:      req.Side,
		Type:      req.Type,
		Status:    exchange.StatusOpen,
		Quantity:  req.Quantity,
		Timestamp: time.Now(),
	}, nil
}

// Helpers

func (a *Adapter) publicGet(ctx context.Context, path string, params url.Values, result interface{}) error {
	u := a.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
	return a.do(req, result)
}

func (a *Adapter) privateGet(ctx context.Context, path string, params url.Values, result interface{}) error {
	u := path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, _ := http.NewRequestWithContext(ctx, "GET", a.baseURL+u, nil)
	a.sign(req, "GET", u, "")
	return a.do(req, result)
}

func (a *Adapter) privatePost(ctx context.Context, path string, body map[string]interface{}, result interface{}) error {
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", a.baseURL+path, strings.NewReader(string(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	a.sign(req, "POST", path, string(jsonBody))
	return a.do(req, result)
}

func (a *Adapter) sign(req *http.Request, method, path, body string) {
	now := strconv.FormatInt(time.Now().UnixMilli(), 10)
	strForSign := now + method + path + body
	mac := hmac.New(sha256.New, []byte(a.config.SecretKey))
	mac.Write([]byte(strForSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	passMac := hmac.New(sha256.New, []byte(a.config.SecretKey))
	passMac.Write([]byte(a.config.Passphrase))
	passphrase := base64.StdEncoding.EncodeToString(passMac.Sum(nil))

	req.Header.Set("KC-API-KEY", a.config.APIKey)
	req.Header.Set("KC-API-SIGN", signature)
	req.Header.Set("KC-API-TIMESTAMP", now)
	req.Header.Set("KC-API-PASSPHRASE", passphrase)
	req.Header.Set("KC-API-KEY-VERSION", "2")
}

func (a *Adapter) do(req *http.Request, result interface{}) error {
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("http %d: %s", resp.StatusCode, string(b))
	}
	return json.Unmarshal(b, result)
}
