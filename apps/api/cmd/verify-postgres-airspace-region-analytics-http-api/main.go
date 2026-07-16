package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/airspaceproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/airspaceregionanalytics"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactiongraph"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/server"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const (
	verificationRegionCode = "azerbaijan"
	verificationWindow     = 5 * time.Minute

	verificationSuccessRunID = "11a1f001-7b41-4f8a-9bb0-100000000001"
	verificationFailedRunID  = "11a1f001-7b41-4f8a-9bb0-100000000002"

	verificationSuccessSource = "airspace-runtime-verification-success-v1"
	verificationFailedSource  = "airspace-runtime-verification-failed-v1"

	selectedAircraftCount           = 4
	selectedObservationCount        = 20
	unknownAltitudeObservationCount = 5
	storedSuccessfulStateCount      = 22
	storedFailedStateCount          = 1
	storedStateCount                = storedSuccessfulStateCount + storedFailedStateCount

	runtimeHTTPTestTimeout     = 60 * time.Second
	runtimeVerificationTimeout = 3 * time.Minute
	fixtureCleanupTimeout      = 60 * time.Second
)

var verificationAircraft = []string{
	"A1B2C3",
	"B2C3D4",
	"C3D4E5",
	"D4E5F6",
}

type verificationSchedule struct {
	WindowStart      time.Time
	AsOfTime         time.Time
	GeneratedAt      time.Time
	SnapshotTimes    []time.Time
	LatestObservedAt time.Time
	FutureObservedAt time.Time
}

type fixtureObservation struct {
	IngestionRunID string
	ICAO24         string
	Callsign       string
	Latitude       float64
	Longitude      float64
	AltitudeMeters *float64
	Velocity       float64
	Heading        float64
	VerticalRate   float64
	OnGround       bool
	ObservedAt     time.Time
	SourceName     string
}

