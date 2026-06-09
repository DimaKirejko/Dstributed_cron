package core_worker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/DimaKirejko/Dstributed_cron/internal/core/domain"
	core_logger "github.com/DimaKirejko/Dstributed_cron/internal/core/logger"
	core_job_types "github.com/DimaKirejko/Dstributed_cron/internal/core/types"
)

type jobExecutor struct {
	logger         *core_logger.Logger
	httpClient     *http.Client
	dbActionRunner DBActionRunner
	config         ExecutorConfig
}

type DBActionRunner interface {
	CreatePartition(ctx context.Context, targetDB string) error
	YearCleanup(ctx context.Context, targetDB string) error
}

func NewExecutor(
	logger *core_logger.Logger,
	httpClient *http.Client,
	dbActionRunner DBActionRunner,
	config ExecutorConfig,
) *jobExecutor {
	return &jobExecutor{
		logger:         logger,
		httpClient:     httpClient,
		dbActionRunner: dbActionRunner,
		config:         config,
	}
}

func (e *jobExecutor) Execute(ctx context.Context, job domain.Job) ExecutionResult {
	switch job.JobType {
	case core_job_types.TypeHTTP:
		return e.executeHTTP(ctx, job)
	case core_job_types.TypeDB:
		return e.executeDB(ctx, job)
	default:
		return ExecutionResult{
			Success:   false,
			Retryable: false,
			Err:       fmt.Errorf("unsupported job type: %s", job.JobType),
		}
	}
}

func (e *jobExecutor) executeHTTP(ctx context.Context, job domain.Job) ExecutionResult {
	parsedURL, err := url.Parse(*job.HttpURL)
	if err != nil {
		return fatalResult(fmt.Errorf("parse http_url: %w", err))
	}

	if err := CheckHTTPAllowlist(parsedURL, e.config.HTTPAllowlist); err != nil {
		return fatalResult(err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		string(*job.HttpMethod),
		parsedURL.String(),
		nil,
	)
	if err != nil {
		return fatalResult(fmt.Errorf("create http request: %w", err))
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return retryableResult(err)
		}

		return retryableResult(fmt.Errorf("perform http request: %w", err))
	}
	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, resp.Body) // дочитуємо відповідь до кінця щоб повернути TCP con у пул

	return classifyHTTPStatus(resp.StatusCode)
}

func CheckHTTPAllowlist(parsedURL *url.URL, HTTPAllowlist []string) error { // ugly func
	if len(HTTPAllowlist) == 0 {
		return fmt.Errorf("not allowed %s", parsedURL.Host)
	}
	for _, allowed := range HTTPAllowlist {
		if parsedURL.Host == allowed {
			return nil
		}
	}

	return fmt.Errorf("not allowed %s", parsedURL.Host)
}

func classifyHTTPStatus(statusCode int) ExecutionResult {
	switch {
	case statusCode >= 200 && statusCode <= 299:
		return ExecutionResult{
			Success:    true,
			Retryable:  false,
			HTTPStatus: &statusCode,
		}

	case statusCode == http.StatusRequestTimeout ||
		statusCode == http.StatusTooManyRequests ||
		statusCode >= 500:
		return ExecutionResult{
			Success:    false,
			Retryable:  true,
			HTTPStatus: &statusCode,
			Err:        fmt.Errorf("http retryable status: %d", statusCode),
		}

	default:
		return ExecutionResult{
			Success:    false,
			Retryable:  false,
			HTTPStatus: &statusCode,
			Err:        fmt.Errorf("http fatal status: %d", statusCode),
		}
	}
}

func (e *jobExecutor) executeDB(ctx context.Context, job domain.Job) ExecutionResult {

	if job.DbAction == nil {
		e.logger.Error("service data consistency is broken 'DbAction = nil'")
		return fatalResult(fmt.Errorf("db_action is required"))
	}

	if job.TargetDB == nil || strings.TrimSpace(*job.TargetDB) == "" {
		e.logger.Error("service data consistency is broken 'TargetDB = nil'")
		return fatalResult(fmt.Errorf("target_db is required"))
	}

	switch {
	case *job.DbAction == core_job_types.DBActiCreatePartition:
		if err := e.dbActionRunner.CreatePartition(ctx, *job.TargetDB); err != nil {
			return classifyDBStatus(err)
		}
	case *job.DbAction == core_job_types.DBActiYearCleanup:
		if err := e.dbActionRunner.YearCleanup(ctx, *job.TargetDB); err != nil {
			return classifyDBStatus(err)
		}

	default:
		return ExecutionResult{
			Success:   false,
			Retryable: false,
			Err:       fmt.Errorf("fatal status: not possible operation"),
		}
	}

	return ExecutionResult{
		Success:   true,
		Retryable: false,
		Err:       nil,
	}
}

func classifyDBStatus(err error) ExecutionResult {
	switch {
	case errors.Is(err, ErrFatalDBAction):
		return fatalResult(err)
	case errors.Is(err, ErrRetryableDBAction):
		return retryableResult(err)
	default:
		return retryableResult(err)
	}
}

func fatalResult(err error) ExecutionResult {
	return ExecutionResult{
		Success:   false,
		Retryable: false,
		Err:       err,
	}
}

func retryableResult(err error) ExecutionResult {
	return ExecutionResult{
		Success:   false,
		Retryable: true,
		Err:       err,
	}
}
