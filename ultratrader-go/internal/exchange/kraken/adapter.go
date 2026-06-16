package kraken

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
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

const defaultBaseURL = "https://api.kraken.com"

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

func (a *Adapter) Name() string { return "kraken" }

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
		Error []string `json:"error"`
		Result map[string]struct {
			Altname string `json:"altname"`
			Base    string `json:"base"`
			Quote   string `json:"quote"`
		} `json:"result"`
	}

	if err := a.publicGet(ctx, "/0/public/AssetPairs", nil, &resp); err != nil {
		return nil, err
	}

	var markets []exchange.Market
	for symbol, p := range resp.Result {
		markets = append(markets, exchange.Market{
			Symbol:     symbol,
			BaseAsset:  p.Base,
			QuoteAsset: p.Quote,
		})
	}
	return markets, nil
}

func (a *Adapter) Balances(ctx context.Context) ([]exchange.Balance, error) {
	var resp struct {
		Error []string           `json:"error"`
		Result map[string]string `json:"result"`
	}

	if err := a.privatePost(ctx, "/0/private/Balance", nil, &resp); err != nil {
		return nil, err
	}

	var balances []exchange.Balance
	for asset, total := range resp.Result {
		val, _ := strconv.ParseFloat(total, 64)
		if val > 0 {
			balances = append(balances, exchange.Balance{
				Asset: asset,
				Total: total,
				Free:  total, // Kraken simple balance doesn't split by default in this endpoint
			})
		}
	}
	return balances, nil
}

func (a *Adapter) PlaceOrder(ctx context.Context, req exchange.OrderRequest) (exchange.Order, error) {
	params := url.Values{}
	params.Set("pair", req.Symbol)
	params.Set("type", strings.ToLower(string(req.Side)))
	params.Set("ordertype", strings.ToLower(string(req.Type)))
	params.Set("volume", req.Quantity)
	if req.Type == exchange.LimitOrder {
		params.Set("price", req.Price)
	}

	var resp struct {
		Error []string `json:"error"`
		Result struct {
			TxID []string `json:"txid"`
		} `json:"result"`
	}

	if err := a.privatePost(ctx, "/0/private/AddOrder", params, &resp); err != nil {
		return exchange.Order{}, err
	}

	id := ""
	if len(resp.Result.TxID) > 0 {
		id = resp.Result.TxID[0]
	}

	return exchange.Order{
		ID:        id,
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
		Error []string `json:"error"`
		Result map[string]struct {
			C []string `json:"c"` // Last trade closed [price, volume]
		} `json:"result"`
	}
	if err := a.publicGet(ctx, "/0/public/Ticker", url.Values{"pair": {symbol}}, &resp); err != nil {
		return "", err
	}
	for _, data := range resp.Result {
		if len(data.C) > 0 {
			return data.C[0], nil
		}
	}
	return "", fmt.Errorf("ticker not found for %s", symbol)
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

func (a *Adapter) privatePost(ctx context.Context, path string, params url.Values, result interface{}) error {
	if params == nil {
		params = url.Values{}
	}
	nonce := fmt.Sprintf("%d", time.Now().UnixNano())
	params.Set("nonce", nonce)

	req, _ := http.NewRequestWithContext(ctx, "POST", a.baseURL+path, strings.NewReader(params.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	sig, err := a.sign(path, nonce, params.Encode())
	if err != nil {
		return err
	}
	req.Header.Set("API-Key", a.config.APIKey)
	req.Header.Set("API-Sign", sig)

	return a.do(req, result)
}

func (a *Adapter) sign(path, nonce, postData string) (string, error) {
	sha := sha256.New()
	sha.Write([]byte(nonce + postData))
	shaSum := sha.Sum(nil)

	secret, err := base64.StdEncoding.DecodeString(a.config.SecretKey)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha512.New, secret)
	mac.Write([]byte(path))
	mac.Write(shaSum)
	macSum := mac.Sum(nil)

	return base64.StdEncoding.EncodeToString(macSum), nil
}

func (a *Adapter) do(req *http.Request, result interface{}) error {
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("kraken error %d: %s", resp.StatusCode, string(b))
	}

	// Check for Kraken specific errors in result
	var baseResp struct {
		Error []string `json:"error"`
	}
	if err := json.Unmarshal(b, &baseResp); err == nil && len(baseResp.Error) > 0 {
		return fmt.Errorf("kraken api error: %s", strings.Join(baseResp.Error, ", "))
	}

	return json.Unmarshal(b, result)
}
