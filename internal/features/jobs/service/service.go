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

	GetJob(
		ctx context.Context,
		id core_job_types.ID,
	) (domain.Job, error)

	ChangeJob(
		ctx context.Context,
		id core_job_types.ID,
		newStatus core_job_types.Status,
	) (domain.Job, error)
}

func NewJobsService(jobsrepository JobRepository) *JobService {
	return &JobService{
		jobRepository: jobsrepository,
	}
}
