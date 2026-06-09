package metrics

import (
	"context"
	"net/http"
	"time"

	core_logger "github.com/DimaKirejko/Dstributed_cron/internal/core/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type Metrics struct {
	registry *prometheus.Registry
	FailJob  prometheus.Gauge
	logger   *core_logger.Logger
	repo     MetricsRepo
}

type MetricsRepo interface {
	GetFailedTasksCount(ctx context.Context) (float64, error)
}

func NewMetrics(logger *core_logger.Logger, repo MetricsRepo) *Metrics {
	registry := prometheus.NewRegistry()

	m := &Metrics{
		registry: registry,
		FailJob: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "distributed_cron",
			Subsystem: "scheduler",
			Name:      "failed_job_indicator", // distributed_cron_scheduler_failed_job_indicator
			Help:      "Indicator for failed job.",
		}),
		logger: logger,
		repo:   repo,
	}

	registry.MustRegister(m.FailJob)

	return m
}

func (m *Metrics) StartFailedJobsIndicator(ctx context.Context) {
	ticker := time.NewTicker(7 * time.Second)

	go func() {
		defer ticker.Stop()

		for {
			count, err := m.repo.GetFailedTasksCount(ctx)
			if err != nil {
				m.logger.Error("failed to get failed count:", zap.Error(err))
			}

			m.FailJob.Set(count)

			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", m.Handler())

		if err := http.ListenAndServe(":2112", mux); err != nil {
			m.logger.Error("metrics server stopped", zap.Error(err))
		}
	}()

}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}
