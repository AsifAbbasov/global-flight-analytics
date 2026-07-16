package airspaceproduction

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactiongraph"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
)

type fakeObservationReader struct {
	observations []Observation
	query        ObservationQuery
	err          error
}

func (reader *fakeObservationReader) ListAirspaceObservations(
	_ context.Context,
	query ObservationQuery,
) ([]Observation, error) {
	reader.query = query
	if reader.err != nil {
		return nil, reader.err
	}
	result := make([]Observation, 0, len(reader.observations))
	for _, observation := range reader.observations {
		result = append(result, observation.Clone())
	}
	return result, nil
}

func TestGetAirspaceRegionAnalyticsComposesProductionPipeline(
	t *testing.T,
) {
	generatedAt := time.Date(
		2026, time.July, 17, 12, 6, 0, 0, time.UTC,
	)
	asOfTime := time.Date(
		2026, time.July, 17, 12, 5, 0, 0, time.UTC,
	)
	altitudeA := 10000.0
	altitudeB := 10300.0

	reader := &fakeObservationReader{
		observations: []Observation{
			{
				StateID:                     "state-a-1",
				ICAO24:                      "abc001",
				Callsign:                    "AZA101",
				Latitude:                    40.10,
				Longitude:                   49.80,
				AltitudeMeters:              &altitudeA,
				AltitudeReference:           interactiongraph.AltitudeReferenceGeometric,
				VelocityMetersPerSecond:     220,
				HeadingDegrees:              90,
				VerticalRateMetersPerSecond: 0,
				ObservedAt:                  asOfTime.Add(-70 * time.Second),
				SourceName:                  "postgres",
			},
			{
				StateID:                     "state-b-1",
				ICAO24:                      "abc002",
				Callsign:                    "AZA202",
				Latitude:                    40.10,
				Longitude:                   49.90,
				AltitudeMeters:              &altitudeB,
				AltitudeReference:           interactiongraph.AltitudeReferenceGeometric,
				VelocityMetersPerSecond:     220,
				HeadingDegrees:              270,
				VerticalRateMetersPerSecond: 0,
				ObservedAt:                  asOfTime.Add(-70 * time.Second),
				SourceName:                  "postgres",
			},
			{
				StateID:                     "state-a-2",
				ICAO24:                      "abc001",
				Callsign:                    "AZA101",
				Latitude:                    40.10,
				Longitude:                   49.84,
				AltitudeMeters:              &altitudeA,
				AltitudeReference:           interactiongraph.AltitudeReferenceGeometric,
				VelocityMetersPerSecond:     220,
				HeadingDegrees:              90,
				VerticalRateMetersPerSecond: 0,
				ObservedAt:                  asOfTime.Add(-10 * time.Second),
				SourceName:                  "postgres",
			},
			{
				StateID:                     "state-b-2",
				ICAO24:                      "abc002",
				Callsign:                    "AZA202",
				Latitude:                    40.10,
				Longitude:                   49.86,
				AltitudeMeters:              &altitudeB,
				AltitudeReference:           interactiongraph.AltitudeReferenceGeometric,
				VelocityMetersPerSecond:     220,
				HeadingDegrees:              270,
				VerticalRateMetersPerSecond: 0,
				ObservedAt:                  asOfTime.Add(-10 * time.Second),
				SourceName:                  "postgres",
			},
		},
	}

	service, err := New(Config{
		ObservationReader: reader,
		RegionResolver:    region.NewService(),
		Now: func() time.Time {
			return generatedAt
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, err := service.GetAirspaceRegionAnalytics(
		context.Background(),
		Request{
			RegionCode: "azerbaijan",
			AsOfTime:   asOfTime,
			Window:     2 * time.Minute,
		},
	)
	if err != nil {
		t.Fatalf("GetAirspaceRegionAnalytics() error = %v", err)
	}

	if result.RegionCode != "AZERBAIJAN" {
		t.Fatalf("RegionCode = %q, want AZERBAIJAN", result.RegionCode)
	}
	if result.Metrics.SnapshotCount != 2 {
		t.Fatalf("SnapshotCount = %d, want 2", result.Metrics.SnapshotCount)
	}
	if result.Occupancy.Metrics.ExpectedBucketCount != 2 {
		t.Fatalf(
			"ExpectedBucketCount = %d, want 2",
			result.Occupancy.Metrics.ExpectedBucketCount,
		)
	}
	if result.Provenance.InputFingerprint == "" {
		t.Fatal("InputFingerprint is empty")
	}
	if reader.query.Limit != maximumObservations+1 {
		t.Fatalf("query limit = %d", reader.query.Limit)
	}
	if !reader.query.WindowEnd.Equal(asOfTime) {
		t.Fatalf("query end = %s, want %s", reader.query.WindowEnd, asOfTime)
	}
}

func TestGetAirspaceRegionAnalyticsRejectsFutureAsOfTime(
	t *testing.T,
) {
	now := time.Date(
		2026, time.July, 17, 12, 0, 0, 0, time.UTC,
	)
	service, err := New(Config{
		ObservationReader: &fakeObservationReader{},
		RegionResolver:    region.NewService(),
		Now: func() time.Time {
			return now
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = service.GetAirspaceRegionAnalytics(
		context.Background(),
		Request{
			RegionCode: "azerbaijan",
			AsOfTime:   now.Add(time.Minute),
			Window:     time.Minute,
		},
	)
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("error = %v, want ErrInvalidRequest", err)
	}
}

func TestGetAirspaceRegionAnalyticsRejectsObservationOverflow(
	t *testing.T,
) {
	now := time.Date(
		2026, time.July, 17, 12, 0, 0, 0, time.UTC,
	)
	reader := &fakeObservationReader{
		observations: []Observation{
			{ICAO24: "abc001", ObservedAt: now.Add(-time.Second)},
			{ICAO24: "abc002", ObservedAt: now.Add(-time.Second)},
		},
	}
	service, err := New(Config{
		ObservationReader:   reader,
		RegionResolver:      region.NewService(),
		MaximumObservations: 1,
		Now: func() time.Time {
			return now
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = service.GetAirspaceRegionAnalytics(
		context.Background(),
		Request{
			RegionCode: "azerbaijan",
			AsOfTime:   now,
			Window:     time.Minute,
		},
	)
	if !errors.Is(err, ErrObservationCapacityExceeded) {
		t.Fatalf(
			"error = %v, want ErrObservationCapacityExceeded",
			err,
		)
	}
}
