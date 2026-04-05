package logging

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestLoggerWithContextAddsCorrelationID(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	logger := &Logger{writer: buf, fields: map[string]any{}}
	ctx := context.WithValue(context.Background(), CorrelationIDKey, "corr-123")
	logger.WithContext(ctx).Info("hello", map[string]any{"component": "test"})
	text := buf.String()
	if !strings.Contains(text, "corr-123") {
		t.Fatalf("expected correlation id in log, got %q", text)
	}
}
