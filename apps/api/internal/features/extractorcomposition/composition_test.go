package extractorcomposition

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/aircraft"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/aircraftprovider"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/geographicalbuilder"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/operationalbuilder"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/temporalbuilder"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/trajectorybuilder"
)

func TestNewRequiresAircraftLookup(t *testing.T) {
	_, err := New(Config{})
	if !errors.Is(err, ErrAircraftLookupRequired) {
		t.Fatalf(
			"New() error = %v, want %v",
			err,
			ErrAircraftLookupRequired,
		)
	}
}

func TestNewWrapsGeographicalBuilderConfigurationError(
	t *testing.T,
) {
	_, err := New(Config{
		AircraftLookup:          &fakeAircraftLookup{},
		GeographicCellPrecision: 7,
	})
	if !errors.Is(
		err,
		geographicalbuilder.ErrInvalidGeographicCellPrecision,
	) {
		t.Fatalf(
			"New() error = %v, want %v",
			err,
			geographicalbuilder.ErrInvalidGeographicCellPrecision,
		)
	}

	var componentErr *ComponentError
	if !errors.As(err, &componentErr) {
		t.Fatalf(
			"New() error = %T, want *ComponentError",
			err,
		)
	}
	if componentErr.Component !=
		ComponentGeographicalBuilder {
		t.Fatalf(
			"component = %q, want %q",
			componentErr.Component,
			ComponentGeographicalBuilder,
		)
	}
}

func TestNewWrapsAircraftProviderConfigurationError(
	t *testing.T,
) {
	_, err := New(Config{
		AircraftLookup:           &fakeAircraftLookup{},
		AircraftPositiveCacheTTL: -time.Second,
	})
	if !errors.Is(
		err,
		aircraftprovider.ErrInvalidPositiveCacheTTL,
	) {
		t.Fatalf(
			"New() error = %v, want %v",
			err,
			aircraftprovider.ErrInvalidPositiveCacheTTL,
		)
	}

	var componentErr *ComponentError
	if !errors.As(err, &componentErr) {
		t.Fatalf(
			"New() error = %T, want *ComponentError",
			err,
		)
	}
	if componentErr.Component != ComponentAircraftProvider {
		t.Fatalf(
			"component = %q, want %q",
			componentErr.Component,
			ComponentAircraftProvider,
		)
	}
}

