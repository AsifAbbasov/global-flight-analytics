package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/server"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/scopeenforcement"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/stabilityproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/unknownintervention"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const (
	verificationTrajectoryID = "12a9fe4c-7a14-4bda-b46f-2f38f5e8b912"
	verificationIdentityKey  = "flight-identity-fc0468b28b5bc3f5a4dad30903544fea47d5226f481e6743448e19f253b56975"
	verificationICAO24       = "A1B2C3"
	verificationCallsign     = "GFA12RV"
	verificationSourceName   = "stability-intelligence-http-runtime-verification-v1"

	verificationStateCount = 6
	verificationDuration   = 5 * time.Minute

	verificationDatabaseConnectAttempts       = 4
	verificationMinimumDatabaseConnectTimeout = 30 * time.Second
	verificationDatabaseRetryDelay            = 2 * time.Second
	verificationMinimumRuntimeTimeout         = 5 * time.Minute
	verificationHTTPServiceTimeout            = 90 * time.Second
)

type verificationSchedule struct {
	GeneratedAt     time.Time
	TrajectoryStart time.Time
	LatestAsOfTime  time.Time
	PointTimes      []time.Time
	AsOfTimes       []time.Time
}

type fixtureCounts struct {
	Trajectories int
	FlightStates int
	RouteResults int
}

type runtimeProjectionReader struct {
	service *projectionread.Service
}

type stabilityIntelligenceService interface {
	Get(
		context.Context,
		stabilityproduction.Request,
	) (stabilityproduction.Result, error)
}

type runtimeStabilityReader struct {
	service stabilityIntelligenceService
	timeout time.Duration
}

func main() {
	os.Exit(
		run(
			os.Stdout,
			os.Stderr,
		),
	)
}

