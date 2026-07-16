package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	verificationSourceName = "projection-intelligence-historical-http-runtime-verification-v1"
	verificationDuration   = 3 * time.Minute
	currentPointCount      = 6
	candidatePointCount    = 9
	routeRecordIDPrefix    = "route-record-"

	minimumVerificationCommandTimeout = 5 * time.Minute
	historicalReadTimeout             = 60 * time.Second
	historicalHTTPTestTimeout         = historicalReadTimeout + 5*time.Second
	fixtureCleanupTimeout             = 60 * time.Second
)

var verificationFlights = []verificationFlight{
	{
		TrajectoryID: "a1111111-1111-4111-8111-111111111111",
		ICAO24:       "B1C001",
		Callsign:     "GFAH00",
		AgeDays:      0,
		PointCount:   currentPointCount,
	},
	{
		TrajectoryID:   "b1111111-1111-4111-8111-111111111111",
		ICAO24:         "B1C101",
		Callsign:       "GFAH01",
		AgeDays:        1,
		PointCount:     candidatePointCount,
		LatitudeShift:  0.0005,
		LongitudeShift: -0.0005,
	},
	{
		TrajectoryID:   "b2222222-2222-4222-8222-222222222222",
		ICAO24:         "B1C102",
		Callsign:       "GFAH02",
		AgeDays:        2,
		PointCount:     candidatePointCount,
		LatitudeShift:  -0.0005,
		LongitudeShift: 0.0005,
	},
	{
		TrajectoryID:   "b3333333-3333-4333-8333-333333333333",
		ICAO24:         "B1C103",
		Callsign:       "GFAH03",
		AgeDays:        3,
		PointCount:     candidatePointCount,
		LatitudeShift:  0.0010,
		LongitudeShift: 0.0010,
	},
	{
		TrajectoryID:   "b4444444-4444-4444-8444-444444444444",
		ICAO24:         "B1C104",
		Callsign:       "GFAH04",
		AgeDays:        4,
		PointCount:     candidatePointCount,
		LatitudeShift:  -0.0010,
		LongitudeShift: -0.0010,
	},
	{
		TrajectoryID:   "b5555555-5555-4555-8555-555555555555",
		ICAO24:         "B1C105",
		Callsign:       "GFAH05",
		AgeDays:        5,
		PointCount:     candidatePointCount,
		LatitudeShift:  0.0008,
		LongitudeShift: -0.0008,
	},
}

type verificationFlight struct {
	TrajectoryID   string
	ICAO24         string
	Callsign       string
	AgeDays        int
	PointCount     int
	LatitudeShift  float64
	LongitudeShift float64
}

type verificationSchedule struct {
	GeneratedAt time.Time
	AsOfTime    time.Time
}

type fixtureCounts struct {
	Trajectories int
	FlightStates int
	RouteResults int
}

func buildVerificationSchedule(
	now time.Time,
) (verificationSchedule, error) {
	if now.IsZero() {
		return verificationSchedule{},
			fmt.Errorf("verification clock is required")
	}

	generatedAt := now.UTC().Truncate(time.Second)
	return verificationSchedule{
		GeneratedAt: generatedAt,
		AsOfTime: generatedAt.Add(
			-time.Minute,
		),
	}, nil
}

func verifySchema(
	ctx context.Context,
	pool *pgxpool.Pool,
) error {
	for _, tableName := range []string{
		"flight_trajectories",
		"flight_states",
		"flight_route_results",
	} {
		var exists bool
		if err := pool.QueryRow(
			ctx,
			`SELECT to_regclass($1) IS NOT NULL;`,
			"public."+tableName,
		).Scan(&exists); err != nil {
			return fmt.Errorf(
				"query table %s: %w",
				tableName,
				err,
			)
		}
		if !exists {
			return fmt.Errorf(
				"required table %s is absent",
				tableName,
			)
		}
	}

	return nil
}