type fixtureCounts struct {
	IngestionRuns int
	FlightStates  int
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
		fmt.Fprintf(stderr, "ERROR: verify Airspace Intelligence runtime schema: %v\n", err)
		return 1
	}

	schedule := buildVerificationSchedule()
	if err := cleanupFixture(ctx, pool); err != nil {
		fmt.Fprintf(stderr, "ERROR: remove stale Airspace Intelligence verification fixture: %v\n", err)
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
		fmt.Fprintf(stderr, "ERROR: insert Airspace Intelligence runtime fixture: %v\n", err)
		return 1
	}

	beforeCounts, err := loadFixtureCounts(ctx, pool)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: count inserted Airspace Intelligence fixture: %v\n", err)
		return 1
	}
	if beforeCounts != (fixtureCounts{IngestionRuns: 2, FlightStates: storedStateCount}) {
		fmt.Fprintf(stderr, "ERROR: unexpected inserted fixture counts: %#v\n", beforeCounts)
		return 1
	}

	resolvedRegion, err := region.NewService().GetByCode(verificationRegionCode)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: resolve verification region: %v\n", err)
		return 1
	}

	observationReader, err := airspaceproduction.NewPostgresObservationReader(pool)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: compose PostgreSQL Airspace Intelligence observation reader: %v\n", err)
		return 1
	}

	loadedObservations, err := verifyPostgresObservationReader(
		ctx,
		observationReader,
		resolvedRegion,
		schedule,
	)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: verify PostgreSQL Airspace Intelligence observation reader: %v\n", err)
		return 1
	}

	service, err := airspaceproduction.New(
		airspaceproduction.Config{
			ObservationReader: observationReader,
			RegionResolver:    region.NewService(),
			Now: func() time.Time {
				return schedule.GeneratedAt
			},
		},
	)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: compose production Airspace Region Analytics service: %v\n", err)
		return 1
	}

	directResult, err := verifyDirectProductionComposition(
		ctx,
		service,
		schedule,
	)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: verify direct production Airspace Region Analytics composition: %v\n", err)
		return 1
	}

	replayedResult, err := service.GetAirspaceRegionAnalytics(
		ctx,
		verificationRequest(schedule),
	)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: replay production Airspace Region Analytics composition: %v\n", err)
		return 1
	}
	if directResult.Provenance.InputFingerprint != replayedResult.Provenance.InputFingerprint {
		fmt.Fprintf(
			stderr,
			"ERROR: deterministic replay fingerprint mismatch: %s != %s\n",
			directResult.Provenance.InputFingerprint,
			replayedResult.Provenance.InputFingerprint,
		)
		return 1
	}

	app := fiber.New()
	v1 := app.Group("/api/v1")
	if err := server.RegisterAirspaceRegionAnalyticsReadRoute(v1, service); err != nil {
		fmt.Fprintf(stderr, "ERROR: register Airspace Region Analytics HTTP route: %v\n", err)
		return 1
	}

	payload, err := verifySuccessEndpoint(app, schedule, directResult)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: verify production Airspace Region Analytics endpoint: %v\n", err)
		return 1
	}

	if err := verifyHTTPErrorContracts(app, schedule); err != nil {
		fmt.Fprintf(stderr, "ERROR: verify Airspace Region Analytics HTTP error contracts: %v\n", err)
		return 1
	}

	if err := cleanupFixture(ctx, pool); err != nil {
		fmt.Fprintf(stderr, "ERROR: clean Airspace Intelligence runtime fixture: %v\n", err)
		return 1
	}

	afterCounts, err := loadFixtureCounts(ctx, pool)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: count Airspace Intelligence fixture after cleanup: %v\n", err)
		return 1
	}
	if afterCounts != (fixtureCounts{}) {
		fmt.Fprintf(stderr, "ERROR: runtime fixture remained after cleanup: %#v\n", afterCounts)
		return 1
	}
	cleanupPending = false

	fmt.Fprintln(stdout, "PostgreSQL Airspace Region Analytics HTTP API Verification")
	fmt.Fprintf(stdout, "Airspace production version: %s\n", payload.Data.Version)
	fmt.Fprintf(stdout, "Region: %s\n", payload.Data.RegionCode)
	fmt.Fprintf(stdout, "Selected PostgreSQL observations: %d of %d stored\n", len(loadedObservations), storedStateCount)
	fmt.Fprintf(stdout, "Minute snapshots: %d\n", payload.Data.Metrics.SnapshotCount)
	fmt.Fprintf(stdout, "Unique aircraft: %d\n", payload.Data.Metrics.UniqueAircraftCount)
	fmt.Fprintf(stdout, "Candidate-backed sector reports: %d\n", payload.Data.Metrics.SectorReportCount)
	fmt.Fprintf(stdout, "Airspace pressure index: %.6f\n", payload.Data.Metrics.AirspacePressureIndex)
	fmt.Fprintf(stdout, "Highest complexity level: %s\n", payload.Data.Metrics.HighestComplexityLevel)
	fmt.Fprintln(stdout, "Schema objects: PASS")
	fmt.Fprintln(stdout, "Deterministic runtime fixture: PASS")
	fmt.Fprintln(stdout, "Successful ingestion-run boundary: PASS")
	fmt.Fprintln(stdout, "Failed ingestion-run exclusion: PASS")
	fmt.Fprintln(stdout, "Future-evidence boundary: PASS")
	fmt.Fprintln(stdout, "Region boundary: PASS")
	fmt.Fprintln(stdout, "PostgreSQL observation reader: PASS")
	fmt.Fprintln(stdout, "Altitude-reference resolution: PASS")
	fmt.Fprintln(stdout, "Direct production composition: PASS")
	fmt.Fprintln(stdout, "Deterministic replay fingerprint: PASS")
	fmt.Fprintln(stdout, "Local Traffic Scene pipeline: PASS")
	fmt.Fprintln(stdout, "Proximity Scanner pipeline: PASS")
	fmt.Fprintln(stdout, "Separation Risk pipeline: PASS")
	fmt.Fprintln(stdout, "Temporal occupancy pipeline: PASS")
	fmt.Fprintln(stdout, "Sector complexity pipeline: PASS")
	fmt.Fprintln(stdout, "Airspace Region Analytics endpoint: PASS")
	fmt.Fprintln(stdout, "Not-found contract: PASS")
	fmt.Fprintln(stdout, "Validation error contract: PASS")
	fmt.Fprintln(stdout, "JSON response contract: PASS")
	fmt.Fprintln(stdout, "Research-only scope guard: PASS")
	fmt.Fprintln(stdout, "Synthetic-sector scope guard: PASS")
	fmt.Fprintln(stdout, "Fixture cleanup: PASS")
	fmt.Fprintln(stdout, "Persistent verification rows: 0")
	fmt.Fprintln(stdout, "Result: PASS")
	return 0
}

