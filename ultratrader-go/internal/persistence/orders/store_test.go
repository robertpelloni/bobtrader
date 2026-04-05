package orders

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendOrderRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "orders.jsonl")
	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}

	err = store.Append(context.Background(), Record{AccountID: "paper-main", Exchange: "paper", OrderID: "ord-1", Symbol: "BTCUSDT", Side: "buy", Type: "market", Status: "filled", Quantity: "0.01"})
	if err != nil {
		t.Fatalf("Append returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if !strings.Contains(string(data), "ord-1") {
		t.Fatalf("expected order record, got %q", string(data))
	}
}