func insertFixture(
	ctx context.Context,
	pool *pgxpool.Pool,
	schedule verificationSchedule,
) error {
	for _, flight := range verificationFlights {
		endTime := schedule.AsOfTime.Add(
			-time.Duration(flight.AgeDays) *
				24 * time.Hour,
		)
		startTime := endTime.Add(
			-time.Duration(flight.PointCount-1) *
				time.Minute,
		)

		if err := insertTrajectory(
			ctx,
			pool,
			flight,
			startTime,
			endTime,
		); err != nil {
			return err
		}
		if err := insertFlightStates(
			ctx,
			pool,
			flight,
			startTime,
		); err != nil {
			return err
		}

		route := buildCompleteRoute(
			flight,
			startTime,
			endTime,
			schedule.GeneratedAt,
		)
		report := routecontract.Validate(route)
		if report.Status !=
			routecontract.ValidationStatusValid {
			return fmt.Errorf(
				"route fixture for %s is invalid: %#v",
				flight.TrajectoryID,
				report.Issues,
			)
		}
		if err := insertRouteResult(
			ctx,
			pool,
			flight,
			route,
			schedule.GeneratedAt,
		); err != nil {
			return err
		}
	}

	return nil
}

func insertTrajectory(
	ctx context.Context,
	pool *pgxpool.Pool,
	flight verificationFlight,
	startTime time.Time,
	endTime time.Time,
) error {
	_, err := pool.Exec(
		ctx,
		`
			INSERT INTO flight_trajectories (
				id,
				identity_key,
				identity_basis,
				split_reason,
				flight_id,
				aircraft_id,
				icao24,
				callsign,
				start_time,
				end_time,
				duration_seconds,
				segment_count,
				point_count,
				coverage_gap_count,
				quality_score,
				source_name
			)
			VALUES (
				$1::uuid,
				$2,
				'callsign_and_start_time',
				'initial_observation',
				NULL,
				NULL,
				$3,
				$4,
				$5,
				$6,
				$7,
				0,
				$8,
				0,
				0.98,
				$9
			);
		`,
		flight.TrajectoryID,
		identityKey(flight.TrajectoryID),
		flight.ICAO24,
		flight.Callsign,
		startTime.UTC(),
		endTime.UTC(),
		int64(endTime.Sub(startTime)/time.Second),
		flight.PointCount,
		verificationSourceName,
	)
	if err != nil {
		return fmt.Errorf(
			"insert trajectory %s: %w",
			flight.TrajectoryID,
			err,
		)
	}

	return nil
}

func insertFlightStates(
	ctx context.Context,
	pool *pgxpool.Pool,
	flight verificationFlight,
	startTime time.Time,
) error {
	for index := 0; index <
		flight.PointCount; index++ {
		latitude, longitude :=
			trackCoordinate(
				index,
				flight.LatitudeShift,
				flight.LongitudeShift,
			)
		observedAt := startTime.Add(
			time.Duration(index) *
				time.Minute,
		)
		altitudeM := 9000 +
			float64(index)*100

		_, err := pool.Exec(
			ctx,
			`
				INSERT INTO flight_states (
					flight_id,
					aircraft_id,
					icao24,
					callsign,
					latitude,
					longitude,
					barometric_altitude_m,
					barometric_altitude_status,
					geometric_altitude_m,
					geometric_altitude_status,
					velocity_mps,
					heading_degrees,
					vertical_rate_mps,
					on_ground,
					origin_country,
					observed_at,
					source_name,
					ingestion_run_id
				)
				VALUES (
					NULL,
					NULL,
					$1,
					$2,
					$3,
					$4,
					CAST($5::double precision AS integer),
					'observed',
					CAST(($5 + 100)::double precision AS integer),
					'observed',
					220,
					63,
					0.5,
					false,
					'Azerbaijan',
					$6,
					$7,
					NULL
				);
			`,
			flight.ICAO24,
			flight.Callsign,
			latitude,
			longitude,
			altitudeM,
			observedAt.UTC(),
			verificationSourceName,
		)
		if err != nil {
			return fmt.Errorf(
				"insert flight state %d for %s: %w",
				index,
				flight.TrajectoryID,
				err,
			)
		}
	}

	return nil
}

