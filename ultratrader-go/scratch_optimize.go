package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/backtest"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy/demo"
)

type RunResult struct {
	Params map[string]float64
	Result backtest.Result
	Score  float64
}

func fetchBinanceCandles(ctx context.Context, symbol, interval string, limit int) ([]marketdata.Candle, error) {
	url := fmt.Sprintf("https://api.binance.us/api/v3/klines?symbol=%s&interval=%s&limit=%d", symbol, interval, limit)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("http error %d: %s", resp.StatusCode, string(body))
	}

	var raw [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	candles := make([]marketdata.Candle, len(raw))
	for i, r := range raw {
		openTimeMs := int64(r[0].(float64))
		t := time.Unix(0, openTimeMs*int64(time.Millisecond))

		candles[i] = marketdata.Candle{
			Symbol:    symbol,
			Open:      r[1].(string),
			High:      r[2].(string),
			Low:       r[3].(string),
			Close:     r[4].(string),
			Volume:    r[5].(string),
			Timestamp: t,
		}
	}

	return candles, nil
}

func generateGrid(paramGrid map[string][]float64) []map[string]float64 {
	var keys []string
	for k := range paramGrid {
		keys = append(keys, k)
	}

	var results []map[string]float64
	var helper func(keyIndex int, currentMap map[string]float64)

	helper = func(keyIndex int, currentMap map[string]float64) {
		if keyIndex == len(keys) {
			copyMap := make(map[string]float64)
			for k, v := range currentMap {
				copyMap[k] = v
			}
			results = append(results, copyMap)
			return
		}

		key := keys[keyIndex]
		values := paramGrid[key]
		for _, v := range values {
			currentMap[key] = v
			helper(keyIndex+1, currentMap)
		}
	}

	helper(0, make(map[string]float64))
	return results
}

func main() {
	ctx := context.Background()

	// Whitelisted high-volatility symbols
	symbols := []string{"SOLUSDT", "DOGEUSDT", "XRPUSDT"}

	// Grid search parameter ranges
	paramGrid := map[string][]float64{
		"rsi_period":     {10.0, 14.0, 18.0},
		"rsi_oversold":   {20.0, 25.0, 30.0},
		"rsi_overbought": {70.0, 75.0, 80.0},
		"bb_period":      {15.0, 20.0, 25.0},
		"bb_multiplier":  {1.5, 2.0, 2.5},
	}

	for _, symbol := range symbols {
		fmt.Printf("\n=========================================\n")
		fmt.Printf("Optimizing Strategy for Symbol: %s\n", symbol)
		fmt.Printf("=========================================\n")

		candles, err := fetchBinanceCandles(ctx, symbol, "15m", 1000)
		if err != nil {
			fmt.Printf("Error fetching candles for %s: %v\n", symbol, err)
			continue
		}

		fmt.Printf("Successfully fetched %d candles of 15m\n", len(candles))

		// Convert candles to ticks
		ticks := make([]marketdata.Tick, len(candles))
		for i, c := range candles {
			ticks[i] = marketdata.Tick{
				Symbol:    c.Symbol,
				Price:     c.Close,
				Timestamp: c.Timestamp,
			}
		}
		history := backtest.NewMemoryHistory(ticks)

		perms := generateGrid(paramGrid)
		if len(perms) == 0 {
			fmt.Printf("Parameter grid is empty\n")
			continue
		}

		resultsCh := make(chan RunResult, len(perms))
		var wg sync.WaitGroup

		workers := runtime.NumCPU()
		jobs := make(chan map[string]float64, len(perms))

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for p := range jobs {
					rsiPeriod := int(p["rsi_period"])
					rsiOversold := p["rsi_oversold"]
					rsiOverbought := p["rsi_overbought"]
					bbPeriod := int(p["bb_period"])
					bbMultiplier := p["bb_multiplier"]

					strat := demo.NewRSIBollingerComposite(
						"paper-main",
						symbol,
						"100.0", // large enough base quantity to trade
						rsiPeriod,
						rsiOversold,
						rsiOverbought,
						bbPeriod,
						bbMultiplier,
						nil,
					)

					opts := backtest.EmulatorOptions{
						MakerFeeRate: 0.001,
						TakerFeeRate: 0.001,
						SlippageRate: 0.0005,
					}

					eng := backtest.NewEngineWithOptions(strat, 10000.0, opts)
					res, err := eng.RunTicks(ctx, history)
					if err != nil {
						continue
					}

					// Score: maximize realized PnL, penalize setups with fewer than 5 trades
					score := res.RealizedPnL
					if res.TotalTrades < 5 {
						score = -10000.0 + float64(res.TotalTrades)
					}

					resultsCh <- RunResult{
						Params: p,
						Result: res,
						Score:  score,
					}
				}
			}()
		}

		for _, p := range perms {
			jobs <- p
		}
		close(jobs)

		wg.Wait()
		close(resultsCh)

		var results []RunResult
		for r := range resultsCh {
			results = append(results, r)
		}

		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})

		if len(results) == 0 {
			fmt.Printf("No results evaluated for %s\n", symbol)
			continue
		}

		// Print top 3 parameter sets
		fmt.Printf("Top 3 Parameter Sets for %s:\n", symbol)
		for i := 0; i < 3 && i < len(results); i++ {
			r := results[i]
			fmt.Printf("Rank %d (Score: %.2f, PnL: %.2f, Trades: %d):\n", i+1, r.Score, r.Result.RealizedPnL, r.Result.TotalTrades)
			fmt.Printf("  rsi_period: %.0f, rsi_oversold: %.1f, rsi_overbought: %.1f, bb_period: %.0f, bb_multiplier: %.1f\n",
				r.Params["rsi_period"], r.Params["rsi_oversold"], r.Params["rsi_overbought"], r.Params["bb_period"], r.Params["bb_multiplier"])
		}
	}
}
