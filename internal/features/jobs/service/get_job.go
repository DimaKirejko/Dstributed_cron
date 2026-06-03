package jobs_service

import (
	"context"
	"fmt"

	"github.com/DimaKirejko/Dstributed_cron/internal/core/domain"
	core_job_types "github.com/DimaKirejko/Dstributed_cron/internal/core/types"
)

func (s *JobService) GetJob(
	ctx context.Context,
	id core_job_types.ID,
) (domain.Job, error) {
	job, err := s.jobRepository.GetJob(ctx, id)
	if err != nil {
		return domain.Job{}, fmt.Errorf("filed to get Job: %w", err)
	}

	return job, err
}
