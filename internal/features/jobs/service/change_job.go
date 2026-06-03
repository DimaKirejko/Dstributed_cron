package jobs_service

import (
	"context"
	"fmt"

	"github.com/DimaKirejko/Dstributed_cron/internal/core/domain"
	core_job_types "github.com/DimaKirejko/Dstributed_cron/internal/core/types"
)

func (s *JobService) CencelJob(
	ctx context.Context,
	id core_job_types.ID,
) (domain.Job, error) {
	job, err := s.jobRepository.ChangeJob(ctx, id, core_job_types.StatusCanceled)
	if err != nil {
		return job, fmt.Errorf("filed to cencel Job: %w", err)
	}

	return job, err
}

func (s *JobService) RerunJob(
	ctx context.Context,
	id core_job_types.ID,
) (domain.Job, error) {
	job, err := s.jobRepository.ChangeJob(ctx, id, core_job_types.StatusQueued)
	if err != nil {
		return job, fmt.Errorf("filed to rerun Job: %w", err)
	}

	return job, err
}
