package metrics

import (
	"context"
	"fmt"

	core_postgres_pool "github.com/DimaKirejko/Dstributed_cron/internal/core/repository/postgres_pgx"
)

type MetricsRepository struct {
	pool core_postgres_pool.PgxPool
}

func NewRepository(
	pool core_postgres_pool.PgxPool,
) *MetricsRepository {
	return &MetricsRepository{
		pool: pool,
	}
}

func (r *MetricsRepository) GetFailedTasksCount(ctx context.Context) (float64, error) {
	query := `
	SELECT COUNT(*)
	FROM cronapp.jobs
	WHERE status = 'failed';
	`

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var FailedTasksCount float64

	row := r.pool.QueryRow(ctx, query)

	if err := row.Scan(&FailedTasksCount); err != nil {
		return 0, fmt.Errorf(
			"failed to create DB job: %v",
			err,
		)
	}

	return FailedTasksCount, nil
}
