package jobs_repository

import (
	"context"
	"fmt"

	"github.com/DimaKirejko/Dstributed_cron/internal/core/domain"
	core_job_types "github.com/DimaKirejko/Dstributed_cron/internal/core/types"
)

func (r *JobRepository) GetJobAttempts(
	ctx context.Context,
	jobID core_job_types.ID,
) ([]domain.Attempt, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT
		id,
		job_id,
		result,
		http_status,
		error_message,
		started_at
	FROM cronapp.attempts
	WHERE job_id = $1
	ORDER BY started_at DESC, id DESC;`

	rows, err := r.pool.Query(ctx, query, jobID)
	if err != nil {
		return nil, fmt.Errorf("query attempts: %w", err)
	}
	defer rows.Close()

	var attempts []domain.Attempt
	for rows.Next() {
		var attempt domain.Attempt
		if err := rows.Scan(
			&attempt.ID,
			&attempt.JobID,
			&attempt.Result,
			&attempt.HTTPStatus,
			&attempt.ErrorMessage,
			&attempt.StartedAt,
		); err != nil {
			return nil, fmt.Errorf("scan attempt: %w", err)
		}

		attempts = append(attempts, attempt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate attempts: %w", err)
	}

	return attempts, nil
}
