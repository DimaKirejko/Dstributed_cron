package jobs_repository

import (
	"context"
	"fmt"

	"github.com/DimaKirejko/Dstributed_cron/internal/core/domain"
	core_job_types "github.com/DimaKirejko/Dstributed_cron/internal/core/types"
)

func (r *JobRepository) Create_job(
	ctx context.Context,
	job domain.Job,
) (core_job_types.ID, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	INSERT INTO cronapp.jobs (
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
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	RETURNING id;`

	row := r.pool.QueryRow(
		ctx,
		query,
		job.JobType,
		job.Status,
		job.DailyRunTime,
		job.Attempt,
		job.MaxRetries,
		job.IsRepetable,
		job.HttpMethod,
		job.HttpURL,
		job.DbAction,
		job.TargetDB,
		job.LastError,
		job.LockedBy,
		job.LockUntil,
		job.UpdatedAt,
		job.FinishedAt,
		job.CreatedAt,
	)

	var taskID core_job_types.ID

	if err := row.Scan(&taskID); err != nil {
		return core_job_types.UninitializedID, fmt.Errorf(
			"failed to create DB job: %v",
			err,
		)
	}

	return taskID, nil
}
