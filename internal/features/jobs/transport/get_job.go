package jobs_transport_http

import (
	"fmt"
	"net/http"
	"time"

	core_errors "github.com/DimaKirejko/Dstributed_cron/internal/core/errors"
	core_logger "github.com/DimaKirejko/Dstributed_cron/internal/core/logger"
	core_error_tamplate "github.com/DimaKirejko/Dstributed_cron/internal/core/response"
	core_job_types "github.com/DimaKirejko/Dstributed_cron/internal/core/types"
	"github.com/gin-gonic/gin"
)

type GetJobDTO struct {
	ID core_job_types.ID `json:"id" binding:"required"`
}

type GetJobRespDTO struct {
	ID           core_job_types.ID        `json:"id"`
	JobType      core_job_types.Type      `json:"job_type"`
	Status       core_job_types.Status    `json:"status"`
	DailyRunTime string                   `json:"daily_run_time"`
	Attempt      int                      `json:"attempt"`
	MaxRetries   int                      `json:"max_retries"`
	IsRepetable  bool                     `json:"is_repetable"`
	HTTPURL      *string                  `json:"http_url"`
	DBAction     *core_job_types.DBAction `json:"db_action"`
	TargetDB     *string                  `json:"target_db"`
	LastError    *string                  `json:"last_error"`
	LockedBy     *int64                   `json:"locked_by"`
	UpdatedAt    *time.Time               `json:"updated_at"`
	FinishedAt   *time.Time               `json:"finished_at"`
	CreatedAt    time.Time                `json:"created_at"`
}

// GetJob returns a job by id.
//
// @Summary Get job
// @Description Returns a job by id.
// @Tags jobs
// @Accept json
// @Produce json
// @Param request body GetJobDTO true "Get job request"
// @Success 200 {object} GetJobRespDTO
// @Failure 400 {object} core_error_tamplate.ErrorResponse
// @Failure 404 {object} core_error_tamplate.ErrorResponse
// @Failure 500 {object} core_error_tamplate.ErrorResponse
// @Router /jobs [get]
func (h *JobsHttpHandler) GetJob(c *gin.Context) {
	var req GetJobDTO
	logger := core_logger.FromContext(c.Request.Context())

	if err := c.ShouldBindJSON(&req); err != nil {
		core_error_tamplate.Error(
			c,
			logger,
			fmt.Errorf("%w: decode json: %v", core_errors.ErrInvalidArgument, err),
			"failed to get job",
		)

		return
	}

	job, err := h.jobsService.GetJob(c.Request.Context(), req.ID)
	if err != nil {
		core_error_tamplate.Error(c, logger, err, "failed to get job")
		return
	}

	resp := GetJobRespDTO{
		ID:           job.Id,
		JobType:      job.JobType,
		Status:       job.Status,
		DailyRunTime: job.DailyRunTime,
		Attempt:      job.Attempt,
		MaxRetries:   job.MaxRetries,
		IsRepetable:  job.IsRepetable,
		HTTPURL:      job.HttpURL,
		DBAction:     job.DbAction,
		TargetDB:     job.TargetDB,
		LastError:    job.LastError,
		LockedBy:     job.LockedBy,
		UpdatedAt:    job.UpdatedAt,
		FinishedAt:   job.FinishedAt,
		CreatedAt:    job.CreatedAt,
	}

	c.JSON(http.StatusOK, resp)

}
