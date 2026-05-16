package jobs_service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/DimaKirejko/Dstributed_cron/internal/core/domain"
	core_errors "github.com/DimaKirejko/Dstributed_cron/internal/core/errors"
	core_job_types "github.com/DimaKirejko/Dstributed_cron/internal/core/types"
)

func (s *JobService) CreateJob(
	ctx context.Context,
	input CreateJobInput,
) (core_job_types.ID, error) {
	if err := validate(input); err != nil {
		return core_job_types.UninitializedID, fmt.Errorf(
			"%w: failed to validate request: %v",
			core_errors.ErrInvalidArgument,
			err,
		)
	}

	job := domain.NewJob(
		input.Type,
		input.DailyRunTime,
		input.MaxRetries,
		input.HTTPMethod,
		input.HTTPURL,
		input.DBAction,
	)

	id, err := s.jobRepository.Create_job(ctx, job)
	if err != nil {
		return core_job_types.UninitializedID, fmt.Errorf("create job: %w", err)
	}

	return id, nil

}

func validate(input CreateJobInput) error {
	if _, err := time.Parse("15:04", input.DailyRunTime); err != nil {
		return fmt.Errorf("%w: 'run_at' is required to be in 'HH:MM' format", core_errors.ErrInvalidArgument)
	}

	switch input.Type {
	case core_job_types.TypeHTTP:
		if input.HTTPMethod == nil {
			return fmt.Errorf("%w: 'http_method' is required for http job", core_errors.ErrInvalidArgument)
		}

		if input.HTTPURL == nil || strings.TrimSpace(*input.HTTPURL) == "" {
			return fmt.Errorf("%w: 'http_url' is required for http job", core_errors.ErrInvalidArgument)
		}

		if input.DBAction != nil {
			return fmt.Errorf("%w: 'db_action' is not allowed for http job", core_errors.ErrInvalidArgument)
		}

	case core_job_types.TypeDB:
		if input.DBAction == nil {
			return fmt.Errorf("%w: 'db_action' is required for db job", core_errors.ErrInvalidArgument)
		}

		if input.HTTPMethod != nil {
			return fmt.Errorf("%w: 'http_method' is not allowed for db job", core_errors.ErrInvalidArgument)
		}

		if input.HTTPURL != nil {
			return fmt.Errorf("%w: 'http_url' is not allowed for db job", core_errors.ErrInvalidArgument)
		}

	default:
		return fmt.Errorf("%w: unsupported job type: %s", core_errors.ErrInvalidArgument, input.Type)
	}

	return nil
}
