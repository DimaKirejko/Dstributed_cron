package jobs_transport_http

import (
	"fmt"
	"net/http"

	core_errors "github.com/DimaKirejko/Dstributed_cron/internal/core/errors"
	core_logger "github.com/DimaKirejko/Dstributed_cron/internal/core/logger"
	core_error_tamplate "github.com/DimaKirejko/Dstributed_cron/internal/core/response"
	core_job_types "github.com/DimaKirejko/Dstributed_cron/internal/core/types"
	jobs_service "github.com/DimaKirejko/Dstributed_cron/internal/features/jobs/service"
	"github.com/gin-gonic/gin"
)

type CreateJobReqDTO struct {
	Type         core_job_types.Type `json:"type" binding:"required,oneof=http db"`
	DailyRunTime string              `json:"run_at" binding:"required"`
	MaxRetries   int                 `json:"max_retries" binding:"omitempty,min=1"`

	HTTPMethod *core_job_types.HTTPMethod `json:"http_method" binding:"omitempty,oneof=GET POST DELETE"`
	HTTPURL    *string                    `json:"http_url"`

	DBAction *core_job_types.DBAction `json:"db_action" binding:"omitempty,oneof=create_partition"`
}

func (h *JobsHttpHandler) CreateJob(c *gin.Context) {
	var req CreateJobReqDTO
	logger := core_logger.FromContext(c.Request.Context())

	if err := c.ShouldBindJSON(&req); err != nil {
		core_error_tamplate.Error(
			c,
			logger,
			fmt.Errorf("%w: decode json: %v", core_errors.ErrInvalidArgument, err),
			"failed to create job",
		)

		return
	}

	jobID, err := h.jobsService.CreateJob(c.Request.Context(), jobs_service.CreateJobInput{
		Type:         req.Type,
		DailyRunTime: req.DailyRunTime,
		MaxRetries:   req.MaxRetries,
		HTTPMethod:   req.HTTPMethod,
		HTTPURL:      req.HTTPURL,
		DBAction:     req.DBAction,
	})
	if err != nil {
		core_error_tamplate.Error(c, logger, err, "failed to create job")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": jobID,
	})

}
