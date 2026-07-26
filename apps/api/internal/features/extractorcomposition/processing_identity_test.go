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
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/geographicalbuilder"
)

func TestDefaultConfigUsesExplicitDefaults(t *testing.T) {
	composition, err := New(
		DefaultConfig(
			&fakeAircraftLookup{
				result: aircraft.Aircraft{ICAO24: "ABC123"},
			},
		),
	)
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

func TestNewRejectsZeroExplicitConfigurationValues(
	t *testing.T,
) {
	lookup := &fakeAircraftLookup{}

	tests := []struct {
		name    string
		config  Config
		wantErr error
	}{
		{
			name: "geographic precision",
			config: DefaultConfig(
				lookup,
			).WithGeographicCellPrecision(0),
			wantErr: ErrGeographicCellPrecisionRequired,
		},
		{
			name: "positive cache duration",
			config: DefaultConfig(
				lookup,
			).WithAircraftCacheDurations(
				0,
				aircraftprovider.DefaultNegativeCacheTTL,
			),
			wantErr: ErrAircraftPositiveCacheDurationRequired,
		},
		{
			name: "negative cache duration",
			config: DefaultConfig(
				lookup,
			).WithAircraftCacheDurations(
				aircraftprovider.DefaultPositiveCacheTTL,
				0,
			),
			wantErr: ErrAircraftNegativeCacheDurationRequired,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(test.config)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf(
					"New() error = %v, want %v",
					err,
					test.wantErr,
				)
			}
		})
	}
}

func TestConfigurationMethodsDoNotMutateSource(
	t *testing.T,
) {
	base := DefaultConfig(&fakeAircraftLookup{})
	custom := base.
		WithGeographicCellPrecision(4).
		WithAircraftCacheDurations(
			time.Minute,
			30*time.Second,
		).
		WithClock(func() time.Time {
			return time.Unix(1, 0)
		})

	baseComposition, err := New(base)
	if err != nil {
		t.Fatalf("New(base) error = %v", err)
	}
	customComposition, err := New(custom)
	if err != nil {
		t.Fatalf("New(custom) error = %v", err)
	}

	if baseComposition.ProcessingIdentity.GeographicCellPrecision !=
		geographicalbuilder.DefaultGeographicCellPrecision ||
		baseComposition.ProcessingIdentity.AircraftPositiveCacheTTL !=
			aircraftprovider.DefaultPositiveCacheTTL {
		t.Fatalf(
			"base configuration was mutated: %#v",
			baseComposition.ProcessingIdentity,
		)
	}
	if customComposition.ProcessingIdentity.GeographicCellPrecision != 4 ||
		customComposition.ProcessingIdentity.AircraftPositiveCacheTTL !=
			time.Minute ||
		customComposition.ProcessingIdentity.AircraftNegativeCacheTTL !=
			30*time.Second {
		t.Fatalf(
			"custom configuration was not applied: %#v",
			customComposition.ProcessingIdentity,
		)
	}
}

func TestGeographicPrecisionChangesInputFingerprint(t *testing.T) {
	first := extractFingerprint(
		t,
		DefaultConfig(
			&fakeAircraftLookup{
				result: aircraft.Aircraft{ICAO24: "ABC123"},
			},
		).WithGeographicCellPrecision(2),
	)
	second := extractFingerprint(
		t,
		DefaultConfig(
			&fakeAircraftLookup{
				result: aircraft.Aircraft{ICAO24: "ABC123"},
			},
		).WithGeographicCellPrecision(3),
	)

	if first == second {
		t.Fatal(
			"different geographic precision produced one input fingerprint",
		)
	}
}

func TestAircraftMetadataChangesInputFingerprint(t *testing.T) {
	first := extractFingerprint(
		t,
		DefaultConfig(
			&fakeAircraftLookup{
				result: aircraft.Aircraft{
					ICAO24: "ABC123",
					Model:  "A320",
				},
			},
		),
	)
	second := extractFingerprint(
		t,
		DefaultConfig(
			&fakeAircraftLookup{
				result: aircraft.Aircraft{
					ICAO24: "ABC123",
					Model:  "A321",
				},
			},
		),
	)

	if first == second {
		t.Fatal(
			"different aircraft metadata produced one input fingerprint",
		)
	}
}

func TestCustomAircraftNotFoundPolicyRequiresVersion(t *testing.T) {
	_, err := New(
		DefaultConfig(
			&fakeAircraftLookup{},
		).WithAircraftNotFoundPolicy(
			"",
			func(error) bool {
				return true
			},
		),
	)
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

	_, err := New(DefaultConfig(lookup))
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
	config = config.WithClock(
		func() time.Time {
			return fixedNow
		},
	)

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

func TestProcessingManifestIsPersistedInFeatureProvenance(t *testing.T) {
	composition, err := New(DefaultConfig(&fakeAircraftLookup{
		result: aircraft.Aircraft{ICAO24: "ABC123", Model: "A320"},
	}))
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
	if features.Provenance.ProcessingIdentityFingerprint !=
		composition.FingerprintIdentity {
		t.Fatalf(
			"processing fingerprint = %q, want %q",
			features.Provenance.ProcessingIdentityFingerprint,
			composition.FingerprintIdentity,
		)
	}
	if features.Provenance.ProcessingIdentity !=
		composition.ProcessingIdentity {
		t.Fatalf(
			"persisted identity = %#v, want %#v",
			features.Provenance.ProcessingIdentity,
			composition.ProcessingIdentity,
		)
	}
}

func TestCompositionSupportsExplicitlyDisabledAircraftEnrichment(t *testing.T) {
	composition, err := New(DefaultConfigWithoutAircraftEnrichment())
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
	if features.Aircraft.Evidence.Status !=
		flightfeatures.AvailabilityStatusUnavailable {
		t.Fatalf("aircraft evidence = %#v", features.Aircraft.Evidence)
	}
	if features.Provenance.AircraftMetadataSourceName != "" ||
		features.Provenance.AircraftMetadataProviderVersion != "" ||
		!features.Provenance.AircraftMetadataRetrievedAt.IsZero() {
		t.Fatalf("unexpected disabled enrichment provenance: %#v", features.Provenance)
	}
	if features.Provenance.ProcessingIdentity.AircraftEnrichmentMode !=
		flightfeatures.AircraftEnrichmentModeDisabled {
		t.Fatalf("processing identity = %#v", features.Provenance.ProcessingIdentity)
	}
}

func TestAircraftCacheModeChangesInputFingerprint(t *testing.T) {
	lookup := &fakeAircraftLookup{
		result: aircraft.Aircraft{ICAO24: "ABC123", Model: "A320"},
	}
	cached := extractFingerprint(t, DefaultConfig(lookup))
	uncached := extractFingerprint(
		t,
		DefaultConfig(&fakeAircraftLookup{
			result: aircraft.Aircraft{ICAO24: "ABC123", Model: "A320"},
		}).WithoutAircraftCache(),
	)
	if cached == uncached {
		t.Fatal("enabled and disabled cache modes produced one input fingerprint")
	}
}
