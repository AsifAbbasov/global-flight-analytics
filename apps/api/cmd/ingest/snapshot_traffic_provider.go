package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/regionalprovider"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/sharedsnapshot"
)

var (
	errSnapshotTrafficSourceNameRequired = errors.New(
		"snapshot traffic source name is required",
	)

	errSnapshotTrafficResultMissing = errors.New(
		"regional traffic result is missing from shared snapshot",
	)

	errSnapshotTrafficResultType = errors.New(
		"regional traffic result has unexpected type",
	)

	errSnapshotTrafficRequestKeyMissing = errors.New(
		"regional traffic request key is missing from shared snapshot",
	)

	errSnapshotTrafficRequestMismatch = errors.New(
		"snapshot traffic request does not match captured snapshot scope",
	)
)

type snapshotTrafficProvider struct {
	sourceName string
	requestKey string
	states     []flightstate.FlightState
}

func newSnapshotTrafficProvider(
	snapshot sharedsnapshot.Snapshot,
	sourceName string,
) (*snapshotTrafficProvider, error) {
	normalizedSourceName := strings.TrimSpace(
		sourceName,
	)
	if normalizedSourceName == "" {
		return nil, errSnapshotTrafficSourceNameRequired
	}

	for _, success := range snapshot.Successes {
		if success.TaskID != sharedsnapshot.TaskIDRegionalTraffic {
			continue
		}

		trafficPayload, ok := success.Payload.(sharedsnapshot.RegionalTrafficPayload)
		if !ok {
			return nil, fmt.Errorf(
				"%w: task_id=%s payload_type=%T",
				errSnapshotTrafficResultType,
				success.TaskID,
				success.Payload,
			)
		}

		if strings.TrimSpace(success.RequestKey) == "" {
			return nil, fmt.Errorf(
				"%w: task_id=%s",
				errSnapshotTrafficRequestKeyMissing,
				success.TaskID,
			)
		}

		return &snapshotTrafficProvider{
			sourceName: normalizedSourceName,
			requestKey: success.RequestKey,
			states: cloneFlightStates(
				trafficPayload.States,
			),
		}, nil
	}

	for _, failure := range snapshot.Failures {
		if failure.TaskID != sharedsnapshot.TaskIDRegionalTraffic {
			continue
		}

		if failure.Err == nil {
			return nil, fmt.Errorf(
				"%w: task_id=%s",
				errSnapshotTrafficResultMissing,
				failure.TaskID,
			)
		}

		return nil, fmt.Errorf(
			"regional traffic shared snapshot task failed: %w",
			failure.Err,
		)
	}

	return nil, errSnapshotTrafficResultMissing
}

func (
	provider *snapshotTrafficProvider,
) SourceName() string {
	if provider == nil {
		return ""
	}

	return provider.sourceName
}

func (
	provider *snapshotTrafficProvider,
) LoadByPoint(
	ctx context.Context,
	latitude float64,
	longitude float64,
	radius int,
) ([]flightstate.FlightState, error) {
	if provider == nil {
		return nil, errSnapshotTrafficResultMissing
	}

	if ctx != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}

	actualRequestKey := regionalprovider.PointRequestKey(
		latitude,
		longitude,
		radius,
	)

	if actualRequestKey != provider.requestKey {
		return nil, fmt.Errorf(
			"%w: expected=%q actual=%q",
			errSnapshotTrafficRequestMismatch,
			provider.requestKey,
			actualRequestKey,
		)
	}

	return cloneFlightStates(
		provider.states,
	), nil
}

func cloneFlightStates(
	states []flightstate.FlightState,
) []flightstate.FlightState {
	if states == nil {
		return nil
	}

	clonedStates := make(
		[]flightstate.FlightState,
		len(states),
	)

	copy(
		clonedStates,
		states,
	)

	return clonedStates
}
