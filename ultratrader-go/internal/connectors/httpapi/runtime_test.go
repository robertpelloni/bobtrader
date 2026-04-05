package httpapi

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/metrics"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/execution"
)

func TestRuntimeStartAndShutdown(t *testing.T) {
	handler := NewHandler(Dependencies{
		StatusProvider:           func() Status { return Status{Name: "ultratrader-go", Ready: true, AccountCount: 1} },
		PortfolioProvider:        func() PortfolioSnapshot { return PortfolioSnapshot{} },
		OrdersProvider:           func() []exchange.Order { return nil },
		ExecutionSummaryProvider: func() execution.Summary { return execution.Summary{} },
		MetricsProvider:          func() metrics.Snapshot { return metrics.Snapshot{} },
	})
	runtime := NewRuntime("127.0.0.1:0", handler)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := runtime.Start(ctx); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	time.Sleep(20 * time.Millisecond)
	resp, err := http.Get("http://" + runtime.Address() + "/healthz")
	if err != nil {
		t.Fatalf("http get failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || len(body) == 0 {
		t.Fatalf("unexpected response: status=%d body=%q", resp.StatusCode, string(body))
	}
	if err := runtime.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}
}
