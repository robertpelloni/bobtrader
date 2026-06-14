package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange/binance"
)

func main() {
	// Load secrets
	secretsData, err := os.ReadFile("config/secrets/binance-production.json")
	if err != nil {
		fmt.Printf("❌ Failed to read secrets: %v\n", err)
		return
	}

	var secrets struct {
		APIKey    string `json:"api_key"`
		SecretKey string `json:"secret_key"`
		Testnet   bool   `json:"testnet"`
	}
	if err := json.Unmarshal(secretsData, &secrets); err != nil {
		fmt.Printf("❌ Failed to parse secrets: %v\n", err)
		return
	}

	fmt.Printf("API Key: %s...\n", secrets.APIKey[:15])

	// Create Binance adapter
	adapter := binance.New(binance.Config{
		APIKey:    secrets.APIKey,
		SecretKey: secrets.SecretKey,
		Testnet:   secrets.Testnet,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test 1: Get price
	fmt.Println("\n=== TEST 1: Get XRP Price ===")
	price, err := adapter.GetTickerPrice(ctx, "XRPUSDT")
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
	} else {
		fmt.Printf("✅ XRP Price: $%s\n", price)
	}

	// Test 2: Get balance
	fmt.Println("\n=== TEST 2: Get Balance ===")
	balances, err := adapter.Balances(ctx)
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
	} else {
		for _, b := range balances {
			if b.Free != "0" {
				fmt.Printf("💰 %s: free=%s locked=%s\n", b.Asset, b.Free, b.Locked)
			}
		}
	}

	// Test 3: Try to sell 1 XRP
	fmt.Println("\n=== TEST 3: Sell 1 XRP ===")
	order, err := adapter.PlaceOrder(ctx, exchange.OrderRequest{
		Symbol:   "XRPUSDT",
		Side:     exchange.Sell,
		Type:     exchange.MarketOrder,
		Quantity: "1",
	})
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
	} else {
		fmt.Printf("✅ Order placed!\n")
		fmt.Printf("   Order ID: %s\n", order.ID)
		fmt.Printf("   Status: %s\n", order.Status)
		fmt.Printf("   Price: $%s\n", order.Price)
		fmt.Printf("   Quantity: %s\n", order.Quantity)
	}
}
