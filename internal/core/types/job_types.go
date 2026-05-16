package core_job_types

import "net/http"

type ID int64

const (
	UninitializedID = -1
)

type Type string

const (
	TypeHTTP Type = "http"
	TypeDB   Type = "db"
)

type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
	StatusCanceled  Status = "canceled"
)

type HTTPMethod string

const (
	HTTPGet  HTTPMethod = http.MethodGet
	HTTPPost HTTPMethod = http.MethodPost
	// HTTPPut    HTTPMethod = http.MethodPut
	// HTTPPatch  HTTPMethod = http.MethodPatch
	HTTPDelete HTTPMethod = http.MethodDelete
)

type DBAction string

const (
	DBActionVacuum DBAction = "create_partition"
)
