package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Logger struct {
	mu      sync.Mutex
	writer  io.Writer
	closers []io.Closer
	fields  map[string]any
}

type entry struct {
	Timestamp string         `json:"timestamp"`
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	Fields    map[string]any `json:"fields,omitempty"`
}

type Config struct {
	Path   string
	Stdout bool
}

type ctxKey string

const CorrelationIDKey ctxKey = "correlation_id"

func New(cfg Config) (*Logger, error) {
	writers := make([]io.Writer, 0, 2)
	closers := make([]io.Closer, 0, 1)
	if cfg.Path != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.Path), 0o755); err != nil {
			return nil, fmt.Errorf("create log directory: %w", err)
		}
		f, err := os.OpenFile(cfg.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, fmt.Errorf("open log file: %w", err)
		}
		writers = append(writers, f)
		closers = append(closers, f)
	}
	if cfg.Stdout || len(writers) == 0 {
		writers = append(writers, os.Stdout)
	}
	return &Logger{writer: io.MultiWriter(writers...), closers: closers, fields: map[string]any{}}, nil
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	var firstErr error
	for _, closer := range l.closers {
		if err := closer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	l.closers = nil
	return firstErr
}

func (l *Logger) WithFields(fields map[string]any) *Logger {
	merged := make(map[string]any, len(l.fields)+len(fields))
	for k, v := range l.fields {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}
	return &Logger{writer: l.writer, closers: l.closers, fields: merged}
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
	if ctx == nil {
		return l
	}
	correlationID, _ := ctx.Value(CorrelationIDKey).(string)
	if correlationID == "" {
		return l
	}
	return l.WithFields(map[string]any{"correlation_id": correlationID})
}

func (l *Logger) Info(msg string, fields map[string]any)  { l.write("info", msg, fields) }
func (l *Logger) Error(msg string, fields map[string]any) { l.write("error", msg, fields) }

func (l *Logger) write(level, msg string, fields map[string]any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	merged := make(map[string]any, len(l.fields)+len(fields))
	for k, v := range l.fields {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}
	payload := entry{Timestamp: time.Now().UTC().Format(time.RFC3339Nano), Level: level, Message: msg, Fields: merged}
	_ = json.NewEncoder(l.writer).Encode(payload)
}