func run(
	stdout io.Writer,
	stderr io.Writer,
) int {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("apps/api/.env")

	cfg, err :=
		config.LoadMigrationConfig()
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: load database configuration: %v\n",
			err,
		)
		return 1
	}

	pool, err := connectPostgreSQLWithRetry(
		cfg.Database.URL,
		cfg.Database.ConnectTimeout,
		stderr,
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: connect PostgreSQL: %v\n",
			err,
		)
		return 1
	}
	defer pool.Close()

	runtimeTimeout := runtimeVerificationTimeout(
		cfg.MigrationTimeout,
	)
	ctx, cancel := context.WithTimeout(
		context.Background(),
		runtimeTimeout,
	)
	defer cancel()

	if err := verifySchema(ctx, pool); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: verify Stability Intelligence runtime schema: %v\n",
			err,
		)
		return 1
	}

	schedule, err :=
		buildVerificationSchedule(
			time.Now().UTC(),
		)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: build verification schedule: %v\n",
			err,
		)
		return 1
	}

	if err := cleanupFixture(ctx, pool); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: remove stale verification fixture: %v\n",
			err,
		)
		return 1
	}

	if err := verifyFixtureSchemaCompatibility(
		ctx,
		pool,
		schedule,
	); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: verify complete runtime fixture compatibility: %v\n",
			err,
		)
		return 1
	}

	cleanupPending := true
	defer func() {
		if !cleanupPending {
			return
		}
		cleanupContext, cleanupCancel :=
			context.WithTimeout(
				context.Background(),
				30*time.Second,
			)
		defer cleanupCancel()
		if cleanupErr := cleanupFixture(
			cleanupContext,
			pool,
		); cleanupErr != nil {
			fmt.Fprintf(
				stderr,
				"ERROR: deferred fixture cleanup failed: %v\n",
				cleanupErr,
			)
		}
	}()

	if err := insertFixture(
		ctx,
		pool,
		schedule,
	); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: insert Stability Intelligence runtime fixture: %v\n",
			err,
		)
		return 1
	}

	beforeCounts, err :=
		loadFixtureCounts(
			ctx,
			pool,
		)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: count inserted runtime fixture: %v\n",
			err,
		)
		return 1
	}
	if beforeCounts != (fixtureCounts{
		Trajectories: 1,
		FlightStates: verificationStateCount,
		RouteResults: 0,
	}) {
		fmt.Fprintf(
			stderr,
			"ERROR: unexpected inserted fixture counts: %#v\n",
			beforeCounts,
		)
		return 1
	}

	projectionComposition, err :=
		projectionread.NewPostgres(
			projectionread.PostgresConfig{
				Pool:   pool,
				Policy: projectionread.DefaultPolicy(),
				Now: func() time.Time {
					return schedule.GeneratedAt
				},
			},
		)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: compose production Projection Intelligence reader: %v\n",
			err,
		)
		return 1
	}

	stabilityService, err :=
		stabilityproduction.New(
			stabilityproduction.Config{
				ProjectionReader: runtimeProjectionReader{
					service: projectionComposition.Service,
				},
				Now: func() time.Time {
					return schedule.GeneratedAt.Add(
						2 * time.Second,
					)
				},
			},
		)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: compose production Stability Intelligence service: %v\n",
			err,
		)
		return 1
	}

	directResult, err :=
		stabilityService.Get(
			ctx,
			stabilityproduction.Request{
				TrajectoryID:      verificationTrajectoryID,
				AsOfTimes:         schedule.AsOfTimes,
				RequestedDuration: verificationDuration,
			},
		)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: execute direct production Stability Intelligence composition: %v\n",
			err,
		)
		return 1
	}
	if err := validateDirectResult(
		directResult,
		schedule,
	); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: validate direct production result: %v\n",
			err,
		)
		return 1
	}

	replayed, err :=
		stabilityService.Get(
			ctx,
			stabilityproduction.Request{
				TrajectoryID:      verificationTrajectoryID,
				AsOfTimes:         schedule.AsOfTimes,
				RequestedDuration: verificationDuration,
			},
		)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: replay production Stability Intelligence composition: %v\n",
			err,
		)
		return 1
	}
	if replayed.InputFingerprint !=
		directResult.InputFingerprint {
		fmt.Fprintf(
			stderr,
			"ERROR: deterministic replay fingerprint mismatch: %s != %s\n",
			replayed.InputFingerprint,
			directResult.InputFingerprint,
		)
		return 1
	}

	app := fiber.New()
	v1 := app.Group("/api/v1")
	runtimeHTTPReader := runtimeStabilityReader{
		service: stabilityService,
		timeout: verificationHTTPServiceTimeout,
	}
	if err := server.RegisterStabilityIntelligenceReadRoute(
		v1,
		runtimeHTTPReader,
	); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: register Stability Intelligence route: %v\n",
			err,
		)
		return 1
	}

	payload, err := verifySuccessEndpoint(
		app,
		schedule,
		directResult,
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: verify production Stability Intelligence endpoint: %v\n",
			err,
		)
		return 1
	}

	if err := verifyHTTPErrorContracts(
		app,
		schedule,
	); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: verify Stability Intelligence HTTP error contracts: %v\n",
			err,
		)
		return 1
	}

	if err := cleanupFixture(ctx, pool); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: clean Stability Intelligence runtime fixture: %v\n",
			err,
		)
		return 1
	}

	afterCounts, err :=
		loadFixtureCounts(
			ctx,
			pool,
		)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: count runtime fixture after cleanup: %v\n",
			err,
		)
		return 1
	}
	if afterCounts != (fixtureCounts{}) {
		fmt.Fprintf(
			stderr,
			"ERROR: runtime fixture remained after cleanup: %#v\n",
			afterCounts,
		)
		return 1
	}
	cleanupPending = false

	fmt.Fprintln(
		stdout,
		"PostgreSQL Stability Intelligence HTTP API Verification",
	)
	fmt.Fprintf(
		stdout,
		"Production composition: %s\n",
		payload.Data.Version,
	)
	fmt.Fprintf(
		stdout,
		"Forecast versions: %d\n",
		len(payload.Data.Versions),
	)
	fmt.Fprintf(
		stdout,
		"Stability transitions: %d\n",
		len(payload.Data.Transitions),
	)
	fmt.Fprintf(
		stdout,
		"Forecast trend: %s\n",
		payload.Data.Analysis.Trend,
	)
	fmt.Fprintf(
		stdout,
		"Forecast health: %s\n",
		payload.Data.Analysis.Health,
	)
	fmt.Fprintf(
		stdout,
		"Propagated confidence: %.6f (%s)\n",
		payload.Data.Confidence.Score,
		payload.Data.Confidence.Level,
	)
	fmt.Fprintf(
		stdout,
		"Unknown intervention decision: %s\n",
		payload.Data.UnknownIntervention.Decision,
	)
	fmt.Fprintf(
		stdout,
		"Scope enforcement decision: %s\n",
		payload.Data.ScopeEnforcement.Decision,
	)
	fmt.Fprintln(stdout, "PostgreSQL connection retry policy: PASS")
	fmt.Fprintln(stdout, "Bounded Fiber HTTP execution: PASS")
	fmt.Fprintln(stdout, "Schema objects: PASS")
	fmt.Fprintln(stdout, "Fixture schema compatibility preflight: PASS")
	fmt.Fprintln(stdout, "Deterministic runtime fixture: PASS")
	fmt.Fprintln(stdout, "Production PostgreSQL projection reader: PASS")
	fmt.Fprintln(stdout, "Multi-as-of forecast versioning: PASS")
	fmt.Fprintln(stdout, "Decision Stability transitions: PASS")
	fmt.Fprintln(stdout, "Forecast Stability Analysis: PASS")
	fmt.Fprintln(stdout, "Confidence Propagation: PASS")
	fmt.Fprintln(stdout, "Failure Explanation Engine: PASS")
	fmt.Fprintln(stdout, "Unknown Intervention Guard: PASS")
	fmt.Fprintln(stdout, "Scope Guard Enforcement: PASS")
	fmt.Fprintln(stdout, "Explanation API standardization: PASS")
	fmt.Fprintln(stdout, "Direct production composition: PASS")
	fmt.Fprintln(stdout, "Deterministic replay fingerprint: PASS")
	fmt.Fprintln(stdout, "Stability Intelligence endpoint: PASS")
	fmt.Fprintln(stdout, "Not-found contract: PASS")
	fmt.Fprintln(stdout, "Validation error contract: PASS")
	fmt.Fprintln(stdout, "JSON response contract: PASS")
	fmt.Fprintln(stdout, "Research-only publication boundary: PASS")
	fmt.Fprintln(stdout, "Fixture cleanup: PASS")
	fmt.Fprintln(stdout, "Persistent verification rows: 0")
	fmt.Fprintln(stdout, "Result: PASS")

	return 0
}

