package domain

import (
	"time"

	core_job_types "github.com/DimaKirejko/Dstributed_cron/internal/core/types"
)

type Job struct {
	Id core_job_types.ID

	JobType core_job_types.Type
	Status  core_job_types.Status

	DailyRunTime string
	Attempt      int
	MaxRetries   int
	IsRepetable  bool

	HttpMethod *core_job_types.HTTPMethod
	HttpURL    *string

	DbAction *core_job_types.DBAction
	TargetDB *string

	LastError *string

	LockedBy  *int64
	LockUntil *time.Time

	UpdatedAt  *time.Time
	FinishedAt *time.Time
	CreatedAt  time.Time
}

const defaultMaxRetries = 3

func NewJob(
	jobType core_job_types.Type,
	dailyRunTime string,
	maxRetries int,
	isRepetable *bool,

	httpMethod *core_job_types.HTTPMethod,
	httpURL *string,

	dbAction *core_job_types.DBAction,
	targetDb *string,
) Job {
	now := time.Now().UTC()

	if maxRetries == 0 {
		maxRetries = defaultMaxRetries
	}

	return Job{
		Id: core_job_types.UninitializedID,

		JobType: jobType,
		Status:  core_job_types.StatusQueued,

		DailyRunTime: dailyRunTime,
		Attempt:      0,
		MaxRetries:   maxRetries,
		IsRepetable:  *isRepetable,

		HttpMethod: httpMethod,
		HttpURL:    httpURL,

		DbAction: dbAction,
		TargetDB: targetDb,

		LastError: nil,

		LockedBy:  nil,
		LockUntil: nil,

		UpdatedAt:  &now,
		FinishedAt: nil,
		CreatedAt:  now,
	}
}
