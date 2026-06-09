package core_scheduler

import (
	"context"

	core_logger "github.com/DimaKirejko/Dstributed_cron/internal/core/logger"
	core_postgres_pool "github.com/DimaKirejko/Dstributed_cron/internal/core/repository/postgres_pgx"
)

type SchedulerRepository struct {
	pool   core_postgres_pool.PgxPool
	config Config
}

func NewSchedulerRepository(pool core_postgres_pool.PgxPool, config Config) *SchedulerRepository {
	return &SchedulerRepository{
		pool:   pool,
		config: config,
	}
}

func (r *SchedulerRepository) QueueDueDailyJobs(ctx context.Context, l *core_logger.Logger) (int64, error) {
	l.Logger.Debug("Start Scheduler check")
	query := r.queueDueDailyJobsQuery()

	tag, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
}

func (r *SchedulerRepository) queueDueDailyJobsQuery() string {
	if r.config.IsTestMode {
		return `
        UPDATE cronapp.jobs
        SET
            status = 'queued',
            scheduled_for = current_date,
            attempt = 0,
            last_error = NULL,
            locked_by = NULL,
            lock_until = NULL,
            updated_at = now(),
            finished_at = NULL
        WHERE
            status IN ('queued', 'succeeded', 'failed')
            AND daily_run_time <= now()::time
        AND (
            scheduled_for IS NULL
            OR scheduled_for < current_date
        );
    `
	}

	return `
        UPDATE cronapp.jobs
        SET
            status = 'queued',
            scheduled_for = CASE
                WHEN scheduled_for IS NULL
                    AND created_at::date = current_date
                    AND created_at::time > daily_run_time
                THEN current_date + 1
                ELSE current_date
            END,
            attempt = 0,
            last_error = NULL,
            locked_by = NULL,
            lock_until = NULL,
            updated_at = now(),
            finished_at = NULL
        WHERE
            status IN ('queued', 'succeeded', 'failed')
            AND daily_run_time <= now()::time
            AND (
                scheduled_for IS NULL
                OR scheduled_for < current_date
            );
    `
}
