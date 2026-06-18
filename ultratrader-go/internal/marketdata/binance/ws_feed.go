package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange/binance"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

// StreamFeed implements marketdata.StreamFeed using Binance WebSocket streams.
type StreamFeed struct {
	adapter *binance.Adapter
	baseURL string

	// Health monitoring
	mu             sync.RWMutex
	connected      bool
	lastTickTime   time.Time
	ticksReceived  int64
	errorsCount    int64
	reconnectCount int64
}

// NewStreamFeed creates a WebSocket-based market data feed.
func NewStreamFeed(adapter *binance.Adapter) *StreamFeed {
	baseURL := "wss://stream.binance.us:9443/ws"
	if adapter.IsTestnet() {
		baseURL = "wss://testnet.binance.vision/ws"
	}
	return &StreamFeed{
		adapter: adapter,
		baseURL: baseURL,
	}
}

func (f *StreamFeed) LatestTick(ctx context.Context, symbol string) (marketdata.Tick, error) {
	price, err := f.adapter.GetTickerPrice(ctx, symbol)
	if err != nil {
		return marketdata.Tick{}, err
	}
	return marketdata.Tick{
		Symbol:    symbol,
		Price:     price,
		Source:    "binance-ws",
		Timestamp: time.Now().UTC(),
	}, nil
}

func (f *StreamFeed) LatestCandle(ctx context.Context, symbol, interval string) (marketdata.Candle, error) {
	return marketdata.Candle{}, fmt.Errorf("use REST feed for historical candles")
}

func (f *StreamFeed) CandleHistory(ctx context.Context, symbol, interval string, limit int) ([]marketdata.Candle, error) {
	klines, err := f.adapter.GetKlines(ctx, symbol, interval, limit)
	if err != nil {
		return nil, err
	}
	candles := make([]marketdata.Candle, len(klines))
	for i, k := range klines {
		candles[i] = marketdata.Candle{
			Symbol:    symbol,
			Interval:  interval,
			Open:      k.Open,
			High:      k.High,
			Low:       k.Low,
			Close:     k.Close,
			Volume:    k.Volume,
			Timestamp: time.UnixMilli(k.OpenTime).UTC(),
		}
	}
	return candles, nil
}

func (f *StreamFeed) SubscribeTicks(ctx context.Context, symbol string, interval time.Duration) marketdata.TickSubscription {
	ch := make(chan marketdata.Tick, 100)
	streamPath := fmt.Sprintf("%s/%s@ticker", f.baseURL, strings.ToLower(symbol))
	go f.connectAndStream(ctx, streamPath, func(msg []byte) {
		if tick, ok := parseTickerMessage(msg); ok {
			f.mu.Lock()
			f.lastTickTime = time.Now()
			f.mu.Unlock()
			atomic.AddInt64(&f.ticksReceived, 1)

			select {
			case ch <- tick:
			default:
				// Channel full, drop tick to keep real-time
			}
		}
	})
	return tickSub{ch: ch}
}

func (f *StreamFeed) SubscribeCandles(ctx context.Context, symbol, interval string) marketdata.CandleSubscription {
	ch := make(chan marketdata.Candle, 100)
	streamPath := fmt.Sprintf("%s/%s@kline_%s", f.baseURL, strings.ToLower(symbol), interval)
	go f.connectAndStream(ctx, streamPath, func(msg []byte) {
		if candle, ok := parseKlineMessage(msg); ok {
			select {
			case ch <- candle:
			default:
			}
		}
	})
	return candleSub{ch: ch}
}

func (f *StreamFeed) SubscribeDepth(ctx context.Context, symbol string) marketdata.DepthSubscription {
	ch := make(chan marketdata.DepthUpdate, 100)
	// Using @depth20 for the top 20 levels of the order book
	streamPath := fmt.Sprintf("%s/%s@depth20", f.baseURL, strings.ToLower(symbol))
	go f.connectAndStream(ctx, streamPath, func(msg []byte) {
		if update, ok := parseDepthMessage(msg, symbol); ok {
			select {
			case ch <- update:
			default:
			}
		}
	})
	return depthSub{ch: ch}
}

func (f *StreamFeed) connectAndStream(ctx context.Context, wsURL string, handler func([]byte)) {
	backoff := 1 * time.Second
	maxBackoff := 60 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		_ = f.dialAndRead(ctx, wsURL, handler)
		if ctx.Err() != nil {
			return
		}

		atomic.AddInt64(&f.errorsCount, 1)
		atomic.AddInt64(&f.reconnectCount, 1)

		f.mu.Lock()
		f.connected = false
		f.mu.Unlock()

		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}

