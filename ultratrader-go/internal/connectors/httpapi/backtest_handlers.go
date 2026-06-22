package httpapi

import (
	"encoding/json"
	"net/http"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/backtest/optimizer"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy/demo"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

type BacktestRequest struct {
	Symbol       string                 `json:"symbol"`
	Interval     string                 `json:"interval"`
	StrategyName string                 `json:"strategy"`
	Params       optimizer.ParameterMap `json:"params"`
	StartTime    int64                  `json:"start_time"` // unix seconds
	EndTime      int64                  `json:"end_time"`   // unix seconds
}

// Very basic strategy factory for now.
func buildStrategy(name string, params optimizer.ParameterMap) (strategy.Strategy, error) {
	if name == "DoubleEMATrend" {
		fast, _ := params["fast_period"].(float64)
		medium, _ := params["medium_period"].(float64)
		slow, _ := params["slow_period"].(float64)
		return demo.NewDoubleEMATrendStrategy("paper", "BTCUSDT", "0.01", int(fast), int(medium), int(slow)), nil
	}
	if name == "RSIReversion" {
		period, _ := params["rsi_period"].(float64)
		oversold, _ := params["oversold"].(float64)
		overbought, _ := params["overbought"].(float64)
		return demo.NewRSIReversion("paper", "BTCUSDT", "0.01", int(period), oversold, overbought), nil
	}
	return nil, nil // return error normally
}

func handleBacktest(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req BacktestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// In a real scenario we'd fetch historical candles from binance adapter here,
		// but since we need context/adapter we'll mock it for the skeleton.
		// For the time being, we will just return a 501. The actual implementation
		// needs the binance adapter or a database of historical candles.
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(`{"error":"backtest endpoint requires historical data provider"}`))
	}
}

type HyperoptRequest struct {
	Symbol       string                                    `json:"symbol"`
	Interval     string                                    `json:"interval"`
	StrategyName string                                    `json:"strategy"`
	ParamRanges  map[string]optimizer.ParamRange           `json:"param_ranges"`
	StartTime    int64                                     `json:"start_time"`
	EndTime      int64                                     `json:"end_time"`
	Method       string                                    `json:"method"` // "grid" or "walkforward"
}

func handleHyperopt(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req HyperoptRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(`{"error":"hyperopt endpoint requires historical data provider"}`))
	}
}
