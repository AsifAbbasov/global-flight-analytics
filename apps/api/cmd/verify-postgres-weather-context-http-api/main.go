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
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/server"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathercontext"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const (
	verificationTrajectoryID      = "ae14b2ef-55e4-4ee1-8fb8-8480a5406825"
	verificationWeatherSnapshotID = "492be80a-c7da-4b24-9388-506fd819881b"
	futureWeatherSnapshotID       = "c4359c98-f31d-4768-a747-45bd838d9330"
	verificationIdentityKey       = "flight-identity-9f2e7d1288444a083f1df81391d4b940c0246dc9d3ac1957f22f58a5882c392f"
	verificationICAO24            = "B4C5D6"
	verificationCallsign          = "GFA10WX"
	verificationSourceName        = "weather-context-http-runtime-verification-v1"
	verificationDuration          = 3 * time.Minute
	boundedPointCount             = 6
	storedPointCount              = 7
	storedWeatherSnapshotCount    = 2
	verificationTemperature       = 18.5
	futureTemperature             = 99.0
	runtimeHTTPTestTimeout        = 60 * time.Second
	runtimeVerificationTimeout    = 2 * time.Minute
	fixtureCleanupTimeout         = 60 * time.Second
)

type verificationSchedule struct {
	GeneratedAt         time.Time
	AsOfTime            time.Time
	TrajectoryStart     time.Time
	TrajectoryEnd       time.Time
	PointTimes          []time.Time
	WeatherObservedAt   time.Time
	WeatherRetrievedAt  time.Time
	FutureWeatherAt     time.Time
	FutureWeatherReadAt time.Time
}

type fixtureCounts struct {
	Trajectories     int
	FlightStates     int
	WeatherSnapshots int
	RouteResults     int
}

type runtimeProjectionReader struct {
	service *projectionread.Service
}

func main() {
	os.Exit(run(os.Stdout, os.Stderr))
}

