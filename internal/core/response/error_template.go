package core_error_tamplate

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	core_errors "github.com/DimaKirejko/Dstributed_cron/internal/core/errors"
	core_logger "github.com/DimaKirejko/Dstributed_cron/internal/core/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func JSON(c *gin.Context, statusCode int, body any) {
	c.JSON(statusCode, body)
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func Error(
	c *gin.Context,
	log *core_logger.Logger,
	err error,
	msg string,
) {
	statusCode, logFunc := mapError(log, err)

	logFunc(msg, zap.Error(err))

	c.JSON(statusCode, ErrorResponse{
		Error:   err.Error(),
		Message: msg,
	})
}

func Panic(
	c *gin.Context,
	log *core_logger.Logger,
	p any,
	msg string,
) {
	err := fmt.Errorf("unexpected panic: %v", p)
	stackString := strings.Split(string(debug.Stack()), "\n")

	log.Error("PANIC stackTrace:", zap.Strings("stack", stackString))
	log.Error(msg, zap.Error(err))

	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error:   err.Error(),
		Message: msg,
	})
}

func mapError(
	log *core_logger.Logger,
	err error,
) (int, func(string, ...zap.Field)) {
	switch {
	case errors.Is(err, core_errors.ErrInvalidArgument):
		return http.StatusBadRequest, log.Warn

	case errors.Is(err, core_errors.ErrNotFound):
		return http.StatusNotFound, log.Warn

	default:
		return http.StatusInternalServerError, log.Error
	}
}
