package binance

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
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange/ratelimit"
)

const defaultBaseURL = "https://api.binance.com"

type Config struct {
	APIKey    string `json:"api_key"`
	SecretKey string `json:"secret_key"`
	BaseURL   string `json:"base_url"`
	Testnet   bool   `json:"testnet"`
}

type Adapter struct {
	config     Config
	httpClient *http.Client
	baseURL    string
	limiter    *ratelimit.Limiter
	orderLimit *ratelimit.Limiter
}

func New(cfg Config) *Adapter {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		if cfg.Testnet {
			baseURL = "https://testnet.binance.vision"
		} else {
			baseURL = defaultBaseURL
		}
	}
	return &Adapter{
		config:     cfg,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
		limiter:    ratelimit.BinanceSpotLimiter(),
		orderLimit: ratelimit.BinanceOrderLimiter(),
	}
}

func (a *Adapter) Name() string { return "binance" }

func (a *Adapter) SetBaseURL(url string) { a.baseURL = url }

func (a *Adapter) IsTestnet() bool { return a.config.Testnet }

func (a *Adapter) Capabilities() []exchange.Capability {
	return []exchange.Capability{
		exchange.CapabilitySpot,
		exchange.CapabilityBalances,
		exchange.CapabilityOrders,
		exchange.CapabilityCandles,
		exchange.CapabilityTickers,
	}
}

// ListMarkets fetches exchange info from Binance and returns spot markets.
func (a *Adapter) ListMarkets(ctx context.Context) ([]exchange.Market, error) {
	var resp struct {
		Symbols []struct {
			Symbol             string `json:"symbol"`
			BaseAsset          string `json:"baseAsset"`
			QuoteAsset         string `json:"quoteAsset"`
			BaseAssetPrecision int    `json:"baseAssetPrecision"`
			QuotePrecision     int    `json:"quotePrecision"`
			Status             string `json:"status"`
		} `json:"symbols"`
	}
	if err := a.publicGet(ctx, "/api/v3/exchangeInfo", nil, &resp); err != nil {
		return nil, fmt.Errorf("get exchange info: %w", err)
	}

	var markets []exchange.Market
	for _, s := range resp.Symbols {
		if s.Status != "TRADING" {
			continue
		}
		markets = append(markets, exchange.Market{
			Symbol:        s.Symbol,
			BaseAsset:     s.BaseAsset,
			QuoteAsset:    s.QuoteAsset,
			PriceScale:    s.QuotePrecision,
			QuantityScale: s.BaseAssetPrecision,
		})
	}
	return markets, nil
}

// Balances fetches spot account balances via signed endpoint.
func (a *Adapter) Balances(ctx context.Context) ([]exchange.Balance, error) {
	if a.config.APIKey == "" {
		return nil, fmt.Errorf("api key required for balances")
	}

	var resp struct {
		Balances []struct {
			Asset  string `json:"asset"`
			Free   string `json:"free"`
			Locked string `json:"locked"`
		} `json:"balances"`
	}
	if err := a.signedGet(ctx, "/api/v3/account", nil, &resp); err != nil {
		return nil, fmt.Errorf("get account: %w", err)
	}

	var balances []exchange.Balance
	for _, b := range resp.Balances {
		free, _ := strconv.ParseFloat(b.Free, 64)
		locked, _ := strconv.ParseFloat(b.Locked, 64)
		if free > 0 || locked > 0 {
			balances = append(balances, exchange.Balance{
				Asset:  b.Asset,
				Free:   b.Free,
				Locked: b.Locked,
			})
		}
	}
	return balances, nil
}

// PlaceOrder places a new order on Binance with rate limiting.
func (a *Adapter) PlaceOrder(ctx context.Context, request exchange.OrderRequest) (exchange.Order, error) {
	if a.config.APIKey == "" {
		return exchange.Order{}, fmt.Errorf("api key required for orders")
	}

	// Order-specific rate limiting
	if err := a.orderLimit.Wait(ctx); err != nil {
		return exchange.Order{}, fmt.Errorf("order rate limit wait cancelled: %w", err)
	}

	params := url.Values{}
	params.Set("symbol", request.Symbol)
	params.Set("side", strings.ToUpper(string(request.Side)))
	params.Set("type", strings.ToUpper(string(request.Type)))
	params.Set("quantity", request.Quantity)
	if request.Type == exchange.LimitOrder {
		if request.Price == "" {
			return exchange.Order{}, fmt.Errorf("price required for limit orders")
		}
		params.Set("price", request.Price)
		params.Set("timeInForce", "GTC")
	}

	var resp struct {
		Symbol        string `json:"symbol"`
		OrderID       int64  `json:"orderId"`
		ClientOrderID string `json:"clientOrderId"`
		TransactTime  int64  `json:"transactTime"`
		Price         string `json:"price"`
		OrigQty       string `json:"origQty"`
		ExecutedQty   string `json:"executedQty"`
		Status        string `json:"status"`
		Type          string `json:"type"`
		Side          string `json:"side"`
	}
	if err := a.signedPost(ctx, "/api/v3/order", params, &resp); err != nil {
		return exchange.Order{}, fmt.Errorf("place order: %w", err)
	}

	return exchange.Order{
		ID:       strconv.FormatInt(resp.OrderID, 10),
		Symbol:   resp.Symbol,
		Side:     request.Side,
		Type:     request.Type,
		Status:   resp.Status,
		Quantity: resp.ExecutedQty,
		Price:    resp.Price,
	}, nil
}