func TestNewBuildsCompleteProductionExtractor(
	t *testing.T,
) {
	fixedNow := time.Date(
		2026,
		time.July,
		14,
		16,
		0,
		0,
		0,
		time.FixedZone("test", 4*60*60),
	)
	lookup := &fakeAircraftLookup{
		result: aircraft.Aircraft{
			ICAO24:       "ABC123",
			Registration: "4K-AAA",
			Model:        "A320-200",
			Manufacturer: "Airbus",
			AircraftType: "A320",
			Airline:      "Example Air",
			Country:      "Azerbaijan",
		},
	}

	composition, err := New(Config{
		AircraftLookup: lookup,
		Now: func() time.Time {
			return fixedNow
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if composition.Extractor == nil {
		t.Fatal("New() returned a nil Extractor")
	}
	if !reflect.DeepEqual(
		composition.Versions,
		CurrentVersions(),
	) {
		t.Fatalf(
			"versions = %#v, want %#v",
			composition.Versions,
			CurrentVersions(),
		)
	}

	item := completeTrajectory()
	original := cloneTrajectoryForTest(item)
	asOfTime := item.EndTime.Add(time.Minute)

	features, err := composition.Extractor.Extract(
		context.Background(),
		extractor.Request{
			Trajectory: item,
			AsOfTime:   asOfTime,
		},
	)
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if features.SchemaVersion !=
		flightfeatures.SchemaVersionV1 ||
		features.TrajectoryID != "trajectory-1" ||
		features.IdentityKey != "identity-1" ||
		features.ICAO24 != "ABC123" ||
		features.Callsign != "TEST123" {
		t.Fatalf(
			"unexpected feature identity: %#v",
			features,
		)
	}
	if !features.ExtractedAt.Equal(fixedNow.UTC()) {
		t.Fatalf(
			"ExtractedAt = %v, want %v",
			features.ExtractedAt,
			fixedNow.UTC(),
		)
	}
	if features.Provenance.ExtractorVersion !=
		extractor.Version {
		t.Fatalf(
			"extractor version = %q, want %q",
			features.Provenance.ExtractorVersion,
			extractor.Version,
		)
	}
	if features.Provenance.InputFingerprint == "" {
		t.Fatal("input fingerprint is empty")
	}

	assertCompleteEvidence(
		t,
		"temporal",
		features.Temporal.Evidence,
		8,
		3,
	)
	assertCompleteEvidence(
		t,
		"geographical",
		features.Geographical.Evidence,
		11,
		3,
	)
	assertCompleteEvidence(
		t,
		"operational",
		features.Operational.Evidence,
		11,
		3,
	)
	assertCompleteEvidence(
		t,
		"trajectory",
		features.Trajectory.Evidence,
		16,
		3,
	)
	assertCompleteEvidence(
		t,
		"aircraft",
		features.Aircraft.Evidence,
		6,
		0,
	)

	if features.Aircraft.Registration != "4K-AAA" ||
		features.Aircraft.Manufacturer != "Airbus" ||
		features.Aircraft.Model != "A320-200" ||
		features.Aircraft.AircraftType != "A320" ||
		features.Aircraft.Airline != "Example Air" ||
		features.Aircraft.Country != "Azerbaijan" {
		t.Fatalf(
			"unexpected aircraft features: %#v",
			features.Aircraft,
		)
	}
	if features.Quality.Status !=
		flightfeatures.ValidationStatusUnvalidated ||
		features.Quality.CompletenessScore != 1 ||
		features.Quality.InputQualityScore != 0.9 ||
		features.Quality.SupportingPointCount != 3 {
		t.Fatalf(
			"unexpected initial quality: %#v",
			features.Quality,
		)
	}
	if lookup.callCount != 1 ||
		lookup.lastICAO24 != "ABC123" {
		t.Fatalf(
			"lookup calls = %d, last ICAO24 = %q",
			lookup.callCount,
			lookup.lastICAO24,
		)
	}
	if !reflect.DeepEqual(item, original) {
		t.Fatalf(
			"input trajectory mutated\nitem=%#v\noriginal=%#v",
			item,
			original,
		)
	}

	secondItem := completeTrajectory()
	secondItem.ID = "trajectory-2"
	secondItem.IdentityKey = "identity-2"
	if _, err := composition.Extractor.Extract(
		context.Background(),
		extractor.Request{
			Trajectory: secondItem,
			AsOfTime: secondItem.EndTime.Add(
				time.Minute,
			),
		},
	); err != nil {
		t.Fatalf(
			"second Extract() error = %v",
			err,
		)
	}
	if lookup.callCount != 1 {
		t.Fatalf(
			"aircraft provider cache missed: calls = %d, want 1",
			lookup.callCount,
		)
	}
}

func TestNewExtractorReturnsDirectProductionExtractor(
	t *testing.T,
) {
	featureExtractor, err := NewExtractor(Config{
		AircraftLookup: &fakeAircraftLookup{
			result: aircraft.Aircraft{
				ICAO24: "ABC123",
			},
		},
	})
	if err != nil {
		t.Fatalf(
			"NewExtractor() error = %v",
			err,
		)
	}
	if featureExtractor == nil {
		t.Fatal(
			"NewExtractor() returned a nil Extractor",
		)
	}
}

func TestCurrentVersionsRemainStable(t *testing.T) {
	want := Versions{
		Composition:         "flight-feature-extractor-composition-v1",
		Extractor:           "flight-feature-extractor-v1",
		AircraftProvider:    "aircraft-feature-provider-v1",
		TemporalBuilder:     "temporal-feature-builder-v1",
		GeographicalBuilder: "geographical-feature-builder-v1",
		OperationalBuilder:  "operational-feature-builder-v1",
		TrajectoryBuilder:   "trajectory-feature-builder-v1",
	}

	if got := CurrentVersions(); !reflect.DeepEqual(got, want) {
		t.Fatalf(
			"CurrentVersions() = %#v, want %#v",
			got,
			want,
		)
	}
	if want.Extractor != extractor.Version ||
		want.AircraftProvider != aircraftprovider.Version ||
		want.TemporalBuilder != temporalbuilder.Version ||
		want.GeographicalBuilder !=
			geographicalbuilder.Version ||
		want.OperationalBuilder !=
			operationalbuilder.Version ||
		want.TrajectoryBuilder !=
			trajectorybuilder.Version {
		t.Fatal(
			"component version constants diverged from the composition manifest",
		)
	}
}

type fakeAircraftLookup struct {
	result     aircraft.Aircraft
	err        error
	callCount  int
	lastICAO24 string
}

func (lookup *fakeAircraftLookup) GetByICAO24(
	ctx context.Context,
	icao24 string,
) (aircraft.Aircraft, error) {
	if err := ctx.Err(); err != nil {
		return aircraft.Aircraft{}, err
	}

	lookup.callCount++
	lookup.lastICAO24 = icao24

	if lookup.err != nil {
		return aircraft.Aircraft{}, lookup.err
	}

	return lookup.result, nil
}

func completeTrajectory() trajectory.FlightTrajectory {
	startTime := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	endTime := startTime.Add(2 * time.Minute)

	points := []trajectory.TrackPoint4D{
		{
			ID:                       "point-1",
			ICAO24:                   "abc123",
			Callsign:                 " TEST123 ",
			Latitude:                 40,
			Longitude:                49,
			BarometricAltitudeStatus: flightstate.AltitudeStatusGround,
			VelocityMPS:              0,
			HeadingDegrees:           350,
			VerticalRateMPS:          0,
			OnGround:                 true,
			ObservedAt:               startTime,
			SourceName:               "source-a",
		},
		{
			ID:                       "point-2",
			ICAO24:                   "abc123",
			Callsign:                 " TEST123 ",
			Latitude:                 40.5,
			Longitude:                49.5,
			BarometricAltitudeM:      1000,
			BarometricAltitudeStatus: flightstate.AltitudeStatusObserved,
			VelocityMPS:              100,
			HeadingDegrees:           10,
			VerticalRateMPS:          5,
			ObservedAt: startTime.Add(
				time.Minute,
			),
			SourceName: "source-a",
		},
		{
			ID:                       "point-3",
			ICAO24:                   "abc123",
			Callsign:                 " TEST123 ",
			Latitude:                 41,
			Longitude:                50,
			BarometricAltitudeM:      2000,
			BarometricAltitudeStatus: flightstate.AltitudeStatusObserved,
			VelocityMPS:              200,
			HeadingDegrees:           20,
			VerticalRateMPS:          10,
			ObservedAt:               endTime,
			SourceName:               "source-b",
		},
	}

	segments := []trajectory.TrajectorySegment{
		{
			ID:              "segment-1",
			TrajectoryID:    "trajectory-1",
			ICAO24:          "abc123",
			Callsign:        " TEST123 ",
			SequenceNumber:  1,
			Status:          trajectory.SegmentStatusObserved,
			QualityScore:    0.9,
			StartTime:       startTime,
			EndTime:         endTime,
			DurationSeconds: 120,
			StartLatitude:   40,
			StartLongitude:  49,
			EndLatitude:     41,
			EndLongitude:    50,
			PointCount:      3,
			SourceName:      "source-a",
		},
	}

	return trajectory.FlightTrajectory{
		ID:               "trajectory-1",
		IdentityKey:      "identity-1",
		FlightID:         "flight-1",
		AircraftID:       "aircraft-1",
		ICAO24:           "abc123",
		Callsign:         " TEST123 ",
		StartTime:        startTime,
		EndTime:          endTime,
		DurationSeconds:  120,
		SegmentCount:     len(segments),
		PointCount:       len(points),
		CoverageGapCount: 0,
		QualityScore:     0.9,
		SourceName:       "source-a",
		Points:           points,
		Segments:         segments,
		CoverageGaps:     nil,
		CreatedAt:        startTime,
		UpdatedAt:        endTime,
	}
}

func cloneTrajectoryForTest(
	item trajectory.FlightTrajectory,
) trajectory.FlightTrajectory {
	cloned := item
	cloned.Points = append(
		[]trajectory.TrackPoint4D(nil),
		item.Points...,
	)
	cloned.Segments = append(
		[]trajectory.TrajectorySegment(nil),
		item.Segments...,
	)
	cloned.CoverageGaps = append(
		[]trajectory.CoverageGap(nil),
		item.CoverageGaps...,
	)

	return cloned
}

func assertCompleteEvidence(
	t *testing.T,
	group string,
	evidence flightfeatures.GroupEvidence,
	fieldCount int,
	supportingPointCount int,
) {
	t.Helper()

	if evidence.Status !=
		flightfeatures.AvailabilityStatusAvailable ||
		evidence.AvailableFieldCount != fieldCount ||
		evidence.TotalFieldCount != fieldCount ||
		evidence.SupportingPointCount !=
			supportingPointCount {
		t.Fatalf(
			"%s evidence = %#v",
			group,
			evidence,
		)
	}
}