func (f *StreamFeed) dialAndRead(ctx context.Context, wsURL string, handler func([]byte)) error {
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second

	conn, _, err := dialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	f.mu.Lock()
	f.connected = true
	f.mu.Unlock()

	// Setup pong handler and heartbeat deadline
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}
			handler(message)
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		}
	}()

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return fmt.Errorf("websocket read error or connection closed")
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return fmt.Errorf("ping: %w", err)
			}
		case <-ctx.Done():
			// Graceful close attempt
			deadline := time.Now().Add(time.Second)
			msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
			_ = conn.WriteControl(websocket.CloseMessage, msg, deadline)
			return ctx.Err()
		}
	}
}

// GetStatus returns the current health of the WebSocket feed.
func (f *StreamFeed) GetStatus() map[string]interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return map[string]interface{}{
		"connected":        f.connected,
		"last_tick":        f.lastTickTime,
		"ticks_received":   atomic.LoadInt64(&f.ticksReceived),
		"errors":           atomic.LoadInt64(&f.errorsCount),
		"reconnects":       atomic.LoadInt64(&f.reconnectCount),
		"uptime_seconds":   time.Since(f.lastTickTime).Seconds(),
	}
}

// Binance WebSocket message structures
type tickerMessage struct {
	EventType string `json:"e"`
	Symbol    string `json:"s"`
	Price     string `json:"c"`
	Quantity  string `json:"Q"` // Last quantity
	EventTime int64  `json:"E"`
}

type tickSub struct{ ch <-chan marketdata.Tick }

func (s tickSub) Chan() <-chan marketdata.Tick { return s.ch }

type candleSub struct{ ch <-chan marketdata.Candle }

func (s candleSub) Chan() <-chan marketdata.Candle { return s.ch }

type depthSub struct{ ch <-chan marketdata.DepthUpdate }

func (s depthSub) Chan() <-chan marketdata.DepthUpdate { return s.ch }

type klineMessage struct {
	EventType string      `json:"e"`
	Symbol    string      `json:"s"`
	EventTime json.Number `json:"E"`
	Kline     struct {
		StartTime int64  `json:"t"`
		EndTime   int64  `json:"T"`
		Interval  string `json:"i"`
		Open      string `json:"o"`
		High      string `json:"h"`
		Low       string `json:"l"`
		Close     string `json:"c"`
		Volume    string `json:"v"`
		Closed    bool   `json:"x"`
	} `json:"k"`
}

func parseTickerMessage(data []byte) (marketdata.Tick, bool) {
	// Use map[string]interface{} for maximum flexibility if struct tags fail
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return marketdata.Tick{}, false
	}

	symbol, _ := raw["s"].(string)
	price, _ := raw["c"].(string)
	if symbol == "" || price == "" {
		return marketdata.Tick{}, false
	}

	qty, _ := raw["Q"].(string)
	if qty == "" {
		qty, _ = raw["q"].(string) // fallback to quote volume
	}

	var eventTime int64
	if et, ok := raw["E"].(float64); ok {
		eventTime = int64(et)
	}

	return marketdata.Tick{
		Symbol:    strings.ToUpper(symbol),
		Price:     price,
		Quantity:  qty,
		Source:    "binance-ws",
		Timestamp: time.UnixMilli(eventTime).UTC(),
	}, true
}

func parseKlineMessage(data []byte) (marketdata.Candle, bool) {
	var msg klineMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return marketdata.Candle{}, false
	}
	if msg.EventType != "kline" {
		return marketdata.Candle{}, false
	}
	return marketdata.Candle{
		Symbol:    msg.Symbol,
		Interval:  msg.Kline.Interval,
		Open:      msg.Kline.Open,
		High:      msg.Kline.High,
		Low:       msg.Kline.Low,
		Close:     msg.Kline.Close,
		Volume:    msg.Kline.Volume,
		Timestamp: time.UnixMilli(msg.Kline.StartTime).UTC(),
	}, true
}

type depthMessage struct {
	LastUpdateID int64       `json:"lastUpdateId"`
	Bids         [][2]string `json:"bids"`
	Asks         [][2]string `json:"asks"`
}

func parseDepthMessage(data []byte, symbol string) (marketdata.DepthUpdate, bool) {
	var msg depthMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return marketdata.DepthUpdate{}, false
	}
	return marketdata.DepthUpdate{
		Symbol:    strings.ToUpper(symbol),
		Bids:      msg.Bids,
		Asks:      msg.Asks,
		Timestamp: time.Now().UTC(),
	}, true
}
