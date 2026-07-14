package main

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routestore"
	"github.com/joho/godotenv"
)

const verificationSource = "postgres-route-store-verification"

func main() {
	os.Exit(run(os.Stdout, os.Stderr))
}

func run(
	stdout *os.File,
	stderr *os.File,
) int {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("apps/api/.env")

	cfg, err := config.LoadMigrationConfig()
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: load database configuration: %v\n",
			err,
		)
		return 1
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		cfg.MigrationTimeout,
	)
	defer cancel()

	pool, err := database.NewPostgresPool(
		cfg.Database.URL,
		cfg.Database.ConnectTimeout,
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: connect postgres: %v\n",
			err,
		)
		return 1
	}
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: begin verification transaction: %v\n",
			err,
		)
		return 1
	}

	transactionOpen := true
	defer func() {
		if transactionOpen {
			_ = tx.Rollback(context.Background())
		}
	}()

	now := time.Now().UTC()
	startTime := now.Add(-2 * time.Minute)
	endTime := now.Add(-time.Minute)
	asOfTime := endTime
	identityKey := "flight-identity-" +
		strings.Repeat("d", 64)

	var trajectoryID string
	if err := tx.QueryRow(
		ctx,
		`
			INSERT INTO flight_trajectories (
				icao24,
				callsign,
				start_time,
				end_time,
				duration_seconds,
				segment_count,
				point_count,
				coverage_gap_count,
				quality_score,
				source_name,
				identity_key,
				identity_basis,
				split_reason
			)
			VALUES (
				'ABC123',
				'VERIFY2',
				$1,
				$2,
				60,
				1,
				2,
				0,
				1,
				$3,
				$4,
				'source_flight_id',
				'initial_observation'
			)
			RETURNING id::text;
		`,
		startTime,
		endTime,
		verificationSource,
		identityKey,
	).Scan(&trajectoryID); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: create transactional verification trajectory: %v\n",
			err,
		)
		return 1
	}

	store, err := routestore.NewPostgresWithExecutor(
		tx,
		func() time.Time {
			return now
		},
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: compose PostgreSQL Route Store: %v\n",
			err,
		)
		return 1
	}

	result := verificationResult(
		trajectoryID,
		identityKey,
		startTime,
		endTime,
		asOfTime,
		now,
	)

	record, err := store.Put(ctx, result)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: put verification route result: %v\n",
			err,
		)
		return 1
	}

	replayed, err := store.Put(ctx, result)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: replay verification route result: %v\n",
			err,
		)
		return 1
	}
	if !reflect.DeepEqual(record, replayed) {
		fmt.Fprintln(
			stderr,
			"ERROR: idempotent route result replay returned a different record",
		)
		return 1
	}

	loaded, err := store.Get(ctx, record.Key)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: get verification route result: %v\n",
			err,
		)
		return 1
	}
	if !reflect.DeepEqual(record, loaded) {
		fmt.Fprintln(
			stderr,
			"ERROR: loaded route result differs from stored record",
		)
		return 1
	}

	latest, err := store.GetLatest(
		ctx,
		trajectoryID,
		routecontract.SchemaVersionV1,
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: get latest verification route result: %v\n",
			err,
		)
		return 1
	}
	if latest.ID != record.ID {
		fmt.Fprintln(
			stderr,
			"ERROR: latest route result does not match stored record",
		)
		return 1
	}

	page, err := store.List(
		ctx,
		routestore.ListQuery{
			TrajectoryID:  trajectoryID,
			SchemaVersion: routecontract.SchemaVersionV1,
			Limit:         1,
		},
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: list verification route results: %v\n",
			err,
		)
		return 1
	}
	if len(page.Records) != 1 ||
		page.HasMore ||
		page.Records[0].ID != record.ID {
		fmt.Fprintf(
			stderr,
			"ERROR: unexpected route result history page: %#v\n",
			page,
		)
		return 1
	}

	var resultCount int
	if err := tx.QueryRow(
		ctx,
		`
			SELECT count(*)
			FROM flight_route_results
			WHERE trajectory_id = $1::uuid;
		`,
		trajectoryID,
	).Scan(&resultCount); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: count transactional verification route results: %v\n",
			err,
		)
		return 1
	}
	if resultCount != 1 {
		fmt.Fprintf(
			stderr,
			"ERROR: transactional route result count = %d, want 1\n",
			resultCount,
		)
		return 1
	}

	if err := tx.Rollback(ctx); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: rollback verification transaction: %v\n",
			err,
		)
		return 1
	}
	transactionOpen = false

	var routeResultPersisted bool
	if err := pool.QueryRow(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM flight_route_results
				WHERE id = $1
			);
		`,
		record.ID,
	).Scan(&routeResultPersisted); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: verify route result rollback: %v\n",
			err,
		)
		return 1
	}
	if routeResultPersisted {
		fmt.Fprintln(
			stderr,
			"ERROR: verification route result remained after rollback",
		)
		return 1
	}

	var trajectoryPersisted bool
	if err := pool.QueryRow(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM flight_trajectories
				WHERE id = $1::uuid
			);
		`,
		trajectoryID,
	).Scan(&trajectoryPersisted); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: verify trajectory rollback: %v\n",
			err,
		)
		return 1
	}
	if trajectoryPersisted {
		fmt.Fprintln(
			stderr,
			"ERROR: verification trajectory remained after rollback",
		)
		return 1
	}

	fmt.Fprintln(
		stdout,
		"PostgreSQL Route Store Verification",
	)
	fmt.Fprintf(
		stdout,
		"Store: %s\n",
		routestore.PostgresVersion,
	)
	fmt.Fprintf(
		stdout,
		"Schema: %s\n",
		routecontract.SchemaVersionV1,
	)
	fmt.Fprintf(
		stdout,
		"Route status: %s\n",
		record.Result.Status,
	)
	fmt.Fprintf(
		stdout,
		"Record identifier: %s\n",
		record.ID,
	)
	fmt.Fprintln(stdout, "Put: PASS")
	fmt.Fprintln(
		stdout,
		"Idempotent replay: PASS",
	)
	fmt.Fprintln(stdout, "Get: PASS")
	fmt.Fprintln(stdout, "GetLatest: PASS")
	fmt.Fprintln(stdout, "List: PASS")
	fmt.Fprintln(
		stdout,
		"Transaction rollback: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Persistent verification rows: 0",
	)
	fmt.Fprintln(stdout, "Result: PASS")

	return 0
}

