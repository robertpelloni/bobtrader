package persistence_test

import (
	"context"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/persistence"
)

func TestDB_InsertTradeExit(t *testing.T) {
	// Use an in-memory SQLite database for fast, isolated tests
	dsn := "file:testdb?mode=memory&cache=shared"
	db, err := persistence.Connect(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to in-memory db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Try to insert a malicious payload as a symbol
	// In an un-parameterized query, this would execute the DROP TABLE command.
	maliciousSymbol := "'; DROP TABLE trade_exits; --"

	id, err := db.InsertTradeExit(ctx, maliciousSymbol, "SELL", 1.5, 45000.0, 0.05)
	if err != nil {
		t.Fatalf("Failed to insert record: %v", err)
	}

	if id <= 0 {
		t.Errorf("Expected positive insert ID, got %d", id)
	}

	// Verify the table still exists and wasn't dropped
	// Try to insert another normal record
	id2, err := db.InsertTradeExit(ctx, "BTC", "SELL", 0.5, 46000.0, 0.02)
	if err != nil {
		t.Fatalf("Table may have been dropped, second insert failed: %v", err)
	}

	if id2 <= id {
		t.Errorf("Expected auto-incrementing ID, got %d followed by %d", id, id2)
	}
}
