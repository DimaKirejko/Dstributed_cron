package jobs_transport_http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/DimaKirejko/Dstributed_cron/internal/core/domain"
	core_errors "github.com/DimaKirejko/Dstributed_cron/internal/core/errors"
	core_logger "github.com/DimaKirejko/Dstributed_cron/internal/core/logger"
	core_error_tamplate "github.com/DimaKirejko/Dstributed_cron/internal/core/response"
	core_job_types "github.com/DimaKirejko/Dstributed_cron/internal/core/types"
	"github.com/gin-gonic/gin"
)

type ChangeJob struct {
	ID core_job_types.ID `json:"id" binding:"required"`
}

type ChangeJobRespDTO struct {
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

// CancelJob cancels a job.
//
// @Summary Cancel job
// @Description Cancels an existing job by id and returns the updated job.
// @Tags jobs
// @Accept json
// @Produce json
// @Param request body ChangeJob true "Cancel job request"
// @Success 200 {object} ChangeJobRespDTO
// @Failure 400 {object} core_error_tamplate.ErrorResponse
// @Failure 404 {object} core_error_tamplate.ErrorResponse
// @Failure 500 {object} core_error_tamplate.ErrorResponse
// @Router /jobs/cencel [patch]
func (h *JobsHttpHandler) CancelJob(c *gin.Context) {
	var req ChangeJob
	logger := core_logger.FromContext(c.Request.Context())

	if err := c.ShouldBindJSON(&req); err != nil {
		core_error_tamplate.Error(
			c,
			logger,
			fmt.Errorf("%w: decode json: %v", core_errors.ErrInvalidArgument, err),
			"failed to cancel job",
		)

		return
	}

	job, err := h.jobsService.CencelJob(c.Request.Context(), req.ID)
	if err != nil {
		core_error_tamplate.Error(
			c,
			logger,
			err,
			"failed to cancel job",
		)

		return
	}

	successRespWithJob(c, job)
}

// RerunJob reruns a job.
//
// @Summary Rerun job
// @Description Schedules an existing job for rerun by id and returns the updated job.
// @Tags jobs
// @Accept json
// @Produce json
// @Param request body ChangeJob true "Rerun job request"
// @Success 200 {object} ChangeJobRespDTO
// @Failure 400 {object} core_error_tamplate.ErrorResponse
// @Failure 404 {object} core_error_tamplate.ErrorResponse
// @Failure 500 {object} core_error_tamplate.ErrorResponse
// @Router /jobs/rerun [patch]
func (h *JobsHttpHandler) RerunJob(c *gin.Context) {
	var req ChangeJob
	logger := core_logger.FromContext(c.Request.Context())

	if err := c.ShouldBindJSON(&req); err != nil {
		core_error_tamplate.Error(
			c,
			logger,
			fmt.Errorf("%w: decode json: %v", core_errors.ErrInvalidArgument, err),
			"failed to rerun job",
		)

		return
	}

	job, err := h.jobsService.RerunJob(c.Request.Context(), req.ID)
	if err != nil {
		core_error_tamplate.Error(
			c,
			logger,
			err,
			"failed to rerun job",
		)

		return
	}

	successRespWithJob(c, job)
}

func successRespWithJob(c *gin.Context, job domain.Job) {
	c.JSON(http.StatusOK, gin.H{
		"id":             job.Id,
		"job_type":       job.JobType,
		"status":         job.Status,
		"daily_run_time": job.DailyRunTime,
		"attempt":        job.Attempt,
		"max_retries":    job.MaxRetries,
		"is_repetable":   job.IsRepetable,
		"http_url":       job.HttpURL,
		"db_action":      job.DbAction,
		"target_db":      job.TargetDB,
		"last_error":     job.LastError,
		"locked_by":      job.LockedBy,
		"updated_at":     job.UpdatedAt,
		"finished_at":    job.FinishedAt,
		"created_at":     job.CreatedAt,
	})
}
