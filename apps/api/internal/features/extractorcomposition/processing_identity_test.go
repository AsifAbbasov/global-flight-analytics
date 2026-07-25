package extractorcomposition

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/aircraft"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/aircraftprovider"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/geographicalbuilder"
)

func TestProcessingIdentityUsesEffectiveDefaults(t *testing.T) {
	composition, err := New(Config{
		AircraftLookup: &fakeAircraftLookup{
			result: aircraft.Aircraft{ICAO24: "ABC123"},
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	identity := composition.ProcessingIdentity
	if identity.GeographicCellPrecision !=
		geographicalbuilder.DefaultGeographicCellPrecision {
		t.Fatalf(
			"precision = %d, want %d",
			identity.GeographicCellPrecision,
			geographicalbuilder.DefaultGeographicCellPrecision,
		)
	}
	if identity.AircraftPositiveCacheTTL !=
		aircraftprovider.DefaultPositiveCacheTTL ||
		identity.AircraftNegativeCacheTTL !=
			aircraftprovider.DefaultNegativeCacheTTL {
		t.Fatalf("cache identity = %#v", identity)
	}
	if identity.AircraftNotFoundPolicyVersion !=
		aircraftprovider.DefaultNotFoundPolicyVersion {
		t.Fatalf(
			"not-found policy version = %q",
			identity.AircraftNotFoundPolicyVersion,
		)
	}
	if !strings.HasPrefix(
		composition.FingerprintIdentity,
		fingerprintIdentityPrefix,
	) || len(composition.FingerprintIdentity) !=
		len(fingerprintIdentityPrefix)+64 {
		t.Fatalf(
			"fingerprint identity = %q",
			composition.FingerprintIdentity,
		)
	}
}

func TestGeographicPrecisionChangesInputFingerprint(t *testing.T) {
	first := extractFingerprint(t, Config{
		AircraftLookup: &fakeAircraftLookup{
			result: aircraft.Aircraft{ICAO24: "ABC123"},
		},
		GeographicCellPrecision: 2,
	})
	second := extractFingerprint(t, Config{
		AircraftLookup: &fakeAircraftLookup{
			result: aircraft.Aircraft{ICAO24: "ABC123"},
		},
		GeographicCellPrecision: 3,
	})

	if first == second {
		t.Fatal(
			"different geographic precision produced one input fingerprint",
		)
	}
}

func TestAircraftMetadataChangesInputFingerprint(t *testing.T) {
	first := extractFingerprint(t, Config{
		AircraftLookup: &fakeAircraftLookup{
			result: aircraft.Aircraft{
				ICAO24: "ABC123",
				Model:  "A320",
			},
		},
	})
	second := extractFingerprint(t, Config{
		AircraftLookup: &fakeAircraftLookup{
			result: aircraft.Aircraft{
				ICAO24: "ABC123",
				Model:  "A321",
			},
		},
	})

	if first == second {
		t.Fatal(
			"different aircraft metadata produced one input fingerprint",
		)
	}
}

func TestCustomAircraftNotFoundPolicyRequiresVersion(t *testing.T) {
	_, err := New(Config{
		AircraftLookup: &fakeAircraftLookup{},
		IsAircraftNotFound: func(error) bool {
			return true
		},
	})
	if !errors.Is(
		err,
		ErrAircraftNotFoundPolicyVersionRequired,
	) {
		t.Fatalf(
			"New() error = %v, want %v",
			err,
			ErrAircraftNotFoundPolicyVersionRequired,
		)
	}
}

func TestNewRejectsTypedNilAircraftLookup(t *testing.T) {
	var lookup *fakeAircraftLookup

	_, err := New(Config{AircraftLookup: lookup})
	if !errors.Is(err, ErrAircraftLookupRequired) {
		t.Fatalf(
			"New() error = %v, want %v",
			err,
			ErrAircraftLookupRequired,
		)
	}
}

func extractFingerprint(
	t *testing.T,
	config Config,
) string {
	t.Helper()

	fixedNow := time.Date(
		2026,
		time.July,
		25,
		12,
		0,
		0,
		0,
		time.UTC,
	)
	config.Now = func() time.Time {
		return fixedNow
	}

	composition, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	item := completeTrajectory()
	features, err := composition.Extractor.Extract(
		context.Background(),
		extractor.Request{
			Trajectory: item,
			AsOfTime:   item.EndTime.Add(time.Minute),
		},
	)
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	return features.Provenance.InputFingerprint
}
