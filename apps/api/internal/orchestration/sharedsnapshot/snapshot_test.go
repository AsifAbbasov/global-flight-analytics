package sharedsnapshot

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	domainweather "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanin"
)

func TestFromEnvelopeRejectsZeroAssembledAt(
	t *testing.T,
) {
	_, err := FromEnvelope(
		time.Time{},
		providerfanin.Envelope[Payload]{},
	)

	if !errors.Is(
		err,
		ErrAssembledAtRequired,
	) {
		t.Fatalf(
			"expected ErrAssembledAtRequired, got %v",
			err,
		)
	}
}

func TestFromEnvelopeRejectsPayloadKindMismatch(
	t *testing.T,
) {
	_, err := FromEnvelope(
		time.Now(),
		providerfanin.Envelope[Payload]{
			Status: providerfanin.BatchStatusSucceeded,

			TotalCount:   1,
			SuccessCount: 1,

			Successes: []providerfanin.Success[Payload]{
				{
					TaskID: TaskIDRegionalTraffic,
					Value: NewCurrentWeatherPayload(
						domainweather.CurrentSnapshot{},
					),
				},
			},
		},
	)

	if !errors.Is(
		err,
		ErrSuccessPayloadKindMismatch,
	) {
		t.Fatalf(
			"expected ErrSuccessPayloadKindMismatch, got %v",
			err,
		)
	}
}

func TestFromEnvelopeRejectsUnsupportedSuccessTask(
	t *testing.T,
) {
	_, err := FromEnvelope(
		time.Now(),
		providerfanin.Envelope[Payload]{
			Status: providerfanin.BatchStatusSucceeded,

			TotalCount:   1,
			SuccessCount: 1,

			Successes: []providerfanin.Success[Payload]{
				{
					TaskID: "unsupported-task",
					Value: NewRegionalTrafficPayload(
						nil,
					),
				},
			},
		},
	)

	if !errors.Is(
		err,
		ErrUnsupportedSuccessTask,
	) {
		t.Fatalf(
			"expected ErrUnsupportedSuccessTask, got %v",
			err,
		)
	}
}

func TestFromEnvelopeCopiesEnvelopeAndNormalizesAssembledAtToUTC(
	t *testing.T,
) {
	location := time.FixedZone(
		"test-zone",
		int((4*time.Hour)/time.Second),
	)

	assembledAt := time.Date(
		2026,
		time.July,
		5,
		15,
		30,
		0,
		0,
		location,
	)

	providerFailure := errors.New(
		"provider request failed",
	)

	sourceStates := []flightstate.FlightState{
		{},
	}

	envelope := providerfanin.Envelope[Payload]{
		Status: providerfanin.BatchStatusPartial,

		TotalCount:   2,
		SuccessCount: 1,
		FailureCount: 1,

		Successes: []providerfanin.Success[Payload]{
			{
				TaskID:     TaskIDRegionalTraffic,
				RequestKey: "regional-traffic",
				Value: NewRegionalTrafficPayload(
					sourceStates,
				),
				Shared: true,
			},
		},

		Failures: []providerfanin.Failure{
			{
				TaskID:     "weather",
				RequestKey: "current-weather",
				Err:        providerFailure,
			},
		},
	}

	snapshot, err := FromEnvelope(
		assembledAt,
		envelope,
	)
	if err != nil {
		t.Fatalf(
			"assemble shared snapshot: %v",
			err,
		)
	}

	if !snapshot.AssembledAt.Equal(
		assembledAt.UTC(),
	) {
		t.Fatalf(
			"unexpected assembled time: got %s, want %s",
			snapshot.AssembledAt,
			assembledAt.UTC(),
		)
	}

	if snapshot.AssembledAt.Location() != time.UTC {
		t.Fatalf(
			"expected UTC assembled time, got %s",
			snapshot.AssembledAt.Location(),
		)
	}

	if snapshot.Status != envelope.Status {
		t.Fatalf(
			"unexpected status: got %q, want %q",
			snapshot.Status,
			envelope.Status,
		)
	}

	if len(snapshot.Successes) != 1 {
		t.Fatalf(
			"unexpected success length: got %d, want 1",
			len(snapshot.Successes),
		)
	}

	trafficPayload, ok := snapshot.Successes[0].Payload.RegionalTraffic()
	if !ok {
		t.Fatalf(
			"expected regional traffic payload, got kind %q",
			snapshot.Successes[0].Payload.Kind(),
		)
	}

	if len(trafficPayload.States) != 1 {
		t.Fatalf(
			"unexpected traffic state count: got %d, want 1",
			len(trafficPayload.States),
		)
	}

	sourcePayload, ok := envelope.Successes[0].Value.RegionalTraffic()
	if !ok {
		t.Fatal(
			"expected source regional traffic payload",
		)
	}

	if &sourcePayload.States[0] == &trafficPayload.States[0] {
		t.Fatal(
			"expected snapshot traffic payload to use an independent backing array",
		)
	}

	envelope.Successes[0].TaskID = "mutated-success"
	envelope.Failures[0].TaskID = "mutated-failure"

	if snapshot.Successes[0].TaskID != TaskIDRegionalTraffic {
		t.Fatal(
			"expected snapshot success slice to be independent from envelope",
		)
	}

	if snapshot.Failures[0].TaskID != "weather" {
		t.Fatal(
			"expected snapshot failure slice to be independent from envelope",
		)
	}

	if !errors.Is(
		snapshot.Failures[0].Err,
		providerFailure,
	) {
		t.Fatalf(
			"unexpected provider failure: %v",
			snapshot.Failures[0].Err,
		)
	}
}

func TestSnapshotCloneReturnsIndependentSlicesAndPayloads(
	t *testing.T,
) {
	original := Snapshot{
		AssembledAt: time.Date(
			2026,
			time.July,
			5,
			12,
			0,
			0,
			0,
			time.UTC,
		),

		Status: providerfanin.BatchStatusSucceeded,

		TotalCount:   1,
		SuccessCount: 1,

		Successes: []Success{
			{
				TaskID:     TaskIDRegionalTraffic,
				RequestKey: "regional-traffic",
				Payload: NewRegionalTrafficPayload(
					[]flightstate.FlightState{
						{},
					},
				),
				Shared: true,
			},
		},
	}

	cloned := original.Clone()

	cloned.Successes[0].TaskID = "mutated"

	if original.Successes[0].TaskID != TaskIDRegionalTraffic {
		t.Fatal(
			"expected original snapshot success slice to remain unchanged",
		)
	}

	originalPayload, ok := original.Successes[0].Payload.RegionalTraffic()
	if !ok {
		t.Fatal(
			"expected original regional traffic payload",
		)
	}

	clonedPayload, ok := cloned.Successes[0].Payload.RegionalTraffic()
	if !ok {
		t.Fatal(
			"expected cloned regional traffic payload",
		)
	}

	if &originalPayload.States[0] == &clonedPayload.States[0] {
		t.Fatal(
			"expected cloned traffic payload to use an independent backing array",
		)
	}
}
