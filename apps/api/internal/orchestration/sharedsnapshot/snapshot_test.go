package sharedsnapshot

import (
	"errors"
	"testing"
	"time"

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

	if !errors.Is(err, ErrAssembledAtRequired) {
		t.Fatalf(
			"expected ErrAssembledAtRequired, got %v",
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
				TaskID:     "traffic",
				RequestKey: "regional-traffic",
				Value:      "traffic-value",
				Shared:     true,
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

	if snapshot.Successes[0].TaskID != "traffic" {
		t.Fatalf(
			"unexpected success task identifier: %q",
			snapshot.Successes[0].TaskID,
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

	if snapshot.Successes[0].TaskID != "traffic" {
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

func TestSnapshotCloneReturnsIndependentSlices(
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

		Successes: []providerfanin.Success{
			{
				TaskID:     "traffic",
				RequestKey: "regional-traffic",
				Value:      "traffic-value",
				Shared:     true,
			},
		},
	}

	cloned := original.Clone()

	cloned.Successes[0].TaskID = "mutated"

	if original.Successes[0].TaskID != "traffic" {
		t.Fatal(
			"expected original snapshot success slice to remain unchanged",
		)
	}
}
