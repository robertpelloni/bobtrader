package logging

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

var correlationCounter uint64

func NewCorrelationContext(parent context.Context, prefix string) (context.Context, string) {
	id := fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixNano(), atomic.AddUint64(&correlationCounter, 1))
	return context.WithValue(parent, CorrelationIDKey, id), id
}
