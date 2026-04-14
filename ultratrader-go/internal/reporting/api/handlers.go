package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/logging"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/portfolio"
)

// Server encapsulates the HTTP handlers for the dashboard and operator interfaces.
type Server struct {
	mux     *http.ServeMux
	tracker *portfolio.Tracker
	feed    marketdata.Feed
	logger  *logging.Logger
}

// NewServer creates a new API server routing instance.
func NewServer(tracker *portfolio.Tracker, feed marketdata.Feed, logger *logging.Logger) *Server {
	if logger == nil {
		logger, _ = logging.New(logging.Config{Stdout: true})
	}

	s := &Server{
		mux:     http.NewServeMux(),
		tracker: tracker,
		feed:    feed,
		logger:  logger,
	}

	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/portfolio/summary", s.handlePortfolioSummary())
	s.mux.HandleFunc("/api/system/health", s.handleHealth())
}

// PortfolioSummaryResponse mirrors the exact payload expected by the legacy React/Vite dashboard.
type PortfolioSummaryResponse struct {
	TotalValue    float64              `json:"total_value"`
	UnrealizedPnL float64              `json:"unrealized_pnl"`
	RealizedPnL   float64              `json:"realized_pnl"`
	OpenPositions int                  `json:"open_positions"`
	Timestamp     int64                `json:"timestamp"`
	Positions     []portfolio.Position `json:"positions"`
}

func (s *Server) handlePortfolioSummary() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		var totalVal, unrealized float64
		var positions []portfolio.Position

		if s.tracker != nil {
			totalVal = s.tracker.TotalMarketValue(ctx, s.feed)
			unrealized = s.tracker.TotalUnrealizedPnL(ctx, s.feed)
			positions = s.tracker.ValuedPositions(ctx, s.feed)
		}

		// Calculate total realized independently
		realized := s.tracker.TotalRealizedPnL()
		openCount := s.tracker.OpenPositionCount()

		resp := PortfolioSummaryResponse{
			TotalValue:    totalVal,
			UnrealizedPnL: unrealized,
			RealizedPnL:   realized,
			OpenPositions: openCount,
			Timestamp:     time.Now().UnixMilli(),
			Positions:     positions,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func (s *Server) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}
}
