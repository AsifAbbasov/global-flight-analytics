package sharedsnapshot

import (
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	domainweather "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanin"
)

var (
	ErrUnsupportedSuccessTask = errors.New(
		"shared snapshot success task is unsupported",
	)

	ErrSuccessValueTypeMismatch = errors.New(
		"shared snapshot success value type mismatch",
	)
)

type Payload interface {
	isSharedSnapshotPayload()
	cloneSharedSnapshotPayload() Payload
}

type RegionalTrafficPayload struct {
	States []flightstate.FlightState
}

func (RegionalTrafficPayload) isSharedSnapshotPayload() {}

func (
	payload RegionalTrafficPayload,
) cloneSharedSnapshotPayload() Payload {
	return RegionalTrafficPayload{
		States: cloneFlightStates(
			payload.States,
		),
	}
}

type CurrentWeatherPayload struct {
	Snapshot domainweather.CurrentSnapshot
}

func (CurrentWeatherPayload) isSharedSnapshotPayload() {}

func (
	payload CurrentWeatherPayload,
) cloneSharedSnapshotPayload() Payload {
	return payload
}

func payloadFromProviderSuccess(
	success providerfanin.Success,
) (Payload, error) {
	switch success.TaskID {
	case TaskIDRegionalTraffic:
		states, ok := success.Value.([]flightstate.FlightState)
		if !ok {
			return nil, fmt.Errorf(
				"%w: task %q expects []flightstate.FlightState, got %T",
				ErrSuccessValueTypeMismatch,
				success.TaskID,
				success.Value,
			)
		}

		return RegionalTrafficPayload{
			States: cloneFlightStates(
				states,
			),
		}, nil

	case TaskIDCurrentWeather:
		snapshot, ok := success.Value.(domainweather.CurrentSnapshot)
		if !ok {
			return nil, fmt.Errorf(
				"%w: task %q expects weather.CurrentSnapshot, got %T",
				ErrSuccessValueTypeMismatch,
				success.TaskID,
				success.Value,
			)
		}

		return CurrentWeatherPayload{
			Snapshot: snapshot,
		}, nil

	default:
		return nil, fmt.Errorf(
			"%w: %q",
			ErrUnsupportedSuccessTask,
			success.TaskID,
		)
	}
}

func clonePayload(
	payload Payload,
) Payload {
	if payload == nil {
		return nil
	}

	return payload.cloneSharedSnapshotPayload()
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