func buildVerificationSchedule() verificationSchedule {
	windowStart := time.Date(2035, time.January, 15, 12, 0, 0, 0, time.UTC)
	asOfTime := windowStart.Add(verificationWindow)
	snapshotTimes := make([]time.Time, 0, int(verificationWindow/time.Minute))
	for snapshotTime := windowStart.Add(time.Minute); !snapshotTime.After(asOfTime); snapshotTime = snapshotTime.Add(time.Minute) {
		snapshotTimes = append(snapshotTimes, snapshotTime)
	}
	return verificationSchedule{
		WindowStart:      windowStart,
		AsOfTime:         asOfTime,
		GeneratedAt:      asOfTime.Add(time.Minute),
		SnapshotTimes:    snapshotTimes,
		LatestObservedAt: snapshotTimes[len(snapshotTimes)-1].Add(-5 * time.Second),
		FutureObservedAt: asOfTime.Add(30 * time.Second),
	}
}

func verificationRequest(schedule verificationSchedule) airspaceproduction.Request {
	return airspaceproduction.Request{
		RegionCode: verificationRegionCode,
		AsOfTime:   schedule.AsOfTime,
		Window:     verificationWindow,
	}
}

func fixtureObservations(schedule verificationSchedule) []fixtureObservation {
	observations := make([]fixtureObservation, 0, storedStateCount)
	for index, snapshotTime := range schedule.SnapshotTimes {
		observedAt := snapshotTime.Add(-5 * time.Second)
		progress := float64(index)
		observations = append(
			observations,
			fixtureObservation{
				IngestionRunID: verificationSuccessRunID,
				ICAO24:         "a1b2c3",
				Callsign:       "GFA1101",
				Latitude:       40.4000,
				Longitude:      49.8000 + 0.0010*progress,
				AltitudeMeters: float64Pointer(9000),
				Velocity:       230,
				Heading:        90,
				ObservedAt:     observedAt,
				SourceName:     verificationSuccessSource,
			},
			fixtureObservation{
				IngestionRunID: verificationSuccessRunID,
				ICAO24:         "b2c3d4",
				Callsign:       "GFA1102",
				Latitude:       40.4000,
				Longitude:      49.8200 - 0.0010*progress,
				AltitudeMeters: float64Pointer(9200),
				Velocity:       230,
				Heading:        270,
				ObservedAt:     observedAt,
				SourceName:     verificationSuccessSource,
			},
			fixtureObservation{
				IngestionRunID: verificationSuccessRunID,
				ICAO24:         "c3d4e5",
				Callsign:       "GFA1103",
				Latitude:       40.4080 - 0.0005*progress,
				Longitude:      49.8100,
				AltitudeMeters: float64Pointer(9500),
				Velocity:       180,
				Heading:        180,
				VerticalRate:   -1,
				ObservedAt:     observedAt,
				SourceName:     verificationSuccessSource,
			},
			fixtureObservation{
				IngestionRunID: verificationSuccessRunID,
				ICAO24:         "d4e5f6",
				Callsign:       "GFA1104",
				Latitude:       40.4050 + 0.0004*progress,
				Longitude:      49.8150,
				AltitudeMeters: nil,
				Velocity:       170,
				Heading:        0,
				VerticalRate:   1,
				ObservedAt:     observedAt,
				SourceName:     verificationSuccessSource,
			},
		)
	}

	observations = append(
		observations,
		fixtureObservation{
			IngestionRunID: verificationSuccessRunID,
			ICAO24:         "a1b2c3",
			Callsign:       "GFA1101",
			Latitude:       40.4000,
			Longitude:      49.8060,
			AltitudeMeters: float64Pointer(9000),
			Velocity:       230,
			Heading:        90,
			ObservedAt:     schedule.FutureObservedAt,
			SourceName:     verificationSuccessSource,
		},
		fixtureObservation{
			IngestionRunID: verificationSuccessRunID,
			ICAO24:         "e5f6a7",
			Callsign:       "GFAOUT1",
			Latitude:       45.5000,
			Longitude:      49.8100,
			AltitudeMeters: float64Pointer(8500),
			Velocity:       200,
			Heading:        90,
			ObservedAt:     schedule.LatestObservedAt,
			SourceName:     verificationSuccessSource,
		},
		fixtureObservation{
			IngestionRunID: verificationFailedRunID,
			ICAO24:         "f1a2b3",
			Callsign:       "GFAFAIL",
			Latitude:       40.4040,
			Longitude:      49.8120,
			AltitudeMeters: float64Pointer(9100),
			Velocity:       240,
			Heading:        270,
			ObservedAt:     schedule.LatestObservedAt,
			SourceName:     verificationFailedSource,
		},
	)
	return observations
}

