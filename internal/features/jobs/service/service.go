package jobs_service

import (
	"context"

	"github.com/DimaKirejko/Dstributed_cron/internal/core/domain"
	core_job_types "github.com/DimaKirejko/Dstributed_cron/internal/core/types"
)

type JobService struct {
	jobRepository JobRepository
}

type JobRepository interface {
	Create_job(
		ctx context.Context,
		job domain.Job,
	) (core_job_types.ID, error)
}

func NewJobsService(jobsrepository JobRepository) *JobService {
	return &JobService{
		jobRepository: jobsrepository,
	}
}
