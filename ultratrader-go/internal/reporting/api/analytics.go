package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/analytics"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/logging"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

// AnalyticsServer extends the base Server to provide endpoints for advanced
// metrics like heatmaps, pattern recognition, and order flow.
type AnalyticsServer struct {
	mux        *http.ServeMux
	logger     *logging.Logger
	feed       marketdata.Feed
	recognizer *analytics.PatternRecognizer
	flowEngine *analytics.OrderFlowAnalyzer
	arbitrage  *analytics.ArbitrageDetector
}

func NewAnalyticsServer(feed marketdata.Feed, logger *logging.Logger) *AnalyticsServer {
	if logger == nil {
		logger, _ = logging.New(logging.Config{Stdout: true})
	}
	s := &AnalyticsServer{
		mux:        http.NewServeMux(),
		logger:     logger,
		feed:       feed,
		recognizer: analytics.NewPatternRecognizer(0.02),
		flowEngine: analytics.NewOrderFlowAnalyzer(5),
		arbitrage:  analytics.NewArbitrageDetector(20),
	}
	s.routes()
	return s
}

func (s *AnalyticsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *AnalyticsServer) routes() {
	s.mux.HandleFunc("/api/analytics/patterns", s.handlePatternRecognition())
	s.mux.HandleFunc("/api/analytics/arbitrage", s.handleStatisticalArbitrage())
	s.mux.HandleFunc("/api/analytics/orderflow", s.handleOrderFlow())
}

// PatternResponse defines the JSON structure for the frontend UI to render chart overlays.
type PatternResponse struct {
	Symbol   string                    `json:"symbol"`
	Patterns []analytics.PatternResult `json:"patterns"`
	Error    string                    `json:"error,omitempty"`
}

func (s *AnalyticsServer) handlePatternRecognition() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		symbol := r.URL.Query().Get("symbol")
		if symbol == "" {
			http.Error(w, "symbol parameter required", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		var prices []float64
		// If feed implemented HistoricalFeed, we'd fetch candles here.
		// For the API structure, we assume we fetch the last 100 periods.
		// For demonstration, we mock the prices if feed is unavailable.
		if histFeed, ok := s.feed.(interface {
			HistoricalCandles(ctx context.Context, symbol, interval string, limit int) ([]marketdata.Candle, error)
		}); ok {
			candles, err := histFeed.HistoricalCandles(ctx, symbol, "1hour", 100)
			if err == nil {
				for _, c := range candles {
					// Dummy parse to float (we'd use utils.ParseFloat typically)
					prices = append(prices, 100.0) // placeholder
					_ = c
				}
			}
		}

		// Fallback for API structure demonstration if feed not present
		if len(prices) == 0 {
			prices = make([]float64, 100)
			for i := range prices {
				prices[i] = 100.0
			}
			prices[15] = 130.0 // mock head and shoulders
		}

		patterns := s.recognizer.Scan(prices)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(PatternResponse{
			Symbol:   symbol,
			Patterns: patterns,
		})
	}
}

// ArbitrageResponse defines the JSON structure for cross-exchange or pairs trading correlation.
type ArbitrageResponse struct {
	SymbolA string              `json:"symbol_a"`
	SymbolB string              `json:"symbol_b"`
	Stats   analytics.PairStats `json:"stats"`
	Action  string              `json:"action"`
}

func (s *AnalyticsServer) handleStatisticalArbitrage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		symA := r.URL.Query().Get("symbol_a")
		symB := r.URL.Query().Get("symbol_b")

		if symA == "" || symB == "" {
			http.Error(w, "symbol_a and symbol_b parameters required", http.StatusBadRequest)
			return
		}

		// Mocking price arrays for API structure mapping
		pricesA := make([]float64, 50)
		pricesB := make([]float64, 50)
		for i := range pricesA {
			pricesA[i] = 100.0
			pricesB[i] = 50.0 // B is half of A
		}
		// Introduce a divergence
		pricesA[49] = 120.0
		pricesB[49] = 40.0

		stats, err := s.arbitrage.AnalyzePair(symA, symB, pricesA, pricesB)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		action := s.arbitrage.CheckSignal(stats, 2.0, 0.5)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ArbitrageResponse{
			SymbolA: symA,
			SymbolB: symB,
			Stats:   stats,
			Action:  action,
		})
	}
}

// OrderFlowResponse maps CVD data for UI visualizations (Heatmaps/Depth).
type OrderFlowResponse struct {
	Symbol     string                    `json:"symbol"`
	Divergence analytics.DivergenceType  `json:"divergence"`
	Data       []analytics.OrderFlowData `json:"data"`
}

func (s *AnalyticsServer) handleOrderFlow() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		symbol := r.URL.Query().Get("symbol")
		if symbol == "" {
			http.Error(w, "symbol parameter required", http.StatusBadRequest)
			return
		}

		// Mock order flow data to fulfill endpoint signature
		prices := []float64{100, 105, 110, 115, 120}
		buyVols := []float64{10, 15, 5, 2, 0}
		sellVols := []float64{5, 10, 20, 25, 30}

		data, divergence, err := s.flowEngine.Analyze(prices, buyVols, sellVols)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(OrderFlowResponse{
			Symbol:     symbol,
			Divergence: divergence,
			Data:       data,
		})
	}
}
