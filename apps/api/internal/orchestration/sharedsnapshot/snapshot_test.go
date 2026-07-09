package sharedsnapshot

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanin"
)

func TestFromEnvelopeRejectsZeroAssembledAt(
	t *testing.T,
) {
	_, err := FromEnvelope(
		time.Time{},
		providerfanin.Envelope{},
	)
	if err == nil {
		t.Fatal(
			"expected zero assembled time to be rejected",
		)
	}

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

func TestFromEnvelopeRejectsRegionalTrafficValueTypeMismatch(
	t *testing.T,
) {
	_, err := FromEnvelope(
		time.Now(),
		providerfanin.Envelope{
			Status: providerfanin.BatchStatusSucceeded,

			TotalCount:   1,
			SuccessCount: 1,

			Successes: []providerfanin.Success{
				{
					TaskID: TaskIDRegionalTraffic,
					Value:  "unexpected-traffic-value",
				},
			},
		},
	)
	if err == nil {
		t.Fatal(
			"expected regional traffic value type mismatch error",
		)
	}

	if !errors.Is(
		err,
		ErrSuccessValueTypeMismatch,
	) {
		t.Fatalf(
			"expected ErrSuccessValueTypeMismatch, got %v",
			err,
		)
	}
}

func TestFromEnvelopeRejectsUnsupportedSuccessTask(
	t *testing.T,
) {
	_, err := FromEnvelope(
		time.Now(),
		providerfanin.Envelope{
			Status: providerfanin.BatchStatusSucceeded,

			TotalCount:   1,
			SuccessCount: 1,

			Successes: []providerfanin.Success{
				{
					TaskID: "unsupported-task",
					Value:  "unsupported-value",
				},
			},
		},
	)
	if err == nil {
		t.Fatal(
			"expected unsupported success task error",
		)
	}

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

	envelope := providerfanin.Envelope{
		Status: providerfanin.BatchStatusPartial,

		TotalCount:   2,
		SuccessCount: 1,
		FailureCount: 1,

		Successes: []providerfanin.Success{
			{
				TaskID:     TaskIDRegionalTraffic,
				RequestKey: "regional-traffic",
				Value: []flightstate.FlightState{
					{},
				},
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

	if snapshot.TotalCount != envelope.TotalCount {
		t.Fatalf(
			"unexpected total count: got %d, want %d",
			snapshot.TotalCount,
			envelope.TotalCount,
		)
	}

	if snapshot.SuccessCount != envelope.SuccessCount {
		t.Fatalf(
			"unexpected success count: got %d, want %d",
			snapshot.SuccessCount,
			envelope.SuccessCount,
		)
	}

	if snapshot.FailureCount != envelope.FailureCount {
		t.Fatalf(
			"unexpected failure count: got %d, want %d",
			snapshot.FailureCount,
			envelope.FailureCount,
		)
	}

	if len(snapshot.Successes) != 1 {
		t.Fatalf(
			"unexpected success length: got %d, want 1",
			len(snapshot.Successes),
		)
	}

	if len(snapshot.Failures) != 1 {
		t.Fatalf(
			"unexpected failure length: got %d, want 1",
			len(snapshot.Failures),
		)
	}

	if snapshot.Successes[0].TaskID != TaskIDRegionalTraffic {
		t.Fatalf(
			"unexpected success task identifier: %q",
			snapshot.Successes[0].TaskID,
		)
	}

	trafficPayload, ok := snapshot.Successes[0].Payload.(RegionalTrafficPayload)
	if !ok {
		t.Fatalf(
			"unexpected success payload type: %T",
			snapshot.Successes[0].Payload,
		)
	}

	if len(trafficPayload.States) != 1 {
		t.Fatalf(
			"unexpected traffic state count: got %d, want 1",
			len(trafficPayload.States),
		)
	}

	sourceStates, ok := envelope.Successes[0].Value.([]flightstate.FlightState)
	if !ok {
		t.Fatalf(
			"unexpected source traffic value type: %T",
			envelope.Successes[0].Value,
		)
	}

	if len(sourceStates) != 1 {
		t.Fatalf(
			"unexpected source traffic state count: got %d, want 1",
			len(sourceStates),
		)
	}

	if &sourceStates[0] == &trafficPayload.States[0] {
		t.Fatal(
			"expected snapshot traffic payload to use an independent backing array",
		)
	}

	if snapshot.Failures[0].TaskID != "weather" {
		t.Fatalf(
			"unexpected failure task identifier: %q",
			snapshot.Failures[0].TaskID,
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
				Payload: RegionalTrafficPayload{
					States: []flightstate.FlightState{
						{},
					},
				},
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

	originalPayload, ok := original.Successes[0].Payload.(RegionalTrafficPayload)
	if !ok {
		t.Fatalf(
			"unexpected original payload type: %T",
			original.Successes[0].Payload,
		)
	}

	clonedPayload, ok := cloned.Successes[0].Payload.(RegionalTrafficPayload)
	if !ok {
		t.Fatalf(
			"unexpected cloned payload type: %T",
			cloned.Successes[0].Payload,
		)
	}

	if len(originalPayload.States) != 1 {
		t.Fatalf(
			"unexpected original traffic state count: got %d, want 1",
			len(originalPayload.States),
		)
	}

	if len(clonedPayload.States) != 1 {
		t.Fatalf(
			"unexpected cloned traffic state count: got %d, want 1",
			len(clonedPayload.States),
		)
	}

	if &originalPayload.States[0] == &clonedPayload.States[0] {
		t.Fatal(
			"expected cloned traffic payload to use an independent backing array",
		)
	}
}