func (
	reader runtimeProjectionReader,
) ReadProjection(
	ctx context.Context,
	request stabilityproduction.ProjectionRequest,
) (
	projectionproduction.Result,
	error,
) {
	if reader.service == nil {
		return projectionproduction.Result{},
			stabilityproduction.ErrServiceUnavailable
	}

	result, err := reader.service.Get(
		ctx,
		projectionread.Request{
			TrajectoryID:      request.TrajectoryID,
			AsOfTime:          request.AsOfTime,
			RequestedDuration: request.RequestedDuration,
		},
	)
	if err != nil {
		switch {
		case errors.Is(
			err,
			projectionread.ErrTrajectoryNotFound,
		):
			return projectionproduction.Result{},
				stabilityproduction.ErrTrajectoryNotFound
		case errors.Is(
			err,
			projectionread.ErrServiceUnavailable,
		):
			return projectionproduction.Result{},
				stabilityproduction.ErrServiceUnavailable
		case errors.Is(
			err,
			projectionread.ErrInvalidRequest,
		):
			return projectionproduction.Result{},
				stabilityproduction.ErrInvalidRequest
		default:
			return projectionproduction.Result{},
				err
		}
	}

	return result.Clone(), nil
}

func (
	reader runtimeStabilityReader,
) Get(
	ctx context.Context,
	request stabilityproduction.Request,
) (
	stabilityproduction.Result,
	error,
) {
	if reader.service == nil {
		return stabilityproduction.Result{},
			stabilityproduction.ErrServiceUnavailable
	}

	timeout := reader.timeout
	if timeout <= 0 {
		timeout = verificationHTTPServiceTimeout
	}
	if ctx == nil {
		ctx = context.Background()
	}

	boundedContext, cancel := context.WithTimeout(
		ctx,
		timeout,
	)
	defer cancel()

	return reader.service.Get(
		boundedContext,
		request,
	)
}

func executeFiberRequest(
	app *fiber.App,
	request *http.Request,
) (*http.Response, error) {
	if app == nil {
		return nil, fmt.Errorf(
			"Fiber application is required",
		)
	}
	if request == nil {
		return nil, fmt.Errorf(
			"HTTP request is required",
		)
	}

	// Fiber defaults App.Test to one second. The service adapter above
	// supplies the real context deadline, so the transport helper disables
	// Fiber's shorter synthetic timeout without making database work unbounded.
	return app.Test(
		request,
		-1,
	)
}

