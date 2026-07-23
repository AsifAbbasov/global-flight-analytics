package postgres

import (
	"context"
	"strings"
	"testing"
)

func TestInsertFlightStateQueryUsesReplayIdentityConflict(
	t *testing.T,
) {
	const expected = `ON CONFLICT (source_name, icao24, observed_at)
	DO NOTHING`
	if !strings.Contains(insertFlightStateQuery, expected) {
		t.Fatalf(
			"flight state insert query is not replay-safe:\n%s",
			insertFlightStateQuery,
		)
	}
}

func TestSaveFlightStatesCountedAllowsEmptyBatchWithoutPool(
	t *testing.T,
) {
	repository := NewFlightStateRepository(nil)

	insertedCount, err := repository.SaveFlightStatesCounted(
		context.Background(),
		nil,
	)
	if err != nil {
		t.Fatalf("save empty counted batch: %v", err)
	}
	if insertedCount != 0 {
		t.Fatalf("inserted count = %d, want 0", insertedCount)
	}
}
