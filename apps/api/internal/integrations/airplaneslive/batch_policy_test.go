package airplaneslive

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/providerbatch"
)

func TestMapStateResponseRejectsMalformedItemsIndividually(
	t *testing.T,
) {
	response := &StateResponse{
		Now: float64(
			time.Date(
				2026,
				time.July,
				23,
				17,
				0,
				0,
				0,
				time.UTC,
			).UnixMilli(),
		),
		Aircraft: []AircraftItem{
			{
				Hex:       "abc123",
				Latitude:  40.4093,
				Longitude: 49.8671,
			},
			{
				Hex:       "",
				Latitude:  40.4093,
				Longitude: 49.8671,
			},
			{
				Hex:       "bad999",
				Latitude:  120,
				Longitude: 49.8671,
			},
		},
	}

	states, evidence, err := MapStateResponseWithEvidence(response)
	if err != nil {
		t.Fatalf("map mixed provider batch: %v", err)
	}
	if len(states) != 1 {
		t.Fatalf("state count=%d, want 1", len(states))
	}
	if evidence != (providerbatch.Evidence{
		Received:          3,
		Accepted:          1,
		RejectedMalformed: 2,
	}) {
		t.Fatalf("unexpected evidence: %+v", evidence)
	}
}

func TestMapStateResponseRejectsCompletelyMalformedBatch(
	t *testing.T,
) {
	response := &StateResponse{
		Now: float64(time.Now().UTC().UnixMilli()),
		Aircraft: []AircraftItem{
			{Latitude: 40.4093, Longitude: 49.8671},
		},
	}

	states, evidence, err := MapStateResponseWithEvidence(response)
	if !errors.Is(err, providerbatch.ErrAllItemsRejected) {
		t.Fatalf("expected all-items-rejected error, got %v", err)
	}
	if len(states) != 0 ||
		evidence.Received != 1 ||
		evidence.RejectedMalformed != 1 {
		t.Fatalf(
			"states=%d evidence=%+v",
			len(states),
			evidence,
		)
	}
}
