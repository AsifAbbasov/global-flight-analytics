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
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionbaseline"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/server"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const (
	verificationTrajectoryID = "91a9fe4c-7a14-4bda-b46f-2f38f5e8b973"
	verificationIdentityKey  = "flight-identity-5c1a9f93eaf9962dd174c5a9b21b848cbf3b2e27ad0f50530d8df3bf0f3c713a"
	verificationICAO24       = "A1B2C3"
	verificationCallsign     = "GFA9RV"
	verificationSourceName   = "projection-intelligence-http-runtime-verification-v1"
	verificationDuration     = 3 * time.Minute
	verificationPointCount   = 6
)

type verificationSchedule struct {
	GeneratedAt     time.Time
	AsOfTime        time.Time
	TrajectoryStart time.Time
	PointTimes      []time.Time
}

type fixtureCounts struct {
	Trajectories int
	FlightStates int
	RouteResults int
}

type runtimeReader struct {
	service *projectionread.Service
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
			"ERROR: connect PostgreSQL: %v\n",
			err,
		)
		return 1
	}
	defer pool.Close()

	if err := verifySchema(
		ctx,
		pool,
	); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: verify Projection Intelligence runtime schema: %v\n",
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

	if err := cleanupFixture(
		ctx,
		pool,
	); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: remove stale verification fixture: %v\n",
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
			"ERROR: insert Projection Intelligence runtime fixture: %v\n",
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
	expectedBeforeCounts := fixtureCounts{
		Trajectories: 1,
		FlightStates: verificationPointCount,
		RouteResults: 0,
	}
	if beforeCounts != expectedBeforeCounts {
		fmt.Fprintf(
			stderr,
			"ERROR: unexpected inserted fixture counts: %#v\n",
			beforeCounts,
		)
		return 1
	}

	composition, err :=
		projectionread.NewPostgres(
			projectionread.PostgresConfig{
				Pool: pool,
				Policy: projectionread.
					DefaultPolicy(),
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

	app := fiber.New()
	v1 := app.Group("/api/v1")
	if err := server.RegisterProjectionIntelligenceReadRoute(
		v1,
		runtimeReader{
			service: composition.Service,
		},
	); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: register Projection Intelligence HTTP route: %v\n",
			err,
		)
		return 1
	}

	payload, err := verifySuccessEndpoint(
		app,
		schedule,
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: verify production Projection Intelligence endpoint: %v\n",
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
			"ERROR: verify Projection Intelligence HTTP error contracts: %v\n",
			err,
		)
		return 1
	}

	if err := cleanupFixture(
		ctx,
		pool,
	); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: clean Projection Intelligence runtime fixture: %v\n",
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
		"PostgreSQL Projection Intelligence HTTP API Verification",
	)
	fmt.Fprintf(
		stdout,
		"Production composition: %s\n",
		projectionproduction.Version,
	)
	fmt.Fprintf(
		stdout,
		"Projection method: %s\n",
		payload.Data.Projection.Method.Name,
	)
	fmt.Fprintf(
		stdout,
		"Forecast points: %d\n",
		len(payload.Data.Projection.Points),
	)
	fmt.Fprintln(
		stdout,
		"Schema objects: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Deterministic verification fixture: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Production PostgreSQL reader: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Observed flight-state hydration: PASS",
	)
	fmt.Fprintln(
		stdout,
		"As-of boundary: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Missing Route Intelligence fallback: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Kinematic projection endpoint: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Projection uncertainty contract: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Projection confidence contract: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Not-found contract: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Validation error contract: PASS",
	)
	fmt.Fprintln(
		stdout,
		"JSON response contract: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Fixture cleanup: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Persistent verification rows: 0",
	)
	fmt.Fprintln(
		stdout,
		"Result: PASS",
	)

	return 0
}

