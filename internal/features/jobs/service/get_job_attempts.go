package jobs_service

import (
	"context"
	"fmt"

	"github.com/DimaKirejko/Dstributed_cron/internal/core/domain"
	core_job_types "github.com/DimaKirejko/Dstributed_cron/internal/core/types"
)

func (s *JobService) GetJobAttempts(
	ctx context.Context,
	jobID core_job_types.ID,
) ([]domain.Attempt, error) {
	attempts, err := s.jobRepository.GetJobAttempts(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("get job attempts: %w", err)
	}

	return attempts, nil
}
