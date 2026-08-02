package observability

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

const metricsServerShutdownTimeout = 5 * time.Second

type MetricsServerConfig struct {
	Address       string
	Registry      *Registry
	Authorization AuthorizationConfig
	Logger        *slog.Logger
}

type MetricsServer struct {
	server   *http.Server
	listener net.Listener
}

func StartMetricsServer(
	ctx context.Context,
	config MetricsServerConfig,
) (*MetricsServer, error) {
	address := strings.TrimSpace(config.Address)
	if address == "" {
		return nil, nil
	}
	if ctx == nil {
		return nil, fmt.Errorf("metrics server context is required")
	}
	if config.Registry == nil {
		return nil, fmt.Errorf("metrics server registry is required")
	}
	if !config.Authorization.Configured {
		return nil, fmt.Errorf(
			"metrics server authorization must be configured when an address is set",
		)
	}
	if config.Logger == nil {
		config.Logger = slog.Default()
	}

	handler, err := StandaloneHandler(
		config.Registry,
		config.Authorization,
	)
	if err != nil {
		return nil, err
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("listen for metrics: %w", err)
	}

	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	result := &MetricsServer{
		server:   server,
		listener: listener,
	}

	go func() {
		serveErr := server.Serve(listener)
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			config.Logger.Error(
				"metrics server stopped unexpectedly",
				"error_type",
				fmt.Sprintf("%T", serveErr),
			)
		}
	}()

	go func() {
		<-ctx.Done()
		shutdownContext, cancel := context.WithTimeout(
			context.Background(),
			metricsServerShutdownTimeout,
		)
		defer cancel()
		if shutdownErr := server.Shutdown(shutdownContext); shutdownErr != nil {
			config.Logger.Error(
				"metrics server shutdown failed",
				"error_type",
				fmt.Sprintf("%T", shutdownErr),
			)
		}
	}()

	return result, nil
}

func (
	server *MetricsServer,
) Address() string {
	if server == nil || server.listener == nil {
		return ""
	}
	return server.listener.Addr().String()
}

func (
	server *MetricsServer,
) Close(
	ctx context.Context,
) error {
	if server == nil || server.server == nil {
		return nil
	}
	if ctx == nil {
		return fmt.Errorf("metrics server shutdown context is required")
	}
	return server.server.Shutdown(ctx)
}
