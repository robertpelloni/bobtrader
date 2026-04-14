package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange/binance"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

// StreamFeed implements marketdata.StreamFeed using Binance combined WebSocket streams.
// For simplicity and zero external dependencies, this implementation uses
// a minimal pure-Go WebSocket client.
type StreamFeed struct {
	adapter *binance.Adapter
	mu      sync.Mutex
	baseURL string
}

// NewStreamFeed creates a WebSocket-based market data feed.
func NewStreamFeed(adapter *binance.Adapter) *StreamFeed {
	baseURL := "wss://stream.binance.com:9443/ws"
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

func (f *StreamFeed) SubscribeTicks(ctx context.Context, symbol string, interval time.Duration) marketdata.TickSubscription {
	ch := make(chan marketdata.Tick, 10)
	streamPath := fmt.Sprintf("%s/%s@ticker", f.baseURL, strings.ToLower(symbol))

	go f.connectAndStream(ctx, streamPath, func(msg []byte) {
		if tick, ok := parseTickerMessage(msg); ok {
			select {
			case ch <- tick:
			default:
			}
		}
	})

	return tickSub{ch: ch}
}

func (f *StreamFeed) SubscribeCandles(ctx context.Context, symbol, interval string) marketdata.CandleSubscription {
	ch := make(chan marketdata.Candle, 10)
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

// connectAndStream connects to a Binance WebSocket stream and processes messages.
// Automatically reconnects on disconnect with exponential backoff.
func (f *StreamFeed) connectAndStream(ctx context.Context, wsURL string, handler func([]byte)) {
	backoff := 1 * time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		err := f.dialAndRead(ctx, wsURL, handler)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			// Connection lost, wait before reconnecting
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
				backoff = backoff * 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				continue
			}
		}
		return
	}
}

// dialAndRead performs a WebSocket upgrade and reads text frames.
// Uses a minimal WebSocket client implemented with standard library.
func (f *StreamFeed) dialAndRead(ctx context.Context, wsURL string, handler func([]byte)) error {
	parsed, err := url.Parse(wsURL)
	if err != nil {
		return fmt.Errorf("parse ws url: %w", err)
	}

	host := parsed.Host
	path := parsed.Path
	if parsed.RawQuery != "" {
		path += "?" + parsed.RawQuery
	}

	// Determine if we need TLS
	useTLS := parsed.Scheme == "wss"
	port := parsed.Port()
	if port == "" {
		if useTLS {
			port = "443"
		} else {
			port = "80"
		}
	}

	// Connect via TCP
	var conn net.Conn
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	if useTLS {
		conn, err = dialTLS(ctx, net.JoinHostPort(host, port))
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, port))
	}
	if err != nil {
		return fmt.Errorf("dial %s: %w", host, err)
	}
	defer conn.Close()

	// Perform WebSocket upgrade
	key := generateWSKey()
	upgradeReq := fmt.Sprintf("GET %s HTTP/1.1\r\n"+
		"Host: %s\r\n"+
		"Upgrade: websocket\r\n"+
		"Connection: Upgrade\r\n"+
		"Sec-WebSocket-Key: %s\r\n"+
		"Sec-WebSocket-Version: 13\r\n\r\n", path, host, key)

	if _, err := conn.Write([]byte(upgradeReq)); err != nil {
		return fmt.Errorf("send upgrade: %w", err)
	}

	// Read HTTP response (upgrade confirmation)
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("read upgrade response: %w", err)
	}
	response := string(buf[:n])
	if !strings.Contains(response, "101") {
		return fmt.Errorf("websocket upgrade failed: %s", response[:min(200, len(response))])
	}

	// Read WebSocket frames
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		frame, err := readWebSocketFrame(conn)
		if err != nil {
			return fmt.Errorf("read frame: %w", err)
		}
		if frame != nil {
			handler(frame)
		}
	}
}

// readWebSocketFrame reads a single WebSocket text frame.
func readWebSocketFrame(conn net.Conn) ([]byte, error) {
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, err
	}

	payloadLen := int(header[1] & 0x7F)
	masked := (header[1] & 0x80) != 0

	if payloadLen == 126 {
		ext := make([]byte, 2)
		if _, err := io.ReadFull(conn, ext); err != nil {
			return nil, err
		}
		payloadLen = int(ext[0])<<8 | int(ext[1])
	} else if payloadLen == 127 {
		ext := make([]byte, 8)
		if _, err := io.ReadFull(conn, ext); err != nil {
			return nil, err
		}
		payloadLen = 0
		for i := 0; i < 8; i++ {
			payloadLen = payloadLen<<8 | int(ext[i])
		}
	}

	var maskKey []byte
	if masked {
		maskKey = make([]byte, 4)
		if _, err := io.ReadFull(conn, maskKey); err != nil {
			return nil, err
		}
	}

	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(conn, payload); err != nil {
		return nil, err
	}

	if masked {
		for i := range payload {
			payload[i] ^= maskKey[i%4]
		}
	}

	// opcode: text=0x1, binary=0x2, close=0x8, ping=0x9, pong=0xA
	opcode := header[0] & 0x0F
	switch opcode {
	case 0x8: // close
		return nil, io.EOF
	case 0x9: // ping — send pong
		pong := make([]byte, 2+payloadLen)
		pong[0] = 0x8A // fin + pong
		pong[1] = byte(payloadLen)
		copy(pong[2:], payload)
		conn.Write(pong)
		return nil, nil
	case 0x1: // text
		return payload, nil
	case 0x2: // binary
		return payload, nil
	default:
		return nil, nil
	}
}

func generateWSKey() string {
	return "dGhlIHNhbXBsZSBub25jZQ==" // Standard test key, Binance accepts any valid base64
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Binance WebSocket message structures

type tickerMessage struct {
	EventType string `json:"e"`
	Symbol    string `json:"s"`
	Price     string `json:"c"`
	EventTime int64  `json:"E"`
}

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
	var msg tickerMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return marketdata.Tick{}, false
	}
	if msg.Symbol == "" {
		return marketdata.Tick{}, false
	}
	return marketdata.Tick{
		Symbol:    msg.Symbol,
		Price:     msg.Price,
		Source:    "binance-ws",
		Timestamp: time.UnixMilli(msg.EventTime).UTC(),
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

// Stub for TLS dial (uses crypto/tls via net/http default transport)
func dialTLS(ctx context.Context, addr string) (net.Conn, error) {
	return (&http.Transport{}).DialTLSContext(ctx, "tcp", addr)
}