func run(stdout io.Writer, stderr io.Writer) int {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("apps/api/.env")

	cfg, err := config.LoadMigrationConfig()
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: load database configuration: %v\n", err)
		return 1
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		runtimeVerificationTimeout,
	)
	defer cancel()

	pool, err := database.NewPostgresPool(
		cfg.Database.URL,
		cfg.Database.ConnectTimeout,
	)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: connect PostgreSQL: %v\n", err)
		return 1
	}
	defer pool.Close()

	if err := verifySchema(ctx, pool); err != nil {
		fmt.Fprintf(stderr, "ERROR: verify Weather Context runtime schema: %v\n", err)
		return 1
	}

	schedule, err := buildVerificationSchedule(time.Now().UTC())
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: build verification schedule: %v\n", err)
		return 1
	}

	if err := cleanupFixture(ctx, pool); err != nil {
		fmt.Fprintf(stderr, "ERROR: remove stale verification fixture: %v\n", err)
		return 1
	}

	cleanupPending := true
	defer func() {
		if !cleanupPending {
			return
		}
		cleanupContext, cleanupCancel := context.WithTimeout(
			context.Background(),
			fixtureCleanupTimeout,
		)
		defer cleanupCancel()
		if cleanupErr := cleanupFixture(cleanupContext, pool); cleanupErr != nil {
			fmt.Fprintf(stderr, "ERROR: deferred fixture cleanup failed: %v\n", cleanupErr)
		}
	}()

	if err := insertFixture(ctx, pool, schedule); err != nil {
		fmt.Fprintf(stderr, "ERROR: insert Weather Context runtime fixture: %v\n", err)
		return 1
	}

	beforeCounts, err := loadFixtureCounts(ctx, pool)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: count inserted runtime fixture: %v\n", err)
		return 1
	}
	expectedBeforeCounts := fixtureCounts{
		Trajectories:     1,
		FlightStates:     storedPointCount,
		WeatherSnapshots: storedWeatherSnapshotCount,
		RouteResults:     0,
	}
	if beforeCounts != expectedBeforeCounts {
		fmt.Fprintf(stderr, "ERROR: unexpected inserted fixture counts: %#v\n", beforeCounts)
		return 1
	}

	projectionComposition, err := projectionread.NewPostgres(
		projectionread.PostgresConfig{
			Pool:   pool,
			Policy: projectionread.DefaultPolicy(),
			Now: func() time.Time {
				return schedule.GeneratedAt
			},
		},
	)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: compose production Projection Intelligence reader: %v\n", err)
		return 1
	}

	if err := verifyProductionDependencies(
		ctx,
		pool,
		projectionComposition.Service,
		schedule,
	); err != nil {
		fmt.Fprintf(stderr, "ERROR: verify production Weather Context dependencies: %v\n", err)
		return 1
	}

	weatherContextReader, err := server.NewWeatherContextPostgresReader(
		pool,
		runtimeProjectionReader{service: projectionComposition.Service},
	)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: compose production Weather Context reader: %v\n", err)
		return 1
	}

	if err := verifyDirectWeatherContext(
		ctx,
		weatherContextReader,
		schedule,
	); err != nil {
		fmt.Fprintf(stderr, "ERROR: verify direct production Weather Context composition: %v\n", err)
		return 1
	}

	app := fiber.New()
	v1 := app.Group("/api/v1")
	if err := server.RegisterWeatherContextReadRoute(v1, weatherContextReader); err != nil {
		fmt.Fprintf(stderr, "ERROR: register Weather Context HTTP route: %v\n", err)
		return 1
	}

	payload, err := verifySuccessEndpoint(app, schedule)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: verify production Weather Context endpoint: %v\n", err)
		return 1
	}

	if err := verifyHTTPErrorContracts(app, schedule); err != nil {
		fmt.Fprintf(stderr, "ERROR: verify Weather Context HTTP error contracts: %v\n", err)
		return 1
	}

	if err := cleanupFixture(ctx, pool); err != nil {
		fmt.Fprintf(stderr, "ERROR: clean Weather Context runtime fixture: %v\n", err)
		return 1
	}

	afterCounts, err := loadFixtureCounts(ctx, pool)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: count runtime fixture after cleanup: %v\n", err)
		return 1
	}
	if afterCounts != (fixtureCounts{}) {
		fmt.Fprintf(stderr, "ERROR: runtime fixture remained after cleanup: %#v\n", afterCounts)
		return 1
	}

	cleanupPending = false

	fmt.Fprintln(stdout, "PostgreSQL Weather Context HTTP API Verification")
	fmt.Fprintf(stdout, "Weather Context version: %s\n", payload.Data.Version)
	fmt.Fprintf(stdout, "Weather samples returned: %d\n", len(payload.Data.Weather.Samples))
	fmt.Fprintf(stdout, "Bounded trajectory points: %d of %d stored\n", payload.Data.Alignment.PointCount, storedPointCount)
	fmt.Fprintf(stdout, "Aligned weather points: %d\n", payload.Data.Alignment.AlignedCount)
	fmt.Fprintf(stdout, "Weather uncertainty status: %s\n", payload.Data.Uncertainty.Status)
	fmt.Fprintf(stdout, "Weather multiplier: %.6f\n", payload.Data.Uncertainty.WeatherMultiplier)
	fmt.Fprintln(stdout, "Schema objects: PASS")
	fmt.Fprintln(stdout, "Deterministic verification fixture: PASS")
	fmt.Fprintln(stdout, "Production PostgreSQL composition: PASS")
	fmt.Fprintln(stdout, "Direct production dependency verification: PASS")
	fmt.Fprintln(stdout, "Direct production Weather Context composition: PASS")
	fmt.Fprintln(stdout, "Production trajectory hydration: PASS")
	fmt.Fprintln(stdout, "Trajectory future-evidence boundary: PASS")
	fmt.Fprintln(stdout, "Weather future-evidence boundary: PASS")
	fmt.Fprintln(stdout, "Weather Feature Contract endpoint: PASS")
	fmt.Fprintln(stdout, "Weather Trust Gate endpoint: PASS")
	fmt.Fprintln(stdout, "Four-dimensional alignment endpoint: PASS")
	fmt.Fprintln(stdout, "Weather Encounter Profile endpoint: PASS")
	fmt.Fprintln(stdout, "Scope-limited weather uncertainty decision: PASS")
	fmt.Fprintln(stdout, "Projection preservation contract: PASS")
	fmt.Fprintln(stdout, "Not-found contract: PASS")
	fmt.Fprintln(stdout, "Validation error contract: PASS")
	fmt.Fprintln(stdout, "JSON response contract: PASS")
	fmt.Fprintln(stdout, "Fixture cleanup: PASS")
	fmt.Fprintln(stdout, "Persistent verification rows: 0")
	fmt.Fprintln(stdout, "Result: PASS")

	return 0
}

