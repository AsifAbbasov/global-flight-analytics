package extractor

import (
	"context"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

type temporalBuilderStub struct {
	features flightfeatures.TemporalFeatures
	err      error
	calls    int
	mutate   bool
}

func (stub *temporalBuilderStub) Build(
	ctx context.Context,
	item trajectory.FlightTrajectory,
) (flightfeatures.TemporalFeatures, error) {
	stub.calls++
	if stub.mutate && len(item.Points) > 0 {
		item.Points[0].Latitude = 90
	}

	return stub.features, stub.err
}

type geographicalBuilderStub struct {
	features           flightfeatures.GeographicalFeatures
	err                error
	calls              int
	firstPointLatitude float64
}

func (stub *geographicalBuilderStub) Build(
	ctx context.Context,
	item trajectory.FlightTrajectory,
) (flightfeatures.GeographicalFeatures, error) {
	stub.calls++
	if len(item.Points) > 0 {
		stub.firstPointLatitude = item.Points[0].Latitude
	}

	return stub.features, stub.err
}

type operationalBuilderStub struct {
	features flightfeatures.OperationalFeatures
	err      error
	calls    int
}

func (stub *operationalBuilderStub) Build(
	ctx context.Context,
	item trajectory.FlightTrajectory,
) (flightfeatures.OperationalFeatures, error) {
	stub.calls++

	return stub.features, stub.err
}

type trajectoryBuilderStub struct {
	features flightfeatures.TrajectoryFeatures
	err      error
	calls    int
}

func (stub *trajectoryBuilderStub) Build(
	ctx context.Context,
	item trajectory.FlightTrajectory,
) (flightfeatures.TrajectoryFeatures, error) {
	stub.calls++

	return stub.features, stub.err
}

type aircraftFeatureProviderStub struct {
	features  flightfeatures.AircraftFeatures
	err       error
	calls     int
	reference AircraftReference
}

func (stub *aircraftFeatureProviderStub) Provide(
	ctx context.Context,
	reference AircraftReference,
) (flightfeatures.AircraftFeatures, error) {
	stub.calls++
	stub.reference = reference

	return stub.features, stub.err
}

func TestNewRejectsMissingRequiredBuilders(t *testing.T) {
	validConfig := Config{
		TemporalBuilder:     &temporalBuilderStub{},
		GeographicalBuilder: &geographicalBuilderStub{},
		OperationalBuilder:  &operationalBuilderStub{},
		TrajectoryBuilder:   &trajectoryBuilderStub{},
	}

	tests := []struct {
		name      string
		configure func(Config) Config
		wantErr   error
	}{
		{
			name: "temporal builder",
			configure: func(config Config) Config {
				config.TemporalBuilder = nil
				return config
			},
			wantErr: ErrTemporalBuilderRequired,
		},
		{
			name: "geographical builder",
			configure: func(config Config) Config {
				config.GeographicalBuilder = nil
				return config
			},
			wantErr: ErrGeographicalBuilderRequired,
		},
		{
			name: "operational builder",
			configure: func(config Config) Config {
				config.OperationalBuilder = nil
				return config
			},
			wantErr: ErrOperationalBuilderRequired,
		},
		{
			name: "trajectory builder",
			configure: func(config Config) Config {
				config.TrajectoryBuilder = nil
				return config
			},
			wantErr: ErrTrajectoryBuilderRequired,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(test.configure(validConfig))
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

func TestExtractorAssemblesFeaturesAndProvenance(t *testing.T) {
	extractedAt := time.Date(
		2026,
		time.July,
		14,
		10,
		45,
		0,
		0,
		time.UTC,
	)
	duplicateLimitation := flightfeatures.FeatureLimitation{
		Code:    "shared-limitation",
		Message: "Shared limitation.",
	}

	temporalBuilder := &temporalBuilderStub{
		features: flightfeatures.TemporalFeatures{
			Evidence: flightfeatures.GroupEvidence{
				Status:               flightfeatures.AvailabilityStatusAvailable,
				AvailableFieldCount:  8,
				TotalFieldCount:      8,
				SupportingPointCount: 4,
				Limitations: []flightfeatures.FeatureLimitation{
					duplicateLimitation,
				},
			},
			DurationSeconds: 180,
		},
	}
	geographicalBuilder := &geographicalBuilderStub{
		features: flightfeatures.GeographicalFeatures{
			Evidence: flightfeatures.GroupEvidence{
				Status:               flightfeatures.AvailabilityStatusPartial,
				AvailableFieldCount:  9,
				TotalFieldCount:      11,
				SupportingPointCount: 4,
				Limitations: []flightfeatures.FeatureLimitation{
					duplicateLimitation,
					{
						Code:    "geographical-partial",
						Message: "Geographical evidence is partial.",
					},
				},
			},
			GreatCircleDistanceKM: 120,
		},
	}
	operationalBuilder := &operationalBuilderStub{
		features: flightfeatures.OperationalFeatures{
			Evidence: flightfeatures.GroupEvidence{
				Status:               flightfeatures.AvailabilityStatusAvailable,
				AvailableFieldCount:  11,
				TotalFieldCount:      11,
				SupportingPointCount: 4,
			},
			MeanVelocityMPS: 210,
		},
	}
	trajectoryBuilder := &trajectoryBuilderStub{
		features: flightfeatures.TrajectoryFeatures{
			Evidence: flightfeatures.GroupEvidence{
				Status:               flightfeatures.AvailabilityStatusAvailable,
				AvailableFieldCount:  16,
				TotalFieldCount:      16,
				SupportingPointCount: 4,
			},
			PointCount:             4,
			TrajectoryQualityScore: 0.75,
		},
	}

	extractor, err := New(Config{
		TemporalBuilder:     temporalBuilder,
		GeographicalBuilder: geographicalBuilder,
		OperationalBuilder:  operationalBuilder,
		TrajectoryBuilder:   trajectoryBuilder,
		Now: func() time.Time {
			return extractedAt
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	request := validRequest()
	result, err := extractor.Extract(
		context.Background(),
		request,
	)
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if result.SchemaVersion != flightfeatures.SchemaVersionV1 ||
		result.TrajectoryID != request.Trajectory.ID ||
		result.IdentityKey != request.Trajectory.IdentityKey ||
		result.ICAO24 != "ABC123" ||
		result.Callsign != "TEST123" {
		t.Fatalf("unexpected result identity: %#v", result)
	}
	if !result.Window.StartTime.Equal(
		request.Trajectory.StartTime.UTC(),
	) || !result.Window.EndTime.Equal(
		request.Trajectory.EndTime.UTC(),
	) || !result.Window.AsOfTime.Equal(
		request.AsOfTime.UTC(),
	) {
		t.Fatalf("unexpected feature window: %#v", result.Window)
	}
	if !result.ExtractedAt.Equal(extractedAt) {
		t.Fatalf(
			"ExtractedAt = %v, want %v",
			result.ExtractedAt,
			extractedAt,
		)
	}
	if result.Temporal.DurationSeconds != 180 ||
		result.Geographical.GreatCircleDistanceKM != 120 ||
		result.Operational.MeanVelocityMPS != 210 ||
		result.Trajectory.PointCount != 4 {
		t.Fatalf("unexpected group features: %#v", result)
	}
	if result.Aircraft.Evidence.Status !=
		flightfeatures.AvailabilityStatusUnavailable {
		t.Fatalf(
			"aircraft evidence status = %q",
			result.Aircraft.Evidence.Status,
		)
	}
	if result.Aircraft.Evidence.TotalFieldCount !=
		aircraftFeatureFieldCount {
		t.Fatalf(
			"aircraft total fields = %d",
			result.Aircraft.Evidence.TotalFieldCount,
		)
	}
	if result.Quality.Status !=
		flightfeatures.ValidationStatusUnvalidated {
		t.Fatalf(
			"quality status = %q",
			result.Quality.Status,
		)
	}

	wantCompleteness := float64(8+9+11+16) /
		float64(8+11+11+16+aircraftFeatureFieldCount)
	if math.Abs(
		result.Quality.CompletenessScore-wantCompleteness,
	) > 1e-12 {
		t.Fatalf(
			"completeness = %v, want %v",
			result.Quality.CompletenessScore,
			wantCompleteness,
		)
	}
	if result.Quality.InputQualityScore != 0.75 ||
		result.Quality.SupportingPointCount != 4 {
		t.Fatalf("unexpected quality: %#v", result.Quality)
	}
	if len(result.Quality.Limitations) != 3 {
		t.Fatalf(
			"limitations = %#v, want three unique limitations",
			result.Quality.Limitations,
		)
	}
	if result.Provenance.ExtractorVersion != Version ||
		!strings.HasPrefix(
			result.Provenance.InputFingerprint,
			fingerprintPrefix,
		) ||
		len(result.Provenance.InputFingerprint) !=
			len(fingerprintPrefix)+64 {
		t.Fatalf(
			"unexpected provenance: %#v",
			result.Provenance,
		)
	}
	wantSources := []string{
		"airplanes.live",
		"open-sky",
		"reconciled",
	}
	if !reflect.DeepEqual(
		result.Provenance.SourceNames,
		wantSources,
	) {
		t.Fatalf(
			"source names = %#v, want %#v",
			result.Provenance.SourceNames,
			wantSources,
		)
	}
	if temporalBuilder.calls != 1 ||
		geographicalBuilder.calls != 1 ||
		operationalBuilder.calls != 1 ||
		trajectoryBuilder.calls != 1 {
		t.Fatalf(
			"unexpected builder calls: temporal=%d geographical=%d operational=%d trajectory=%d",
			temporalBuilder.calls,
			geographicalBuilder.calls,
			operationalBuilder.calls,
			trajectoryBuilder.calls,
		)
	}
}

func TestExtractorUsesAircraftFeatureProvider(t *testing.T) {
	provider := &aircraftFeatureProviderStub{
		features: flightfeatures.AircraftFeatures{
			Evidence: flightfeatures.GroupEvidence{
				Status:              flightfeatures.AvailabilityStatusAvailable,
				AvailableFieldCount: aircraftFeatureFieldCount,
				TotalFieldCount:     aircraftFeatureFieldCount,
			},
			Registration: "4K-AZ01",
			Model:        "Example",
		},
	}
	extractor := newTestExtractor(t, Config{
		AircraftFeatureProvider: provider,
	})

	result, err := extractor.Extract(
		context.Background(),
		validRequest(),
	)
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if result.Aircraft.Registration != "4K-AZ01" ||
		result.Aircraft.Model != "Example" {
		t.Fatalf(
			"unexpected aircraft features: %#v",
			result.Aircraft,
		)
	}
	wantReference := AircraftReference{
		AircraftID: "aircraft-one",
		ICAO24:     "ABC123",
		Callsign:   "TEST123",
	}
	if provider.calls != 1 ||
		provider.reference != wantReference {
		t.Fatalf(
			"provider calls=%d reference=%#v, want %#v",
			provider.calls,
			provider.reference,
			wantReference,
		)
	}
}

func TestExtractorRejectsInvalidRequestsBeforeBuilders(
	t *testing.T,
) {
	tests := []struct {
		name    string
		mutate  func(*Request)
		wantErr error
	}{
		{
			name: "trajectory id",
			mutate: func(request *Request) {
				request.Trajectory.ID = ""
			},
			wantErr: ErrTrajectoryIDRequired,
		},
		{
			name: "identity key",
			mutate: func(request *Request) {
				request.Trajectory.IdentityKey = ""
			},
			wantErr: ErrIdentityKeyRequired,
		},
		{
			name: "icao24",
			mutate: func(request *Request) {
				request.Trajectory.ICAO24 = "bad"
			},
			wantErr: ErrInvalidICAO24,
		},
		{
			name: "start time",
			mutate: func(request *Request) {
				request.Trajectory.StartTime = time.Time{}
			},
			wantErr: ErrTrajectoryStartTimeRequired,
		},
		{
			name: "end time",
			mutate: func(request *Request) {
				request.Trajectory.EndTime = time.Time{}
			},
			wantErr: ErrTrajectoryEndTimeRequired,
		},
		{
			name: "invalid window",
			mutate: func(request *Request) {
				request.Trajectory.EndTime =
					request.Trajectory.StartTime.Add(-time.Second)
			},
			wantErr: ErrInvalidTrajectoryWindow,
		},
		{
			name: "as-of time",
			mutate: func(request *Request) {
				request.AsOfTime = time.Time{}
			},
			wantErr: ErrAsOfTimeRequired,
		},
		{
			name: "as-of before trajectory end",
			mutate: func(request *Request) {
				request.AsOfTime =
					request.Trajectory.EndTime.Add(-time.Second)
			},
			wantErr: ErrAsOfBeforeTrajectoryEnd,
		},
		{
			name: "trajectory evidence",
			mutate: func(request *Request) {
				request.Trajectory.Points = nil
				request.Trajectory.Segments = nil
			},
			wantErr: ErrTrajectoryEvidenceRequired,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			temporalBuilder := &temporalBuilderStub{}
			extractor := newTestExtractor(t, Config{
				TemporalBuilder: temporalBuilder,
			})
			request := validRequest()
			test.mutate(&request)

			_, err := extractor.Extract(
				context.Background(),
				request,
			)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf(
					"Extract() error = %v, want %v",
					err,
					test.wantErr,
				)
			}
			if temporalBuilder.calls != 0 {
				t.Fatalf(
					"builder called %d times for invalid request",
					temporalBuilder.calls,
				)
			}
		})
	}
}

func TestExtractorWrapsBuilderErrorWithGroup(t *testing.T) {
	buildErr := errors.New("geographical builder failure")
	extractor := newTestExtractor(t, Config{
		GeographicalBuilder: &geographicalBuilderStub{
			err: buildErr,
		},
	})

	_, err := extractor.Extract(
		context.Background(),
		validRequest(),
	)
	if !errors.Is(err, buildErr) {
		t.Fatalf("Extract() error = %v, want wrapped build error", err)
	}

	var groupErr *GroupBuildError
	if !errors.As(err, &groupErr) {
		t.Fatalf(
			"Extract() error = %T, want *GroupBuildError",
			err,
		)
	}
	if groupErr.Group != flightfeatures.FeatureGroupGeographical {
		t.Fatalf(
			"group = %q, want %q",
			groupErr.Group,
			flightfeatures.FeatureGroupGeographical,
		)
	}
}

func TestExtractorPreservesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	extractor := newTestExtractor(t, Config{})

	_, err := extractor.Extract(ctx, validRequest())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Extract() error = %v, want context.Canceled",
			err,
		)
	}
}

func TestExtractorGivesEachBuilderIndependentTrajectoryCopy(
	t *testing.T,
) {
	temporalBuilder := &temporalBuilderStub{
		mutate: true,
	}
	geographicalBuilder := &geographicalBuilderStub{}
	extractor := newTestExtractor(t, Config{
		TemporalBuilder:     temporalBuilder,
		GeographicalBuilder: geographicalBuilder,
	})
	request := validRequest()
	originalLatitude := request.Trajectory.Points[0].Latitude

	_, err := extractor.Extract(
		context.Background(),
		request,
	)
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if geographicalBuilder.firstPointLatitude != originalLatitude {
		t.Fatalf(
			"geographical builder received latitude %v, want %v",
			geographicalBuilder.firstPointLatitude,
			originalLatitude,
		)
	}
	if request.Trajectory.Points[0].Latitude != originalLatitude {
		t.Fatalf(
			"input trajectory was mutated: latitude=%v want=%v",
			request.Trajectory.Points[0].Latitude,
			originalLatitude,
		)
	}
}

func TestExtractorResultDoesNotShareBuilderLimitationSlices(
	t *testing.T,
) {
	limitations := []flightfeatures.FeatureLimitation{
		{
			Code:    "original",
			Message: "Original limitation.",
		},
	}
	temporalBuilder := &temporalBuilderStub{
		features: flightfeatures.TemporalFeatures{
			Evidence: flightfeatures.GroupEvidence{
				Limitations: limitations,
			},
		},
	}
	extractor := newTestExtractor(t, Config{
		TemporalBuilder: temporalBuilder,
	})

	result, err := extractor.Extract(
		context.Background(),
		validRequest(),
	)
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	limitations[0].Code = "changed"

	if result.Temporal.Evidence.Limitations[0].Code != "original" {
		t.Fatal(
			"extractor result shares builder limitation slice",
		)
	}
}

func newTestExtractor(
	t *testing.T,
	overrides Config,
) *Extractor {
	t.Helper()

	config := Config{
		TemporalBuilder: &temporalBuilderStub{
			features: flightfeatures.TemporalFeatures{
				Evidence: flightfeatures.GroupEvidence{
					Status:              flightfeatures.AvailabilityStatusAvailable,
					TotalFieldCount:     8,
					AvailableFieldCount: 8,
				},
			},
		},
		GeographicalBuilder: &geographicalBuilderStub{
			features: flightfeatures.GeographicalFeatures{
				Evidence: flightfeatures.GroupEvidence{
					Status:              flightfeatures.AvailabilityStatusAvailable,
					TotalFieldCount:     11,
					AvailableFieldCount: 11,
				},
			},
		},
		OperationalBuilder: &operationalBuilderStub{
			features: flightfeatures.OperationalFeatures{
				Evidence: flightfeatures.GroupEvidence{
					Status:              flightfeatures.AvailabilityStatusAvailable,
					TotalFieldCount:     11,
					AvailableFieldCount: 11,
				},
			},
		},
		TrajectoryBuilder: &trajectoryBuilderStub{
			features: flightfeatures.TrajectoryFeatures{
				Evidence: flightfeatures.GroupEvidence{
					Status:              flightfeatures.AvailabilityStatusAvailable,
					TotalFieldCount:     16,
					AvailableFieldCount: 16,
				},
				TrajectoryQualityScore: 0.8,
			},
		},
		Now: func() time.Time {
			return time.Date(
				2026,
				time.July,
				14,
				10,
				0,
				0,
				0,
				time.UTC,
			)
		},
	}

	if overrides.TemporalBuilder != nil {
		config.TemporalBuilder = overrides.TemporalBuilder
	}
	if overrides.GeographicalBuilder != nil {
		config.GeographicalBuilder = overrides.GeographicalBuilder
	}
	if overrides.OperationalBuilder != nil {
		config.OperationalBuilder = overrides.OperationalBuilder
	}
	if overrides.TrajectoryBuilder != nil {
		config.TrajectoryBuilder = overrides.TrajectoryBuilder
	}
	if overrides.AircraftFeatureProvider != nil {
		config.AircraftFeatureProvider =
			overrides.AircraftFeatureProvider
	}
	if overrides.Now != nil {
		config.Now = overrides.Now
	}

	extractor, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return extractor
}

func validRequest() Request {
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
	end := start.Add(3 * time.Minute)

	return Request{
		Trajectory: trajectory.FlightTrajectory{
			ID:              "trajectory-one",
			IdentityKey:     "flight-identity-example",
			IdentityBasis:   trajectory.FlightIdentityBasisAircraftAndStartTime,
			SplitReason:     trajectory.FlightSplitReasonInitialObservation,
			FlightID:        "flight-one",
			AircraftID:      "aircraft-one",
			ICAO24:          "abc123",
			Callsign:        " TEST123 ",
			StartTime:       start,
			EndTime:         end,
			DurationSeconds: 180,
			SegmentCount:    1,
			PointCount:      4,
			QualityScore:    0.75,
			SourceName:      "reconciled",
			Points: []trajectory.TrackPoint4D{
				{
					ID:         "point-one",
					Latitude:   40.4,
					Longitude:  49.8,
					ObservedAt: start,
					SourceName: "open-sky",
				},
				{
					ID:         "point-two",
					Latitude:   40.5,
					Longitude:  49.9,
					ObservedAt: start.Add(time.Minute),
					SourceName: "airplanes.live",
				},
				{
					ID:         "point-three",
					Latitude:   40.6,
					Longitude:  50.0,
					ObservedAt: start.Add(2 * time.Minute),
					SourceName: "open-sky",
				},
				{
					ID:         "point-four",
					Latitude:   40.7,
					Longitude:  50.1,
					ObservedAt: end,
					SourceName: "open-sky",
				},
			},
			Segments: []trajectory.TrajectorySegment{
				{
					ID:             "segment-one",
					TrajectoryID:   "trajectory-one",
					SequenceNumber: 1,
					Status:         trajectory.SegmentStatusObserved,
					StartTime:      start,
					EndTime:        end,
					PointCount:     4,
					SourceName:     "reconciled",
				},
			},
			CreatedAt: start.Add(4 * time.Minute),
			UpdatedAt: start.Add(5 * time.Minute),
		},
		AsOfTime: end,
	}
}
