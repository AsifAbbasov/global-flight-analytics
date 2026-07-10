package sharedsnapshot

import (
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	domainweather "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanin"
)

type PayloadKind string

const (
	PayloadKindRegionalTraffic PayloadKind = "regional-traffic"
	PayloadKindCurrentWeather  PayloadKind = "current-weather"
)

var (
	ErrUnsupportedSuccessTask = errors.New(
		"shared snapshot success task is unsupported",
	)

	ErrSuccessPayloadKindMismatch = errors.New(
		"shared snapshot success payload kind mismatch",
	)
)

type RegionalTrafficPayload struct {
	States []flightstate.FlightState
}

type CurrentWeatherPayload struct {
	Snapshot domainweather.CurrentSnapshot
}

type Payload struct {
	kind PayloadKind

	regionalTraffic RegionalTrafficPayload
	currentWeather  CurrentWeatherPayload
}

func (Payload) RequestCoalescingValue() {}

func NewRegionalTrafficPayload(
	states []flightstate.FlightState,
) Payload {
	return Payload{
		kind: PayloadKindRegionalTraffic,
		regionalTraffic: RegionalTrafficPayload{
			States: cloneFlightStates(
				states,
			),
		},
	}
}

func NewCurrentWeatherPayload(
	snapshot domainweather.CurrentSnapshot,
) Payload {
	return Payload{
		kind: PayloadKindCurrentWeather,
		currentWeather: CurrentWeatherPayload{
			Snapshot: snapshot,
		},
	}
}

func (
	payload Payload,
) Kind() PayloadKind {
	return payload.kind
}

func (
	payload Payload,
) RegionalTraffic() (
	RegionalTrafficPayload,
	bool,
) {
	if payload.kind != PayloadKindRegionalTraffic {
		return RegionalTrafficPayload{},
			false
	}

	return RegionalTrafficPayload{
		States: cloneFlightStates(
			payload.regionalTraffic.States,
		),
	}, true
}

func (
	payload Payload,
) CurrentWeather() (
	CurrentWeatherPayload,
	bool,
) {
	if payload.kind != PayloadKindCurrentWeather {
		return CurrentWeatherPayload{},
			false
	}

	return payload.currentWeather,
		true
}

func (
	payload Payload,
) Clone() Payload {
	switch payload.kind {
	case PayloadKindRegionalTraffic:
		return NewRegionalTrafficPayload(
			payload.regionalTraffic.States,
		)

	case PayloadKindCurrentWeather:
		return NewCurrentWeatherPayload(
			payload.currentWeather.Snapshot,
		)

	default:
		return Payload{}
	}
}

func payloadFromProviderSuccess(
	success providerfanin.Success[Payload],
) (Payload, error) {
	expectedKind, err := expectedPayloadKind(
		success.TaskID,
	)
	if err != nil {
		return Payload{},
			err
	}

	if success.Value.Kind() != expectedKind {
		return Payload{},
			fmt.Errorf(
				"%w: task=%q expected=%q actual=%q",
				ErrSuccessPayloadKindMismatch,
				success.TaskID,
				expectedKind,
				success.Value.Kind(),
			)
	}

	return success.Value.Clone(),
		nil
}

func expectedPayloadKind(
	taskID string,
) (PayloadKind, error) {
	switch taskID {
	case TaskIDRegionalTraffic:
		return PayloadKindRegionalTraffic,
			nil

	case TaskIDCurrentWeather:
		return PayloadKindCurrentWeather,
			nil

	default:
		return "",
			fmt.Errorf(
				"%w: %q",
				ErrUnsupportedSuccessTask,
				taskID,
			)
	}
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
