package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	core_logger "github.com/DimaKirejko/Dstributed_cron/internal/core/logger"
	core_postgres_pool "github.com/DimaKirejko/Dstributed_cron/internal/core/repository/postgres_pgx"
	core_http_server "github.com/DimaKirejko/Dstributed_cron/internal/core/transport/server"
	jobs_repository "github.com/DimaKirejko/Dstributed_cron/internal/features/jobs/repository"
	jobs_service "github.com/DimaKirejko/Dstributed_cron/internal/features/jobs/service"
	jobs_transport_http "github.com/DimaKirejko/Dstributed_cron/internal/features/jobs/transport"
	"github.com/DimaKirejko/Dstributed_cron/internal/features/metrics"
	core_scheduler "github.com/DimaKirejko/Dstributed_cron/internal/features/scheduler"
	core_worker "github.com/DimaKirejko/Dstributed_cron/internal/features/worker"
	"go.uber.org/zap"

	_ "github.com/DimaKirejko/Dstributed_cron/docs"
)

// @DISTRIBUTED_CRON (DC)
// @version 1.0
// @description TDO REST-API schema
// @host 127.0.0.1:5050

// BasePath ??
func main() {
	zone, err := time.LoadLocation("UTC") // REWORK
	if err != nil {
		fmt.Errorf("load time zone: %s: %w", "UTC", err)
		panic("time error")
	}

	time.Local = zone

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	defer cancel()

	logger, err := core_logger.NewLoger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Println("failed to init application logger:", err)
		os.Exit(1)
	}

	defer func() {
		logger.Close()
	}()

	logger.Debug("initializing postgres connection pool")
	pool, err := core_postgres_pool.NewPool(ctx, core_postgres_pool.NewConfigMust())

	if err != nil {
		logger.Fatal("Failed to init postgres connection pool", zap.Error(err))
	}
	defer pool.Close()

	logger.Debug("initializing feature", zap.String("feature", "metrics"))
	startMetrics(logger, ctx, pool)

	logger.Debug("initializing feature", zap.String("feature", "scheduler"))
	schedulerRepository := core_scheduler.NewSchedulerRepository(pool, core_scheduler.NewConfigMust())
	scheduler := core_scheduler.NewScheduler(schedulerRepository, logger, 5*time.Second) // need add time in config!
	scheduler.Start(ctx)

	logger.Debug("initializing feature", zap.String("feature", "worker pool"))
	workerRepository := core_worker.NewWorkerRepository(pool)
	httpClient := core_worker.NewHTTPClient(10 * time.Second) // need add time in config!
	dbRunner := core_worker.NewDBActionRepository(pool)
	config := core_worker.NewConfigMust()
	executor := core_worker.NewExecutor(logger, httpClient, dbRunner, config)
	workerPool := core_worker.NewPool(10, workerRepository, executor, logger, 1*time.Second, 60*time.Second) // need add time in config!
	workerPool.Start(ctx)

	logger.Debug("initializing feature", zap.String("feature", "Jobs"))
	jobRepository := jobs_repository.NewJobRepository(pool)
	jobService := jobs_service.NewJobsService(jobRepository)
	jobTransport := jobs_transport_http.NewJobsHttpHandler(jobService)

	logger.Debug("initializing HTTP server")
	httpConfig := core_http_server.NewConfigMust()
	httpServer := core_http_server.NewHTTPServer(
		httpConfig,
		logger,
	)
	httpServer.RegisterRoutes(jobTransport.Route()...)
	if err := httpServer.Run(ctx); err != nil {
		logger.Error("HTTP server RUN error", zap.Error(err))
	}

}

func startMetrics(
	logger *core_logger.Logger,
	ctx context.Context,
	pool *core_postgres_pool.Pool,
) {
	repo := metrics.NewRepository(pool)
	m := metrics.NewMetrics(logger, repo)
	m.StartFailedJobsIndicator(ctx)
}
