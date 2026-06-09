package core_worker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DimaKirejko/Dstributed_cron/internal/core/domain"
	core_errors "github.com/DimaKirejko/Dstributed_cron/internal/core/errors"
	core_postgres_pool "github.com/DimaKirejko/Dstributed_cron/internal/core/repository/postgres_pgx"
	core_job_types "github.com/DimaKirejko/Dstributed_cron/internal/core/types"
	"github.com/jackc/pgx/v5"
)

type WorkerRepository struct {
	pool core_postgres_pool.PgxPool
}

func NewWorkerRepository(pool core_postgres_pool.PgxPool) *WorkerRepository {
	return &WorkerRepository{
		pool: pool,
	}
}

func (r *WorkerRepository) ClaimJob(
	ctx context.Context,
	workerID int64,
	leaseDuration time.Duration,
) (domain.Job, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	WITH picked AS (
		SELECT id
		FROM cronapp.jobs
		WHERE
			(
				status = 'queued'
				AND scheduled_for = current_date
				AND daily_run_time <= now()::time
			)
			OR (
				status = 'running'
				AND lock_until < now()
			)
		ORDER BY daily_run_time, id
		FOR UPDATE SKIP LOCKED
		LIMIT 1
	)
	UPDATE cronapp.jobs j
	SET
		status = 'running',
		locked_by = $1,
		lock_until = now() + $2::interval,
		updated_at = now()
	FROM picked
	WHERE j.id = picked.id
	RETURNING
		j.id,
		j.type,
		j.status,
		j.daily_run_time,
		j.attempt,
		j.max_retries,
		j.is_repetable,
		j.http_method,
		j.http_url,
		j.db_action,
		j.target_db,
		j.last_error,
		j.locked_by,
		j.lock_until,
		j.updated_at,
		j.finished_at,
		j.created_at;`

	var job domain.Job

	err := r.pool.QueryRow(ctx, query, workerID, leaseDuration).Scan(
		&job.Id,
		&job.JobType,
		&job.Status,
		&job.DailyRunTime,
		&job.Attempt,
		&job.MaxRetries,
		&job.IsRepetable,
		&job.HttpMethod,
		&job.HttpURL,
		&job.DbAction,
		&job.TargetDB,
		&job.LastError,
		&job.LockedBy,
		&job.LockUntil,
		&job.UpdatedAt,
		&job.FinishedAt,
		&job.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Job{}, false, nil
		}

		return domain.Job{}, false, fmt.Errorf("claim job: %w", err)
	}

	return job, true, nil
}

func (r *WorkerRepository) CompleteJob(
	ctx context.Context,
	jobID core_job_types.ID,
	workerID int64,
	IsRepetable bool,
) error {
	var newStatus core_job_types.Status
	if IsRepetable != true {
		newStatus = core_job_types.StatusCanceled
	} else {
		newStatus = core_job_types.StatusSucceeded
	}

	query := `
	UPDATE cronapp.jobs
	SET
		status = $1,
		last_error = NULL,
		locked_by = NULL,
		lock_until = NULL,
		finished_at = now(),
		updated_at = now()
	WHERE
		id = $2
		AND locked_by = $3
		AND status = 'running';`

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	return r.execOwnedJobUpdate(ctx, query, newStatus, jobID, workerID)
}

func (r *WorkerRepository) FailJob(
	ctx context.Context,
	jobID core_job_types.ID,
	workerID int64,
	cause error,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	UPDATE cronapp.jobs
	SET
		status = 'failed',
		last_error = $3,
		locked_by = NULL,
		lock_until = NULL,
		finished_at = now(),
		updated_at = now()
	WHERE
		id = $1
		AND locked_by = $2
		AND status = 'running';`

	return r.execOwnedJobUpdate(ctx, query, jobID, workerID, errorText(cause))
}

func (r *WorkerRepository) ReleaseForRetry(
	ctx context.Context,
	jobID core_job_types.ID,
	workerID int64,
	cause error,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	UPDATE cronapp.jobs
	SET
		status = 'queued',
		attempt = attempt + 1,
		last_error = $3,
		locked_by = NULL,
		lock_until = NULL,
		updated_at = now()
	WHERE
		id = $1
		AND locked_by = $2
		AND status = 'running';`

	return r.execOwnedJobUpdate(ctx, query, jobID, workerID, errorText(cause))
}

func (r *WorkerRepository) InsertAttempt(ctx context.Context, attempt Attempt) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	INSERT INTO cronapp.attempts (
		job_id,
		result,
		http_status,
		error_message,
		started_at
	)
	VALUES ($1, $2, $3, $4, $5);`

	_, err := r.pool.Exec(
		ctx,
		query,
		attempt.JobID,
		attempt.Result,
		attempt.HTTPStatus,
		attempt.ErrorMessage,
		attempt.StartedAt,
	)
	if err != nil {
		return fmt.Errorf("insert attempt: %w", err)
	}

	return nil
}

func (r *WorkerRepository) execOwnedJobUpdate(
	ctx context.Context,
	query string,
	args ...any,
) error {
	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("job is not owned by worker or not running: %w", core_errors.ErrNotFound)
	}

	return nil
}

func errorText(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}
