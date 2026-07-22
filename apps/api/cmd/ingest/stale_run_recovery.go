package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	errStaleRunRecoveryRepositoryRequired = errors.New(
		"stale ingestion run recovery repository is required",
	)
	errStaleRunRecoveryContextRequired = errors.New(
		"stale ingestion run recovery context is required",
	)
	errStaleRunRecoveryTimeRequired = errors.New(
		"stale ingestion run recovery time is required",
	)
	errStaleRunRecoveryThresholdInvalid = errors.New(
		"stale ingestion run recovery threshold must be greater than zero",
	)
	errStaleRunRecoveryTimeoutInvalid = errors.New(
		"stale ingestion run recovery timeout must be greater than zero",
	)
)

const staleIngestionRunRecoveryMessage = "ingestion process stopped before terminal status was recorded"

type staleRunRecoveryRepository interface {
	RecoverStaleRunning(
		ctx context.Context,
		staleBefore time.Time,
		recoveredAt time.Time,
		errorMessage string,
	) (int64, error)
}

func recoverStaleIngestionRuns(
	ctx context.Context,
	repository staleRunRecoveryRepository,
	now time.Time,
	staleAfter time.Duration,
	timeout time.Duration,
) (int64, error) {
	if ctx == nil {
		return 0, errStaleRunRecoveryContextRequired
	}
	if repository == nil {
		return 0, errStaleRunRecoveryRepositoryRequired
	}
	if now.IsZero() {
		return 0, errStaleRunRecoveryTimeRequired
	}
	if staleAfter <= 0 {
		return 0, errStaleRunRecoveryThresholdInvalid
	}
	if timeout <= 0 {
		return 0, errStaleRunRecoveryTimeoutInvalid
	}

	recoveredAt := now.UTC()
	staleBefore := recoveredAt.Add(-staleAfter)
	recoveryContext, cancel := context.WithTimeout(
		context.WithoutCancel(ctx),
		timeout,
	)
	defer cancel()

	recoveredCount, err := repository.RecoverStaleRunning(
		recoveryContext,
		staleBefore,
		recoveredAt,
		staleIngestionRunRecoveryMessage,
	)
	if err != nil {
		return 0, fmt.Errorf(
			"recover stale ingestion runs: %w",
			err,
		)
	}

	return recoveredCount, nil
}