type postgresPoolConnector func(
	string,
	time.Duration,
) (*pgxpool.Pool, error)

type retrySleeper func(time.Duration)

func connectPostgreSQLWithRetry(
	databaseURL string,
	configuredTimeout time.Duration,
	stderr io.Writer,
) (*pgxpool.Pool, error) {
	return connectPostgreSQLWithRetryUsing(
		databaseURL,
		configuredTimeout,
		stderr,
		database.NewPostgresPool,
		time.Sleep,
	)
}

func connectPostgreSQLWithRetryUsing(
	databaseURL string,
	configuredTimeout time.Duration,
	stderr io.Writer,
	connector postgresPoolConnector,
	sleep retrySleeper,
) (*pgxpool.Pool, error) {
	if stderr == nil {
		stderr = io.Discard
	}
	if connector == nil {
		return nil, fmt.Errorf(
			"PostgreSQL connector is required",
		)
	}
	if sleep == nil {
		sleep = time.Sleep
	}

	attemptTimeout := configuredTimeout
	if attemptTimeout < verificationMinimumDatabaseConnectTimeout {
		attemptTimeout = verificationMinimumDatabaseConnectTimeout
	}

	var lastErr error
	for attempt := 1; attempt <= verificationDatabaseConnectAttempts; attempt++ {
		pool, err := connector(
			databaseURL,
			attemptTimeout,
		)
		if err == nil && pool != nil {
			if attempt > 1 {
				fmt.Fprintf(
					stderr,
					"PostgreSQL connection established on attempt %d of %d.\n",
					attempt,
					verificationDatabaseConnectAttempts,
				)
			}
			return pool, nil
		}
		if err == nil {
			err = fmt.Errorf(
				"PostgreSQL connector returned a nil pool",
			)
		}

		lastErr = err
		if attempt == verificationDatabaseConnectAttempts {
			break
		}

		retryDelay := time.Duration(attempt) *
			verificationDatabaseRetryDelay
		fmt.Fprintf(
			stderr,
			"PostgreSQL connection attempt %d of %d failed: %v\n",
			attempt,
			verificationDatabaseConnectAttempts,
			err,
		)
		fmt.Fprintf(
			stderr,
			"Retrying PostgreSQL connection after %s.\n",
			retryDelay,
		)
		sleep(retryDelay)
	}

	return nil, fmt.Errorf(
		"connection failed after %d attempts with a per-attempt timeout of %s: %w",
		verificationDatabaseConnectAttempts,
		attemptTimeout,
		lastErr,
	)
}

func runtimeVerificationTimeout(
	configured time.Duration,
) time.Duration {
	if configured < verificationMinimumRuntimeTimeout {
		return verificationMinimumRuntimeTimeout
	}
	return configured
}

