package binance

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange/binance"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

// StreamFeed implements marketdata.StreamFeed using Binance WebSocket streams.
type StreamFeed struct {
	adapter *binance.Adapter
	mu      sync.Mutex
	baseURL string
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
// Uses the exact same approach as the working standalone test.
func (f *StreamFeed) dialAndRead(ctx context.Context, wsURL string, handler func([]byte)) error {
	parsed, err := url.Parse(wsURL)
	if err != nil {
		return fmt.Errorf("parse ws url: %w", err)
	}
	host := parsed.Hostname()
	path := parsed.Path
	if parsed.RawQuery != "" {
		path += "?" + parsed.RawQuery
	}

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
	addr := net.JoinHostPort(host, port)
	if useTLS {
		conn, err = dialTLS(ctx, addr)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
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
		"Sec-WebSocket-Version: 13\r\n\r\n", path, net.JoinHostPort(host, port), key)

	if _, err := conn.Write([]byte(upgradeReq)); err != nil {
		return fmt.Errorf("send upgrade: %w", err)
	}

	// Read the HTTP upgrade response directly from conn.
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("read upgrade response: %w", err)
	}
	response := string(buf[:n])
	if !strings.Contains(response, "101") {
		return fmt.Errorf("websocket upgrade failed: %s", response[:min(200, len(response))])
	}

	// Determine the frame reader: if there's leftover data after the
	// HTTP \r\n\r\n terminator, combine it with the connection so that
	// readWSFrame sees contiguous bytes.
	var frameReader io.Reader = conn
	headerEnd := strings.Index(response, "\r\n\r\n")
	if headerEnd >= 0 && headerEnd+4 < n {
		leftover := buf[headerEnd+4 : n]
		if len(leftover) > 0 {
			frameReader = io.MultiReader(bytes.NewReader(leftover), conn)
		}
	}

	// Continue reading WebSocket frames from the frame reader.
	frameCount := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		frame, err := readWSFrame(frameReader, conn)
		if err != nil {
			return fmt.Errorf("read frame (after %d frames): %w", frameCount, err)
		}
		if frame != nil {
			handler(frame)
			frameCount++
		}
	}
}



// readWSFrame reads a single WebSocket frame from an io.Reader.
// The w io.Writer is used to send pong responses.
func readWSFrame(r io.Reader, w io.Writer) ([]byte, error) {
	header := make([]byte, 2)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	payloadLen := int(header[1] & 0x7F)
	masked := (header[1] & 0x80) != 0

	if payloadLen == 126 {
		ext := make([]byte, 2)
		if _, err := io.ReadFull(r, ext); err != nil {
			return nil, err
		}
		payloadLen = int(ext[0])<<8 | int(ext[1])
	} else if payloadLen == 127 {
		ext := make([]byte, 8)
		if _, err := io.ReadFull(r, ext); err != nil {
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
		if _, err := io.ReadFull(r, maskKey); err != nil {
			return nil, err
		}
	}

	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}

	if masked {
		for i := range payload {
			payload[i] ^= maskKey[i%4]
		}
	}

	opcode := header[0] & 0x0F
	switch opcode {
	case 0x8: // close
		return nil, io.EOF
	case 0x9: // ping — send pong
		pong := make([]byte, 2+payloadLen)
		pong[0] = 0x8A // fin + pong
		pong[1] = byte(payloadLen)
		copy(pong[2:], payload)
		w.Write(pong)
		return nil, nil
	case 0x1: // text
		return payload, nil
	case 0x2: // binary
		return payload, nil
	default:
		return nil, nil
	}
}

// readWSFrameConn reads a single WebSocket frame from a net.Conn (backward-compat wrapper).
func readWSFrameConn(conn net.Conn) ([]byte, error) {
	return readWSFrame(conn, conn)
}

// parseWSFrame parses a WebSocket frame from raw bytes (for leftover data after HTTP upgrade).
func parseWSFrame(data []byte) ([]byte, bool) {
	if len(data) < 2 {
		return nil, false
	}
	payloadLen := int(data[1] & 0x7F)
	offset := 2
	if payloadLen == 126 {
		if len(data) < 4 {
			return nil, false
		}
		payloadLen = int(data[2])<<8 | int(data[3])
		offset = 4
	} else if payloadLen == 127 {
		if len(data) < 10 {
			return nil, false
		}
		payloadLen = 0
		for i := 0; i < 8; i++ {
			payloadLen = payloadLen<<8 | int(data[2+i])
		}
		offset = 10
	}

	masked := (data[1] & 0x80) != 0
	if masked {
		offset += 4 // skip mask key
	}

	if offset+payloadLen > len(data) {
		return nil, false
	}

	payload := data[offset : offset+payloadLen]
	if masked {
		maskKey := data[offset-4 : offset]
		for i := range payload {
			payload[i] ^= maskKey[i%4]
		}
	}

	opcode := data[0] & 0x0F
	if opcode == 0x1 || opcode == 0x2 {
		return payload, true
	}
	return nil, false
}

func generateWSKey() string {
	return "dGhlIHNhbXBsZSBub25jZQ=="
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
	Quantity  string `json:"q"`
	EventTime int64  `json:"E"`
}

type tickSub struct{ ch <-chan marketdata.Tick }

func (s tickSub) Chan() <-chan marketdata.Tick { return s.ch }

type candleSub struct{ ch <-chan marketdata.Candle }

func (s candleSub) Chan() <-chan marketdata.Candle { return s.ch }

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
		Quantity:  msg.Quantity,
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

// dialTLS creates a TLS connection using crypto/tls.
func dialTLS(ctx context.Context, addr string) (net.Conn, error) {
	host, _, _ := net.SplitHostPort(addr)
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	return tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{
		ServerName: host,
	})
}
