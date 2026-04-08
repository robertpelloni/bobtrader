package binance

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

func TestAdapter_Name(t *testing.T) {
	a := New(Config{})
	if a.Name() != "binance" {
		t.Errorf("expected name binance, got %s", a.Name())
	}
}

func TestAdapter_Capabilities(t *testing.T) {
	a := New(Config{})
	caps := a.Capabilities()
	if len(caps) == 0 {
		t.Errorf("expected non-empty capabilities")
	}
}

func TestAdapter_GetTickerPrice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/ticker/price" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("symbol") != "BTCUSDT" {
			t.Errorf("unexpected symbol: %s", r.URL.Query().Get("symbol"))
		}
		resp := map[string]string{"symbol": "BTCUSDT", "price": "65000.00"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	a := New(Config{})
	a.baseURL = server.URL

	price, err := a.GetTickerPrice(context.Background(), "BTCUSDT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if price != "65000.00" {
		t.Errorf("expected price 65000.00, got %s", price)
	}
}

func TestAdapter_GetKlines(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("symbol") != "BTCUSDT" {
			t.Errorf("unexpected symbol: %s", r.URL.Query().Get("symbol"))
		}
		if r.URL.Query().Get("interval") != "1m" {
			t.Errorf("unexpected interval: %s", r.URL.Query().Get("interval"))
		}
		resp := [][]interface{}{
			{1609459200000, "65000.00", "65100.00", "64900.00", "65050.00", "100.5", 1609459259999, "6528250.00", 150, "50.25", "3264125.00", "0"},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	a := New(Config{})
	a.baseURL = server.URL

	klines, err := a.GetKlines(context.Background(), "BTCUSDT", "1m", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(klines) != 1 {
		t.Fatalf("expected 1 kline, got %d", len(klines))
	}
	if klines[0].Open != "65000.00" {
		t.Errorf("expected open 65000.00, got %s", klines[0].Open)
	}
	if klines[0].Close != "65050.00" {
		t.Errorf("expected close 65050.00, got %s", klines[0].Close)
	}
	if klines[0].High != "65100.00" {
		t.Errorf("expected high 65100.00, got %s", klines[0].High)
	}
	if klines[0].Low != "64900.00" {
		t.Errorf("expected low 64900.00, got %s", klines[0].Low)
	}
}

func TestAdapter_ListMarkets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"symbols": []map[string]interface{}{
				{"symbol": "BTCUSDT", "baseAsset": "BTC", "quoteAsset": "USDT", "baseAssetPrecision": 8, "quotePrecision": 2, "status": "TRADING"},
				{"symbol": "ETHUSDT", "baseAsset": "ETH", "quoteAsset": "USDT", "baseAssetPrecision": 8, "quotePrecision": 2, "status": "TRADING"},
				{"symbol": "DELISTED", "baseAsset": "DEL", "quoteAsset": "USDT", "baseAssetPrecision": 8, "quotePrecision": 2, "status": "BREAK"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	a := New(Config{})
	a.baseURL = server.URL

	markets, err := a.ListMarkets(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(markets) != 2 {
		t.Errorf("expected 2 active markets (excluding DELISTED), got %d", len(markets))
	}
}

func TestAdapter_Signing(t *testing.T) {
	a := New(Config{SecretKey: "testsecret"})
	sig := a.sign("symbol=BTCUSDT&timestamp=1234567890000")
	if sig == "" {
		t.Errorf("expected non-empty signature")
	}
	sig2 := a.sign("symbol=BTCUSDT&timestamp=1234567890000")
	if sig != sig2 {
		t.Errorf("signature should be deterministic")
	}
}

func TestAdapter_TestnetURL(t *testing.T) {
	a := New(Config{Testnet: true})
	if a.baseURL != "https://testnet.binance.vision" {
		t.Errorf("expected testnet URL, got %s", a.baseURL)
	}
}

func TestAdapter_Balances_NoAPIKey(t *testing.T) {
	a := New(Config{})
	_, err := a.Balances(context.Background())
	if err == nil {
		t.Errorf("expected error when API key is empty")
	}
}

func TestAdapter_PlaceOrder_NoAPIKey(t *testing.T) {
	a := New(Config{})
	_, err := a.PlaceOrder(context.Background(), exchange.OrderRequest{
		Symbol: "BTCUSDT", Side: exchange.Buy, Type: exchange.MarketOrder, Quantity: "0.001",
	})
	if err == nil {
		t.Errorf("expected error when API key is empty")
	}
}
