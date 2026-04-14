package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/logging"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Accept connections from any origin for development
		// In production, restrict this to specific dashboard domains
		return true
	},
}

// Client represents a single connected websocket client.
type Client struct {
	conn *websocket.Conn
	send chan []byte
}

// StreamHub manages all active websocket connections and broadcasts messages.
type StreamHub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	logger     *logging.Logger
	mu         sync.Mutex
}

// NewStreamHub creates a new hub.
func NewStreamHub(logger *logging.Logger) *StreamHub {
	if logger == nil {
		logger, _ = logging.New(logging.Config{Stdout: true})
	}
	return &StreamHub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
	}
}

// Run starts the hub's main event loop for adding/removing clients and broadcasting.
func (h *StreamHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.logger.Info("Websocket client connected", nil)
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			h.logger.Info("Websocket client disconnected", nil)
		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

// BroadcastJSON serializes an object and sends it to all connected clients.
func (h *StreamHub) BroadcastJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	h.broadcast <- data
	return nil
}

// HandleWebSocket upgrades the HTTP connection and registers the client.
func (h *StreamHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade websocket", map[string]any{"error": err.Error()})
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
	}
	h.register <- client

	// Start pump routines
	go h.writePump(client)
	go h.readPump(client)
}

func (h *StreamHub) writePump(c *Client) {
	ticker := time.NewTicker(54 * time.Second) // Ping interval
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Batch queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *StreamHub) readPump(c *Client) {
	defer func() {
		h.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(60 * time.Second)); return nil })
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error("Websocket error", map[string]any{"error": err.Error()})
			}
			break
		}
		// In a read-only telemetry dashboard, we just discard client messages.
	}
}
