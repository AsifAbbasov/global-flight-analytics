package main

import (
	"context"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanout"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/sharedsnapshot"
)

type sharedSnapshotRunConfig struct {
	Executor providerfanout.Executor

	TrafficSource sharedsnapshot.RegionalTrafficSource

	Latitude  float64
	Longitude float64
	Radius    int
}

func runSharedSnapshot(
	ctx context.Context,
	config sharedSnapshotRunConfig,
) (sharedsnapshot.Snapshot, error) {
	trafficTask, err := sharedsnapshot.BuildRegionalTrafficTask(
		sharedsnapshot.RegionalTrafficTaskConfig{
			TrafficSource: config.TrafficSource,
			Provider:      providerpolicy.ProviderAirplanesLive,
			Latitude:      config.Latitude,
			Longitude:     config.Longitude,
			Radius:        config.Radius,
		},
	)
	if err != nil {
		return sharedsnapshot.Snapshot{}, fmt.Errorf(
			"build shared snapshot regional traffic task: %w",
			err,
		)
	}

	runtime, err := sharedsnapshot.NewRuntime(
		sharedsnapshot.RuntimeConfig{
			Executor: config.Executor,
		},
	)
	if err != nil {
		return sharedsnapshot.Snapshot{}, fmt.Errorf(
			"create shared snapshot runtime: %w",
			err,
		)
	}

	snapshot, err := runtime.Run(
		ctx,
		[]providerfanout.Task{
			trafficTask,
		},
	)
	if err != nil {
		return sharedsnapshot.Snapshot{}, fmt.Errorf(
			"run shared snapshot runtime: %w",
			err,
		)
	}

	return snapshot, nil
}