func insertFixture(
	ctx context.Context,
	pool *pgxpool.Pool,
	schedule verificationSchedule,
) error {
	if err := insertIngestionRun(
		ctx,
		pool,
		verificationSuccessRunID,
		verificationSuccessSource,
		"success",
		schedule.WindowStart.Add(-2*time.Minute),
		schedule.GeneratedAt,
		storedSuccessfulStateCount,
		"",
	); err != nil {
		return err
	}
	if err := insertIngestionRun(
		ctx,
		pool,
		verificationFailedRunID,
		verificationFailedSource,
		"failed",
		schedule.WindowStart.Add(-time.Minute),
		schedule.GeneratedAt,
		storedFailedStateCount,
		"intentional failed-run fixture",
	); err != nil {
		return err
	}

	for index, observation := range fixtureObservations(schedule) {
		if err := insertFlightState(ctx, pool, observation); err != nil {
			return fmt.Errorf("insert verification flight state %d: %w", index, err)
		}
	}
	return nil
}

func insertIngestionRun(
	ctx context.Context,
	pool *pgxpool.Pool,
	id string,
	sourceName string,
	status string,
	startedAt time.Time,
	finishedAt time.Time,
	recordCount int,
	errorMessage string,
) error {
	_, err := pool.Exec(
		ctx,
		`
			INSERT INTO ingestion_runs (
				id,
				source_name,
				region_id,
				started_at,
				finished_at,
				status,
				records_received,
				records_inserted,
				records_updated,
				error_message
			)
			VALUES (
				$1::uuid,
				$2,
				NULL,
				$3,
				$4,
				$5,
				$6,
				$6,
				0,
				NULLIF($7, '')
			);
		`,
		id,
		sourceName,
		startedAt,
		finishedAt,
		status,
		recordCount,
		errorMessage,
	)
	if err != nil {
		return fmt.Errorf("insert verification ingestion run %s: %w", id, err)
	}
	return nil
}