func insertRouteResult(
	ctx context.Context,
	pool *pgxpool.Pool,
	flight verificationFlight,
	route routecontract.Result,
	storedAt time.Time,
) error {
	payload, err := json.Marshal(route)
	if err != nil {
		return fmt.Errorf(
			"marshal route result for %s: %w",
			flight.TrajectoryID,
			err,
		)
	}

	_, err = pool.Exec(
		ctx,
		`
			INSERT INTO flight_route_results (
				id,
				trajectory_id,
				schema_version,
				as_of_time,
				as_of_time_unix_nano,
				input_fingerprint,
				route_status,
				confidence_level,
				validation_warning_count,
				route_json,
				stored_at,
				stored_at_unix_nano
			)
			VALUES (
				$1,
				$2::uuid,
				$3,
				$4,
				$5,
				$6,
				$7,
				$8,
				0,
				$9::jsonb,
				$10,
				$11
			);
		`,
		routeRecordID(route),
		flight.TrajectoryID,
		string(route.SchemaVersion),
		route.Window.AsOfTime.UTC(),
		route.Window.AsOfTime.UTC().UnixNano(),
		route.Provenance.InputFingerprint,
		string(route.Status),
		string(route.Confidence.Level),
		payload,
		storedAt.UTC(),
		storedAt.UTC().UnixNano(),
	)
	if err != nil {
		return fmt.Errorf(
			"insert route result for %s: %w",
			flight.TrajectoryID,
			err,
		)
	}

	return nil
}

func buildCompleteRoute(
	flight verificationFlight,
	startTime time.Time,
	endTime time.Time,
	generatedAt time.Time,
) routecontract.Result {
	originEvidence := routeEvidence(
		"synthetic origin evidence",
		endTime,
	)
	destinationEvidence := routeEvidence(
		"synthetic destination evidence",
		endTime,
	)

	return routecontract.Result{
		SchemaVersion: routecontract.SchemaVersionV1,
		Status:        routecontract.RouteStatusComplete,
		TrajectoryID:  flight.TrajectoryID,
		IdentityKey:   identityKey(flight.TrajectoryID),
		ICAO24:        flight.ICAO24,
		Callsign:      flight.Callsign,
		Window: routecontract.RouteWindow{
			StartTime: startTime.UTC(),
			EndTime:   endTime.UTC(),
			AsOfTime:  endTime.UTC(),
		},
		Origin: &routecontract.EndpointInference{
			Role: routecontract.EndpointRoleOrigin,
			Airport: routecontract.AirportReference{
				ICAOCode:  "ZAAA",
				IATACode:  "ZAA",
				Name:      "Synthetic Origin Airport",
				City:      "Synthetic Origin",
				Country:   "Test",
				Latitude:  40.4000,
				Longitude: 49.8000,
				Timezone:  "UTC",
			},
			DistanceKM: 0,
			Confidence: routeConfidence(
				0.95,
				1,
				"origin_fixture_confidence",
			),
			Evidence: []routecontract.Evidence{
				originEvidence,
			},
			Limitations: []routecontract.Limitation{},
		},
		Destination: &routecontract.EndpointInference{
			Role: routecontract.EndpointRoleDestination,
			Airport: routecontract.AirportReference{
				ICAOCode:  "ZBBB",
				IATACode:  "ZBB",
				Name:      "Synthetic Destination Airport",
				City:      "Synthetic Destination",
				Country:   "Test",
				Latitude:  40.5600,
				Longitude: 50.1200,
				Timezone:  "UTC",
			},
			DistanceKM: 0,
			Confidence: routeConfidence(
				0.95,
				1,
				"destination_fixture_confidence",
			),
			Evidence: []routecontract.Evidence{
				destinationEvidence,
			},
			Limitations: []routecontract.Limitation{},
		},
		Summary: routecontract.RouteSummary{
			GreatCircleDistanceKM: 32,
			SameAirport:           false,
		},
		Confidence: routeConfidence(
			0.95,
			2,
			"route_fixture_confidence",
		),
		Limitations: []routecontract.Limitation{},
		Provenance: routecontract.Provenance{
			ResolverVersion: "projection-historical-runtime-route-v1",
			InputFingerprint: fingerprint(
				"route",
				flight.TrajectoryID,
				endTime.UTC().
					Format(time.RFC3339Nano),
			),
			TrajectoryUpdatedAt: endTime.UTC(),
			SourceNames: []string{
				verificationSourceName,
			},
		},
		GeneratedAt: generatedAt.UTC(),
	}
}

func routeEvidence(
	summary string,
	observedAt time.Time,
) routecontract.Evidence {
	return routecontract.Evidence{
		Type: routecontract.
			EvidenceTypeTrajectoryEndpointProximity,
		SourceName:    verificationSourceName,
		SourceVersion: "projection-historical-runtime-route-v1",
		Score:         0.95,
		Weight:        1,
		ObservedAt:    observedAt.UTC(),
		Summary:       summary,
		Attributes:    []routecontract.EvidenceAttribute{},
	}
}

