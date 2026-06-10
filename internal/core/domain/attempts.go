package domain

import (
	"time"

	core_job_types "github.com/DimaKirejko/Dstributed_cron/internal/core/types"
)

type Attempt struct {
	ID           int64
	JobID        core_job_types.ID
	Result       string
	HTTPStatus   *int
	ErrorMessage *string
	StartedAt    time.Time
}
