package weathercontext

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	domainweather "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionbaseline"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatheralignment"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatherencounter"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathertrust"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatheruncertainty"
)

func TestServiceGetComposesBoundedWeatherContext(
	t *testing.T,
) {
	t.Parallel()

	asOfTime := time.Date(
		2026,
		time.July,
		16,
		12,
		0,
		0,
		0,
		time.UTC,
	)
	trajectoryID := "b06ac65d-914a-4daa-8d08-a75ae97b7292"
	projectionGeneratedAt := asOfTime.Add(time.Second)
	generatedAt := asOfTime.Add(2 * time.Second)

	trajectoryReader := &fakeTrajectoryReader{
		result: testTrajectory(
			trajectoryID,
			asOfTime,
		),
	}
	weatherReader := &fakeWeatherSnapshotReader{
		result: testWeatherSnapshot(asOfTime),
	}
	projectionReader := &fakeProjectionReader{
		result: testProjection(
			t,
			trajectoryID,
			asOfTime,
			projectionGeneratedAt,
		),
	}

	service, err := NewService(
		Config{
			TrajectoryReader:      trajectoryReader,
			WeatherSnapshotReader: weatherReader,
			ProjectionReader:      projectionReader,
			TrustPolicy:           weathertrust.DefaultPolicy(),
			AlignmentPolicy:       weatheralignment.DefaultPolicy(),
			EncounterPolicy:       weatherencounter.DefaultPolicy(),
			UncertaintyPolicy:     weatheruncertainty.DefaultPolicy(),
			Now: func() time.Time {
				return generatedAt
			},
		},
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	result, err := service.Get(
		context.Background(),
		Request{
			TrajectoryID:      trajectoryID,
			AsOfTime:          asOfTime,
			RequestedDuration: 10 * time.Minute,
		},
	)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if err := result.Validate(); err != nil {
		t.Fatalf("result.Validate() error = %v", err)
	}

	if result.Alignment.PointCount != 2 {
		t.Fatalf(
			"Alignment.PointCount = %d, want 2 bounded points",
			result.Alignment.PointCount,
		)
	}
	if weatherReader.request.Latitude != 40.20 ||
		weatherReader.request.Longitude != 49.90 {
		t.Fatalf(
			"weather request coordinates = (%v, %v), want latest bounded point",
			weatherReader.request.Latitude,
			weatherReader.request.Longitude,
		)
	}
	if !weatherReader.request.AsOfTime.Equal(asOfTime) {
		t.Fatalf(
			"weather request as-of time = %s, want %s",
			weatherReader.request.AsOfTime,
			asOfTime,
		)
	}
	if projectionReader.request.TrajectoryID != trajectoryID ||
		!projectionReader.request.AsOfTime.Equal(asOfTime) ||
		projectionReader.request.RequestedDuration !=
			10*time.Minute {
		t.Fatalf(
			"projection request = %#v",
			projectionReader.request,
		)
	}
	if result.Uncertainty.Status !=
		weatheruncertainty.StatusWithheld {
		t.Fatalf(
			"Uncertainty.Status = %q, want %q for surface-only weather",
			result.Uncertainty.Status,
			weatheruncertainty.StatusWithheld,
		)
	}
	if !result.GeneratedAt.Equal(generatedAt) {
		t.Fatalf(
			"GeneratedAt = %s, want %s",
			result.GeneratedAt,
			generatedAt,
		)
	}
}

func TestServiceGetPreservesNotFound(
	t *testing.T,
) {
	t.Parallel()

	service, err := NewService(
		Config{
			TrajectoryReader: &fakeTrajectoryReader{
				err: ErrTrajectoryNotFound,
			},
			WeatherSnapshotReader: &fakeWeatherSnapshotReader{},
			ProjectionReader:      &fakeProjectionReader{},
			TrustPolicy:           weathertrust.DefaultPolicy(),
			AlignmentPolicy:       weatheralignment.DefaultPolicy(),
			EncounterPolicy:       weatherencounter.DefaultPolicy(),
			UncertaintyPolicy:     weatheruncertainty.DefaultPolicy(),
		},
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	_, err = service.Get(
		context.Background(),
		Request{
			TrajectoryID:      "b06ac65d-914a-4daa-8d08-a75ae97b7292",
			AsOfTime:          time.Now().UTC(),
			RequestedDuration: time.Minute,
		},
	)
	if !errors.Is(err, ErrTrajectoryNotFound) {
		t.Fatalf(
			"Get() error = %v, want ErrTrajectoryNotFound",
			err,
		)
	}
}

type fakeTrajectoryReader struct {
	result trajectory.FlightTrajectory
	err    error
}

func (
	reader *fakeTrajectoryReader,
) GetTrajectoryByID(
	context.Context,
	string,
) (trajectory.FlightTrajectory, error) {
	return reader.result, reader.err
}

type fakeWeatherSnapshotReader struct {
	request WeatherSnapshotRequest
	result  domainweather.CurrentSnapshot
	err     error
}

func (
	reader *fakeWeatherSnapshotReader,
) GetLatestSnapshot(
	_ context.Context,
	request WeatherSnapshotRequest,
) (domainweather.CurrentSnapshot, error) {
	reader.request = request
	return reader.result, reader.err
}

type fakeProjectionReader struct {
	request ProjectionRequest
	result  projectionproduction.Result
	err     error
}

func (
	reader *fakeProjectionReader,
) GetProjection(
	_ context.Context,
	request ProjectionRequest,
) (projectionproduction.Result, error) {
	reader.request = request
	return reader.result.Clone(), reader.err
}

func testTrajectory(
	trajectoryID string,
	asOfTime time.Time,
) trajectory.FlightTrajectory {
	points := []trajectory.TrackPoint4D{
		testGroundPoint(
			"point-a",
			asOfTime.Add(-10*time.Minute),
			40.10,
			49.80,
		),
		testGroundPoint(
			"point-b",
			asOfTime.Add(-time.Minute),
			40.20,
			49.90,
		),
		testGroundPoint(
			"future-point",
			asOfTime.Add(time.Minute),
			55.00,
			60.00,
		),
	}

	return trajectory.FlightTrajectory{
		ID:               trajectoryID,
		IdentityKey:      "weather-context-test-identity",
		IdentityBasis:    trajectory.FlightIdentityBasisSourceFlightID,
		SplitReason:      trajectory.FlightSplitReasonInitialObservation,
		FlightID:         "4baec5a2-c7d6-4df4-8169-4e457a32f14f",
		AircraftID:       "bd19e090-8bc0-4fe7-bcc6-b33c80a03514",
		ICAO24:           "4A1234",
		Callsign:         "AHY123",
		StartTime:        points[0].ObservedAt,
		EndTime:          points[len(points)-1].ObservedAt,
		DurationSeconds:  11 * 60,
		PointCount:       len(points),
		QualityScore:     0.90,
		SourceName:       "weather-context-test",
		Points:           points,
		Segments:         []trajectory.TrajectorySegment{},
		CoverageGaps:     []trajectory.CoverageGap{},
		CreatedAt:        asOfTime.Add(-time.Hour),
		UpdatedAt:        asOfTime.Add(time.Minute),
		SegmentCount:     0,
		CoverageGapCount: 0,
	}
}

func testGroundPoint(
	id string,
	observedAt time.Time,
	latitude float64,
	longitude float64,
) trajectory.TrackPoint4D {
	return trajectory.TrackPoint4D{
		ID:                       id,
		FlightStateID:            id,
		FlightID:                 "4baec5a2-c7d6-4df4-8169-4e457a32f14f",
		AircraftID:               "bd19e090-8bc0-4fe7-bcc6-b33c80a03514",
		ICAO24:                   "4A1234",
		Callsign:                 "AHY123",
		Latitude:                 latitude,
		Longitude:                longitude,
		BarometricAltitudeM:      0,
		BarometricAltitudeStatus: flightstate.AltitudeStatusGround,
		GeometricAltitudeM:       0,
		GeometricAltitudeStatus:  flightstate.AltitudeStatusGround,
		VelocityMPS:              0,
		HeadingDegrees:           0,
		VerticalRateMPS:          0,
		OnGround:                 true,
		OriginCountry:            "Azerbaijan",
		ObservedAt:               observedAt.UTC(),
		SourceName:               "weather-context-test",
	}
}

func testWeatherSnapshot(
	asOfTime time.Time,
) domainweather.CurrentSnapshot {
	return domainweather.CurrentSnapshot{
		Provider:                 domainweather.ProviderOpenMeteo,
		Latitude:                 40.20,
		Longitude:                49.90,
		ObservedAt:               asOfTime.Add(-2 * time.Minute),
		RetrievedAt:              asOfTime.Add(-time.Minute),
		TemperatureCelsius:       25,
		RelativeHumidityPercent:  55,
		PrecipitationMillimeters: 0,
		RainMillimeters:          0,
		WeatherCode:              1,
		CloudCoverPercent:        20,
		SurfacePressureHPA:       1012,
		WindSpeedMetersPerSecond: 4,
		WindDirectionDegrees:     180,
		WindGustsMetersPerSecond: 6,
	}
}

func testProjection(
	t *testing.T,
	trajectoryID string,
	asOfTime time.Time,
	generatedAt time.Time,
) projectionproduction.Result {
	t.Helper()

	confidence := projectioncontract.Confidence{
		Score: 0.80,
		Level: projectioncontract.
			ConfidenceLevelHigh,
		Reasons: []projectioncontract.ConfidenceReason{
			{
				Code:         "test_projection_confidence",
				Message:      "Test projection confidence.",
				Contribution: 0.80,
			},
		},
	}
	observedAt := asOfTime.Add(-time.Minute)
	projection := projectioncontract.Result{
		SchemaVersion: projectioncontract.SchemaVersionV1,
		Status:        projectioncontract.ResultStatusLimited,
		TrajectoryID:  trajectoryID,
		Method: projectioncontract.Method{
			Name:          projectionbaseline.MethodName,
			Version:       projectionbaseline.Version,
			DecisionClass: projectioncontract.DecisionClassPhysicsDerived,
		},
		Horizon: projectioncontract.Horizon{
			AsOfTime: asOfTime,
			EndTime:  asOfTime.Add(10 * time.Minute),
			Step:     5 * time.Minute,
		},
		Points: []projectioncontract.ProjectionPoint{
			{
				Sequence:     0,
				ForecastTime: asOfTime.Add(5 * time.Minute),
				Position: projectioncontract.Position{
					Latitude:  40.25,
					Longitude: 49.95,
				},
				Uncertainty: projectioncontract.Uncertainty{
					HorizontalRadiusM: 500,
				},
				Confidence: confidence,
			},
			{
				Sequence:     1,
				ForecastTime: asOfTime.Add(10 * time.Minute),
				Position: projectioncontract.Position{
					Latitude:  40.30,
					Longitude: 50.00,
				},
				Uncertainty: projectioncontract.Uncertainty{
					HorizontalRadiusM: 700,
				},
				Confidence: confidence,
			},
		},
		Confidence: confidence,
		Limitations: []projectioncontract.Limitation{
			{
				Code:    "test_projection_limited",
				Message: "Test projection is intentionally limited.",
				Scope:   "result",
			},
		},
		Explanations: []projectioncontract.Explanation{
			{
				Code:    "test_projection",
				Message: "Deterministic projection fixture.",
			},
		},
		ScopeGuard: projectioncontract.ScopeGuardResearchOnly,
		Provenance: projectioncontract.Provenance{
			InputFingerprint: "sha256:" + strings.Repeat("b", 64),
			Inputs: []projectioncontract.InputReference{
				{
					Name:           "trajectory",
					Classification: projectioncontract.InputClassificationObserved,
					SourceName:     "weather-context-test",
					ObservedAt:     observedAt,
					RetrievedAt:    generatedAt,
				},
			},
			LatestInputObservedAt: observedAt,
		},
		GeneratedAt: generatedAt,
	}

	result := projectionproduction.Result{
		Version:          projectionproduction.Version,
		Strategy:         projectionproduction.StrategyKinematic,
		FallbackReason:   "test_kinematic_fallback",
		ArrivalStatus:    projectionproduction.ArrivalStatusSkipped,
		Projection:       projection,
		Notices:          []projectionproduction.Notice{{Code: "test_fallback", Message: "Test kinematic fallback."}},
		InputFingerprint: "sha256:" + strings.Repeat("c", 64),
		GeneratedAt:      generatedAt,
	}
	if err := result.Validate(); err != nil {
		t.Fatalf("test projection is invalid: %v", err)
	}
	return result
}
