package postgres

import (
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
)

func TestBuildAirportPageUsesLastReturnedDuplicateNameRow(t *testing.T) {
	t.Parallel()

	records := []airportRecord{
		{ID: "11111111-1111-1111-1111-111111111111", Item: airport.Airport{Name: "Alpha", ICAOCode: "AAAA"}},
		{ID: "22222222-2222-2222-2222-222222222222", Item: airport.Airport{Name: "Alpha", ICAOCode: "AAAB"}},
		{ID: "33333333-3333-3333-3333-333333333333", Item: airport.Airport{Name: "Bravo", ICAOCode: "BBBB"}},
	}

	page := buildAirportPage(records, 2)
	if len(page.Items) != 2 {
		t.Fatalf("item count = %d, want 2", len(page.Items))
	}
	if page.NextCursor == nil {
		t.Fatal("expected next cursor")
	}
	if page.NextCursor.Name != "Alpha" ||
		page.NextCursor.ID != "22222222-2222-2222-2222-222222222222" {
		t.Fatalf("next cursor = %#v", page.NextCursor)
	}
}

func TestBuildAirportPageOmitsCursorWithoutLookaheadRow(t *testing.T) {
	t.Parallel()

	page := buildAirportPage([]airportRecord{
		{ID: "11111111-1111-1111-1111-111111111111", Item: airport.Airport{Name: "Alpha"}},
	}, 1)
	if page.NextCursor != nil {
		t.Fatalf("unexpected next cursor: %#v", page.NextCursor)
	}
}

func TestParseAirportCursorIDRejectsInvalidUUID(t *testing.T) {
	t.Parallel()

	_, err := parseAirportCursorID("not-a-uuid")
	if !errors.Is(err, airport.ErrListCursorInvalid) {
		t.Fatalf("expected cursor error, got %v", err)
	}
}
