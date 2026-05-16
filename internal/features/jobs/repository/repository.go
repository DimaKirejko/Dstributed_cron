package jobs_repository

import core_postgres_pool "github.com/DimaKirejko/Dstributed_cron/internal/core/repository/postgres_pgx"

type JobRepository struct {
	pool core_postgres_pool.PgxPool
}

func NewJobRepository(
	pool core_postgres_pool.PgxPool,
) *JobRepository {
	return &JobRepository{
		pool: pool,
	}
}
