package scheduler

import (
	"context"
	"time"
)

type Service struct {
	scheduler *Scheduler
	interval  time.Duration
}

func NewService(scheduler *Scheduler, interval time.Duration) *Service {
	return &Service{scheduler: scheduler, interval: interval}
}

func (s *Service) Start(ctx context.Context) {
	if s.interval <= 0 {
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
				_ = s.scheduler.RunOnce(ctx)
			}
		}
	}()
}
