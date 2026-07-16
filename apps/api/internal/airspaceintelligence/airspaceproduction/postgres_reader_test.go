package airspaceproduction

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactiongraph"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestSelectAltitudePrefersObservedGeometricAltitude(
	t *testing.T,
) {
	value, reference := selectAltitude(
		pgtype.Float8{Float64: 9000, Valid: true},
		flightstate.AltitudeStatusObserved,
		pgtype.Float8{Float64: 9100, Valid: true},
		flightstate.AltitudeStatusObserved,
	)
	if value == nil || *value != 9100 {
		t.Fatalf("altitude = %v, want 9100", value)
	}
	if reference != interactiongraph.AltitudeReferenceGeometric {
		t.Fatalf("reference = %q", reference)
	}
}

func TestSelectAltitudeFallsBackToObservedBarometricAltitude(
	t *testing.T,
) {
	value, reference := selectAltitude(
		pgtype.Float8{Float64: 9000, Valid: true},
		flightstate.AltitudeStatusObserved,
		pgtype.Float8{},
		flightstate.AltitudeStatusUnavailable,
	)
	if value == nil || *value != 9000 {
		t.Fatalf("altitude = %v, want 9000", value)
	}
	if reference != interactiongraph.AltitudeReferenceBarometric {
		t.Fatalf("reference = %q", reference)
	}
}

func TestValidateObservationQueryRejectsInvalidWindow(
	t *testing.T,
) {
	now := time.Now().UTC()
	err := validateObservationQuery(ObservationQuery{
		Bounds: region.Bounds{
			MinLatitude:  38,
			MaxLatitude:  42,
			MinLongitude: 44.5,
			MaxLongitude: 51,
		},
		WindowStart: now,
		WindowEnd:   now,
		Limit:       1,
	})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("error = %v, want ErrInvalidRequest", err)
	}
}
