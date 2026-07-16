package main

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/airspaceproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactiongraph"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
	"time"
)

func TestBuildVerificationSchedule(t *testing.T) {
	t.Parallel()

	schedule := buildVerificationSchedule()
	if schedule.WindowStart.Location() != time.UTC ||
		schedule.AsOfTime.Location() != time.UTC ||
		schedule.GeneratedAt.Location() != time.UTC {
		t.Fatalf("schedule is not UTC: %#v", schedule)
	}
	if schedule.AsOfTime.Sub(schedule.WindowStart) != verificationWindow {
		t.Fatalf(
			"window = %s, want %s",
			schedule.AsOfTime.Sub(schedule.WindowStart),
			verificationWindow,
		)
	}
	if len(schedule.SnapshotTimes) != 5 {
		t.Fatalf("snapshot count = %d, want 5", len(schedule.SnapshotTimes))
	}
	if !schedule.LatestObservedAt.Equal(
		schedule.SnapshotTimes[len(schedule.SnapshotTimes)-1].Add(-5 * time.Second),
	) {
		t.Fatalf("latest observed at = %s", schedule.LatestObservedAt)
	}
	if !schedule.FutureObservedAt.After(schedule.AsOfTime) {
		t.Fatalf("future observation is not after as-of time")
	}
}

func TestFixtureObservations(t *testing.T) {
	t.Parallel()

	schedule := buildVerificationSchedule()
	observations := fixtureObservations(schedule)
	if len(observations) != storedStateCount {
		t.Fatalf(
			"fixture state count = %d, want %d",
			len(observations),
			storedStateCount,
		)
	}

	successful := 0
	failed := 0
	unknownAltitude := 0
	future := 0
	outsideRegion := 0
	selectedAircraft := make(map[string]struct{})
	for _, observation := range observations {
		switch observation.IngestionRunID {
		case verificationSuccessRunID:
			successful++
		case verificationFailedRunID:
			failed++
		default:
			t.Fatalf("unexpected ingestion run ID: %s", observation.IngestionRunID)
		}
		if observation.AltitudeMeters == nil {
			unknownAltitude++
		}
		if observation.ObservedAt.After(schedule.AsOfTime) {
			future++
		}
		if observation.Latitude > 42 || observation.Latitude < 38 ||
			observation.Longitude > 51 || observation.Longitude < 44.5 {
			outsideRegion++
		}
		if observation.SourceName == verificationSuccessSource &&
			!observation.ObservedAt.After(schedule.AsOfTime) &&
			observation.Latitude >= 38 && observation.Latitude <= 42 &&
			observation.Longitude >= 44.5 && observation.Longitude <= 51 {
			selectedAircraft[strings.ToUpper(observation.ICAO24)] = struct{}{}
		}
	}
	if successful != storedSuccessfulStateCount || failed != storedFailedStateCount {
		t.Fatalf("unexpected run counts: success=%d failed=%d", successful, failed)
	}
	if unknownAltitude != unknownAltitudeObservationCount {
		t.Fatalf(
			"unknown-altitude count = %d, want %d",
			unknownAltitude,
			unknownAltitudeObservationCount,
		)
	}
	if future != 1 || outsideRegion != 1 {
		t.Fatalf("boundary rows: future=%d outside=%d", future, outsideRegion)
	}
	if len(selectedAircraft) != selectedAircraftCount {
		t.Fatalf(
			"selected aircraft = %d, want %d",
			len(selectedAircraft),
			selectedAircraftCount,
		)
	}
}

func TestAirspaceRequestURL(t *testing.T) {
	t.Parallel()

	schedule := buildVerificationSchedule()
	requestURL := airspaceRequestURL(
		verificationRegionCode,
		schedule.AsOfTime,
		verificationWindow,
	)
	parsed, err := url.Parse(requestURL)
	if err != nil {
		t.Fatalf("parse request URL: %v", err)
	}
	if parsed.Path != "/api/v1/airspace/regions/azerbaijan/analytics" {
		t.Fatalf("path = %q", parsed.Path)
	}
	if parsed.Query().Get("as_of_time") != schedule.AsOfTime.Format(time.RFC3339Nano) {
		t.Fatalf("as_of_time = %q", parsed.Query().Get("as_of_time"))
	}
	if parsed.Query().Get("window_seconds") != "300" {
		t.Fatalf("window_seconds = %q", parsed.Query().Get("window_seconds"))
	}
}