func insertFlightState(
	ctx context.Context,
	pool *pgxpool.Pool,
	observation fixtureObservation,
) error {
	var geometricAltitude any
	geometricStatus := "unknown"
	if observation.AltitudeMeters != nil {
		geometricAltitude = *observation.AltitudeMeters
		geometricStatus = "observed"
	}

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
				NULL,
				'unavailable',
				CAST($5::double precision AS integer),
				$6,
				$7,
				$8,
				$9,
				$10,
				'Azerbaijan',
				$11,
				$12,
				$13::uuid
			);
		`,
		observation.ICAO24,
		observation.Callsign,
		observation.Latitude,
		observation.Longitude,
		geometricAltitude,
		geometricStatus,
		observation.Velocity,
		observation.Heading,
		observation.VerticalRate,
		observation.OnGround,
		observation.ObservedAt,
		observation.SourceName,
		observation.IngestionRunID,
	)
	if err != nil {
		return err
	}
	return nil
}

func verifySchema(ctx context.Context, pool *pgxpool.Pool) error {
	var ingestionRunsPresent bool
	var flightStatesPresent bool
	if err := pool.QueryRow(
		ctx,
		`
			SELECT
				to_regclass('ingestion_runs') IS NOT NULL,
				to_regclass('flight_states') IS NOT NULL;
		`,
	).Scan(&ingestionRunsPresent, &flightStatesPresent); err != nil {
		return fmt.Errorf("inspect required Airspace Intelligence tables: %w", err)
	}
	if !ingestionRunsPresent || !flightStatesPresent {
		return fmt.Errorf(
			"required tables are missing: ingestion_runs=%t flight_states=%t",
			ingestionRunsPresent,
			flightStatesPresent,
		)
	}

	requiredColumns := []string{
		"ingestion_run_id",
		"barometric_altitude_status",
		"geometric_altitude_status",
		"vertical_rate_mps",
		"observed_at",
		"source_name",
	}
	var columnCount int
	if err := pool.QueryRow(
		ctx,
		`
			SELECT COUNT(*)::int
			FROM information_schema.columns
			WHERE table_name = 'flight_states'
			  AND column_name = ANY($1::text[]);
		`,
		requiredColumns,
	).Scan(&columnCount); err != nil {
		return fmt.Errorf("inspect Airspace Intelligence flight-state columns: %w", err)
	}
	if columnCount != len(requiredColumns) {
		return fmt.Errorf(
			"required flight-state column count = %d, want %d",
			columnCount,
			len(requiredColumns),
		)
	}
	return nil
}

func verifyPostgresObservationReader(
	ctx context.Context,
	reader *airspaceproduction.PostgresObservationReader,
	resolvedRegion region.Region,
	schedule verificationSchedule,
) ([]airspaceproduction.Observation, error) {
	observations, err := reader.ListAirspaceObservations(
		ctx,
		airspaceproduction.ObservationQuery{
			Bounds:      resolvedRegion.Bounds,
			WindowStart: schedule.WindowStart.Add(-90 * time.Second),
			WindowEnd:   schedule.AsOfTime,
			Limit:       1000,
		},
	)
	if err != nil {
		return nil, err
	}
	if len(observations) != selectedObservationCount {
		return nil, fmt.Errorf(
			"selected observation count = %d, want %d",
			len(observations),
			selectedObservationCount,
		)
	}

	unknownAltitudeCount := 0
	seenAircraft := make(map[string]struct{})
	for _, observation := range observations {
		if observation.SourceName != verificationSuccessSource {
			return nil, fmt.Errorf(
				"non-successful ingestion source leaked into reader: %q",
				observation.SourceName,
			)
		}
		if observation.ObservedAt.After(schedule.AsOfTime) {
			return nil, fmt.Errorf(
				"future observation leaked into reader: %s",
				observation.ObservedAt,
			)
		}
		if observation.Latitude < resolvedRegion.Bounds.MinLatitude ||
			observation.Latitude > resolvedRegion.Bounds.MaxLatitude ||
			observation.Longitude < resolvedRegion.Bounds.MinLongitude ||
			observation.Longitude > resolvedRegion.Bounds.MaxLongitude {
			return nil, fmt.Errorf(
				"out-of-region observation leaked into reader: %#v",
				observation,
			)
		}
		seenAircraft[observation.ICAO24] = struct{}{}
		if observation.AltitudeMeters == nil {
			unknownAltitudeCount++
			if observation.AltitudeReference != interactiongraph.AltitudeReferenceUnknown {
				return nil, fmt.Errorf(
					"unknown altitude reference = %q",
					observation.AltitudeReference,
				)
			}
		} else if observation.AltitudeReference != interactiongraph.AltitudeReferenceGeometric {
			return nil, fmt.Errorf(
				"known altitude reference = %q, want geometric",
				observation.AltitudeReference,
			)
		}
	}
	if len(seenAircraft) != selectedAircraftCount {
		return nil, fmt.Errorf(
			"selected unique aircraft = %d, want %d",
			len(seenAircraft),
			selectedAircraftCount,
		)
	}
	if unknownAltitudeCount != unknownAltitudeObservationCount {
		return nil, fmt.Errorf(
			"unknown-altitude observations = %d, want %d",
			unknownAltitudeCount,
			unknownAltitudeObservationCount,
		)
	}
	return observations, nil
}

func verifyDirectProductionComposition(
	ctx context.Context,
	service *airspaceproduction.Service,
	schedule verificationSchedule,
) (airspaceregionanalytics.Result, error) {
	result, err := service.GetAirspaceRegionAnalytics(
		ctx,
		verificationRequest(schedule),
	)
	if err != nil {
		return airspaceregionanalytics.Result{}, err
	}

	report := airspaceregionanalytics.Validate(
		result,
		airspaceregionanalytics.DefaultPolicy(),
	)
	if report.Status != airspaceregionanalytics.ValidationStatusValid {
		return airspaceregionanalytics.Result{}, fmt.Errorf(
			"regional analytics validation status = %q issues=%v",
			report.Status,
			report.Issues,
		)
	}
	if result.RegionCode != strings.ToUpper(verificationRegionCode) ||
		!result.WindowStart.Equal(schedule.WindowStart) ||
		!result.WindowEnd.Equal(schedule.AsOfTime) ||
		!result.GeneratedAt.Equal(schedule.GeneratedAt) {
		return airspaceregionanalytics.Result{}, fmt.Errorf(
			"unexpected result identity or temporal boundary: %#v",
			result,
		)
	}
	if result.Status != airspaceregionanalytics.ResultStatusLimited {
		return airspaceregionanalytics.Result{}, fmt.Errorf(
			"result status = %q, want limited because unknown-altitude evidence is intentional",
			result.Status,
		)
	}
	if result.ScopeGuard != airspaceregionanalytics.ScopeGuardResearchOnly {
		return airspaceregionanalytics.Result{}, fmt.Errorf(
			"scope guard = %q",
			result.ScopeGuard,
		)
	}
	if len(result.Provenance.InputFingerprint) != 64 ||
		!isHexFingerprint(result.Provenance.InputFingerprint) {
		return airspaceregionanalytics.Result{}, fmt.Errorf(
			"input fingerprint is invalid: %q",
			result.Provenance.InputFingerprint,
		)
	}
	if !equalStrings(result.Provenance.SourceNames, []string{verificationSuccessSource}) {
		return airspaceregionanalytics.Result{}, fmt.Errorf(
			"unexpected provenance source names: %#v",
			result.Provenance.SourceNames,
		)
	}
	if !result.Provenance.LatestObservedAt.Equal(schedule.LatestObservedAt) {
		return airspaceregionanalytics.Result{}, fmt.Errorf(
			"latest observed at = %s, want %s",
			result.Provenance.LatestObservedAt,
			schedule.LatestObservedAt,
		)
	}

	metrics := result.Metrics
	if metrics.SnapshotCount != len(schedule.SnapshotTimes) ||
		metrics.BucketCount != len(schedule.SnapshotTimes) ||
		metrics.UniqueAircraftCount != selectedAircraftCount ||
		metrics.AircraftObservationCount != selectedObservationCount ||
		metrics.CurrentAircraftCount != selectedAircraftCount ||
		metrics.UnknownAltitudeCount != unknownAltitudeObservationCount ||
		metrics.TemporalCoverage != 1 {
		return airspaceregionanalytics.Result{}, fmt.Errorf(
			"unexpected regional metrics: %#v",
			metrics,
		)
	}
	if metrics.SectorReportCount == 0 ||
		metrics.AirspacePressureIndex <= 0 ||
		metrics.PeakAirspacePressureIndex <= 0 ||
		metrics.HighestComplexityLevel == airspaceregionanalytics.ComplexityLevelNone ||
		metrics.OccupancyTrend == airspaceregionanalytics.OccupancyTrendUnavailable {
		return airspaceregionanalytics.Result{}, fmt.Errorf(
			"regional complexity or pressure evidence is unavailable: %#v",
			metrics,
		)
	}
	if metrics.IndeterminateRiskCount == 0 ||
		metrics.ContextualRiskCount+metrics.ElevatedRiskCount+metrics.HighRiskCount == 0 {
		return airspaceregionanalytics.Result{}, fmt.Errorf(
			"expected both indeterminate and determinate risk evidence: %#v",
			metrics,
		)
	}
	if result.Occupancy.Metrics.BucketCount != len(schedule.SnapshotTimes) ||
		result.Occupancy.Metrics.ExpectedBucketCount != len(schedule.SnapshotTimes) ||
		len(result.Occupancy.Buckets) != len(schedule.SnapshotTimes) {
		return airspaceregionanalytics.Result{}, fmt.Errorf(
			"unexpected temporal occupancy index: %#v",
			result.Occupancy,
		)
	}
	for _, bucket := range result.Occupancy.Buckets {
		if bucket.Metrics.AircraftCount != selectedAircraftCount || len(bucket.Cells) == 0 {
			return airspaceregionanalytics.Result{}, fmt.Errorf(
				"unexpected occupancy bucket: %#v",
				bucket,
			)
		}
	}
	for _, sector := range result.SectorComplexity {
		if sector.CandidatePairCount == 0 ||
			sector.Score <= 0 ||
			sector.Level == airspaceregionanalytics.ComplexityLevelNone ||
			len(sector.Components) == 0 {
			return airspaceregionanalytics.Result{}, fmt.Errorf(
				"unexpected sector complexity report: %#v",
				sector,
			)
		}
	}
	if !containsLimitation(result.Limitations, "research_only_not_operational_airspace_management") ||
		!containsLimitation(result.Limitations, "synthetic_grid_not_official_sectors") ||
		!containsLimitation(result.Limitations, "unknown_altitude_occupancy_present") {
		return airspaceregionanalytics.Result{}, fmt.Errorf(
			"mandatory result limitations are missing: %#v",
			result.Limitations,
		)
	}
	return result.Clone(), nil
}

func verifySuccessEndpoint(
	app *fiber.App,
	schedule verificationSchedule,
	directResult airspaceregionanalytics.Result,
) (response.SuccessResponse[dto.AirspaceRegionAnalyticsResponse], error) {
	request := httptest.NewRequest(
		http.MethodGet,
		airspaceRequestURL(verificationRegionCode, schedule.AsOfTime, verificationWindow),
		nil,
	)
	httpResponse, err := app.Test(
		request,
		int(runtimeHTTPTestTimeout.Milliseconds()),
	)
	if err != nil {
		return response.SuccessResponse[dto.AirspaceRegionAnalyticsResponse]{}, err
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode != fiber.StatusOK {
		body, _ := io.ReadAll(httpResponse.Body)
		return response.SuccessResponse[dto.AirspaceRegionAnalyticsResponse]{}, fmt.Errorf(
			"status = %d, want %d, body = %s",
			httpResponse.StatusCode,
			fiber.StatusOK,
			body,
		)
	}

	var payload response.SuccessResponse[dto.AirspaceRegionAnalyticsResponse]
	if err := json.NewDecoder(httpResponse.Body).Decode(&payload); err != nil {
		return response.SuccessResponse[dto.AirspaceRegionAnalyticsResponse]{}, err
	}
	if err := validateSuccessPayload(payload, schedule, directResult); err != nil {
		return response.SuccessResponse[dto.AirspaceRegionAnalyticsResponse]{}, err
	}
	return payload, nil
}

func validateSuccessPayload(
	payload response.SuccessResponse[dto.AirspaceRegionAnalyticsResponse],
	schedule verificationSchedule,
	directResult airspaceregionanalytics.Result,
) error {
	data := payload.Data
	if !payload.Success {
		return fmt.Errorf("success response flag is false")
	}
	if data.Version != airspaceproduction.Version ||
		data.SchemaVersion != string(airspaceregionanalytics.SchemaVersionV1) ||
		data.Status != string(airspaceregionanalytics.ResultStatusLimited) ||
		data.RegionCode != strings.ToUpper(verificationRegionCode) {
		return fmt.Errorf("unexpected HTTP response identity: %#v", data)
	}
	if !data.WindowStart.Equal(schedule.WindowStart) ||
		!data.WindowEnd.Equal(schedule.AsOfTime) ||
		!data.GeneratedAt.Equal(schedule.GeneratedAt) {
		return fmt.Errorf("unexpected HTTP response times: %#v", data)
	}
	if data.Provenance.InputFingerprint != directResult.Provenance.InputFingerprint ||
		!isHexFingerprint(data.Provenance.InputFingerprint) ||
		len(data.Provenance.SceneFingerprints) != len(schedule.SnapshotTimes) ||
		len(data.Provenance.ScanFingerprints) != len(schedule.SnapshotTimes) ||
		len(data.Provenance.RiskFingerprints) != len(schedule.SnapshotTimes) ||
		!equalStrings(data.Provenance.SourceNames, []string{verificationSuccessSource}) ||
		!data.Provenance.LatestObservedAt.Equal(schedule.LatestObservedAt) {
		return fmt.Errorf("unexpected HTTP provenance: %#v", data.Provenance)
	}
	for _, fingerprint := range append(
		append(
			append([]string(nil), data.Provenance.SceneFingerprints...),
			data.Provenance.ScanFingerprints...,
		),
		data.Provenance.RiskFingerprints...,
	) {
		if !isHexFingerprint(fingerprint) {
			return fmt.Errorf("invalid upstream fingerprint: %q", fingerprint)
		}
	}
	if data.Occupancy.BucketDurationSeconds != 60 ||
		len(data.Occupancy.Buckets) != len(schedule.SnapshotTimes) ||
		data.Occupancy.Metrics.BucketCount != len(schedule.SnapshotTimes) ||
		data.Occupancy.Metrics.ExpectedBucketCount != len(schedule.SnapshotTimes) ||
		data.Occupancy.Metrics.UniqueAircraftCount != selectedAircraftCount ||
		data.Occupancy.Metrics.AircraftObservationCount != selectedObservationCount ||
		data.Occupancy.Metrics.UnknownAltitudeCount != unknownAltitudeObservationCount ||
		data.Occupancy.Metrics.TemporalCoverage != 1 {
		return fmt.Errorf("unexpected HTTP occupancy response: %#v", data.Occupancy)
	}
	if data.Metrics.SnapshotCount != len(schedule.SnapshotTimes) ||
		data.Metrics.UniqueAircraftCount != selectedAircraftCount ||
		data.Metrics.CurrentAircraftCount != selectedAircraftCount ||
		data.Metrics.IndeterminateRiskCount == 0 ||
		data.Metrics.ContextualRiskCount+data.Metrics.ElevatedRiskCount+data.Metrics.HighRiskCount == 0 ||
		data.Metrics.AirspacePressureIndex <= 0 ||
		data.Metrics.HighestComplexityLevel == string(airspaceregionanalytics.ComplexityLevelNone) {
		return fmt.Errorf("unexpected HTTP regional metrics: %#v", data.Metrics)
	}
	if len(data.SectorComplexity) == 0 {
		return fmt.Errorf("HTTP sector complexity response is empty")
	}
	for _, sector := range data.SectorComplexity {
		if sector.CandidatePairCount == 0 || sector.Score <= 0 || len(sector.Components) == 0 {
			return fmt.Errorf("unexpected HTTP sector complexity: %#v", sector)
		}
	}
	if data.ScopeGuard != string(airspaceregionanalytics.ScopeGuardResearchOnly) ||
		!containsResponseLimitation(data.Limitations, "research_only_not_operational_airspace_management") ||
		!containsResponseLimitation(data.Limitations, "synthetic_grid_not_official_sectors") {
		return fmt.Errorf("HTTP scope protections are missing: %#v", data)
	}
	return nil
}

func verifyHTTPErrorContracts(app *fiber.App, schedule verificationSchedule) error {
	if err := expectError(
		app,
		airspaceRequestURL("unknown-region", schedule.AsOfTime, verificationWindow),
		fiber.StatusNotFound,
		"AIRSPACE_ANALYTICS_REGION_NOT_FOUND",
	); err != nil {
		return err
	}

	invalidWindowURL := airspaceRequestURL(
		verificationRegionCode,
		schedule.AsOfTime,
		61*time.Second,
	)
	if err := expectError(
		app,
		invalidWindowURL,
		fiber.StatusBadRequest,
		"INVALID_AIRSPACE_ANALYTICS_WINDOW",
	); err != nil {
		return err
	}

	values := url.Values{}
	values.Set("as_of_time", "not-a-time")
	values.Set("window_seconds", "300")
	invalidTimeURL := "/api/v1/airspace/regions/" +
		verificationRegionCode + "/analytics?" + values.Encode()
	if err := expectError(
		app,
		invalidTimeURL,
		fiber.StatusBadRequest,
		"INVALID_AIRSPACE_ANALYTICS_AS_OF_TIME",
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
	httpResponse, err := app.Test(
		request,
		int(runtimeHTTPTestTimeout.Milliseconds()),
	)
	if err != nil {
		return err
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
		return err
	}
	if payload.Success || payload.Error.Code != expectedCode || payload.Error.Message == "" {
		return fmt.Errorf("unexpected error response: %#v", payload)
	}
	return nil
}

func airspaceRequestURL(
	regionCode string,
	asOfTime time.Time,
	window time.Duration,
) string {
	values := url.Values{}
	values.Set("as_of_time", asOfTime.UTC().Format(time.RFC3339Nano))
	values.Set("window_seconds", fmt.Sprintf("%d", int64(window/time.Second)))
	return "/api/v1/airspace/regions/" + url.PathEscape(regionCode) +
		"/analytics?" + values.Encode()
}

func cleanupFixture(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(
		ctx,
		`
			DELETE FROM flight_states
			WHERE ingestion_run_id IN ($1::uuid, $2::uuid);
		`,
		verificationSuccessRunID,
		verificationFailedRunID,
	); err != nil {
		return fmt.Errorf("delete verification flight states: %w", err)
	}
	if _, err := pool.Exec(
		ctx,
		`
			DELETE FROM ingestion_runs
			WHERE id IN ($1::uuid, $2::uuid);
		`,
		verificationSuccessRunID,
		verificationFailedRunID,
	); err != nil {
		return fmt.Errorf("delete verification ingestion runs: %w", err)
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
			FROM ingestion_runs
			WHERE id IN ($1::uuid, $2::uuid);
		`,
		verificationSuccessRunID,
		verificationFailedRunID,
	).Scan(&result.IngestionRuns); err != nil {
		return fixtureCounts{}, fmt.Errorf("count verification ingestion runs: %w", err)
	}
	if err := pool.QueryRow(
		ctx,
		`
			SELECT COUNT(*)::int
			FROM flight_states
			WHERE ingestion_run_id IN ($1::uuid, $2::uuid);
		`,
		verificationSuccessRunID,
		verificationFailedRunID,
	).Scan(&result.FlightStates); err != nil {
		return fixtureCounts{}, fmt.Errorf("count verification flight states: %w", err)
	}
	return result, nil
}

func float64Pointer(value float64) *float64 {
	return &value
}

func isHexFingerprint(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func equalStrings(left []string, right []string) bool {
	leftCopy := append([]string(nil), left...)
	rightCopy := append([]string(nil), right...)
	sort.Strings(leftCopy)
	sort.Strings(rightCopy)
	if len(leftCopy) != len(rightCopy) {
		return false
	}
	for index := range leftCopy {
		if leftCopy[index] != rightCopy[index] {
			return false
		}
	}
	return true
}

func containsLimitation(
	limitations []airspaceregionanalytics.Limitation,
	code string,
) bool {
	for _, limitation := range limitations {
		if limitation.Code == code {
			return true
		}
	}
	return false
}

func containsResponseLimitation(
	limitations []dto.AirspaceLimitationResponse,
	code string,
) bool {
	for _, limitation := range limitations {
		if limitation.Code == code {
			return true
		}
	}
	return false
}
