package core_http_middleware

import (
	"time"

	core_logger "github.com/DimaKirejko/Dstributed_cron/internal/core/logger"
	core_error_tamplate "github.com/DimaKirejko/Dstributed_cron/internal/core/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	xRequestIDHeader = "X-Request-ID"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(xRequestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Request.Header.Set(xRequestIDHeader, requestID)
		c.Header(xRequestIDHeader, requestID)

		c.Next()
	}
}

func Logger(log *core_logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(xRequestIDHeader)

		l := log.With(
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("url", c.Request.URL.String()),
			zap.String("path", c.FullPath()),
			zap.String("client_ip", c.ClientIP()),
		)

		ctx := core_logger.ToContext(c.Request.Context(), l)
		c.Request = c.Request.WithContext(ctx)

		c.Set("logger", l)

		c.Next()
	}
}

func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := core_logger.FromContext(c.Request.Context())

		startedAt := time.Now()

		log.Debug(
			">>> incoming HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.Time("time", startedAt.UTC()),
		)

		c.Next()

		log.Debug(
			"<<< done HTTP request",
			zap.Int("status_code", c.Writer.Status()),
			zap.Int("response_size", c.Writer.Size()),
			zap.Duration("latency", time.Since(startedAt)),
		)
	}
}

func Panic() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := core_logger.FromContext(c.Request.Context())

		defer func() {
			if p := recover(); p != nil {
				core_error_tamplate.Panic(
					c,
					log,
					p,
					"during handle HTTP request got unexpected panic",
				)

				c.Abort()
			}
		}()

		c.Next()
	}
}
