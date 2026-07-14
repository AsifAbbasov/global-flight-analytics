package routecontext

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type trajectoryReaderStub struct {
	item  trajectory.FlightTrajectory
	err   error
	calls int
	icao  string
}

func (stub *trajectoryReaderStub) GetLatestTrajectoryByICAO24(
	ctx context.Context,
	icao24 string,
) (trajectory.FlightTrajectory, error) {
	stub.calls++
	stub.icao = icao24

	return stub.item, stub.err
}

type airportListerStub struct {
	items []airport.Airport
	err   error
	calls int
}

func (stub *airportListerStub) List(
	ctx context.Context,
) ([]airport.Airport, error) {
	stub.calls++

	return append([]airport.Airport(nil), stub.items...), stub.err
}

func TestServiceBuildsProbableRouteContext(t *testing.T) {
	generatedAt := time.Date(
		2026,
		time.July,
		14,
		10,
		30,
		0,
		0,
		time.UTC,
	)
	trajectoryReader := &trajectoryReaderStub{
		item: trajectory.FlightTrajectory{
			ID:               "trajectory-one",
			ICAO24:           "ABC123",
			QualityScore:     0.9,
			CoverageGapCount: 1,
			Segments: []trajectory.TrajectorySegment{
				{
					ID:             "segment-two",
					SequenceNumber: 2,
					Status:         trajectory.SegmentStatusObserved,
					StartLatitude:  40.4,
					StartLongitude: 49.8,
					EndLatitude:    41.68,
					EndLongitude:   44.95,
					PointCount:     8,
					StartTime:      generatedAt.Add(-20 * time.Minute),
				},
				{
					ID:             "segment-one",
					SequenceNumber: 1,
					Status:         trajectory.SegmentStatusObserved,
					StartLatitude:  40.47,
					StartLongitude: 50.05,
					EndLatitude:    40.4,
					EndLongitude:   49.8,
					PointCount:     10,
					StartTime:      generatedAt.Add(-40 * time.Minute),
				},
			},
		},
	}
	airportLister := &airportListerStub{
		items: []airport.Airport{
			{
				ICAOCode:  "UBBB",
				IATACode:  "GYD",
				Name:      "Heydar Aliyev International Airport",
				City:      "Baku",
				Country:   "Azerbaijan",
				Latitude:  40.4675,
				Longitude: 50.0467,
			},
			{
				ICAOCode:  "UGTB",
				IATACode:  "TBS",
				Name:      "Tbilisi International Airport",
				City:      "Tbilisi",
				Country:   "Georgia",
				Latitude:  41.6692,
				Longitude: 44.9547,
			},
		},
	}
	originalAirports := append(
		[]airport.Airport(nil),
		airportLister.items...,
	)

	service := New(Config{
		TrajectoryReader: trajectoryReader,
		AirportLister:    airportLister,
		Now: func() time.Time {
			return generatedAt
		},
	})

	result, err := service.GetByICAO24(
		context.Background(),
		" abc123 ",
	)
	if err != nil {
		t.Fatalf("expected route context, got %v", err)
	}

	if result.ICAO24 != "ABC123" ||
		result.TrajectoryID != "trajectory-one" ||
		!result.GeneratedAt.Equal(generatedAt) {
		t.Fatalf("unexpected context identity: %#v", result)
	}
	if result.Origin == nil ||
		result.Origin.Airport.ICAOCode != "UBBB" {
		t.Fatalf("unexpected origin: %#v", result.Origin)
	}
	if result.Destination == nil ||
		result.Destination.Airport.ICAOCode != "UGTB" {
		t.Fatalf(
			"unexpected destination: %#v",
			result.Destination,
		)
	}
	if result.Confidence.Level == ConfidenceLevelNone ||
		result.Confidence.Score <= 0 {
		t.Fatalf(
			"expected positive route confidence, got %#v",
			result.Confidence,
		)
	}
	if !hasNotice(
		result.Limitations,
		"trajectory_coverage_gaps",
	) {
		t.Fatalf(
			"expected coverage-gap limitation, got %#v",
			result.Limitations,
		)
	}
	if trajectoryReader.calls != 1 ||
		trajectoryReader.icao != "ABC123" ||
		airportLister.calls != 1 {
		t.Fatalf(
			"unexpected dependency calls: trajectory=%d icao=%q airports=%d",
			trajectoryReader.calls,
			trajectoryReader.icao,
			airportLister.calls,
		)
	}
	if !reflect.DeepEqual(
		airportLister.items,
		originalAirports,
	) {
		t.Fatal("expected airport input not to be mutated")
	}
}

func TestServiceCachesAirportCatalogBetweenRequests(t *testing.T) {
	now := time.Date(
		2026,
		time.July,
		14,
		11,
		0,
		0,
		0,
		time.UTC,
	)
	trajectoryReader := &trajectoryReaderStub{
		item: trajectory.FlightTrajectory{
			ID:           "trajectory-one",
			ICAO24:       "ABC123",
			QualityScore: 0.8,
			Segments: []trajectory.TrajectorySegment{
				{
					SequenceNumber: 1,
					Status:         trajectory.SegmentStatusObserved,
					StartLatitude:  40.4675,
					StartLongitude: 50.0467,
					EndLatitude:    41.6692,
					EndLongitude:   44.9547,
					PointCount:     8,
				},
			},
		},
	}
	airportLister := &airportListerStub{
		items: []airport.Airport{
			{
				ICAOCode:  "UBBB",
				Latitude:  40.4675,
				Longitude: 50.0467,
			},
			{
				ICAOCode:  "UGTB",
				Latitude:  41.6692,
				Longitude: 44.9547,
			},
		},
	}
	service := New(Config{
		TrajectoryReader: trajectoryReader,
		AirportLister:    airportLister,
		AirportCacheTTL:  time.Hour,
		Now: func() time.Time {
			return now
		},
	})

	for index := 0; index < 2; index++ {
		_, err := service.GetByICAO24(
			context.Background(),
			"ABC123",
		)
		if err != nil {
			t.Fatalf(
				"request %d expected context, got %v",
				index,
				err,
			)
		}
	}

	if airportLister.calls != 1 {
		t.Fatalf(
			"expected one airport catalog load, got %d",
			airportLister.calls,
		)
	}
	if trajectoryReader.calls != 2 {
		t.Fatalf(
			"expected two trajectory reads, got %d",
			trajectoryReader.calls,
		)
	}
}