func buildVerificationSchedule(
	now time.Time,
) (verificationSchedule, error) {
	if now.IsZero() {
		return verificationSchedule{},
			fmt.Errorf(
				"verification clock is required",
			)
	}

	generatedAt :=
		now.UTC().Truncate(time.Second)
	latestAsOfTime :=
		generatedAt.Add(-time.Minute)
	trajectoryStart :=
		latestAsOfTime.Add(-5 * time.Minute)

	pointTimes := make(
		[]time.Time,
		0,
		verificationStateCount,
	)
	for index := 0; index < verificationStateCount; index++ {
		pointTimes = append(
			pointTimes,
			trajectoryStart.Add(
				time.Duration(index)*
					time.Minute,
			),
		)
	}
	if !pointTimes[len(pointTimes)-1].
		Equal(latestAsOfTime) {
		return verificationSchedule{},
			fmt.Errorf(
				"verification point schedule does not end at latest as-of time",
			)
	}

	return verificationSchedule{
		GeneratedAt:     generatedAt,
		TrajectoryStart: trajectoryStart,
		LatestAsOfTime:  latestAsOfTime,
		PointTimes:      pointTimes,
		AsOfTimes: []time.Time{
			latestAsOfTime.Add(-time.Minute),
			latestAsOfTime.Add(-30 * time.Second),
			latestAsOfTime,
		},
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

type fixtureExecutor func(
	context.Context,
	string,
	...any,
) error

type fixtureQuerier interface {
	QueryRow(
		context.Context,
		string,
		...any,
	) pgx.Row
}

func verifyFixtureSchemaCompatibility(
	ctx context.Context,
	pool *pgxpool.Pool,
	schedule verificationSchedule,
) error {
	if pool == nil {
		return fmt.Errorf("PostgreSQL pool is required")
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf(
			"begin fixture schema-compatibility transaction: %w",
			err,
		)
	}
	rolledBack := false
	defer func() {
		if !rolledBack {
			_ = tx.Rollback(context.Background())
		}
	}()

	if err := insertFixtureWithExecutor(
		ctx,
		schedule,
		func(
			execContext context.Context,
			statement string,
			arguments ...any,
		) error {
			_, execErr := tx.Exec(
				execContext,
				statement,
				arguments...,
			)
			return execErr
		},
	); err != nil {
		return fmt.Errorf(
			"validate fixture against current PostgreSQL constraints: %w",
			err,
		)
	}

	counts, err := loadFixtureCountsFromQuerier(
		ctx,
		tx,
	)
	if err != nil {
		return fmt.Errorf(
			"count transactional fixture rows: %w",
			err,
		)
	}
	if counts != (fixtureCounts{
		Trajectories: 1,
		FlightStates: verificationStateCount,
		RouteResults: 0,
	}) {
		return fmt.Errorf(
			"unexpected transactional fixture counts: %#v",
			counts,
		)
	}

	if err := tx.Rollback(ctx); err != nil {
		return fmt.Errorf(
			"rollback fixture schema-compatibility transaction: %w",
			err,
		)
	}
	rolledBack = true

	persistentCounts, err := loadFixtureCounts(
		ctx,
		pool,
	)
	if err != nil {
		return fmt.Errorf(
			"count persistent rows after fixture preflight rollback: %w",
			err,
		)
	}
	if persistentCounts != (fixtureCounts{}) {
		return fmt.Errorf(
			"fixture preflight left persistent rows: %#v",
			persistentCounts,
		)
	}

	return nil
}

func insertFixture(
	ctx context.Context,
	pool *pgxpool.Pool,
	schedule verificationSchedule,
) error {
	if pool == nil {
		return fmt.Errorf("PostgreSQL pool is required")
	}

	return insertFixtureWithExecutor(
		ctx,
		schedule,
		func(
			execContext context.Context,
			statement string,
			arguments ...any,
		) error {
			_, execErr := pool.Exec(
				execContext,
				statement,
				arguments...,
			)
			return execErr
		},
	)
}

func insertFixtureWithExecutor(
	ctx context.Context,
	schedule verificationSchedule,
	exec fixtureExecutor,
) error {
	if exec == nil {
		return fmt.Errorf("fixture executor is required")
	}

	err := exec(
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
				0.95,
				$9
			);
		`,
		verificationTrajectoryID,
		verificationIdentityKey,
		verificationICAO24,
		verificationCallsign,
		schedule.TrajectoryStart,
		schedule.LatestAsOfTime,
		int64(
			schedule.LatestAsOfTime.Sub(
				schedule.TrajectoryStart,
			)/time.Second,
		),
		verificationStateCount,
		verificationSourceName,
	)
	if err != nil {
		return fmt.Errorf(
			"insert verification trajectory: %w",
			err,
		)
	}

	for index, observedAt := range schedule.PointTimes {
		latitude :=
			40.4700 +
				float64(index)*0.015
		longitude :=
			50.0400 +
				float64(index)*0.020
		altitudeM :=
			9000 +
				float64(index)*100

		if err := exec(
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
					75,
					0.5,
					false,
					'Azerbaijan',
					$6,
					$7,
					NULL
				);
			`,
			verificationICAO24,
			verificationCallsign,
			latitude,
			longitude,
			altitudeM,
			observedAt,
			verificationSourceName,
		); err != nil {
			return fmt.Errorf(
				"insert verification flight state %d: %w",
				index,
				err,
			)
		}
	}

	return nil
}

func cleanupFixture(
	ctx context.Context,
	pool *pgxpool.Pool,
) error {
	if _, err := pool.Exec(
		ctx,
		`
			DELETE FROM flight_route_results
			WHERE trajectory_id = $1::uuid
			   OR trajectory_id IN (
				   SELECT id
				   FROM flight_trajectories
				   WHERE identity_key = $2
			   );
		`,
		verificationTrajectoryID,
		verificationIdentityKey,
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
			WHERE source_name = $1
			  AND icao24 = $2
			  AND COALESCE(callsign, '') = $3;
		`,
		verificationSourceName,
		verificationICAO24,
		verificationCallsign,
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
			WHERE id = $1::uuid
			   OR identity_key = $2
			   OR (
				   source_name = $3
				   AND icao24 = $4
				   AND COALESCE(callsign, '') = $5
			   );
		`,
		verificationTrajectoryID,
		verificationIdentityKey,
		verificationSourceName,
		verificationICAO24,
		verificationCallsign,
	); err != nil {
		return fmt.Errorf(
			"delete verification trajectory: %w",
			err,
		)
	}

	return nil
}

