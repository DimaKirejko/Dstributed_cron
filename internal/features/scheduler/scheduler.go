package core_scheduler

import (
	"context"
	"errors"
	"time"

	core_logger "github.com/DimaKirejko/Dstributed_cron/internal/core/logger"
	"go.uber.org/zap"
)

type Scheduler struct {
	repo         Repository
	tickInterval time.Duration
	logger       *core_logger.Logger
}

type Repository interface {
	QueueDueDailyJobs(ctx context.Context, logger *core_logger.Logger) (int64, error)
}

func NewScheduler(repo Repository, logger *core_logger.Logger, tickInterval time.Duration) *Scheduler {
	return &Scheduler{
		repo:         repo,
		logger:       logger,
		tickInterval: tickInterval,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	go func() {
		if err := s.runTick(ctx); err != nil && !errors.Is(err, context.Canceled) {
			s.logger.Error("scheduler stopped", zap.Error(err))
		} // need rechek mb need restart policy
	}()
}

func (s *Scheduler) runTick(ctx context.Context) error {
	ticker := time.NewTicker(s.tickInterval)
	defer ticker.Stop()

	for {
		queued, err := s.repo.QueueDueDailyJobs(ctx, s.logger)
		if err != nil {
			s.logger.Error("queue due jobs", zap.Error(err))
		}

		if queued > 0 {
			s.logger.Info("queued due jobs", zap.Int64("count", queued))
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
