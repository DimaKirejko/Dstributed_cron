package jobs_transport_http

import (
	"context"
	"net/http"

	core_http_server "github.com/DimaKirejko/Dstributed_cron/internal/core/transport/server"
	core_job_types "github.com/DimaKirejko/Dstributed_cron/internal/core/types"
	jobs_service "github.com/DimaKirejko/Dstributed_cron/internal/features/jobs/service"
)

type JobsHttpHandler struct {
	jobsService JobsService
}

type JobsService interface {
	CreateJob(
		ctx context.Context,
		input jobs_service.CreateJobInput,
	) (core_job_types.ID, error)
}

func NewJobsHttpHandler(jobsService JobsService) *JobsHttpHandler {
	return &JobsHttpHandler{
		jobsService: jobsService,
	}
}

func (h *JobsHttpHandler) Route() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/jobs",
			Handler: h.CreateJob,
		},
	}
}
