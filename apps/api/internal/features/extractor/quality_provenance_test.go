package extractor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestBuildInitialQualitySeparatesOptionalCoverage(t *testing.T) {
	features := flightfeatures.FlightFeatures{
		Temporal:     flightfeatures.TemporalFeatures{Evidence: qualityProvenanceEvidence(8)},
		Geographical: flightfeatures.GeographicalFeatures{Evidence: qualityProvenanceEvidence(11)},
		Operational:  flightfeatures.OperationalFeatures{Evidence: qualityProvenanceEvidence(11)},
		Trajectory: flightfeatures.TrajectoryFeatures{
			Evidence:               qualityProvenanceEvidence(16),
			TrajectoryQualityScore: 0.9,
		},
		Aircraft: flightfeatures.AircraftFeatures{
			Evidence: flightfeatures.GroupEvidence{
				Status:          flightfeatures.AvailabilityStatusUnavailable,
				TotalFieldCount: 6,
			},
		},
	}

	quality, err := buildInitialQuality(features, validRequest().Trajectory)
	if err != nil {
		t.Fatalf("buildInitialQuality() error = %v", err)
	}
	if quality.CompletenessScore != 1 {
		t.Fatalf("required completeness = %v, want 1", quality.CompletenessScore)
	}
	if quality.OptionalCoverageScore != 0 {
		t.Fatalf("optional coverage = %v, want 0", quality.OptionalCoverageScore)
	}
}

func TestExtractorDoesNotInventTrajectoryUpdateTimestamp(t *testing.T) {
	extractor := newTestExtractor(t, Config{})
	request := validRequest()
	request.Trajectory.CreatedAt = time.Time{}
	request.Trajectory.UpdatedAt = time.Time{}

	features, err := extractor.Extract(context.Background(), request)
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}
	if !features.Provenance.TrajectoryCreatedAt.IsZero() ||
		!features.Provenance.TrajectoryUpdatedAt.IsZero() {
		t.Fatalf("invented trajectory timestamps: %#v", features.Provenance)
	}
	if features.Provenance.TrajectoryUpdatedAt.Equal(request.Trajectory.EndTime) {
		t.Fatal("trajectory end time was reused as update provenance")
	}
}

func TestExtractorRecordsExplicitAircraftMetadataProvenance(t *testing.T) {
	provider := &aircraftFeatureProviderStub{
		features: flightfeatures.AircraftFeatures{
			Evidence:     qualityProvenanceEvidence(6),
			Registration: "4K-AZ01",
		},
	}
	extractor := newTestExtractor(t, Config{
		AircraftFeatureProvider:         provider,
		AircraftMetadataSourceName:      "aircraft-source",
		AircraftMetadataProviderVersion: "aircraft-provider-v9",
	})

	features, err := extractor.Extract(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}
	if features.Provenance.AircraftMetadataSourceName != "aircraft-source" ||
		features.Provenance.AircraftMetadataProviderVersion != "aircraft-provider-v9" ||
		features.Provenance.AircraftMetadataRetrievedAt.IsZero() {
		t.Fatalf("aircraft provenance = %#v", features.Provenance)
	}
}

func TestFingerprintIncludesAircraftMetadataIdentity(t *testing.T) {
	item := validRequest().Trajectory
	first, err := fingerprintExtractionInput(
		item,
		flightfeatures.AircraftFeatures{},
		Version,
		"source-one",
		"provider-v1",
	)
	if err != nil {
		t.Fatalf("first fingerprint error = %v", err)
	}
	second, err := fingerprintExtractionInput(
		item,
		flightfeatures.AircraftFeatures{},
		Version,
		"source-two",
		"provider-v1",
	)
	if err != nil {
		t.Fatalf("second fingerprint error = %v", err)
	}
	if first == second {
		t.Fatal("aircraft metadata source identity did not change fingerprint")
	}
}

func TestNewRequiresAircraftMetadataIdentityForConfiguredProvider(t *testing.T) {
	_, err := New(Config{
		TemporalBuilder:         &temporalBuilderStub{},
		GeographicalBuilder:     &geographicalBuilderStub{},
		OperationalBuilder:      &operationalBuilderStub{},
		TrajectoryBuilder:       &trajectoryBuilderStub{},
		AircraftFeatureProvider: &aircraftFeatureProviderStub{},
	})
	if !errors.Is(err, ErrAircraftMetadataSourceNameRequired) {
		t.Fatalf("New() error = %v, want %v", err, ErrAircraftMetadataSourceNameRequired)
	}
}

func qualityProvenanceEvidence(total int) flightfeatures.GroupEvidence {
	return flightfeatures.GroupEvidence{
		Status:              flightfeatures.AvailabilityStatusAvailable,
		AvailableFieldCount: total,
		TotalFieldCount:     total,
	}
}
