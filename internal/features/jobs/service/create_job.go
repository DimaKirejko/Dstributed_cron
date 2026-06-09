package jobs_service

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/DimaKirejko/Dstributed_cron/internal/core/domain"
	core_errors "github.com/DimaKirejko/Dstributed_cron/internal/core/errors"
	core_job_types "github.com/DimaKirejko/Dstributed_cron/internal/core/types"
	core_worker "github.com/DimaKirejko/Dstributed_cron/internal/features/worker"
)

var validIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func (s *JobService) CreateJob(
	ctx context.Context,
	input *CreateJobInput,
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
		input.IsRepetable,
		input.HTTPMethod,
		input.HTTPURL,
		input.DBAction,
		input.TargetDB,
	)

	id, err := s.jobRepository.Create_job(ctx, job)
	if err != nil {
		return core_job_types.UninitializedID, fmt.Errorf("create job: %w", err)
	}

	return id, nil

}

func validate(input *CreateJobInput) error {
	if _, err := time.Parse("15:04", input.DailyRunTime); err != nil {
		return fmt.Errorf("%w: 'run_at' is required to be in 'HH:MM' format", core_errors.ErrInvalidArgument)
	}

	if input.IsRepetable == nil {
		defaultIsRepetable := true
		input.IsRepetable = &defaultIsRepetable
	}

	switch input.Type {
	case core_job_types.TypeHTTP:
		if input.HTTPMethod == nil {
			return fmt.Errorf("%w: 'http_method' is required for http job", core_errors.ErrInvalidArgument)
		}

		if input.HTTPURL == nil || strings.TrimSpace(*input.HTTPURL) == "" {
			return fmt.Errorf("%w: 'http_url' is required for http job", core_errors.ErrInvalidArgument)
		}

		if err := validateHTTPURL(*input.HTTPURL); err != nil {
			return fmt.Errorf("%w: 'http_url' is not valid: %s", core_errors.ErrInvalidArgument, *input.HTTPURL)
		}

		if input.DBAction != nil {
			return fmt.Errorf("%w: 'db_action' is not allowed for http job", core_errors.ErrInvalidArgument)
		}

	case core_job_types.TypeDB:
		if input.DBAction == nil {
			return fmt.Errorf("%w: 'db_action' is required for db job", core_errors.ErrInvalidArgument)
		}

		if !isValidDBAction(*input.DBAction) {
			return fmt.Errorf("%w: unsupported db_action: %s", core_errors.ErrInvalidArgument, *input.DBAction)
		}

		if input.TargetDB == nil || strings.TrimSpace(*input.TargetDB) == "" {
			return fmt.Errorf("%w: 'target_db' is required for db job", core_errors.ErrInvalidArgument)
		}

		if err := checkSchemaTableRelation(*input.TargetDB); err != nil {
			return fmt.Errorf("'target_db' must be in 'schema.table' format: %w", err)
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

func checkSchemaTableRelation(input string) error {
	parts := strings.Split(input, ".")
	if len(parts) != 2 {
		return fmt.Errorf("target table must be in schema.table format: %q", input)
	}

	if !validIdentifier.MatchString(parts[0]) {
		return fmt.Errorf("invalid target schema: %q", parts[0])
	}

	if !validIdentifier.MatchString(parts[1]) {
		return fmt.Errorf("invalid target table: %q", parts[1])
	}

	return nil
}

func isValidDBAction(action core_job_types.DBAction) bool {
	for _, valid := range core_job_types.AllDBActions {
		if action == valid {
			return true
		}
	}

	return false
}

func validateHTTPURL(HTTPURL string) error {
	parsedURL, err := url.Parse(HTTPURL)
	if err != nil {
		return err
	}

	config := core_worker.NewConfigMust()

	if err := core_worker.CheckHTTPAllowlist(parsedURL, config.HTTPAllowlist); err != nil {
		return err
	}

	return nil
}
