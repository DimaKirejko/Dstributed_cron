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
	MaxRetries   int                 `json:"max_retries" binding:"omitempty,min=1,max=5"`
	IsRepetable  *bool               `json:"is_repetable"`

	HTTPMethod *core_job_types.HTTPMethod `json:"http_method" binding:"omitempty,oneof=GET POST DELETE PUT PATCH"`
	HTTPURL    *string                    `json:"http_url"`

	DBAction *core_job_types.DBAction `json:"db_action" binding:"omitempty"`
	TargetDB *string                  `json:"target_db" binding:"omitempty"`
}

type CreateJobRespDTO struct {
	ID core_job_types.ID `json:"id"`
}

// CreateJob creates a new job.
//
// @Summary Create job
// @Description Creates a new distributed cron job. For type "http", provide http_method and http_url. For type "db", provide db_action and target_db.
// @Tags jobs
// @Accept json
// @Produce json
// @Param request body CreateJobReqDTO true "Create job request"
// @Success 201 {object} CreateJobRespDTO
// @Failure 400 {object} core_error_tamplate.ErrorResponse
// @Failure 500 {object} core_error_tamplate.ErrorResponse
// @Router /jobs [post]
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

	jobID, err := h.jobsService.CreateJob(c.Request.Context(), &jobs_service.CreateJobInput{
		Type:         req.Type,
		DailyRunTime: req.DailyRunTime,
		MaxRetries:   req.MaxRetries,
		IsRepetable:  req.IsRepetable,
		HTTPMethod:   req.HTTPMethod,
		HTTPURL:      req.HTTPURL,
		DBAction:     req.DBAction,
		TargetDB:     req.TargetDB,
	})
	if err != nil {
		core_error_tamplate.Error(c, logger, err, "failed to create job")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": jobID,
	})

}
