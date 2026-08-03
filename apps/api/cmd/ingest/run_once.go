package main

import (
	"context"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/ingestdaemon"
)

func runSingleIngestionCycle(
	ctx context.Context,
	runCycle ingestdaemon.CycleRunner,
	now ingestdaemon.Clock,
	observe ingestdaemon.Observer,
) error {
	if ctx == nil {
		return ingestdaemon.ErrContextRequired
	}
	if runCycle == nil {
		return ingestdaemon.ErrCycleRunnerRequired
	}
	if now == nil {
		now = func() time.Time {
			return time.Now().UTC()
		}
	}

	startedAt := now().UTC()
	cycleErr := runCycle(ctx)
	finishedAt := now().UTC()

	if observe != nil {
		observe(
			ingestdaemon.CycleResult{
				Number:     1,
				StartedAt:  startedAt,
				FinishedAt: finishedAt,
				Err:        cycleErr,
				NextDelay:  0,
			},
		)
	}

	if cycleErr != nil {
		return fmt.Errorf(
			"run single traffic ingestion cycle: %w",
			cycleErr,
		)
	}

	return nil
}
