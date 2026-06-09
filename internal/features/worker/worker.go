package core_worker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DimaKirejko/Dstributed_cron/internal/core/domain"
	core_logger "github.com/DimaKirejko/Dstributed_cron/internal/core/logger"
	core_job_types "github.com/DimaKirejko/Dstributed_cron/internal/core/types"
	"go.uber.org/zap"
)

const (
	AttemptResultSuccess = "success"
	AttemptResultRetry   = "retry"
	AttemptResultFailed  = "failed"
)

type Repository interface {
	ClaimJob(
		ctx context.Context,
		workerID int64,
		leaseDuration time.Duration,
	) (domain.Job, bool, error)

	CompleteJob(
		ctx context.Context,
		jobID core_job_types.ID,
		workerID int64,
		IsRepetable bool,
	) error

	FailJob(
		ctx context.Context,
		jobID core_job_types.ID,
		workerID int64,
		cause error,
	) error

	ReleaseForRetry(
		ctx context.Context,
		jobID core_job_types.ID,
		workerID int64,
		cause error,
	) error

	InsertAttempt(
		ctx context.Context,
		attempt Attempt,
	) error
}

type Executor interface {
	Execute(ctx context.Context, job domain.Job) ExecutionResult
}

type ExecutionResult struct {
	Success    bool
	Retryable  bool
	HTTPStatus *int
	Err        error
}

type Attempt struct {
	JobID        core_job_types.ID
	Result       string
	HTTPStatus   *int
	ErrorMessage *string
	StartedAt    time.Time
}

type Worker struct {
	id            int64
	repo          Repository
	executor      Executor
	logger        *core_logger.Logger
	pollInterval  time.Duration
	leaseDuration time.Duration
}

func NewWorker(
	id int64,
	repo Repository,
	executor Executor,
	logger *core_logger.Logger,
	pollInterval time.Duration,
	leaseDuration time.Duration,
) *Worker {
	return &Worker{
		id:            id,
		repo:          repo,
		executor:      executor,
		logger:        logger.With(zap.Int64("worker_id", id)),
		pollInterval:  pollInterval,
		leaseDuration: leaseDuration,
	}
}

func (w *Worker) Start(ctx context.Context) { // worker supervisor
	go func() {
		for {
			err := w.runSafe(ctx)
			if errors.Is(err, context.Canceled) {
				return
			}

			w.logger.Error("worker stopped, restarting", zap.Error(err))

			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
		}
	}()
}

func (w *Worker) runSafe(ctx context.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("worker panic: %v", r)
		}
	}()

	return w.run(ctx)
}

func (w *Worker) run(ctx context.Context) error {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		w.logger.Debug("started check:")
		job, ok, err := w.repo.ClaimJob(ctx, w.id, w.leaseDuration)
		w.logger.Info("JOB has been taken", zap.Int("job id", int(job.Id)))
		if err != nil {
			w.logger.Error("claim job", zap.Error(err))
		}

		if ok {
			w.handleJob(ctx, job)
			continue
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (w *Worker) handleJob(ctx context.Context, job domain.Job) {
	startedAt := time.Now().UTC()
	result := w.executor.Execute(ctx, job)

	attempt := Attempt{
		JobID:        job.Id,
		Result:       attemptResult(result),
		HTTPStatus:   result.HTTPStatus,
		ErrorMessage: errorMessage(result.Err, w.logger),
		StartedAt:    startedAt,
	}

	if err := w.repo.InsertAttempt(ctx, attempt); err != nil {
		w.logger.Error("insert attempt", zap.Int64("job_id", int64(job.Id)), zap.Error(err))
	}

	if result.Success {
		w.logger.Info("successful completion", zap.Int64("job_id", int64(job.Id)))
		if err := w.repo.CompleteJob(ctx, job.Id, w.id, job.IsRepetable); err != nil {
			w.logger.Error("complete job", zap.Int64("job_id", int64(job.Id)), zap.Error(err))
		}

		return
	}

	if result.Retryable && job.Attempt < job.MaxRetries {
		workerDelay()
		if err := w.repo.ReleaseForRetry(ctx, job.Id, w.id, result.Err); err != nil {
			w.logger.Error("release job for retry", zap.Int64("job_id", int64(job.Id)), zap.Error(err))
		}

		return
	}

	if err := w.repo.FailJob(ctx, job.Id, w.id, result.Err); err != nil {
		w.logger.Error("FILED JOB", zap.Int64("job_id", int64(job.Id)), zap.Error(err))
	}
}

func attemptResult(result ExecutionResult) string {
	if result.Success {
		return AttemptResultSuccess
	}

	if result.Retryable {
		return AttemptResultRetry
	}

	return AttemptResultFailed
}

func errorMessage(err error, logger *core_logger.Logger) *string {
	if err == nil {
		return nil
	}

	logger.Error("error while executing the job", zap.Error(err))
	msg := err.Error()
	return &msg
}

func workerDelay() {
	time.Sleep(1 * time.Second)
}
