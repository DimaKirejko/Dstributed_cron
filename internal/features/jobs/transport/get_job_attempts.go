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

type GetJobAttemptsReqDTO struct {
	JobID core_job_types.ID `form:"job_id" binding:"required"`
}

type JobAttemptRespDTO struct {
	ID           int64             `json:"id"`
	JobID        core_job_types.ID `json:"job_id"`
	Result       string            `json:"result"`
	HTTPStatus   *int              `json:"http_status"`
	ErrorMessage *string           `json:"error_message"`
	StartedAt    time.Time         `json:"started_at"`
}

type GetJobAttemptsRespDTO struct {
	Attempts []JobAttemptRespDTO `json:"attempts"`
}

// GetJobAttempts returns job attempt history.
//
// @Summary Get job attempts
// @Description Returns all attempts for a job by job_id.
// @Tags jobs
// @Produce json
// @Param job_id query int true "Job id"
// @Success 200 {object} GetJobAttemptsRespDTO
// @Failure 400 {object} core_error_tamplate.ErrorResponse
// @Failure 500 {object} core_error_tamplate.ErrorResponse
// @Router /jobs/attempts [get]
func (h *JobsHttpHandler) GetJobAttempts(c *gin.Context) {
	var req GetJobAttemptsReqDTO
	logger := core_logger.FromContext(c.Request.Context())

	if err := c.ShouldBindQuery(&req); err != nil {
		core_error_tamplate.Error(
			c,
			logger,
			fmt.Errorf("%w: decode query: %v", core_errors.ErrInvalidArgument, err),
			"failed to get job attempts",
		)

		return
	}

	attempts, err := h.jobsService.GetJobAttempts(c.Request.Context(), req.JobID)
	if err != nil {
		core_error_tamplate.Error(c, logger, err, "failed to get job attempts")
		return
	}

	resp := GetJobAttemptsRespDTO{
		Attempts: make([]JobAttemptRespDTO, 0, len(attempts)),
	}

	for _, attempt := range attempts {
		resp.Attempts = append(resp.Attempts, JobAttemptRespDTO{
			ID:           attempt.ID,
			JobID:        attempt.JobID,
			Result:       attempt.Result,
			HTTPStatus:   attempt.HTTPStatus,
			ErrorMessage: attempt.ErrorMessage,
			StartedAt:    attempt.StartedAt,
		})
	}

	c.JSON(http.StatusOK, resp)
}