func routeConfidence(
	score float64,
	evidenceCount int,
	code string,
) routecontract.Confidence {
	return routecontract.Confidence{
		Score: score,
		Level: routecontract.ConfidenceLevelForScore(
			score,
		),
		EvidenceCount: evidenceCount,
		Reasons: []routecontract.ConfidenceReason{
			{
				Code:         code,
				Message:      "Synthetic runtime verification confidence.",
				Contribution: score,
			},
		},
	}
}

func trackCoordinate(
	index int,
	latitudeShift float64,
	longitudeShift float64,
) (float64, float64) {
	return 40.4000 +
			float64(index)*0.0200 +
			latitudeShift,
		49.8000 +
			float64(index)*0.0400 +
			longitudeShift
}

func cleanupFixture(
	ctx context.Context,
	pool *pgxpool.Pool,
) error {
	if _, err := pool.Exec(
		ctx,
		`
			DELETE FROM flight_route_results
			WHERE trajectory_id::text =
				ANY($1::text[]);
		`,
		verificationTrajectoryIDs(),
	); err != nil {
		return fmt.Errorf(
			"delete verification route results: %w",
			err,
		)
	}

	if _, err := pool.Exec(
		ctx,
		`
			DELETE FROM flight_states
			WHERE source_name = $1;
		`,
		verificationSourceName,
	); err != nil {
		return fmt.Errorf(
			"delete verification flight states: %w",
			err,
		)
	}

	if _, err := pool.Exec(
		ctx,
		`
			DELETE FROM flight_trajectories
			WHERE id::text =
				ANY($1::text[]);
		`,
		verificationTrajectoryIDs(),
	); err != nil {
		return fmt.Errorf(
			"delete verification trajectories: %w",
			err,
		)
	}

	return nil
}

func loadFixtureCounts(
	ctx context.Context,
	pool *pgxpool.Pool,
) (fixtureCounts, error) {
	var result fixtureCounts

	if err := pool.QueryRow(
		ctx,
		`
			SELECT COUNT(*)::int
			FROM flight_trajectories
			WHERE source_name = $1;
		`,
		verificationSourceName,
	).Scan(&result.Trajectories); err != nil {
		return fixtureCounts{},
			fmt.Errorf(
				"count verification trajectories: %w",
				err,
			)
	}

	if err := pool.QueryRow(
		ctx,
		`
			SELECT COUNT(*)::int
			FROM flight_states
			WHERE source_name = $1;
		`,
		verificationSourceName,
	).Scan(&result.FlightStates); err != nil {
		return fixtureCounts{},
			fmt.Errorf(
				"count verification flight states: %w",
				err,
			)
	}

	if err := pool.QueryRow(
		ctx,
		`
			SELECT COUNT(*)::int
			FROM flight_route_results
			WHERE trajectory_id::text =
				ANY($1::text[]);
		`,
		verificationTrajectoryIDs(),
	).Scan(&result.RouteResults); err != nil {
		return fixtureCounts{},
			fmt.Errorf(
				"count verification route results: %w",
				err,
			)
	}

	return result, nil
}

func verificationTrajectoryIDs() []string {
	result := make(
		[]string,
		0,
		len(verificationFlights),
	)
	for _, flight := range verificationFlights {
		result = append(
			result,
			flight.TrajectoryID,
		)
	}

	return result
}

