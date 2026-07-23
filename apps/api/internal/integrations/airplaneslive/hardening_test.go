package airplaneslive

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestNullableTelemetryDoesNotBecomeObservedZero(
	t *testing.T,
) {
	var response StateResponse
	if err := json.Unmarshal(
		[]byte(`{
			"now": 1720526400000,
			"ac": [
				{
					"hex": "abc123",
					"lat": 40.4093,
					"lon": 49.8671,
					"alt_baro": null,
					"gs": null,
					"track": null,
					"baro_rate": null,
					"seen": null
				},
				{
					"hex": "def456",
					"lat": 40.5,
					"lon": 49.9,
					"alt_baro": "unknown"
				}
			]
		}`),
		&response,
	); err != nil {
		t.Fatalf("decode nullable airplanes.live response: %v", err)
	}

	states := MapStateResponse(&response)
	if len(states) != 2 {
		t.Fatalf("mapped states=%d want=2", len(states))
	}
	for _, state := range states {
		if state.VelocityAvailable ||
			state.HeadingAvailable ||
			state.VerticalRateAvailable ||
			state.OnGroundAvailable {
			t.Fatalf("nullable telemetry became available: %+v", state)
		}
		if !state.TelemetryAvailabilityKnown {
			t.Fatal("provider availability semantics must be explicit")
		}
		if state.ObservedAt.IsZero() {
			t.Fatal("valid response time should remain usable when seen is absent")
		}
	}
}

func TestObservedTelemetryPreservesAvailability(
	t *testing.T,
) {
	responseTime := float64(1720526400000)
	states := MapStateResponse(&StateResponse{
		Now: responseTime,
		Aircraft: []AircraftItem{
			{
				Hex:       "abc123",
				Latitude:  40.4093,
				Longitude: 49.8671,
				AltBaro: BarometricAltitude{
					Feet: 1000,
					Kind: BarometricAltitudeKindObserved,
				},
				GroundSpeed: OptionalFloat64{Value: 200, Available: true},
				Track:       OptionalFloat64{Value: 90, Available: true},
				BaroRate:    OptionalFloat64{Value: 600, Available: true},
				Seen:        OptionalFloat64{Value: 2.5, Available: true},
			},
		},
	})

	state := states[0]
	if !state.VelocityAvailable ||
		!state.HeadingAvailable ||
		!state.VerticalRateAvailable ||
		!state.OnGroundAvailable {
		t.Fatalf("observed telemetry lost availability: %+v", state)
	}
	expected := time.UnixMilli(int64(responseTime)).Add(-2500 * time.Millisecond).UTC()
	if !state.ObservedAt.Equal(expected) {
		t.Fatalf("observed_at=%s want=%s", state.ObservedAt, expected)
	}
}

func TestObservationTimeFailsClosedOnOverflow(
	t *testing.T,
) {
	if actual := observationTime(int64BoundaryFloat64, OptionalFloat64{}); !actual.IsZero() {
		t.Fatalf("overflow response time produced %s", actual)
	}

	base := float64(1720526400000)
	actual := observationTime(
		base,
		OptionalFloat64{
			Value:     int64BoundaryFloat64,
			Available: true,
		},
	)
	expected := time.UnixMilli(int64(base)).UTC()
	if !actual.Equal(expected) {
		t.Fatalf("overflow seen changed base time: got=%s want=%s", actual, expected)
	}
}

func TestProviderRejectsNilClientAndNilReceiverFailsClosed(
	t *testing.T,
) {
	if provider := NewProvider(nil); provider != nil {
		t.Fatalf("NewProvider(nil)=%v want=nil", provider)
	}

	var provider *Provider
	_, err := provider.LoadByPoint(
		context.Background(),
		40.4093,
		49.8671,
		100,
	)
	if !errors.Is(err, ErrClientRequired) {
		t.Fatalf("nil provider error=%v want=%v", err, ErrClientRequired)
	}
}
