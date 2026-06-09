package jobs_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/DimaKirejko/Dstributed_cron/internal/core/domain"
	core_errors "github.com/DimaKirejko/Dstributed_cron/internal/core/errors"
	core_job_types "github.com/DimaKirejko/Dstributed_cron/internal/core/types"
	"github.com/jackc/pgx/v5"
)

func (r *JobRepository) GetJob(
	ctx context.Context,
	id core_job_types.ID,
) (domain.Job, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT 
		id,
		type,
		status, 
		daily_run_time, 
		attempt, 
		max_retries,
		is_repetable,
		http_method, 
		http_url, 
		db_action, 
		target_db, 
		last_error, 
		locked_by, 
		lock_until, 
		updated_at, 
		finished_at, 
		created_at
	from cronapp.jobs
	WHERE id=$1;`

	row := r.pool.QueryRow(ctx, query, id)

	var job domain.Job

	err := row.Scan(
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
			return domain.Job{}, fmt.Errorf("job with id = '%d' : %w", id, core_errors.ErrNotFound)
		}

		return domain.Job{}, fmt.Errorf("scan error: = '%d' : %w", id, err)
	}

	return job, nil
}
