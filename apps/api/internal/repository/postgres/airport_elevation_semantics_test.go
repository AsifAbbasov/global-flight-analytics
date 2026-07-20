package postgres

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestApplyAirportElevationDatabaseValuePreservesNullAndObservedZero(t *testing.T) {
	var unknown airport.Airport
	applyAirportElevationDatabaseValue(&unknown, pgtype.Int4{})
	if unknown.ElevationM != 0 || unknown.ElevationAvailable {
		t.Fatalf("unknown elevation became observed: %#v", unknown)
	}

	var seaLevel airport.Airport
	applyAirportElevationDatabaseValue(&seaLevel, pgtype.Int4{Int32: 0, Valid: true})
	if seaLevel.ElevationM != 0 || !seaLevel.ElevationAvailable {
		t.Fatalf("observed sea-level elevation was lost: %#v", seaLevel)
	}
}

func TestAirportRepositoryDoesNotCoalesceUnknownElevationToZero(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file path")
	}

	content, err := os.ReadFile(filepath.Join(filepath.Dir(currentFile), "airport_repository.go"))
	if err != nil {
		t.Fatalf("read airport repository source: %v", err)
	}
	text := string(content)
	if strings.Contains(text, "COALESCE(a.elevation_ft, 0)") {
		t.Fatal("airport repository still collapses NULL elevation to zero")
	}
	if strings.Count(text, "a.elevation_ft,") != 2 {
		t.Fatal("both airport queries must select nullable elevation directly")
	}
}
