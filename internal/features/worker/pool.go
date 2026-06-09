package core_worker

import (
	"context"
	"time"

	core_logger "github.com/DimaKirejko/Dstributed_cron/internal/core/logger"
	"go.uber.org/zap"
)

type Pool struct {
	workers []*Worker
	logger  *core_logger.Logger
}

func NewPool(
	size int,
	repo Repository,
	executor Executor,
	logger *core_logger.Logger,
	pollInterval time.Duration,
	leaseDuration time.Duration,
) *Pool {
	if size <= 0 {
		size = 1
	}

	workers := make([]*Worker, 0, size)

	for i := 0; i < size; i++ {
		worker := NewWorker(
			int64(i+1),
			repo,
			executor,
			logger,
			pollInterval,
			leaseDuration,
		)

		workers = append(workers, worker)
	}

	return &Pool{
		workers: workers,
		logger:  logger,
	}
}

func (p *Pool) Start(ctx context.Context) {
	for _, worker := range p.workers {
		worker.Start(ctx)
	}

	p.logger.Info("worker pool started", zap.Int("size", len(p.workers)))
}
