package extractor

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

type cancelingAircraftProvider struct {
	cancel context.CancelFunc
}

func (provider cancelingAircraftProvider) Provide(
	context.Context,
	AircraftReference,
) (flightfeatures.AircraftFeatures, error) {
	provider.cancel()
	return flightfeatures.AircraftFeatures{
		Evidence: flightfeatures.GroupEvidence{
			Status:          flightfeatures.AvailabilityStatusUnavailable,
			TotalFieldCount: flightfeatures.CurrentGroupFieldCount(flightfeatures.FeatureGroupAircraft),
		},
	}, nil
}

func TestExtractorRejectsNilContext(t *testing.T) {
	extractor := newTestExtractor(t, Config{})

	_, err := extractor.Extract(nil, validRequest())
	if !errors.Is(err, ErrContextRequired) {
		t.Fatalf("Extract() error = %v, want %v", err, ErrContextRequired)
	}
}

func TestNewRejectsTypedNilBuilder(t *testing.T) {
	var typedNil *temporalBuilderStub

	_, err := New(Config{
		TemporalBuilder:     typedNil,
		GeographicalBuilder: &geographicalBuilderStub{},
		OperationalBuilder:  &operationalBuilderStub{},
		TrajectoryBuilder:   &trajectoryBuilderStub{},
	})
	if !errors.Is(err, ErrTemporalBuilderRequired) {
		t.Fatalf("New() error = %v, want %v", err, ErrTemporalBuilderRequired)
	}
}

func TestNewTreatsTypedNilOptionalAircraftProviderAsUnavailable(t *testing.T) {
	var typedNil *aircraftFeatureProviderStub
	extractor := newTestExtractor(t, Config{AircraftFeatureProvider: typedNil})

	features, err := extractor.Extract(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}
	if features.Aircraft.Evidence.Status != flightfeatures.AvailabilityStatusUnavailable {
		t.Fatalf("aircraft status = %q", features.Aircraft.Evidence.Status)
	}
}

func TestExtractorStopsAfterAircraftProviderCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	extractor := newTestExtractor(t, Config{
		AircraftFeatureProvider: cancelingAircraftProvider{cancel: cancel},
	})

	_, err := extractor.Extract(ctx, validRequest())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Extract() error = %v, want context.Canceled", err)
	}
}

func TestExtractorRejectsNestedEvidenceAfterAsOfTime(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Request)
		wantErr error
	}{
		{
			name: "point",
			mutate: func(request *Request) {
				request.Trajectory.Points[0].ObservedAt = request.AsOfTime.Add(time.Nanosecond)
			},
			wantErr: ErrTrajectoryPointAfterAsOf,
		},
		{
			name: "segment start",
			mutate: func(request *Request) {
				request.Trajectory.Segments[0].StartTime = request.AsOfTime.Add(time.Nanosecond)
			},
			wantErr: ErrTrajectorySegmentAfterAsOf,
		},
		{
			name: "segment end",
			mutate: func(request *Request) {
				request.Trajectory.Segments[0].EndTime = request.AsOfTime.Add(time.Nanosecond)
			},
			wantErr: ErrTrajectorySegmentAfterAsOf,
		},
		{
			name: "coverage gap",
			mutate: func(request *Request) {
				request.Trajectory.CoverageGaps = []trajectory.CoverageGap{{
					StartTime: request.AsOfTime,
					EndTime:   request.AsOfTime.Add(time.Nanosecond),
				}}
			},
			wantErr: ErrCoverageGapAfterAsOf,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			extractor := newTestExtractor(t, Config{})
			request := validRequest()
			test.mutate(&request)

			_, err := extractor.Extract(context.Background(), request)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Extract() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestFingerprintNormalizesSemanticAircraftIdentity(t *testing.T) {
	first := validRequest().Trajectory
	second := first
	second.ICAO24 = "  ABC123 "
	second.Callsign = "TEST123"

	firstFingerprint, err := fingerprintTrajectory(first)
	if err != nil {
		t.Fatalf("first fingerprint error = %v", err)
	}
	secondFingerprint, err := fingerprintTrajectory(second)
	if err != nil {
		t.Fatalf("second fingerprint error = %v", err)
	}
	if firstFingerprint != secondFingerprint {
		t.Fatalf("semantic fingerprints differ: %q != %q", firstFingerprint, secondFingerprint)
	}
}

func TestBuildInitialQualityRejectsInvalidEvidenceCounts(t *testing.T) {
	features := flightfeatures.FlightFeatures{
		Temporal: flightfeatures.TemporalFeatures{
			Evidence: flightfeatures.GroupEvidence{
				AvailableFieldCount: 2,
				TotalFieldCount:     1,
			},
		},
		Trajectory: flightfeatures.TrajectoryFeatures{TrajectoryQualityScore: 0.8},
	}

	_, err := buildInitialQuality(features, trajectory.FlightTrajectory{})
	if !errors.Is(err, ErrInvalidEvidenceFieldCount) {
		t.Fatalf("buildInitialQuality() error = %v, want %v", err, ErrInvalidEvidenceFieldCount)
	}
}

func TestBuildInitialQualityRejectsNonFiniteInputQuality(t *testing.T) {
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		features := flightfeatures.FlightFeatures{
			Trajectory: flightfeatures.TrajectoryFeatures{TrajectoryQualityScore: value},
		}

		_, err := buildInitialQuality(features, trajectory.FlightTrajectory{})
		if !errors.Is(err, ErrInvalidInputQualityScore) {
			t.Fatalf("value=%v error=%v, want %v", value, err, ErrInvalidInputQualityScore)
		}
	}
}
