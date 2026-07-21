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

	content, err := os.ReadFile(filepath.Join(filepath.Dir(currentFile), "airport_read_queries.go"))
	if err != nil {
		t.Fatalf("read airport query owner source: %v", err)
	}
	text := string(content)
	if strings.Contains(text, "COALESCE(a.elevation_ft, 0)") {
		t.Fatal("airport queries still collapse NULL elevation to zero")
	}
	if strings.Count(text, "a.elevation_ft,") != 1 {
		t.Fatal("canonical Airport select columns must own nullable elevation exactly once")
	}
	if strings.Count(text, "SELECT ` + airportSelectColumns") != 3 {
		t.Fatal("all Airport read queries must share the canonical select columns")
	}
}