func TestServiceRejectsInvalidICAO24BeforeDependencies(t *testing.T) {
	trajectoryReader := &trajectoryReaderStub{}
	airportLister := &airportListerStub{}
	service := New(Config{
		TrajectoryReader: trajectoryReader,
		AirportLister:    airportLister,
	})

	_, err := service.GetByICAO24(
		context.Background(),
		"bad",
	)
	if !errors.Is(err, ErrInvalidICAO24) {
		t.Fatalf("expected invalid ICAO24, got %v", err)
	}
	if trajectoryReader.calls != 0 || airportLister.calls != 0 {
		t.Fatal("expected dependencies not to run")
	}
}

func TestServiceReturnsUnavailableContextWithoutUsableSegments(
	t *testing.T,
) {
	service := New(Config{
		TrajectoryReader: &trajectoryReaderStub{
			item: trajectory.FlightTrajectory{
				ID:     "trajectory-one",
				ICAO24: "ABC123",
				Segments: []trajectory.TrajectorySegment{
					{
						Status:         trajectory.SegmentStatusInvalid,
						StartLatitude:  40,
						StartLongitude: 50,
						EndLatitude:    41,
						EndLongitude:   51,
					},
				},
			},
		},
		AirportLister: &airportListerStub{
			items: []airport.Airport{
				{
					ICAOCode:  "UBBB",
					Latitude:  40.4675,
					Longitude: 50.0467,
				},
			},
		},
	})

	result, err := service.GetByICAO24(
		context.Background(),
		"ABC123",
	)
	if err != nil {
		t.Fatalf("expected unavailable context, got %v", err)
	}
	if result.Origin != nil || result.Destination != nil {
		t.Fatalf(
			"expected no candidates, got origin=%#v destination=%#v",
			result.Origin,
			result.Destination,
		)
	}
	if result.Confidence.Level != ConfidenceLevelNone {
		t.Fatalf(
			"expected none confidence, got %#v",
			result.Confidence,
		)
	}
}

func TestServiceRejectsAirportOutsideMaximumDistance(t *testing.T) {
	service := New(Config{
		TrajectoryReader: &trajectoryReaderStub{
			item: trajectory.FlightTrajectory{
				ID:           "trajectory-one",
				ICAO24:       "ABC123",
				QualityScore: 1,
				Segments: []trajectory.TrajectorySegment{
					{
						SequenceNumber: 1,
						Status:         trajectory.SegmentStatusObserved,
						StartLatitude:  0,
						StartLongitude: 0,
						EndLatitude:    0.1,
						EndLongitude:   0.1,
						PointCount:     10,
					},
				},
			},
		},
		AirportLister: &airportListerStub{
			items: []airport.Airport{
				{
					ICAOCode:  "FAR1",
					Latitude:  20,
					Longitude: 20,
				},
			},
		},
		MaximumCandidateDistanceKM: 50,
	})

	result, err := service.GetByICAO24(
		context.Background(),
		"ABC123",
	)
	if err != nil {
		t.Fatalf("expected context, got %v", err)
	}
	if result.Origin != nil || result.Destination != nil {
		t.Fatalf(
			"expected distant airport to be rejected, got %#v",
			result,
		)
	}
	if !hasNotice(
		result.Limitations,
		"origin_candidate_unavailable",
	) || !hasNotice(
		result.Limitations,
		"destination_candidate_unavailable",
	) {
		t.Fatalf(
			"expected missing-candidate limitations, got %#v",
			result.Limitations,
		)
	}
}

func TestServiceReducesConfidenceForSameAirport(t *testing.T) {
	service := New(Config{
		TrajectoryReader: &trajectoryReaderStub{
			item: trajectory.FlightTrajectory{
				ID:           "trajectory-one",
				ICAO24:       "ABC123",
				QualityScore: 1,
				Segments: []trajectory.TrajectorySegment{
					{
						SequenceNumber: 1,
						Status:         trajectory.SegmentStatusObserved,
						StartLatitude:  40.4675,
						StartLongitude: 50.0467,
						EndLatitude:    40.468,
						EndLongitude:   50.047,
						PointCount:     10,
					},
				},
			},
		},
		AirportLister: &airportListerStub{
			items: []airport.Airport{
				{
					ICAOCode:  "UBBB",
					Latitude:  40.4675,
					Longitude: 50.0467,
				},
			},
		},
	})

	result, err := service.GetByICAO24(
		context.Background(),
		"ABC123",
	)
	if err != nil {
		t.Fatalf("expected context, got %v", err)
	}
	if !hasNotice(
		result.Confidence.Reasons,
		"same_airport_candidate",
	) {
		t.Fatalf(
			"expected same-airport reason, got %#v",
			result.Confidence.Reasons,
		)
	}
}

func hasNotice(
	notices []Notice,
	code string,
) bool {
	for _, notice := range notices {
		if notice.Code == code {
			return true
		}
	}

	return false
}