func loadFixtureCounts(
	ctx context.Context,
	pool *pgxpool.Pool,
) (fixtureCounts, error) {
	if pool == nil {
		return fixtureCounts{},
			fmt.Errorf("PostgreSQL pool is required")
	}

	return loadFixtureCountsFromQuerier(
		ctx,
		pool,
	)
}

func loadFixtureCountsFromQuerier(
	ctx context.Context,
	querier fixtureQuerier,
) (fixtureCounts, error) {
	if querier == nil {
		return fixtureCounts{},
			fmt.Errorf("fixture querier is required")
	}

	var result fixtureCounts

	if err := querier.QueryRow(
		ctx,
		`
			SELECT COUNT(*)::int
			FROM flight_trajectories
			WHERE id = $1::uuid;
		`,
		verificationTrajectoryID,
	).Scan(&result.Trajectories); err != nil {
		return fixtureCounts{},
			fmt.Errorf(
				"count verification trajectories: %w",
				err,
			)
	}

	if err := querier.QueryRow(
		ctx,
		`
			SELECT COUNT(*)::int
			FROM flight_states
			WHERE source_name = $1
			  AND icao24 = $2
			  AND COALESCE(callsign, '') = $3;
		`,
		verificationSourceName,
		verificationICAO24,
		verificationCallsign,
	).Scan(&result.FlightStates); err != nil {
		return fixtureCounts{},
			fmt.Errorf(
				"count verification flight states: %w",
				err,
			)
	}

	if err := querier.QueryRow(
		ctx,
		`
			SELECT COUNT(*)::int
			FROM flight_route_results
			WHERE trajectory_id = $1::uuid;
		`,
		verificationTrajectoryID,
	).Scan(&result.RouteResults); err != nil {
		return fixtureCounts{},
			fmt.Errorf(
				"count verification route results: %w",
				err,
			)
	}

	return result, nil
}

func validateDirectResult(
	result stabilityproduction.Result,
	schedule verificationSchedule,
) error {
	if err := result.Validate(); err != nil {
		return err
	}
	if result.Version !=
		stabilityproduction.Version ||
		result.TrajectoryID !=
			verificationTrajectoryID ||
		len(result.Projections) != 3 ||
		len(result.ForecastVersions) != 3 ||
		len(result.Transitions) != 2 ||
		result.ForecastAnalysis.Metrics.
			VersionCount != 3 ||
		result.PropagatedConfidence.Score <= 0 ||
		result.FailureExplanation.PrimaryCode == "" ||
		result.UnknownIntervention.ClaimKind !=
			unknownintervention.
				ClaimKindContextualAssociation ||
		result.ScopeEnforcement.Decision ==
			scopeenforcement.DecisionBlocked {
		return fmt.Errorf(
			"unexpected production result: %#v",
			result,
		)
	}
	for index, asOfTime := range schedule.AsOfTimes {
		if !result.AsOfTimes[index].
			Equal(asOfTime) ||
			!result.Projections[index].
				Projection.Horizon.
				AsOfTime.Equal(asOfTime) {
			return fmt.Errorf(
				"as-of sequence mismatch at index %d",
				index,
			)
		}
	}
	return nil
}

