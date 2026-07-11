package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	reconciliationWorkerMaxAttemptsEnvironmentVariable       = "RECONCILIATION_WORKER_MAX_ATTEMPTS"
	reconciliationWorkerRetryBaseDelayEnvironmentVariable    = "RECONCILIATION_WORKER_RETRY_BASE_DELAY"
	reconciliationWorkerRetryMaximumDelayEnvironmentVariable = "RECONCILIATION_WORKER_RETRY_MAXIMUM_DELAY"
	reconciliationWorkerStaleAfterEnvironmentVariable        = "RECONCILIATION_WORKER_STALE_AFTER"
	reconciliationWorkerMaximumTasksEnvironmentVariable      = "RECONCILIATION_WORKER_MAXIMUM_TASKS"
)

const (
	defaultReconciliationWorkerMaxAttempts       = 5
	defaultReconciliationWorkerRetryBaseDelay    = 30 * time.Second
	defaultReconciliationWorkerRetryMaximumDelay = 15 * time.Minute
	defaultReconciliationWorkerStaleAfter        = 10 * time.Minute
	defaultReconciliationWorkerMaximumTasks      = 100
)

type ReconciliationWorkerConfig struct {
	Database                    PostgresConfig
	TrajectoryMaxTimeGap        time.Duration
	TrajectoryMaxGroundSpeedMPS float64
	MaxAttempts                 int
	RetryBaseDelay              time.Duration
	RetryMaximumDelay           time.Duration
	StaleAfter                  time.Duration
	MaximumTasks                int
}

func LoadReconciliationWorkerConfig() (
	ReconciliationWorkerConfig,
	error,
) {
	databaseURL, err := requiredTrimmedStringEnvironmentVariable(
		databaseURLEnvironmentVariable,
	)
	if err != nil {
		return ReconciliationWorkerConfig{}, fmt.Errorf(
			"load database url: %w",
			err,
		)
	}

	databaseConnectTimeout, err := requiredPositiveDurationEnvironmentVariable(
		databaseConnectTimeoutEnvironmentVariable,
	)
	if err != nil {
		return ReconciliationWorkerConfig{}, fmt.Errorf(
			"load database connect timeout: %w",
			err,
		)
	}

	trajectoryMaxTimeGap, err := requiredNonNegativeDurationEnvironmentVariable(
		trajectoryMaxTimeGapEnvironmentVariable,
	)
	if err != nil {
		return ReconciliationWorkerConfig{}, fmt.Errorf(
			"load trajectory maximum time gap: %w",
			err,
		)
	}

	trajectoryMaxGroundSpeedMPS, err := requiredNonNegativeFiniteFloat64EnvironmentVariable(
		trajectoryMaxGroundSpeedMPSEnvironmentVariable,
	)
	if err != nil {
		return ReconciliationWorkerConfig{}, fmt.Errorf(
			"load trajectory maximum ground speed: %w",
			err,
		)
	}

	maxAttempts, err := reconciliationWorkerOptionalPositiveInteger(
		reconciliationWorkerMaxAttemptsEnvironmentVariable,
		defaultReconciliationWorkerMaxAttempts,
	)
	if err != nil {
		return ReconciliationWorkerConfig{}, err
	}

	retryBaseDelay, err := reconciliationWorkerOptionalPositiveDuration(
		reconciliationWorkerRetryBaseDelayEnvironmentVariable,
		defaultReconciliationWorkerRetryBaseDelay,
	)
	if err != nil {
		return ReconciliationWorkerConfig{}, err
	}

	retryMaximumDelay, err := reconciliationWorkerOptionalPositiveDuration(
		reconciliationWorkerRetryMaximumDelayEnvironmentVariable,
		defaultReconciliationWorkerRetryMaximumDelay,
	)
	if err != nil {
		return ReconciliationWorkerConfig{}, err
	}

	if retryMaximumDelay < retryBaseDelay {
		return ReconciliationWorkerConfig{}, fmt.Errorf(
			"%s must not be less than %s",
			reconciliationWorkerRetryMaximumDelayEnvironmentVariable,
			reconciliationWorkerRetryBaseDelayEnvironmentVariable,
		)
	}

	staleAfter, err := reconciliationWorkerOptionalPositiveDuration(
		reconciliationWorkerStaleAfterEnvironmentVariable,
		defaultReconciliationWorkerStaleAfter,
	)
	if err != nil {
		return ReconciliationWorkerConfig{}, err
	}

	maximumTasks, err := reconciliationWorkerOptionalPositiveInteger(
		reconciliationWorkerMaximumTasksEnvironmentVariable,
		defaultReconciliationWorkerMaximumTasks,
	)
	if err != nil {
		return ReconciliationWorkerConfig{}, err
	}

	return ReconciliationWorkerConfig{
		Database: PostgresConfig{
			URL:            databaseURL,
			ConnectTimeout: databaseConnectTimeout,
		},
		TrajectoryMaxTimeGap:        trajectoryMaxTimeGap,
		TrajectoryMaxGroundSpeedMPS: trajectoryMaxGroundSpeedMPS,
		MaxAttempts:                 maxAttempts,
		RetryBaseDelay:              retryBaseDelay,
		RetryMaximumDelay:           retryMaximumDelay,
		StaleAfter:                  staleAfter,
		MaximumTasks:                maximumTasks,
	}, nil
}

func reconciliationWorkerOptionalPositiveInteger(
	name string,
	defaultValue int,
) (int, error) {
	rawValue := strings.TrimSpace(
		os.Getenv(
			name,
		),
	)
	if rawValue == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(
		rawValue,
	)
	if err != nil {
		return 0, fmt.Errorf(
			"%s must be an integer: %w",
			name,
			err,
		)
	}

	if value <= 0 {
		return 0, fmt.Errorf(
			"%s must be greater than zero",
			name,
		)
	}

	return value, nil
}

func reconciliationWorkerOptionalPositiveDuration(
	name string,
	defaultValue time.Duration,
) (time.Duration, error) {
	rawValue := strings.TrimSpace(
		os.Getenv(
			name,
		),
	)
	if rawValue == "" {
		return defaultValue, nil
	}

	value, err := time.ParseDuration(
		rawValue,
	)
	if err != nil {
		return 0, fmt.Errorf(
			"%s must be a duration: %w",
			name,
			err,
		)
	}

	if value <= 0 {
		return 0, fmt.Errorf(
			"%s must be greater than zero",
			name,
		)
	}

	return value, nil
}