func (reader runtimeProjectionReader) GetProjectionIntelligence(
	ctx context.Context,
	request handlers.ProjectionIntelligenceReadRequest,
) (projectionproduction.Result, error) {
	if reader.service == nil {
		return projectionproduction.Result{}, handlers.ErrProjectionIntelligenceServiceUnavailable
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
		case errors.Is(err, projectionread.ErrTrajectoryNotFound):
			return projectionproduction.Result{}, handlers.ErrProjectionIntelligenceNotFound
		case errors.Is(err, projectionread.ErrServiceUnavailable):
			return projectionproduction.Result{}, handlers.ErrProjectionIntelligenceServiceUnavailable
		case errors.Is(err, projectionread.ErrInvalidRequest):
			return projectionproduction.Result{}, handlers.ErrProjectionIntelligenceInvalidRequest
		default:
			return projectionproduction.Result{}, err
		}
	}

	return result.Clone(), nil
}

func verifyProductionDependencies(
	ctx context.Context,
	pool *pgxpool.Pool,
	projectionService *projectionread.Service,
	schedule verificationSchedule,
) error {
	trajectorySource, err := projectionread.NewPostgresDataSource(
		projectionread.PostgresDataSourceConfig{
			Pool: pool,
			Policy: projectionread.
				DefaultPolicy().DataSource,
		},
	)
	if err != nil {
		return fmt.Errorf("compose production trajectory data source: %w", err)
	}

	loadedTrajectory, err := trajectorySource.LoadCurrentTrajectory(
		ctx,
		verificationTrajectoryID,
		schedule.AsOfTime,
	)
	if err != nil {
		return fmt.Errorf("load bounded production trajectory: %w", err)
	}
	if len(loadedTrajectory.Points) != boundedPointCount ||
		loadedTrajectory.PointCount != boundedPointCount {
		return fmt.Errorf(
			"bounded production trajectory points = %d/%d, want %d",
			len(loadedTrajectory.Points),
			loadedTrajectory.PointCount,
			boundedPointCount,
		)
	}
	for _, point := range loadedTrajectory.Points {
		if point.ObservedAt.After(schedule.AsOfTime) {
			return fmt.Errorf(
				"future trajectory point leaked through production data source: %s",
				point.ObservedAt,
			)
		}
	}

	latestLatitude, latestLongitude := verificationCoordinates(
		boundedPointCount - 1,
	)
	snapshotReader, err := weathercontext.NewPostgresSnapshotReader(
		pool,
		weathercontext.DefaultPostgresSnapshotPolicy(),
	)
	if err != nil {
		return fmt.Errorf("compose production weather snapshot reader: %w", err)
	}
	snapshot, err := snapshotReader.GetLatestSnapshot(
		ctx,
		weathercontext.WeatherSnapshotRequest{
			Latitude:  latestLatitude,
			Longitude: latestLongitude,
			AsOfTime:  schedule.AsOfTime,
		},
	)
	if err != nil {
		return fmt.Errorf("load bounded production weather snapshot: %w", err)
	}
	if snapshot.TemperatureCelsius != verificationTemperature ||
		snapshot.ObservedAt.After(schedule.AsOfTime) ||
		snapshot.RetrievedAt.After(schedule.AsOfTime) {
		return fmt.Errorf(
			"production weather snapshot crossed the analytical boundary: %#v",
			snapshot,
		)
	}

	if projectionService == nil {
		return fmt.Errorf("production projection service is nil")
	}
	projection, err := projectionService.Get(
		ctx,
		projectionread.Request{
			TrajectoryID:      verificationTrajectoryID,
			AsOfTime:          schedule.AsOfTime,
			RequestedDuration: verificationDuration,
		},
	)
	if err != nil {
		return fmt.Errorf("load bounded production projection: %w", err)
	}
	if err := projection.Validate(); err != nil {
		return fmt.Errorf("validate production projection: %w", err)
	}
	if projection.Projection.TrajectoryID != verificationTrajectoryID ||
		!projection.Projection.Horizon.AsOfTime.Equal(schedule.AsOfTime) ||
		len(projection.Projection.Points) == 0 {
		return fmt.Errorf("production projection identity or horizon is invalid")
	}

	return nil
}