func TestFingerprintAndStringHelpers(t *testing.T) {
	t.Parallel()

	valid := strings.Repeat("a", 64)
	if !isHexFingerprint(valid) {
		t.Fatalf("valid fingerprint was rejected")
	}
	if isHexFingerprint("not-a-fingerprint") {
		t.Fatalf("invalid fingerprint was accepted")
	}
	if !equalStrings([]string{"b", "a"}, []string{"a", "b"}) {
		t.Fatalf("equalStrings did not ignore order")
	}
	if equalStrings([]string{"a"}, []string{"a", "b"}) {
		t.Fatalf("equalStrings accepted different lengths")
	}
}

type verificationObservationReader struct {
	observations []airspaceproduction.Observation
}

func (reader verificationObservationReader) ListAirspaceObservations(
	context.Context,
	airspaceproduction.ObservationQuery,
) ([]airspaceproduction.Observation, error) {
	result := make([]airspaceproduction.Observation, 0, len(reader.observations))
	for _, observation := range reader.observations {
		result = append(result, observation.Clone())
	}
	return result, nil
}

func TestVerificationFixtureBuildsExpectedPipeline(t *testing.T) {
	t.Parallel()

	schedule := buildVerificationSchedule()
	observations := make([]airspaceproduction.Observation, 0, selectedObservationCount)
	for _, item := range fixtureObservations(schedule) {
		if item.IngestionRunID != verificationSuccessRunID ||
			item.ObservedAt.After(schedule.AsOfTime) ||
			item.Latitude < 38 || item.Latitude > 42 ||
			item.Longitude < 44.5 || item.Longitude > 51 {
			continue
		}
		reference := interactiongraph.AltitudeReferenceUnknown
		if item.AltitudeMeters != nil {
			reference = interactiongraph.AltitudeReferenceGeometric
		}
		observations = append(observations, airspaceproduction.Observation{
			StateID:                     strings.ToUpper(item.ICAO24) + item.ObservedAt.Format(time.RFC3339Nano),
			ICAO24:                      strings.ToUpper(item.ICAO24),
			Callsign:                    item.Callsign,
			Latitude:                    item.Latitude,
			Longitude:                   item.Longitude,
			AltitudeMeters:              item.AltitudeMeters,
			AltitudeReference:           reference,
			VelocityMetersPerSecond:     item.Velocity,
			HeadingDegrees:              item.Heading,
			VerticalRateMetersPerSecond: item.VerticalRate,
			OnGround:                    item.OnGround,
			ObservedAt:                  item.ObservedAt,
			SourceName:                  item.SourceName,
		})
	}
	if len(observations) != selectedObservationCount {
		t.Fatalf("selected fixture observations = %d", len(observations))
	}

	service, err := airspaceproduction.New(airspaceproduction.Config{
		ObservationReader: verificationObservationReader{observations: observations},
		RegionResolver:    region.NewService(),
		Now: func() time.Time {
			return schedule.GeneratedAt
		},
	})
	if err != nil {
		t.Fatalf("create verification service: %v", err)
	}
	result, err := verifyDirectProductionComposition(
		context.Background(),
		service,
		schedule,
	)
	if err != nil {
		t.Fatalf("verify pipeline fixture: %v", err)
	}
	if result.Metrics.SnapshotCount != 5 ||
		result.Metrics.UniqueAircraftCount != selectedAircraftCount ||
		result.Metrics.IndeterminateRiskCount == 0 {
		t.Fatalf("unexpected pipeline metrics: %#v", result.Metrics)
	}
}