func (
	reader runtimeReader,
) GetProjectionIntelligence(
	ctx context.Context,
	request handlers.ProjectionIntelligenceReadRequest,
) (
	projectionproduction.Result,
	error,
) {
	if reader.service == nil {
		return projectionproduction.Result{},
			handlers.
				ErrProjectionIntelligenceServiceUnavailable
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
				handlers.
					ErrProjectionIntelligenceNotFound

		case errors.Is(
			err,
			projectionread.ErrServiceUnavailable,
		):
			return projectionproduction.Result{},
				handlers.
					ErrProjectionIntelligenceServiceUnavailable

		case errors.Is(
			err,
			projectionread.ErrInvalidRequest,
		):
			return projectionproduction.Result{},
				handlers.
					ErrProjectionIntelligenceInvalidRequest

		default:
			return projectionproduction.Result{},
				err
		}
	}

	return result.Clone(), nil
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
		now.UTC().Truncate(
			time.Second,
		)
	asOfTime :=
		generatedAt.Add(
			-time.Minute,
		)
	trajectoryStart :=
		asOfTime.Add(
			-5 * time.Minute,
		)

	pointTimes := make(
		[]time.Time,
		0,
		verificationPointCount,
	)
	for index := 0; index <
		verificationPointCount; index++ {
		pointTimes = append(
			pointTimes,
			trajectoryStart.Add(
				time.Duration(index)*
					time.Minute,
			),
		)
	}

	lastPointIndex := len(pointTimes) - 1
	if !pointTimes[lastPointIndex].Equal(
		asOfTime,
	) {
		return verificationSchedule{},
			fmt.Errorf(
				"verification point schedule does not end at the analytical time",
			)
	}

	return verificationSchedule{
		GeneratedAt:     generatedAt,
		AsOfTime:        asOfTime,
		TrajectoryStart: trajectoryStart,
		PointTimes:      pointTimes,
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
		).Scan(
			&exists,
		); err != nil {
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
				0.95,
				$9
			);
		`,
		verificationTrajectoryID,
		verificationIdentityKey,
		verificationICAO24,
		verificationCallsign,
		schedule.TrajectoryStart,
		schedule.AsOfTime,
		int64(
			schedule.AsOfTime.Sub(
				schedule.TrajectoryStart,
			)/time.Second,
		),
		verificationPointCount,
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

		if _, err := pool.Exec(
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
			WHERE trajectory_id = $1::uuid;
		`,
		verificationTrajectoryID,
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
			WHERE id = $1::uuid;
		`,
		verificationTrajectoryID,
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
	var result fixtureCounts

	if err := pool.QueryRow(
		ctx,
		`
			SELECT COUNT(*)::int
			FROM flight_trajectories
			WHERE id = $1::uuid;
		`,
		verificationTrajectoryID,
	).Scan(
		&result.Trajectories,
	); err != nil {
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
			WHERE source_name = $1
			  AND icao24 = $2
			  AND COALESCE(callsign, '') = $3;
		`,
		verificationSourceName,
		verificationICAO24,
		verificationCallsign,
	).Scan(
		&result.FlightStates,
	); err != nil {
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
			WHERE trajectory_id = $1::uuid;
		`,
		verificationTrajectoryID,
	).Scan(
		&result.RouteResults,
	); err != nil {
		return fixtureCounts{},
			fmt.Errorf(
				"count verification route results: %w",
				err,
			)
	}

	return result, nil
}

func verifySuccessEndpoint(
	app *fiber.App,
	schedule verificationSchedule,
) (
	response.SuccessResponse[dto.ProjectionIntelligenceResponse],
	error,
) {
	requestURL := projectionRequestURL(
		verificationTrajectoryID,
		schedule.AsOfTime,
		verificationDuration,
	)
	request :=
		httptest.NewRequest(
			http.MethodGet,
			requestURL,
			nil,
		)
	httpResponse, err :=
		app.Test(request)
	if err != nil {
		return response.SuccessResponse[dto.ProjectionIntelligenceResponse]{},
			fmt.Errorf(
				"execute Projection Intelligence request: %w",
				err,
			)
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode !=
		fiber.StatusOK {
		body, _ := io.ReadAll(
			httpResponse.Body,
		)
		return response.SuccessResponse[dto.ProjectionIntelligenceResponse]{},
			fmt.Errorf(
				"status = %d, want %d, body = %s",
				httpResponse.StatusCode,
				fiber.StatusOK,
				body,
			)
	}

	var payload response.SuccessResponse[dto.ProjectionIntelligenceResponse]
	if err := json.NewDecoder(
		httpResponse.Body,
	).Decode(&payload); err != nil {
		return response.SuccessResponse[dto.ProjectionIntelligenceResponse]{},
			fmt.Errorf(
				"decode Projection Intelligence response: %w",
				err,
			)
	}

	if err := validateSuccessPayload(
		payload,
		schedule,
	); err != nil {
		return response.SuccessResponse[dto.ProjectionIntelligenceResponse]{},
			err
	}

	return payload, nil
}

func validateSuccessPayload(
	payload response.SuccessResponse[dto.ProjectionIntelligenceResponse],
	schedule verificationSchedule,
) error {
	data := payload.Data

	if !payload.Success {
		return fmt.Errorf(
			"success response flag is false",
		)
	}
	if data.Version !=
		projectionproduction.Version {
		return fmt.Errorf(
			"production version = %q, want %q",
			data.Version,
			projectionproduction.Version,
		)
	}
	if data.Strategy !=
		string(
			projectionproduction.
				StrategyKinematic,
		) {
		return fmt.Errorf(
			"strategy = %q, want kinematic baseline",
			data.Strategy,
		)
	}
	if data.FallbackReason !=
		"historical_neighbors_unavailable" {
		return fmt.Errorf(
			"fallback reason = %q, want historical_neighbors_unavailable",
			data.FallbackReason,
		)
	}
	if data.ArrivalStatus !=
		string(
			projectionproduction.
				ArrivalStatusWithheld,
		) {
		return fmt.Errorf(
			"arrival status = %q, want withheld",
			data.ArrivalStatus,
		)
	}
	if data.Projection.TrajectoryID !=
		verificationTrajectoryID {
		return fmt.Errorf(
			"trajectory ID = %q, want %q",
			data.Projection.TrajectoryID,
			verificationTrajectoryID,
		)
	}
	if data.Projection.Method.Name !=
		projectionbaseline.MethodName {
		return fmt.Errorf(
			"projection method = %q, want %q",
			data.Projection.Method.Name,
			projectionbaseline.MethodName,
		)
	}
	if !data.Projection.Horizon.AsOfTime.Equal(
		schedule.AsOfTime,
	) ||
		data.Projection.Horizon.DurationSeconds !=
			int64(
				verificationDuration/
					time.Second,
			) {
		return fmt.Errorf(
			"unexpected projection horizon: %#v",
			data.Projection.Horizon,
		)
	}
	if len(data.Projection.Points) != 6 {
		return fmt.Errorf(
			"forecast point count = %d, want 6",
			len(data.Projection.Points),
		)
	}
	lastForecastIndex :=
		len(data.Projection.Points) - 1
	if !data.Projection.Points[0].
		ForecastTime.Equal(
		schedule.AsOfTime.Add(
			30*time.Second,
		),
	) ||
		!data.Projection.Points[lastForecastIndex].
			ForecastTime.Equal(
			schedule.AsOfTime.Add(
				verificationDuration,
			),
		) {
		return fmt.Errorf(
			"forecast timestamps do not cover the configured horizon",
		)
	}
	if data.Projection.Points[0].
		Uncertainty.HorizontalRadiusM <= 0 ||
		data.Projection.Points[lastForecastIndex].
			Uncertainty.HorizontalRadiusM <=
			data.Projection.Points[0].
				Uncertainty.HorizontalRadiusM {
		return fmt.Errorf(
			"horizontal uncertainty did not grow across the forecast horizon",
		)
	}
	if data.Projection.Confidence.Score <= 0 ||
		data.Projection.Confidence.Level ==
			"none" {
		return fmt.Errorf(
			"projection confidence is unavailable: %#v",
			data.Projection.Confidence,
		)
	}
	if data.Projection.ScopeGuard !=
		"research_only_not_for_operational_use" {
		return fmt.Errorf(
			"unexpected scope guard: %q",
			data.Projection.ScopeGuard,
		)
	}
	if data.Projection.Arrival != nil {
		return fmt.Errorf(
			"arrival estimate must be withheld without a complete route",
		)
	}
	if data.Evidence.NeighborSelection == nil ||
		data.Evidence.NeighborSelection.Status !=
			"unavailable" ||
		data.Evidence.PatternConfidence != nil ||
		data.Evidence.Freshness != nil ||
		data.Evidence.RouteFrequency != nil {
		return fmt.Errorf(
			"unexpected historical evidence contract: %#v",
			data.Evidence,
		)
	}
	if !hasNotice(
		data.Notices,
		"historical_neighbors_unavailable",
	) {
		return fmt.Errorf(
			"historical-neighbor fallback notice is missing",
		)
	}
	if !strings.HasPrefix(
		data.InputFingerprint,
		"sha256:",
	) ||
		!strings.HasPrefix(
			data.Projection.Provenance.
				InputFingerprint,
			"sha256:",
		) {
		return fmt.Errorf(
			"deterministic fingerprints are missing",
		)
	}

	return nil
}

func verifyHTTPErrorContracts(
	app *fiber.App,
	schedule verificationSchedule,
) error {
	missingTrajectoryID :=
		"2d736f02-3d37-49ae-bf7f-3c719537c04d"
	if err := expectError(
		app,
		projectionRequestURL(
			missingTrajectoryID,
			schedule.AsOfTime,
			verificationDuration,
		),
		fiber.StatusNotFound,
		"PROJECTION_INTELLIGENCE_NOT_FOUND",
	); err != nil {
		return err
	}

	invalidDurationURL :=
		projectionRequestURL(
			verificationTrajectoryID,
			schedule.AsOfTime,
			0,
		)
	if err := expectError(
		app,
		invalidDurationURL,
		fiber.StatusBadRequest,
		"INVALID_PROJECTION_DURATION",
	); err != nil {
		return err
	}

	return nil
}

func expectError(
	app *fiber.App,
	requestURL string,
	expectedStatus int,
	expectedCode string,
) error {
	request :=
		httptest.NewRequest(
			http.MethodGet,
			requestURL,
			nil,
		)
	httpResponse, err :=
		app.Test(request)
	if err != nil {
		return fmt.Errorf(
			"execute expected-error request: %w",
			err,
		)
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode !=
		expectedStatus {
		body, _ := io.ReadAll(
			httpResponse.Body,
		)
		return fmt.Errorf(
			"status = %d, want %d, body = %s",
			httpResponse.StatusCode,
			expectedStatus,
			body,
		)
	}

	var payload response.ErrorResponse
	if err := json.NewDecoder(
		httpResponse.Body,
	).Decode(&payload); err != nil {
		return fmt.Errorf(
			"decode expected-error response: %w",
			err,
		)
	}
	if payload.Success ||
		payload.Error.Code !=
			expectedCode {
		return fmt.Errorf(
			"unexpected error payload: %#v",
			payload,
		)
	}

	return nil
}

func projectionRequestURL(
	trajectoryID string,
	asOfTime time.Time,
	duration time.Duration,
) string {
	values := url.Values{}
	values.Set(
		"as_of_time",
		asOfTime.UTC().Format(
			time.RFC3339Nano,
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
		"/projection-intelligence?" +
		values.Encode()
}

func hasNotice(
	items []dto.ProjectionIntelligenceNotice,
	code string,
) bool {
	for _, item := range items {
		if item.Code == code {
			return true
		}
	}

	return false
}