func validateFixturePolicyCoverage(
	policy projectionread.Policy,
) error {
	if err := policy.Validate(); err != nil {
		return fmt.Errorf(
			"validate production policy: %w",
			err,
		)
	}

	candidateCount := len(verificationFlights) - 1
	if candidateCount <
		policy.Neighbors.SelectionLimit {
		return fmt.Errorf(
			"historical fixture candidate count %d is below the neighbor selection limit %d",
			candidateCount,
			policy.Neighbors.SelectionLimit,
		)
	}
	if candidateCount <
		policy.Pattern.TargetNeighborCount {
		return fmt.Errorf(
			"historical fixture candidate count %d is below the pattern target %d",
			candidateCount,
			policy.Pattern.TargetNeighborCount,
		)
	}
	if candidateCount <
		policy.Freshness.TargetRecentNeighborCount {
		return fmt.Errorf(
			"historical fixture candidate count %d is below the freshness support target %d",
			candidateCount,
			policy.Freshness.TargetRecentNeighborCount,
		)
	}
	if len(verificationFlights) <
		policy.RouteFrequency.MinimumObservationCount {
		return fmt.Errorf(
			"historical fixture route observations %d are below the route-frequency minimum %d",
			len(verificationFlights),
			policy.RouteFrequency.MinimumObservationCount,
		)
	}

	seenDays := make(map[int]struct{})
	for index, flight := range verificationFlights {
		if index == 0 {
			if flight.AgeDays != 0 {
				return fmt.Errorf(
					"current fixture trajectory must have zero age days",
				)
			}
			seenDays[flight.AgeDays] = struct{}{}
			continue
		}
		if flight.AgeDays <= 0 {
			return fmt.Errorf(
				"historical fixture trajectory %s must have positive age days",
				flight.TrajectoryID,
			)
		}

		age := time.Duration(flight.AgeDays) *
			24 * time.Hour
		if age >
			policy.Freshness.RecentNeighborAgeLimit {
			return fmt.Errorf(
				"historical fixture trajectory %s exceeds the recent-neighbor age limit",
				flight.TrajectoryID,
			)
		}
		if age >
			policy.Neighbors.MaximumCandidateAge {
			return fmt.Errorf(
				"historical fixture trajectory %s exceeds the candidate-age limit",
				flight.TrajectoryID,
			)
		}
		if _, exists := seenDays[flight.AgeDays]; exists {
			return fmt.Errorf(
				"historical fixture repeats age day %d",
				flight.AgeDays,
			)
		}
		seenDays[flight.AgeDays] = struct{}{}
	}

	if len(seenDays) <
		policy.RouteFrequency.MinimumDistinctDayCount {
		return fmt.Errorf(
			"historical fixture distinct days %d are below the route-frequency minimum %d",
			len(seenDays),
			policy.RouteFrequency.MinimumDistinctDayCount,
		)
	}

	return nil
}

func expectedFixtureCounts() fixtureCounts {
	totalStates := 0
	for _, flight := range verificationFlights {
		totalStates += flight.PointCount
	}

	return fixtureCounts{
		Trajectories: len(verificationFlights),
		FlightStates: totalStates,
		RouteResults: len(verificationFlights),
	}
}

func validateFixtureRouteRecordIDs(
	schedule verificationSchedule,
) error {
	seen := make(
		map[string]struct{},
		len(verificationFlights),
	)

	for _, flight := range verificationFlights {
		endTime := schedule.AsOfTime.Add(
			-time.Duration(flight.AgeDays) *
				24 * time.Hour,
		)
		startTime := endTime.Add(
			-time.Duration(flight.PointCount-1) *
				time.Minute,
		)
		route := buildCompleteRoute(
			flight,
			startTime,
			endTime,
			schedule.GeneratedAt,
		)
		recordID := routeRecordID(route)

		if !strings.HasPrefix(
			recordID,
			routeRecordIDPrefix,
		) ||
			len(recordID) !=
				len(routeRecordIDPrefix)+
					sha256.Size*2 {
			return fmt.Errorf(
				"route record identifier %q does not satisfy the production identifier contract",
				recordID,
			)
		}
		if _, exists := seen[recordID]; exists {
			return fmt.Errorf(
				"duplicate route record identifier %q",
				recordID,
			)
		}
		seen[recordID] = struct{}{}
	}

	return nil
}

func routeRecordID(
	route routecontract.Result,
) string {
	compositeKey := fmt.Sprintf(
		"%s\x00%s\x00%s",
		strings.TrimSpace(
			route.TrajectoryID,
		),
		route.SchemaVersion,
		route.Window.AsOfTime.UTC().
			Format(time.RFC3339Nano),
	)
	digest := sha256.Sum256(
		[]byte(
			compositeKey +
				"\x00" +
				strings.TrimSpace(
					route.Provenance.
						InputFingerprint,
				),
		),
	)

	return routeRecordIDPrefix +
		hex.EncodeToString(
			digest[:],
		)
}

func identityKey(
	trajectoryID string,
) string {
	return "flight-identity-" +
		strings.TrimPrefix(
			fingerprint(
				"identity",
				trajectoryID,
			),
			"sha256:",
		)
}

func fingerprint(
	parts ...string,
) string {
	digest := sha256.Sum256(
		[]byte(
			strings.Join(parts, "|"),
		),
	)

	return "sha256:" +
		hex.EncodeToString(
			digest[:],
		)
}
