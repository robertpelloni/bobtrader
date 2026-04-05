package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

type testRunner struct{ count atomic.Int32 }

func (r *testRunner) RunOnce(_ context.Context) error { r.count.Add(1); return nil }

func TestServiceStartWithNonPositiveIntervalDoesNotBlock(t *testing.T) {
	service := NewService(nil, 0)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	service.Start(ctx)
	select {
	case <-time.After(10 * time.Millisecond):
	}
}

func TestServiceStartRunsRepeatedly(t *testing.T) {
	runner := &testRunner{}
	service := NewService(runner, 10*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	service.Start(ctx)
	time.Sleep(35 * time.Millisecond)
	if runner.count.Load() < 2 {
		t.Fatalf("expected repeated scheduler runs, got %d", runner.count.Load())
	}
}