func verifyDirectWeatherContext(
	ctx context.Context,
	reader handlers.WeatherContextReader,
	schedule verificationSchedule,
) error {
	if reader == nil {
		return fmt.Errorf("production Weather Context reader is nil")
	}
	result, err := reader.GetWeatherContext(
		ctx,
		handlers.WeatherContextReadRequest{
			TrajectoryID:      verificationTrajectoryID,
			AsOfTime:          schedule.AsOfTime,
			RequestedDuration: verificationDuration,
		},
	)
	if err != nil {
		return fmt.Errorf("load direct production Weather Context: %w", err)
	}
	if err := result.Validate(); err != nil {
		return fmt.Errorf("validate direct production Weather Context: %w", err)
	}
	if len(result.Weather.Samples) != 1 {
		return fmt.Errorf(
			"direct weather sample count = %d, want 1",
			len(result.Weather.Samples),
		)
	}
	if result.Weather.Samples[0].Features.TemperatureCelsius == nil ||
		*result.Weather.Samples[0].Features.TemperatureCelsius != verificationTemperature {
		return fmt.Errorf("direct Weather Context selected the wrong weather snapshot")
	}
	if result.Alignment.PointCount != boundedPointCount {
		return fmt.Errorf(
			"direct Weather Context point count = %d, want %d",
			result.Alignment.PointCount,
			boundedPointCount,
		)
	}
	if result.Uncertainty.Status != "withheld" ||
		result.Uncertainty.WeatherMultiplier != 1 ||
		len(result.Uncertainty.PointAdjustments) != 0 {
		return fmt.Errorf(
			"direct Weather Context violated the surface-weather uncertainty guard: %#v",
			result.Uncertainty,
		)
	}
	return nil
}

func buildVerificationSchedule(now time.Time) (verificationSchedule, error) {
	if now.IsZero() {
		return verificationSchedule{}, fmt.Errorf("verification clock is required")
	}

	generatedAt := now.UTC().Truncate(time.Second)
	asOfTime := generatedAt.Add(-time.Minute)
	trajectoryStart := asOfTime.Add(-5 * time.Minute)
	pointTimes := make([]time.Time, 0, storedPointCount)
	for index := 0; index < boundedPointCount; index++ {
		pointTimes = append(pointTimes, trajectoryStart.Add(time.Duration(index)*time.Minute))
	}
	pointTimes = append(pointTimes, asOfTime.Add(30*time.Second))

	if !pointTimes[boundedPointCount-1].Equal(asOfTime) {
		return verificationSchedule{}, fmt.Errorf("bounded point schedule does not end at the analytical time")
	}
	if !pointTimes[len(pointTimes)-1].After(asOfTime) {
		return verificationSchedule{}, fmt.Errorf("future trajectory point is not after the analytical time")
	}

	return verificationSchedule{
		GeneratedAt:         generatedAt,
		AsOfTime:            asOfTime,
		TrajectoryStart:     trajectoryStart,
		TrajectoryEnd:       asOfTime,
		PointTimes:          pointTimes,
		WeatherObservedAt:   asOfTime.Add(-30 * time.Second),
		WeatherRetrievedAt:  asOfTime.Add(-15 * time.Second),
		FutureWeatherAt:     asOfTime.Add(30 * time.Second),
		FutureWeatherReadAt: asOfTime.Add(45 * time.Second),
	}, nil
}

