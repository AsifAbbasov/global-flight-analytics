package postgres

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestAltitudeDatabaseValueRejectsNonFiniteObservedAltitude(t *testing.T) {
	t.Parallel()

	_, _, err := altitudeDatabaseValue(
		math.NaN(),
		flightstate.AltitudeStatusObserved,
	)
	if !errors.Is(err, ErrAltitudeMetersNotFinite) {
		t.Fatalf("expected ErrAltitudeMetersNotFinite, got %v", err)
	}
}

func TestAltitudeDatabaseValueRejectsUnsupportedStatus(t *testing.T) {
	t.Parallel()

	_, _, err := altitudeDatabaseValue(
		1000,
		flightstate.AltitudeStatus("unsupported"),
	)
	if err == nil {
		t.Fatal("expected unsupported altitude status error")
	}
	if !strings.Contains(err.Error(), "unsupported altitude status") {
		t.Fatalf("expected unsupported altitude status error, got %v", err)
	}
}

func TestAltitudeDatabaseValuePreservesExplicitInvalidStatus(t *testing.T) {
	t.Parallel()

	value, status, err := altitudeDatabaseValue(
		math.NaN(),
		flightstate.AltitudeStatusInvalid,
	)
	if err != nil {
		t.Fatalf("prepare explicit invalid altitude: %v", err)
	}
	if value.Valid || status != string(flightstate.AltitudeStatusInvalid) {
		t.Fatalf("unexpected invalid altitude value=%#v status=%q", value, status)
	}
}
