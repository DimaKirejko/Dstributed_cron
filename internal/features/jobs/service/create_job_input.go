package jobs_service

import (
	core_job_types "github.com/DimaKirejko/Dstributed_cron/internal/core/types"
)

type CreateJobInput struct {
	Type         core_job_types.Type
	DailyRunTime string
	MaxRetries   int

	HTTPMethod *core_job_types.HTTPMethod
	HTTPURL    *string

	DBAction *core_job_types.DBAction
}