func verificationResult(
	trajectoryID string,
	identityKey string,
	startTime time.Time,
	endTime time.Time,
	asOfTime time.Time,
	generatedAt time.Time,
) routecontract.Result {
	origin := verificationEndpoint(
		routecontract.EndpointRoleOrigin,
		"UBBB",
		"GYD",
		40.4675,
		50.0467,
		startTime,
		0.90,
	)
	destination := verificationEndpoint(
		routecontract.EndpointRoleDestination,
		"UGTB",
		"TBS",
		41.6692,
		44.9547,
		endTime,
		0.85,
	)

	return routecontract.Result{
		SchemaVersion: routecontract.SchemaVersionV1,
		Status:        routecontract.RouteStatusComplete,
		TrajectoryID:  trajectoryID,
		IdentityKey:   identityKey,
		FlightID:      "verification-flight",
		AircraftID:    "verification-aircraft",
		ICAO24:        "ABC123",
		Callsign:      "VERIFY2",
		Window: routecontract.RouteWindow{
			StartTime: startTime,
			EndTime:   endTime,
			AsOfTime:  asOfTime,
		},
		Origin:      origin,
		Destination: destination,
		Summary: routecontract.RouteSummary{
			GreatCircleDistanceKM: 448.8,
			SameAirport:           false,
		},
		Confidence: routecontract.Confidence{
			Score: 0.85,
			Level: routecontract.
				ConfidenceLevelHigh,
			EvidenceCount: 2,
			Reasons: []routecontract.ConfidenceReason{
				{
					Code:         "both_route_endpoints_available",
					Message:      "Both route endpoints are supported by selected endpoint evidence.",
					Contribution: 0.85,
				},
			},
		},
		Limitations: []routecontract.Limitation{
			{
				Code:    "probable_route_only",
				Message: "Route endpoints are inferred and are not filed flight-plan data.",
				Scope:   "route",
			},
		},
		Provenance: routecontract.Provenance{
			ResolverVersion: "route-resolver-v1",
			InputFingerprint: "sha256:" +
				strings.Repeat("e", 64),
			TrajectoryUpdatedAt: endTime,
			SourceNames: []string{
				"ourairports",
				"trajectory",
			},
		},
		GeneratedAt: generatedAt,
	}
}

func verificationEndpoint(
	role routecontract.EndpointRole,
	icaoCode string,
	iataCode string,
	latitude float64,
	longitude float64,
	observedAt time.Time,
	score float64,
) *routecontract.EndpointInference {
	return &routecontract.EndpointInference{
		Role: role,
		Airport: routecontract.AirportReference{
			ICAOCode:   icaoCode,
			IATACode:   iataCode,
			Name:       icaoCode + " Airport",
			City:       "City",
			Country:    "Country",
			Latitude:   latitude,
			Longitude:  longitude,
			ElevationM: 10,
			Timezone:   "UTC",
		},
		DistanceKM: 2,
		Confidence: routecontract.Confidence{
			Score: score,
			Level: routecontract.
				ConfidenceLevelForScore(
					score,
				),
			EvidenceCount: 1,
			Reasons: []routecontract.ConfidenceReason{
				{
					Code: string(role) +
						"_airport_proximity",
					Message:      "Endpoint proximity evidence.",
					Contribution: score,
				},
			},
		},
		Evidence: []routecontract.Evidence{
			{
				Type: routecontract.
					EvidenceTypeTrajectoryEndpointProximity,
				SourceName:    "trajectory_endpoint",
				SourceVersion: "airport-candidate-resolver-v1",
				Score:         score,
				Weight:        1,
				ObservedAt:    observedAt,
				Summary:       "Trajectory endpoint supports the selected airport.",
				Attributes: []routecontract.EvidenceAttribute{
					{
						Key:   "distance_km",
						Value: "2.000000",
					},
					{
						Key:   "rank",
						Value: "1",
					},
				},
			},
		},
		Limitations: []routecontract.Limitation{
			{
				Code:    "probable_endpoint_only",
				Message: "Endpoint is inferred.",
				Scope:   string(role),
			},
		},
	}
}
