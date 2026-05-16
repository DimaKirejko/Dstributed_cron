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
	"go.uber.org/zap"
)

func main() {
	zone, err := time.LoadLocation("UTC")
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

} //// почати реалізовувати фічу додавання джоб
