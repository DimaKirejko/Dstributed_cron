package jobs_transport_http

import (
	"fmt"
	"net/http"

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
		"http_url":       job.HttpURL,
		"db_action":      job.DbAction,
		"last_error":     job.LastError,
		"locked_by":      job.LockedBy,
		"updated_at":     job.UpdatedAt,
		"finished_at":    job.FinishedAt,
		"created_at":     job.CreatedAt,
	})
}
