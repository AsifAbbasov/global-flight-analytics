package main

import (
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func TestBuildEvidenceScheduleUsesAdjacentClosedPeriods(
	t *testing.T,
) {
	now := time.Date(
		2026,
		time.July,
		15,
		12,
		37,
		45,
		123456789,
		time.FixedZone("Asia/Baku", 4*60*60),
	)

	schedule, err := buildEvidenceSchedule(
		now,
	)
	if err != nil {
		t.Fatalf(
			"build evidence schedule: %v",
			err,
		)
	}

	expectedAsOf := now.UTC()
	expectedBoundary := expectedAsOf.Truncate(
		time.Hour,
	)

	if !schedule.AsOfTime.Equal(
		expectedAsOf,
	) ||
		!schedule.GeneratedAt.Equal(
			expectedAsOf,
		) {
		t.Fatalf(
			"unexpected as-of or generated time: %#v",
			schedule,
		)
	}
	if !schedule.ClosedBoundary.Equal(
		expectedBoundary,
	) {
		t.Fatalf(
			"closed boundary = %s, want %s",
			schedule.ClosedBoundary,
			expectedBoundary,
		)
	}
	if schedule.CurrentEnd.Sub(
		schedule.CurrentStart,
	) != 2*time.Hour ||
		schedule.PreviousEnd.Sub(
			schedule.PreviousStart,
		) != 2*time.Hour {
		t.Fatalf(
			"unexpected period durations: %#v",
			schedule,
		)
	}
	if !schedule.PreviousEnd.Equal(
		schedule.CurrentStart,
	) {
		t.Fatalf(
			"periods are not adjacent: %#v",
			schedule,
		)
	}
	if !schedule.CurrentEnd.Equal(
		expectedBoundary,
	) {
		t.Fatalf(
			"current end = %s, want %s",
			schedule.CurrentEnd,
			expectedBoundary,
		)
	}
}

func TestBuildEvidenceScheduleRejectsZeroTime(
	t *testing.T,
) {
	if _, err := buildEvidenceSchedule(
		time.Time{},
	); err == nil {
		t.Fatal(
			"expected zero verification time to be rejected",
		)
	}
}

func TestBuildEvidenceFixtureIsDeterministicAndNonZero(
	t *testing.T,
) {
	schedule, err := buildEvidenceSchedule(
		time.Date(
			2026,
			time.July,
			15,
			12,
			30,
			0,
			0,
			time.UTC,
		),
	)
	if err != nil {
		t.Fatalf(
			"build schedule: %v",
			err,
		)
	}

	fixture, err := buildEvidenceFixture(
		schedule,
	)
	if err != nil {
		t.Fatalf(
			"build fixture: %v",
			err,
		)
	}

	if len(fixture.Flights) != 7 ||
		len(fixture.FlightIDs) != 7 ||
		len(fixture.TrajectoryIDs) != 7 ||
		len(fixture.RouteRecordIDs) != 7 ||
		len(fixture.Observations) != 15 ||
		len(fixture.ObservationIDs) != 15 {
		t.Fatalf(
			"unexpected fixture sizes: %#v",
			fixture,
		)
	}

	uuidPattern := regexp.MustCompile(
		`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
	)
	routeRecordPattern := regexp.MustCompile(
		`^route-record-[0-9a-f]{64}$`,
	)
	fingerprintPattern := regexp.MustCompile(
		`^sha256:[0-9a-f]{64}$`,
	)

	for index, flight := range fixture.Flights {
		if !uuidPattern.MatchString(
			flight.FlightID,
		) ||
			!uuidPattern.MatchString(
				flight.TrajectoryID,
			) {
			t.Fatalf(
				"fixture UUID[%d] is invalid: %#v",
				index,
				flight,
			)
		}
		if !routeRecordPattern.MatchString(
			flight.RouteRecordID,
		) ||
			!fingerprintPattern.MatchString(
				flight.RouteFingerprint,
			) {
			t.Fatalf(
				"route identity[%d] is invalid: %#v",
				index,
				flight,
			)
		}

		var routeResult routecontract.Result
		if err := json.Unmarshal(
			flight.RouteJSON,
			&routeResult,
		); err != nil {
			t.Fatalf(
				"decode route fixture[%d]: %v",
				index,
				err,
			)
		}
		if routeResult.SchemaVersion !=
			routecontract.SchemaVersionV1 ||
			routeResult.Status !=
				routecontract.RouteStatusComplete ||
			routeResult.Origin == nil ||
			routeResult.Destination == nil ||
			routeResult.Origin.Airport.ICAOCode !=
				fixtureOriginICAO ||
			routeResult.Destination.Airport.ICAOCode !=
				fixtureDestinationICAO ||
			routeResult.Confidence.Score != 0.90 {
			t.Fatalf(
				"unexpected route fixture[%d]: %#v",
				index,
				routeResult,
			)
		}
	}
}

func TestEvidenceMetricExpectationsCoverAllSourceFamilies(
	t *testing.T,
) {
	expectations :=
		evidenceMetricExpectations()
	if len(expectations) != 5 {
		t.Fatalf(
			"expectation count = %d, want 5",
			len(expectations),
		)
	}

	seen := make(
		map[historicalcontract.MetricName]bool,
	)
	for _, expectation := range expectations {
		seen[expectation.Name] = true
		if len(expectation.CurrentPoints) != 2 ||
			len(expectation.PreviousPoints) != 2 ||
			expectation.CurrentTotal <=
				expectation.PreviousTotal {
			t.Fatalf(
				"invalid expectation: %#v",
				expectation,
			)
		}
	}

	required := []historicalcontract.MetricName{
		historicalcontract.MetricNameFlightCount,
		historicalcontract.MetricNameTrajectoryCount,
		historicalcontract.MetricNameObservationCount,
		historicalcontract.MetricNameAirportDepartures,
		historicalcontract.MetricNameRouteObservations,
	}
	for _, metricName := range required {
		if !seen[metricName] {
			t.Fatalf(
				"metric %s is not covered",
				metricName,
			)
		}
	}
}

func TestRuntimeVerificationBoundsAndMigrationIdentityArePinned(
	t *testing.T,
) {
	if expectedMigrationVersion != "015" ||
		expectedMigrationName !=
			"create_historical_aggregate_results" ||
		len(expectedMigrationChecksum) != 64 {
		t.Fatal(
			"migration 015 identity is not pinned",
		)
	}
	if evidenceDatasetLimit < 15 ||
		evidenceMaximumBucketCount < 4 ||
		evidenceMaximumWindowCount < 2 {
		t.Fatal(
			"runtime verification bounds are too small",
		)
	}
}
