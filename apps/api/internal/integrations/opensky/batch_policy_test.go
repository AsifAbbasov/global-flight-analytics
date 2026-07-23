package opensky

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/providerbatch"
)

type batchPolicyStatesClient struct {
	result StatesResult
}

func (client *batchPolicyStatesClient) GetStates(
	context.Context,
	StatesRequest,
) (StatesResult, error) {
	return client.result, nil
}

func TestProviderRejectsMalformedStateWithoutDiscardingValidState(
	t *testing.T,
) {
	now := time.Date(
		2026,
		time.July,
		23,
		17,
		30,
		0,
		0,
		time.UTC,
	)
	latitude := 40.4093
	longitude := 49.8671

	provider, err := NewProvider(&batchPolicyStatesClient{
		result: StatesResult{
			States: []StateVector{
				{
					ICAO24:       "abc123",
					SnapshotTime: now,
					LastContact:  now,
					TimePosition: &now,
					Latitude:     &latitude,
					Longitude:    &longitude,
				},
				{
					ICAO24: "bad999",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	states, evidence, err := provider.LoadByPointWithBatchEvidence(
		context.Background(),
		40.4093,
		49.8671,
		100,
	)
	if err != nil {
		t.Fatalf("load mixed batch: %v", err)
	}
	if len(states) != 1 {
		t.Fatalf("state count=%d, want 1", len(states))
	}
	if evidence != (providerbatch.Evidence{
		Received:          2,
		Accepted:          1,
		RejectedMalformed: 1,
	}) {
		t.Fatalf("unexpected evidence: %+v", evidence)
	}
}

func TestProviderReturnsTypedErrorWhenEveryStateIsRejected(
	t *testing.T,
) {
	provider, err := NewProvider(&batchPolicyStatesClient{
		result: StatesResult{
			States: []StateVector{
				{ICAO24: "bad999"},
			},
		},
	})
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	_, evidence, err := provider.LoadByPointWithBatchEvidence(
		context.Background(),
		40.4093,
		49.8671,
		100,
	)
	if !errors.Is(err, providerbatch.ErrAllItemsRejected) {
		t.Fatalf("expected all-items-rejected error, got %v", err)
	}
	if evidence.Received != 1 ||
		evidence.RejectedMalformed != 1 {
		t.Fatalf("unexpected evidence: %+v", evidence)
	}
}