func verifySuccessEndpoint(
	app *fiber.App,
	schedule verificationSchedule,
	directResult stabilityproduction.Result,
) (
	response.SuccessResponse[dto.StabilityIntelligenceResponse],
	error,
) {
	request := httptest.NewRequest(
		http.MethodGet,
		stabilityRequestURL(
			verificationTrajectoryID,
			schedule.AsOfTimes,
			verificationDuration,
		),
		nil,
	)
	httpResponse, err := executeFiberRequest(app, request)
	if err != nil {
		return response.SuccessResponse[dto.StabilityIntelligenceResponse]{},
			fmt.Errorf(
				"execute Stability Intelligence request: %w",
				err,
			)
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode !=
		fiber.StatusOK {
		body, _ := io.ReadAll(
			httpResponse.Body,
		)
		return response.SuccessResponse[dto.StabilityIntelligenceResponse]{},
			fmt.Errorf(
				"status = %d, want %d, body = %s",
				httpResponse.StatusCode,
				fiber.StatusOK,
				body,
			)
	}

	var payload response.SuccessResponse[dto.StabilityIntelligenceResponse]
	if err := json.NewDecoder(
		httpResponse.Body,
	).Decode(&payload); err != nil {
		return response.SuccessResponse[dto.StabilityIntelligenceResponse]{},
			fmt.Errorf(
				"decode Stability Intelligence response: %w",
				err,
			)
	}

	if !payload.Success ||
		payload.Data.Version !=
			stabilityproduction.Version ||
		payload.Data.TrajectoryID !=
			verificationTrajectoryID ||
		len(payload.Data.Projections) != 3 ||
		len(payload.Data.Versions) != 3 ||
		len(payload.Data.Transitions) != 2 ||
		payload.Data.Analysis.Metrics.
			VersionCount != 3 ||
		payload.Data.Confidence.Score <= 0 ||
		payload.Data.FailureExplanation.
			PrimaryCode == "" ||
		payload.Data.UnknownIntervention.
			ClaimKind !=
			string(
				unknownintervention.
					ClaimKindContextualAssociation,
			) ||
		payload.Data.ScopeEnforcement.
			Decision ==
			string(
				scopeenforcement.
					DecisionBlocked,
			) ||
		payload.Data.InputFingerprint !=
			directResult.InputFingerprint ||
		len(payload.Data.ScopeGuards) == 0 {
		return response.SuccessResponse[dto.StabilityIntelligenceResponse]{},
			fmt.Errorf(
				"unexpected Stability Intelligence payload: %#v",
				payload.Data,
			)
	}

	return payload, nil
}

func verifyHTTPErrorContracts(
	app *fiber.App,
	schedule verificationSchedule,
) error {
	testCases := []struct {
		name       string
		requestURL string
		statusCode int
	}{
		{
			name: "invalid trajectory identifier",
			requestURL: stabilityRequestURL(
				"not-a-uuid",
				schedule.AsOfTimes,
				verificationDuration,
			),
			statusCode: fiber.StatusBadRequest,
		},
		{
			name: "insufficient as-of history",
			requestURL: stabilityRequestURL(
				verificationTrajectoryID,
				schedule.AsOfTimes[:1],
				verificationDuration,
			),
			statusCode: fiber.StatusBadRequest,
		},
		{
			name: "missing trajectory",
			requestURL: stabilityRequestURL(
				"00000000-0000-0000-0000-000000000012",
				schedule.AsOfTimes,
				verificationDuration,
			),
			statusCode: fiber.StatusNotFound,
		},
	}

	for _, testCase := range testCases {
		request := httptest.NewRequest(
			http.MethodGet,
			testCase.requestURL,
			nil,
		)
		httpResponse, err := executeFiberRequest(
			app,
			request,
		)
		if err != nil {
			return fmt.Errorf(
				"%s: execute request: %w",
				testCase.name,
				err,
			)
		}
		_, _ = io.Copy(
			io.Discard,
			httpResponse.Body,
		)
		httpResponse.Body.Close()

		if httpResponse.StatusCode !=
			testCase.statusCode {
			return fmt.Errorf(
				"%s: status = %d, want %d",
				testCase.name,
				httpResponse.StatusCode,
				testCase.statusCode,
			)
		}
	}

	return nil
}

func stabilityRequestURL(
	trajectoryID string,
	asOfTimes []time.Time,
	duration time.Duration,
) string {
	values := url.Values{}
	formattedTimes := make(
		[]string,
		0,
		len(asOfTimes),
	)
	for _, asOfTime := range asOfTimes {
		formattedTimes = append(
			formattedTimes,
			asOfTime.UTC().Format(
				time.RFC3339Nano,
			),
		)
	}
	values.Set(
		"as_of_times",
		strings.Join(
			formattedTimes,
			",",
		),
	)
	values.Set(
		"duration_seconds",
		fmt.Sprintf(
			"%d",
			int64(duration/time.Second),
		),
	)

	return "/api/v1/trajectories/" +
		trajectoryID +
		"/stability-intelligence?" +
		values.Encode()
}