func verifySchema(ctx context.Context, pool *pgxpool.Pool) error {
	for _, tableName := range []string{
		"flight_trajectories",
		"flight_states",
		"flight_route_results",
		"weather_snapshots",
	} {
		var exists bool
		if err := pool.QueryRow(
			ctx,
			`SELECT to_regclass($1) IS NOT NULL;`,
			"public."+tableName,
		).Scan(&exists); err != nil {
			return fmt.Errorf("query table %s: %w", tableName, err)
		}
		if !exists {
			return fmt.Errorf("required table %s is absent", tableName)
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
		schedule.TrajectoryEnd,
		int64(schedule.TrajectoryEnd.Sub(schedule.TrajectoryStart)/time.Second),
		boundedPointCount,
		verificationSourceName,
	)
	if err != nil {
		return fmt.Errorf("insert verification trajectory: %w", err)
	}

	for index, observedAt := range schedule.PointTimes {
		latitude, longitude := verificationCoordinates(index)
		onGround := index < 2
		barometricAltitude := 0.0
		geometricAltitude := 0.0
		altitudeStatus := "ground"
		velocity := 0.0
		heading := 75.0
		verticalRate := 0.0
		if !onGround {
			barometricAltitude = 8500 + float64(index)*100
			geometricAltitude = barometricAltitude + 100
			altitudeStatus = "observed"
			velocity = 220
			verticalRate = 0.5
		}

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
					$6,
					CAST($7::double precision AS integer),
					$6,
					$8,
					$9,
					$10,
					$11,
					'Azerbaijan',
					$12,
					$13,
					NULL
				);
			`,
			verificationICAO24,
			verificationCallsign,
			latitude,
			longitude,
			barometricAltitude,
			altitudeStatus,
			geometricAltitude,
			velocity,
			heading,
			verticalRate,
			onGround,
			observedAt,
			verificationSourceName,
		); err != nil {
			return fmt.Errorf("insert verification flight state %d: %w", index, err)
		}
	}

	latestLatitude, latestLongitude := verificationCoordinates(boundedPointCount - 1)
	if err := insertWeatherSnapshot(
		ctx,
		pool,
		verificationWeatherSnapshotID,
		latestLatitude,
		latestLongitude,
		schedule.WeatherObservedAt,
		schedule.WeatherRetrievedAt,
		verificationTemperature,
	); err != nil {
		return err
	}
	if err := insertWeatherSnapshot(
		ctx,
		pool,
		futureWeatherSnapshotID,
		latestLatitude,
		latestLongitude,
		schedule.FutureWeatherAt,
		schedule.FutureWeatherReadAt,
		futureTemperature,
	); err != nil {
		return err
	}

	return nil
}

func insertWeatherSnapshot(
	ctx context.Context,
	pool *pgxpool.Pool,
	id string,
	latitude float64,
	longitude float64,
	observedAt time.Time,
	retrievedAt time.Time,
	temperature float64,
) error {
	_, err := pool.Exec(
		ctx,
		`
			INSERT INTO weather_snapshots (
				id,
				provider,
				latitude,
				longitude,
				observed_at,
				retrieved_at,
				temperature_celsius,
				relative_humidity_percent,
				precipitation_mm,
				rain_mm,
				weather_code,
				cloud_cover_percent,
				surface_pressure_hpa,
				wind_speed_mps,
				wind_direction_degrees,
				wind_gusts_mps,
				metadata_json
			)
			VALUES (
				$1::uuid,
				'open_meteo',
				$2,
				$3,
				$4,
				$5,
				$6,
				70,
				4,
				3,
				61,
				90,
				1005,
				20,
				270,
				30,
				jsonb_build_object('verification_source', $7::text)
			);
		`,
		id,
		latitude,
		longitude,
		observedAt,
		retrievedAt,
		temperature,
		verificationSourceName,
	)
	if err != nil {
		return fmt.Errorf("insert verification weather snapshot %s: %w", id, err)
	}
	return nil
}

func cleanupFixture(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(
		ctx,
		`DELETE FROM flight_route_results WHERE trajectory_id = $1::uuid;`,
		verificationTrajectoryID,
	); err != nil {
		return fmt.Errorf("delete verification route results: %w", err)
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
		return fmt.Errorf("delete verification flight states: %w", err)
	}

	if _, err := pool.Exec(
		ctx,
		`DELETE FROM flight_trajectories WHERE id = $1::uuid;`,
		verificationTrajectoryID,
	); err != nil {
		return fmt.Errorf("delete verification trajectory: %w", err)
	}

	if _, err := pool.Exec(
		ctx,
		`
			DELETE FROM weather_snapshots
			WHERE id IN ($1::uuid, $2::uuid)
			   OR metadata_json->>'verification_source' = $3;
		`,
		verificationWeatherSnapshotID,
		futureWeatherSnapshotID,
		verificationSourceName,
	); err != nil {
		return fmt.Errorf("delete verification weather snapshots: %w", err)
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
		`SELECT COUNT(*)::int FROM flight_trajectories WHERE id = $1::uuid;`,
		verificationTrajectoryID,
	).Scan(&result.Trajectories); err != nil {
		return fixtureCounts{}, fmt.Errorf("count verification trajectories: %w", err)
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
	).Scan(&result.FlightStates); err != nil {
		return fixtureCounts{}, fmt.Errorf("count verification flight states: %w", err)
	}

	if err := pool.QueryRow(
		ctx,
		`
			SELECT COUNT(*)::int
			FROM weather_snapshots
			WHERE id IN ($1::uuid, $2::uuid)
			   OR metadata_json->>'verification_source' = $3;
		`,
		verificationWeatherSnapshotID,
		futureWeatherSnapshotID,
		verificationSourceName,
	).Scan(&result.WeatherSnapshots); err != nil {
		return fixtureCounts{}, fmt.Errorf("count verification weather snapshots: %w", err)
	}

	if err := pool.QueryRow(
		ctx,
		`SELECT COUNT(*)::int FROM flight_route_results WHERE trajectory_id = $1::uuid;`,
		verificationTrajectoryID,
	).Scan(&result.RouteResults); err != nil {
		return fixtureCounts{}, fmt.Errorf("count verification route results: %w", err)
	}

	return result, nil
}

func verifySuccessEndpoint(
	app *fiber.App,
	schedule verificationSchedule,
) (response.SuccessResponse[dto.WeatherContextResponse], error) {
	request := httptest.NewRequest(
		http.MethodGet,
		weatherContextRequestURL(
			verificationTrajectoryID,
			schedule.AsOfTime,
			verificationDuration,
		),
		nil,
	)
	httpResponse, err := executeHTTPTest(app, request)
	if err != nil {
		return response.SuccessResponse[dto.WeatherContextResponse]{}, fmt.Errorf("execute Weather Context request: %w", err)
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != fiber.StatusOK {
		body, _ := io.ReadAll(httpResponse.Body)
		return response.SuccessResponse[dto.WeatherContextResponse]{}, fmt.Errorf(
			"status = %d, want %d, body = %s",
			httpResponse.StatusCode,
			fiber.StatusOK,
			body,
		)
	}

	var payload response.SuccessResponse[dto.WeatherContextResponse]
	if err := json.NewDecoder(httpResponse.Body).Decode(&payload); err != nil {
		return response.SuccessResponse[dto.WeatherContextResponse]{}, fmt.Errorf("decode Weather Context response: %w", err)
	}
	if err := validateSuccessPayload(payload, schedule); err != nil {
		return response.SuccessResponse[dto.WeatherContextResponse]{}, err
	}
	return payload, nil
}

func validateSuccessPayload(
	payload response.SuccessResponse[dto.WeatherContextResponse],
	schedule verificationSchedule,
) error {
	data := payload.Data
	if !payload.Success {
		return fmt.Errorf("success response flag is false")
	}
	if data.Version != dto.WeatherContextResponseVersion {
		return fmt.Errorf("Weather Context response version = %q, want %q", data.Version, dto.WeatherContextResponseVersion)
	}
	if data.TrajectoryID != verificationTrajectoryID {
		return fmt.Errorf("trajectory ID = %q, want %q", data.TrajectoryID, verificationTrajectoryID)
	}
	if !data.AsOfTime.Equal(schedule.AsOfTime) || data.GeneratedAt.Before(schedule.AsOfTime) {
		return fmt.Errorf("unexpected aggregate time boundary: as_of=%s generated_at=%s", data.AsOfTime, data.GeneratedAt)
	}
	if len(data.Weather.Samples) != 1 {
		return fmt.Errorf("weather sample count = %d, want 1", len(data.Weather.Samples))
	}

	sample := data.Weather.Samples[0]
	if sample.Features.TemperatureCelsius == nil ||
		*sample.Features.TemperatureCelsius != verificationTemperature {
		return fmt.Errorf("selected temperature = %v, want %.1f", sample.Features.TemperatureCelsius, verificationTemperature)
	}
	if *sample.Features.TemperatureCelsius == futureTemperature ||
		sample.ValidAt.After(schedule.AsOfTime) ||
		sample.RetrievedAt.After(schedule.AsOfTime) ||
		data.Weather.LatestAvailableAt.After(schedule.AsOfTime) {
		return fmt.Errorf("future weather evidence leaked into response: %#v", sample)
	}
	if sample.Source.Provider != "open_meteo" || sample.Features.PresentCount < 8 {
		return fmt.Errorf("weather feature contract is incomplete: %#v", sample)
	}

	if !data.Trust.Usable || data.Trust.Score <= 0 ||
		(data.Trust.Decision != "allowed" && data.Trust.Decision != "limited") {
		return fmt.Errorf("weather trust decision is unusable: %#v", data.Trust)
	}
	if !containsString(data.Trust.AllowedScopes, "surface_context") ||
		containsString(data.Trust.AllowedScopes, "projection_uncertainty") {
		return fmt.Errorf("unexpected surface-weather usage scopes: %#v", data.Trust.AllowedScopes)
	}

	if data.Alignment.PointCount != boundedPointCount {
		return fmt.Errorf("bounded point count = %d, want %d", data.Alignment.PointCount, boundedPointCount)
	}
	if data.Alignment.AlignedCount <= 0 || data.Alignment.AlignedCount >= data.Alignment.PointCount ||
		data.Alignment.Status != "limited" {
		return fmt.Errorf("expected partial weather alignment, got %#v", data.Alignment)
	}
	for _, match := range data.Alignment.Matches {
		if match.TrajectoryObservedAt.After(schedule.AsOfTime) {
			return fmt.Errorf("future trajectory point leaked into alignment: %#v", match)
		}
	}

	if data.Encounter.EncounterPointCount != data.Alignment.AlignedCount ||
		data.Encounter.EncounterPointCount <= 0 ||
		data.Encounter.Status != "limited" {
		return fmt.Errorf("unexpected Weather Encounter Profile: %#v", data.Encounter)
	}

	if data.Uncertainty.Status != "withheld" {
		return fmt.Errorf("surface weather uncertainty must be withheld, got %#v", data.Uncertainty)
	}
	if data.Uncertainty.WeatherMultiplier != 1 ||
		data.Uncertainty.SeverityScore != 0 ||
		len(data.Uncertainty.PointAdjustments) != 0 {
		return fmt.Errorf("withheld uncertainty modified projection evidence: %#v", data.Uncertainty)
	}
	if len(data.Uncertainty.AdjustedProjection.Points) == 0 ||
		data.Uncertainty.AdjustedProjection.TrajectoryID != verificationTrajectoryID {
		return fmt.Errorf("withheld uncertainty did not preserve the projection contract")
	}

	for name, fingerprint := range map[string]string{
		"aggregate":   data.InputFingerprint,
		"weather":     data.Weather.InputFingerprint,
		"trust":       data.Trust.InputFingerprint,
		"alignment":   data.Alignment.InputFingerprint,
		"encounter":   data.Encounter.InputFingerprint,
		"uncertainty": data.Uncertainty.InputFingerprint,
	} {
		if !strings.HasPrefix(fingerprint, "sha256:") {
			return fmt.Errorf("%s fingerprint is missing: %q", name, fingerprint)
		}
	}

	return nil
}

func verifyHTTPErrorContracts(app *fiber.App, schedule verificationSchedule) error {
	missingTrajectoryID := "15458fd0-3080-4ecf-a6ed-a2287095ce1c"
	if err := expectError(
		app,
		weatherContextRequestURL(missingTrajectoryID, schedule.AsOfTime, verificationDuration),
		fiber.StatusNotFound,
		"WEATHER_CONTEXT_NOT_FOUND",
	); err != nil {
		return err
	}

	if err := expectError(
		app,
		weatherContextRequestURL(verificationTrajectoryID, schedule.AsOfTime, 0),
		fiber.StatusBadRequest,
		"INVALID_WEATHER_CONTEXT_DURATION",
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
	request := httptest.NewRequest(http.MethodGet, requestURL, nil)
	httpResponse, err := executeHTTPTest(app, request)
	if err != nil {
		return fmt.Errorf("execute expected-error request: %w", err)
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != expectedStatus {
		body, _ := io.ReadAll(httpResponse.Body)
		return fmt.Errorf(
			"status = %d, want %d, body = %s",
			httpResponse.StatusCode,
			expectedStatus,
			body,
		)
	}

	var payload response.ErrorResponse
	if err := json.NewDecoder(httpResponse.Body).Decode(&payload); err != nil {
		return fmt.Errorf("decode expected-error response: %w", err)
	}
	if payload.Success || payload.Error.Code != expectedCode {
		return fmt.Errorf("unexpected error payload: %#v", payload)
	}
	return nil
}

func executeHTTPTest(
	app *fiber.App,
	request *http.Request,
) (*http.Response, error) {
	if app == nil {
		return nil, fmt.Errorf("Fiber application is required")
	}
	if request == nil {
		return nil, fmt.Errorf("HTTP request is required")
	}
	return app.Test(
		request,
		int(runtimeHTTPTestTimeout/time.Millisecond),
	)
}

func weatherContextRequestURL(
	trajectoryID string,
	asOfTime time.Time,
	duration time.Duration,
) string {
	values := url.Values{}
	values.Set("as_of_time", asOfTime.UTC().Format(time.RFC3339Nano))
	values.Set("duration_seconds", fmt.Sprintf("%d", int64(duration/time.Second)))
	return "/api/v1/trajectories/" + trajectoryID + "/weather-context?" + values.Encode()
}

func verificationCoordinates(index int) (float64, float64) {
	return 40.4700 + float64(index)*0.015, 50.0400 + float64(index)*0.020
}

func containsString(items []string, expected string) bool {
	for _, item := range items {
		if item == expected {
			return true
		}
	}
	return false
}
