package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/reporting/api"
)

func TestWebSocketStreamHub(t *testing.T) {
	hub := api.NewStreamHub(nil)
	go hub.Run()

	server := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect a client
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect websocket: %v", err)
	}
	defer conn.Close()

	// Wait briefly for client registration
	time.Sleep(100 * time.Millisecond)

	// Broadcast a message
	msg := map[string]string{"type": "candle", "symbol": "BTC"}
	err = hub.BroadcastJSON(msg)
	if err != nil {
		t.Fatalf("Failed to broadcast json: %v", err)
	}

	// Read message from client
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	messageType, p, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	if messageType != websocket.TextMessage {
		t.Errorf("Expected text message type, got %d", messageType)
	}

	var received map[string]string
	if err := json.Unmarshal(p, &received); err != nil {
		t.Fatalf("Failed to parse json: %v", string(p))
	}

	if received["symbol"] != "BTC" {
		t.Errorf("Expected symbol BTC, got %s", received["symbol"])
	}
}
