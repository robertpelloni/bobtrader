package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/config"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange/binance"
)

func main() {
	cfgPath := "config/live-binance-testnet.json"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	// Find the binance account
	var acct *config.AccountConfig
	for i := range cfg.Accounts {
		if cfg.Accounts[i].Exchange == "binance" && cfg.Accounts[i].Enabled {
			acct = &cfg.Accounts[i]
			break
		}
	}
	if acct == nil {
		fmt.Fprintln(os.Stderr, "no binance account found in config")
		os.Exit(1)
	}

	fmt.Printf("Account: %s (%s)\n", acct.Name, acct.ID)
	fmt.Printf("API Key: %s...%s\n", acct.APIKey[:8], acct.APIKey[len(acct.APIKey)-4:])
	fmt.Printf("Testnet: %v\n\n", acct.Testnet)

	// Create adapter with production URL (your keys are for Binance.US, not testnet)
	adapter := binance.New(binance.Config{
		APIKey:    acct.APIKey,
		SecretKey: acct.SecretKey,
		Testnet:   false, // Your keys are real Binance.US keys
	})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// ── Test 1: Fetch BTC price ───────────────────────────────
	fmt.Println("═══ TEST 1: Fetch BTC/USDT Price ═══")
	ticker, err := adapter.GetTickerPrice(ctx, "BTCUSDT")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	} else {
		fmt.Printf("✅ BTC/USDT = $%s\n\n", ticker)
	}

	// ── Test 2: Fetch ETH price ───────────────────────────────
	fmt.Println("═══ TEST 2: Fetch ETH/USDT Price ═══")
	ticker, err = adapter.GetTickerPrice(ctx, "ETHUSDT")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	} else {
		fmt.Printf("✅ ETH/USDT = $%s\n\n", ticker)
	}

	// ── Test 3: Fetch candles ─────────────────────────────────
	fmt.Println("═══ TEST 3: Fetch BTC Candles ═══")
	candles, err := adapter.GetKlines(ctx, "BTCUSDT", "1m", 5)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	} else {
		fmt.Printf("✅ Got %d candles:\n", len(candles))
		for i, c := range candles {
			fmt.Printf("   [%d] Close=%s  Volume=%s\n", i+1, c.Close, c.Volume)
		}
		fmt.Println()
	}

	// ── Test 4: Fetch account balances ────────────────────────
	fmt.Println("═══ TEST 4: Fetch Account Balances ═══")
	balances, err := adapter.Balances(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	} else {
		nonZero := 0
		for _, b := range balances {
			if b.Free != "0" && b.Free != "0.00" && b.Free != "" {
				nonZero++
				fmt.Printf("   💰 %-8s  free=%s  locked=%s\n", b.Asset, b.Free, b.Locked)
			}
		}
		if nonZero == 0 {
			fmt.Println("   (no non-zero balances found)")
		}
		fmt.Printf("\n✅ Total assets returned: %d (non-zero: %d)\n\n", len(balances), nonZero)
	}

	// ── Summary ───────────────────────────────────────────────
	fmt.Println("═══ SUMMARY ═══")
	fmt.Println("If all 4 tests passed ✅, your API keys work correctly.")
	fmt.Println("The bot can read prices, candles, and your account balance.")
	fmt.Println()
	fmt.Println("⚠️  NO orders were placed. This was read-only.")
	fmt.Println("⚠️  To actually trade, run: go run ./cmd/ultratrader --config config/live-binance-testnet.json")
	fmt.Println()
	fmt.Println("NOTE: Your config says testnet=true but your keys are for Binance.US (production).")
	fmt.Println("To trade with real money, use: config/live-binance-production.json")
}
