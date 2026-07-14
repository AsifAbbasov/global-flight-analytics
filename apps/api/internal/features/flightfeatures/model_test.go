package flightfeatures

import (
	"reflect"
	"testing"
	"time"
)

func TestFlightFeaturesCloneDoesNotShareMutableSlices(
	t *testing.T,
) {
	features := FlightFeatures{
		SchemaVersion: SchemaVersionV1,
		TrajectoryID:  "trajectory-one",
		ICAO24:        "ABC123",
		Temporal: TemporalFeatures{
			Evidence: GroupEvidence{
				Limitations: []FeatureLimitation{
					{
						Code:    "temporal-limitation",
						Message: "Temporal evidence is limited.",
					},
				},
			},
		},
		Geographical: GeographicalFeatures{
			Evidence: GroupEvidence{
				Limitations: []FeatureLimitation{
					{
						Code:    "geographical-limitation",
						Message: "Geographical evidence is limited.",
					},
				},
			},
		},
		Operational: OperationalFeatures{
			Evidence: GroupEvidence{
				Limitations: []FeatureLimitation{
					{
						Code:    "operational-limitation",
						Message: "Operational evidence is limited.",
					},
				},
			},
		},
		Trajectory: TrajectoryFeatures{
			Evidence: GroupEvidence{
				Limitations: []FeatureLimitation{
					{
						Code:    "trajectory-limitation",
						Message: "Trajectory evidence is limited.",
					},
				},
			},
		},
		Aircraft: AircraftFeatures{
			Evidence: GroupEvidence{
				Limitations: []FeatureLimitation{
					{
						Code:    "aircraft-limitation",
						Message: "Aircraft evidence is limited.",
					},
				},
			},
		},
		Quality: FeatureQuality{
			Limitations: []FeatureLimitation{
				{
					Code:    "quality-limitation",
					Message: "Feature quality is limited.",
				},
			},
		},
		Provenance: FeatureProvenance{
			SourceNames: []string{"provider-one", "provider-two"},
		},
	}

	cloned := features.Clone()

	cloned.Temporal.Evidence.Limitations[0].Code = "changed"
	cloned.Geographical.Evidence.Limitations[0].Code = "changed"
	cloned.Operational.Evidence.Limitations[0].Code = "changed"
	cloned.Trajectory.Evidence.Limitations[0].Code = "changed"
	cloned.Aircraft.Evidence.Limitations[0].Code = "changed"
	cloned.Quality.Limitations[0].Code = "changed"
	cloned.Provenance.SourceNames[0] = "changed"

	if features.Temporal.Evidence.Limitations[0].Code !=
		"temporal-limitation" {
		t.Fatal("Clone() shared temporal limitations")
	}
	if features.Geographical.Evidence.Limitations[0].Code !=
		"geographical-limitation" {
		t.Fatal("Clone() shared geographical limitations")
	}
	if features.Operational.Evidence.Limitations[0].Code !=
		"operational-limitation" {
		t.Fatal("Clone() shared operational limitations")
	}
	if features.Trajectory.Evidence.Limitations[0].Code !=
		"trajectory-limitation" {
		t.Fatal("Clone() shared trajectory limitations")
	}
	if features.Aircraft.Evidence.Limitations[0].Code !=
		"aircraft-limitation" {
		t.Fatal("Clone() shared aircraft limitations")
	}
	if features.Quality.Limitations[0].Code !=
		"quality-limitation" {
		t.Fatal("Clone() shared quality limitations")
	}
	if features.Provenance.SourceNames[0] != "provider-one" {
		t.Fatal("Clone() shared provenance source names")
	}
}

func TestFlightFeaturesClonePreservesScalarValues(
	t *testing.T,
) {
	start := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	features := FlightFeatures{
		SchemaVersion: SchemaVersionV1,
		TrajectoryID:  "trajectory-one",
		IdentityKey:   "flight-identity-example",
		FlightID:      "flight-one",
		AircraftID:    "aircraft-one",
		ICAO24:        "ABC123",
		Callsign:      "TEST123",
		Window: FeatureWindow{
			StartTime: start,
			EndTime:   start.Add(time.Hour),
			AsOfTime:  start.Add(time.Hour),
		},
		ExtractedAt: start.Add(2 * time.Hour),
		Temporal: TemporalFeatures{
			DurationSeconds: 3600,
		},
		Geographical: GeographicalFeatures{
			GreatCircleDistanceKM: 100,
		},
		Operational: OperationalFeatures{
			MeanVelocityMPS: 200,
		},
		Trajectory: TrajectoryFeatures{
			PointCount: 10,
		},
		Aircraft: AircraftFeatures{
			Model: "Example model",
		},
		Quality: FeatureQuality{
			Status:            ValidationStatusUnvalidated,
			CompletenessScore: 0.8,
		},
		Provenance: FeatureProvenance{
			ExtractorVersion: "extractor-v1",
		},
	}

	cloned := features.Clone()

	if !reflect.DeepEqual(features, cloned) {
		t.Fatalf(
			"Clone() result differs from source\nsource: %#v\nclone: %#v",
			features,
			cloned,
		)
	}
}

func TestContractEnumsRemainStable(t *testing.T) {
	availabilityStatuses := []AvailabilityStatus{
		AvailabilityStatusAvailable,
		AvailabilityStatusPartial,
		AvailabilityStatusUnavailable,
	}
	expectedAvailabilityStatuses := []AvailabilityStatus{
		"available",
		"partial",
		"unavailable",
	}
	if !reflect.DeepEqual(
		availabilityStatuses,
		expectedAvailabilityStatuses,
	) {
		t.Fatalf(
			"availability statuses changed: %#v",
			availabilityStatuses,
		)
	}

	validationStatuses := []ValidationStatus{
		ValidationStatusUnvalidated,
		ValidationStatusValid,
		ValidationStatusLimited,
		ValidationStatusInvalid,
	}
	expectedValidationStatuses := []ValidationStatus{
		"unvalidated",
		"valid",
		"limited",
		"invalid",
	}
	if !reflect.DeepEqual(
		validationStatuses,
		expectedValidationStatuses,
	) {
		t.Fatalf(
			"validation statuses changed: %#v",
			validationStatuses,
		)
	}
}