// QueryOrder fetches the current status of an order from Binance.
func (a *Adapter) QueryOrder(ctx context.Context, symbol, orderID string) (OrderStatus, error) {
	if a.config.APIKey == "" {
		return OrderStatus{}, fmt.Errorf("api key required for order queries")
	}

	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("orderId", orderID)

	var resp struct {
		Symbol        string `json:"symbol"`
		OrderID       int64  `json:"orderId"`
		ClientOrderID string `json:"clientOrderId"`
		Price         string `json:"price"`
		OrigQty       string `json:"origQty"`
		ExecutedQty   string `json:"executedQty"`
		Status        string `json:"status"`
		Type          string `json:"type"`
		Side          string `json:"side"`
		Time          int64  `json:"time"`
	}
	if err := a.signedGet(ctx, "/api/v3/order", params, &resp); err != nil {
		return OrderStatus{}, fmt.Errorf("query order: %w", err)
	}

	return OrderStatus{
		ID:              strconv.FormatInt(resp.OrderID, 10),
		Symbol:          resp.Symbol,
		Side:            exchange.OrderSide(strings.ToLower(resp.Side)),
		Type:            exchange.OrderType(strings.ToLower(resp.Type)),
		Status:          resp.Status,
		Quantity:        resp.OrigQty,
		ExecutedQty:     resp.ExecutedQty,
		Price:           resp.Price,
		TransactionTime: time.UnixMilli(resp.Time),
	}, nil
}

// OrderStatus represents the current state of an order on Binance.
type OrderStatus struct {
	ID              string
	Symbol          string
	Side            exchange.OrderSide
	Type            exchange.OrderType
	Status          string
	Quantity        string
	ExecutedQty     string
	Price           string
	TransactionTime time.Time
}
func (a *Adapter) GetTickerPrice(ctx context.Context, symbol string) (string, error) {
	params := url.Values{}
	params.Set("symbol", symbol)

	var resp struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	if err := a.publicGet(ctx, "/api/v3/ticker/price", params, &resp); err != nil {
		return "", fmt.Errorf("get ticker price: %w", err)
	}
	return resp.Price, nil
}

// GetKlines fetches candle/kline data for a symbol.
func (a *Adapter) GetKlines(ctx context.Context, symbol, interval string, limit int) ([]Kline, error) {
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("interval", interval)
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	var raw [][]interface{}
	if err := a.publicGet(ctx, "/api/v3/klines", params, &raw); err != nil {
		return nil, fmt.Errorf("get klines: %w", err)
	}

	var klines []Kline
	for _, item := range raw {
		if len(item) < 12 {
			continue
		}
		k := Kline{
			OpenTime:  int64(item[0].(float64)),
			Open:      item[1].(string),
			High:      item[2].(string),
			Low:       item[3].(string),
			Close:     item[4].(string),
			Volume:    item[5].(string),
			CloseTime: int64(item[6].(float64)),
		}
		klines = append(klines, k)
	}
	return klines, nil
}

type Kline struct {
	OpenTime  int64
	Open      string
	High      string
	Low       string
	Close     string
	Volume    string
	CloseTime int64
}

// HTTP helpers

func (a *Adapter) publicGet(ctx context.Context, path string, params url.Values, result interface{}) error {
	return a.doRequest(ctx, http.MethodGet, path, params, false, result)
}

func (a *Adapter) signedGet(ctx context.Context, path string, params url.Values, result interface{}) error {
	return a.doRequest(ctx, http.MethodGet, path, params, true, result)
}

func (a *Adapter) signedPost(ctx context.Context, path string, params url.Values, result interface{}) error {
	return a.doRequest(ctx, http.MethodPost, path, params, true, result)
}

func (a *Adapter) doRequest(ctx context.Context, method, path string, params url.Values, signed bool, result interface{}) error {
	// Rate limiting: wait for an available token
	if err := a.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit wait cancelled: %w", err)
	}
	if params == nil {
		params = url.Values{}
	}

	if signed {
		params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
		params.Set("recvWindow", "5000")
		signature := a.sign(params.Encode())
		params.Set("signature", signature)
	}

	u := a.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	if a.config.APIKey != "" {
		req.Header.Set("X-MBX-APIKEY", a.config.APIKey)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}
		if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Msg != "" {
			return fmt.Errorf("binance api error %d: %s", apiErr.Code, apiErr.Msg)
		}
		return fmt.Errorf("http %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}
	return nil
}

func (a *Adapter) sign(payload string) string {
	if a.config.SecretKey == "" {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(a.config.SecretKey))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
