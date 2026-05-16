package core_http_server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	core_logger "github.com/DimaKirejko/Dstributed_cron/internal/core/logger"
	core_http_middleware "github.com/DimaKirejko/Dstributed_cron/internal/core/transport/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HTTPServer struct {
	engine *gin.Engine
	config Config
	log    *core_logger.Logger
}

func NewHTTPServer(
	config Config,
	log *core_logger.Logger,
) *HTTPServer {
	engine := gin.New()

	engine.Use(
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(log),
		core_http_middleware.Panic(),
		core_http_middleware.Trace(),
	)

	return &HTTPServer{
		engine: engine,
		config: config,
		log:    log,
	}
}

func (s *HTTPServer) RegisterRoutes(routes ...Route) {
	for _, route := range routes {
		s.engine.Handle(route.Method, route.Path, route.Handler)
	}

}

func (s *HTTPServer) Run(ctx context.Context) error {
	server := &http.Server{
		Addr:    s.config.Addr,
		Handler: s.engine,
	}

	errCh := make(chan error, 1)

	go func() {
		s.log.Info("starting HTTP server", zap.String("addr", s.config.Addr))

		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("listen and serve HTTP: %w", err)
		}

	case <-ctx.Done():
		s.log.Warn("shutdown HTTP server..")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil { //Shutdown gracefully
			_ = server.Close() //Close immediately

			return fmt.Errorf("Brutal Shutdown HTTP server %w", err)
		}

		s.log.Warn("HTTP server stopped")
	}

	return nil
}
