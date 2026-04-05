package scheduler

import (
	"context"
	"testing"
	"time"
)

func TestServiceStartWithNonPositiveIntervalDoesNotBlock(t *testing.T) {
	service := NewService(nil, 0)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	service.Start(ctx)
	select {
	case <-time.After(10 * time.Millisecond):
	}
}
