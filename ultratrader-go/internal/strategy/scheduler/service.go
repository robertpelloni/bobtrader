package scheduler

import (
	"context"
	"time"
)

type runner interface {
	RunOnce(ctx context.Context) error
}

type Service struct {
	runner   runner
	interval time.Duration
}

func NewService(r runner, interval time.Duration) *Service {
	return &Service{runner: r, interval: interval}
}

func (s *Service) Start(ctx context.Context) {
	if s.runner == nil || s.interval <= 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = s.runner.RunOnce(ctx)
			}
		}
	}()
}
