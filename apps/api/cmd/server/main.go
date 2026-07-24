package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/buildinfo"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database"
	applogger "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/logger"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/server"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const serverShutdownTimeout = 10 * time.Second

var (
	errServerContextRequired = errors.New(
		"server lifecycle context is required",
	)
	errServerListenRequired = errors.New(
		"server listen function is required",
	)
	errServerShutdownRequired = errors.New(
		"server shutdown function is required",
	)
	errServerAddressRequired = errors.New(
		"server listen address is required",
	)
	errServerShutdownTimeoutInvalid = errors.New(
		"server shutdown timeout must be greater than zero",
	)
	errServerStoppedUnexpectedly = errors.New(
		"server listener stopped unexpectedly",
	)
	errServerListenerStopTimeout = errors.New(
		"server listener did not stop within the shutdown timeout",
	)
	errServerConfigurationLoad = errors.New(
		"server configuration load failed",
	)
	errServerDatabaseConnection = errors.New(
		"server database connection failed",
	)
	errServerInitialization = errors.New(
		"server initialization failed",
	)
	errServerListen = errors.New(
		"server listen failed",
	)
	errServerShutdown = errors.New(
		"server shutdown failed",
	)
)

type serverLifecycle struct {
	Listen func(
		address string,
	) error
	ShutdownWithTimeout func(
		timeout time.Duration,
	) error
}

func main() {
	_ = godotenv.Load()

	log := applogger.New()
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	if err := run(
		ctx,
		log,
	); err != nil {
		log.Error(
			"api server stopped with error",
			"failure_code",
			serverFailureCode(
				err,
			),
			"error_type",
			fmt.Sprintf(
				"%T",
				err,
			),
		)
		os.Exit(1)
	}
}

func run(
	ctx context.Context,
	log *slog.Logger,
) error {
	if ctx == nil {
		return errServerContextRequired
	}
	if log == nil {
		log = slog.Default()
	}

	cfg, err := config.LoadServerConfig()
	if err != nil {
		return fmt.Errorf(
			"%w: %w",
			errServerConfigurationLoad,
			err,
		)
	}

	trustedProxyConfig, err :=
		config.LoadTrustedProxyConfig()
	if err != nil {
		return fmt.Errorf(
			"%w: %w",
			errServerConfigurationLoad,
			err,
		)
	}

	dbPool, err := openServerDatabase(
		cfg,
		log,
	)
	if err != nil {
		return err
	}
	if dbPool != nil {
		defer func() {
			dbPool.Close()
			log.Info(
				"postgres connection closed",
			)
		}()
	}

	app, err := server.New(
		server.Config{
			DatabasePool:     dbPool,
			Logger:           log,
			OpenMeteoTimeout: cfg.OpenMeteoTimeout,
			Protection: server.ProtectionConfig{
				AllowedOrigins:        cfg.APIProtection.AllowedOrigins,
				BodyLimitBytes:        cfg.APIProtection.BodyLimitBytes,
				ReadTimeout:           cfg.APIProtection.ReadTimeout,
				WriteTimeout:          cfg.APIProtection.WriteTimeout,
				IdleTimeout:           cfg.APIProtection.IdleTimeout,
				RateLimitMax:          cfg.APIProtection.RateLimitMax,
				RateLimitWindow:       cfg.APIProtection.RateLimitWindow,
				ClientIPHeader:        trustedProxyConfig.ClientIPHeader,
				TrustedProxyRanges:    trustedProxyConfig.TrustedProxyRanges,
				MutationKeyDigest:     cfg.APIProtection.MutationKeyDigest,
				MutationKeyConfigured: cfg.APIProtection.MutationKeyConfigured,
			},
		},
	)
	if err != nil {
		return fmt.Errorf(
			"%w: %w",
			errServerInitialization,
			err,
		)
	}

	address := ":" + cfg.Port
	build := buildinfo.Current()
	log.Info(
		"api server starting",
		"address",
		address,
		"version",
		build.Version,
		"revision",
		build.Revision,
		"built_at",
		build.BuiltAt,
		"trusted_proxy_client_identity",
		len(trustedProxyConfig.TrustedProxyRanges) > 0,
	)

	if err := serve(
		ctx,
		serverLifecycle{
			Listen:              app.Listen,
			ShutdownWithTimeout: app.ShutdownWithTimeout,
		},
		address,
		serverShutdownTimeout,
	); err != nil {
		return err
	}

	log.Info(
		"api server stopped",
	)
	return nil
}

func openServerDatabase(
	cfg config.ServerConfig,
	log *slog.Logger,
) (
	*pgxpool.Pool,
	error,
) {
	if cfg.Database == nil {
		log.Warn(
			"database url is not set; starting api without database connection",
		)
		return nil, nil
	}

	dbPool, err := database.NewPostgresPool(
		cfg.Database.URL,
		cfg.Database.ConnectTimeout,
	)
	if err != nil {
		return nil,
			fmt.Errorf(
				"%w: %w",
				errServerDatabaseConnection,
				err,
			)
	}

	log.Info(
		"postgres connection established",
	)
	return dbPool, nil
}

func serve(
	ctx context.Context,
	lifecycle serverLifecycle,
	address string,
	shutdownTimeout time.Duration,
) error {
	if ctx == nil {
		return errServerContextRequired
	}
	if lifecycle.Listen == nil {
		return errServerListenRequired
	}
	if lifecycle.ShutdownWithTimeout == nil {
		return errServerShutdownRequired
	}

	normalizedAddress := strings.TrimSpace(
		address,
	)
	if normalizedAddress == "" {
		return errServerAddressRequired
	}
	if shutdownTimeout <= 0 {
		return errServerShutdownTimeoutInvalid
	}

	listenErrors := make(
		chan error,
		1,
	)
	go func() {
		listenErrors <- lifecycle.Listen(
			normalizedAddress,
		)
	}()

	select {
	case listenErr := <-listenErrors:
		if listenErr == nil {
			return errServerStoppedUnexpectedly
		}
		return fmt.Errorf(
			"%w: %w",
			errServerListen,
			listenErr,
		)

	case <-ctx.Done():
	}

	if err := lifecycle.ShutdownWithTimeout(
		shutdownTimeout,
	); err != nil {
		return fmt.Errorf(
			"%w: %w",
			errServerShutdown,
			err,
		)
	}

	stopTimer := time.NewTimer(
		shutdownTimeout,
	)
	defer stopTimer.Stop()

	select {
	case <-listenErrors:
		return nil

	case <-stopTimer.C:
		return errServerListenerStopTimeout
	}
}

func serverFailureCode(
	err error,
) string {
	switch {
	case errors.Is(
		err,
		errServerConfigurationLoad,
	):
		return "SERVER_CONFIGURATION_LOAD_FAILED"

	case errors.Is(
		err,
		errServerDatabaseConnection,
	):
		return "SERVER_DATABASE_CONNECTION_FAILED"

	case errors.Is(
		err,
		errServerInitialization,
	):
		return "SERVER_INITIALIZATION_FAILED"

	case errors.Is(
		err,
		errServerListen,
	):
		return "SERVER_LISTEN_FAILED"

	case errors.Is(
		err,
		errServerShutdown,
	):
		return "SERVER_SHUTDOWN_FAILED"

	case errors.Is(
		err,
		errServerListenerStopTimeout,
	):
		return "SERVER_LISTENER_STOP_TIMEOUT"

	default:
		return "SERVER_LIFECYCLE_FAILED"
	}
}
