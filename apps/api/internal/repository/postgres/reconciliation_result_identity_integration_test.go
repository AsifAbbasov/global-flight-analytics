package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/reconciliation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const derivedIdentityTestDatabaseURL = "TEST_DATABASE_URL"

var derivedIdentitySchemaCounter uint64

type derivedIdentityFixture struct {
	pool                 *pgxpool.Pool
	adminPool            *pgxpool.Pool
	schemaName           string
	qualityRepository    *DataQualityRepository
	trajectoryRepository *TrajectoryRepository
	state                flightstate.FlightState
	qualityTaskID        string
	trajectoryTaskID     string
}

func TestReconciledQualitySaveIsIdempotentAndRefreshesResult(
	t *testing.T,
) {
	fixture := newDerivedIdentityFixture(
		t,
	)
	defer fixture.close(
		t,
	)

	firstQuality := makeDerivedIdentityQuality(
		0.70,
		"first",
	)
	secondQuality := makeDerivedIdentityQuality(
		0.95,
		"second",
	)

	for _, quality := range []dataquality.DataQuality{
		firstQuality,
		secondQuality,
	} {
		if err := fixture.qualityRepository.SaveReconciledFlightStateQuality(
			context.Background(),
			fixture.qualityTaskID,
			1,
			fixture.state,
			quality,
		); err != nil {
			t.Fatalf(
				"save reconciled quality: %v",
				err,
			)
		}
	}

	var count int
	var score float64
	var warningCode string

	err := fixture.pool.QueryRow(
		context.Background(),
		`
			SELECT
				COUNT(*),
				MAX(score)::double precision,
				MAX(warnings_json->0->>'Code')
			FROM data_quality_reports
			WHERE reconciliation_task_id = $1;
		`,
		fixture.qualityTaskID,
	).Scan(
		&count,
		&score,
		&warningCode,
	)
	if err != nil {
		t.Fatalf(
			"load reconciled quality identity: %v",
			err,
		)
	}

	if count != 1 {
		t.Fatalf(
			"expected one reconciled quality row, got %d",
			count,
		)
	}

	if score != secondQuality.Score {
		t.Fatalf(
			"expected refreshed score %.2f, got %.2f",
			secondQuality.Score,
			score,
		)
	}

	if warningCode != "second" {
		t.Fatalf(
			"expected refreshed warning code second, got %q",
			warningCode,
		)
	}
}

func TestReconciledTrajectorySaveIsIdempotentAndReplacesChildren(
	t *testing.T,
) {
	fixture := newDerivedIdentityFixture(
		t,
	)
	defer fixture.close(
		t,
	)

	firstTrajectory := makeDerivedIdentityTrajectory(
		1,
	)
	secondTrajectory := makeDerivedIdentityTrajectory(
		2,
	)

	for _, item := range []trajectory.FlightTrajectory{
		firstTrajectory,
		secondTrajectory,
	} {
		if err := fixture.trajectoryRepository.SaveReconciledTrajectory(
			context.Background(),
			fixture.trajectoryTaskID,
			1,
			item,
		); err != nil {
			t.Fatalf(
				"save reconciled trajectory: %v",
				err,
			)
		}
	}

	var trajectoryID string
	var trajectoryCount int
	var storedSegmentCount int

	err := fixture.pool.QueryRow(
		context.Background(),
		`
			SELECT
				id::text,
				segment_count
			FROM flight_trajectories
			WHERE reconciliation_task_id = $1;
		`,
		fixture.trajectoryTaskID,
	).Scan(
		&trajectoryID,
		&storedSegmentCount,
	)
	if err != nil {
		t.Fatalf(
			"load reconciled trajectory identity: %v",
			err,
		)
	}

	err = fixture.pool.QueryRow(
		context.Background(),
		`
			SELECT COUNT(*)
			FROM flight_trajectories
			WHERE reconciliation_task_id = $1;
		`,
		fixture.trajectoryTaskID,
	).Scan(
		&trajectoryCount,
	)
	if err != nil {
		t.Fatalf(
			"count reconciled trajectory identity: %v",
			err,
		)
	}

	if trajectoryCount != 1 {
		t.Fatalf(
			"expected one reconciled trajectory row, got %d",
			trajectoryCount,
		)
	}

	if storedSegmentCount != 2 {
		t.Fatalf(
			"expected refreshed parent segment count 2, got %d",
			storedSegmentCount,
		)
	}

	var childCount int

	err = fixture.pool.QueryRow(
		context.Background(),
		`
			SELECT COUNT(*)
			FROM trajectory_segments
			WHERE trajectory_id = $1;
		`,
		trajectoryID,
	).Scan(
		&childCount,
	)
	if err != nil {
		t.Fatalf(
			"load reconciled trajectory children: %v",
			err,
		)
	}

	if childCount != 2 {
		t.Fatalf(
			"expected exactly 2 replacement segments, got %d",
			childCount,
		)
	}
}

func TestReconciledWritesRejectStaleAttemptOwnership(
	t *testing.T,
) {
	fixture := newDerivedIdentityFixture(
		t,
	)
	defer fixture.close(
		t,
	)

	_, err := fixture.pool.Exec(
		context.Background(),
		`
			UPDATE derived_reconciliation_tasks
			SET
				attempt_count = 2,
				updated_at = now()
			WHERE id IN ($1, $2);
		`,
		fixture.qualityTaskID,
		fixture.trajectoryTaskID,
	)
	if err != nil {
		t.Fatalf(
			"advance reconciliation attempts: %v",
			err,
		)
	}

	err = fixture.qualityRepository.SaveReconciledFlightStateQuality(
		context.Background(),
		fixture.qualityTaskID,
		1,
		fixture.state,
		makeDerivedIdentityQuality(
			0.8,
			"stale",
		),
	)
	if !errors.Is(
		err,
		reconciliation.ErrTaskTransitionRejected,
	) {
		t.Fatalf(
			"expected stale quality attempt rejection, got %v",
			err,
		)
	}

	err = fixture.trajectoryRepository.SaveReconciledTrajectory(
		context.Background(),
		fixture.trajectoryTaskID,
		1,
		makeDerivedIdentityTrajectory(
			1,
		),
	)
	if !errors.Is(
		err,
		reconciliation.ErrTaskTransitionRejected,
	) {
		t.Fatalf(
			"expected stale trajectory attempt rejection, got %v",
			err,
		)
	}
}

func TestOrdinaryDerivedWritesRemainAppendOnly(
	t *testing.T,
) {
	fixture := newDerivedIdentityFixture(
		t,
	)
	defer fixture.close(
		t,
	)

	quality := makeDerivedIdentityQuality(
		0.8,
		"ordinary",
	)
	item := makeDerivedIdentityTrajectory(
		1,
	)

	for index := 0; index < 2; index++ {
		if err := fixture.qualityRepository.SaveFlightStateQuality(
			context.Background(),
			fixture.state,
			quality,
		); err != nil {
			t.Fatalf(
				"save ordinary quality: %v",
				err,
			)
		}

		if err := fixture.trajectoryRepository.SaveTrajectory(
			context.Background(),
			item,
		); err != nil {
			t.Fatalf(
				"save ordinary trajectory: %v",
				err,
			)
		}
	}

	var qualityCount int
	var trajectoryCount int

	err := fixture.pool.QueryRow(
		context.Background(),
		`
			SELECT COUNT(*)
			FROM data_quality_reports
			WHERE reconciliation_task_id IS NULL
				AND state_id = $1;
		`,
		fixture.state.ID,
	).Scan(
		&qualityCount,
	)
	if err != nil {
		t.Fatalf(
			"count ordinary quality rows: %v",
			err,
		)
	}

	err = fixture.pool.QueryRow(
		context.Background(),
		`
			SELECT COUNT(*)
			FROM flight_trajectories
			WHERE reconciliation_task_id IS NULL
				AND icao24 = $1;
		`,
		item.ICAO24,
	).Scan(
		&trajectoryCount,
	)
	if err != nil {
		t.Fatalf(
			"count ordinary trajectory rows: %v",
			err,
		)
	}

	if qualityCount != 2 {
		t.Fatalf(
			"expected two ordinary quality rows, got %d",
			qualityCount,
		)
	}

	if trajectoryCount != 2 {
		t.Fatalf(
			"expected two ordinary trajectory rows, got %d",
			trajectoryCount,
		)
	}
}

func newDerivedIdentityFixture(
	t *testing.T,
) *derivedIdentityFixture {
	t.Helper()

	databaseURL := os.Getenv(
		derivedIdentityTestDatabaseURL,
	)
	if databaseURL == "" {
		t.Skipf(
			"%s is not set; skipping PostgreSQL integration test",
			derivedIdentityTestDatabaseURL,
		)
	}

	ctx := context.Background()
	schemaName := fmt.Sprintf(
		"derived_identity_test_%d_%d",
		time.Now().UnixNano(),
		atomic.AddUint64(
			&derivedIdentitySchemaCounter,
			1,
		),
	)

	adminPool, err := pgxpool.New(
		ctx,
		databaseURL,
	)
	if err != nil {
		t.Fatalf(
			"connect derived identity admin postgres: %v",
			err,
		)
	}

	_, err = adminPool.Exec(
		ctx,
		"CREATE SCHEMA "+pgx.Identifier{schemaName}.Sanitize(),
	)
	if err != nil {
		adminPool.Close()
		t.Fatalf(
			"create derived identity schema: %v",
			err,
		)
	}

	poolConfig, err := pgxpool.ParseConfig(
		databaseURL,
	)
	if err != nil {
		adminPool.Close()
		t.Fatalf(
			"parse derived identity postgres config: %v",
			err,
		)
	}

	poolConfig.ConnConfig.RuntimeParams["search_path"] =
		schemaName + ",public"

	pool, err := pgxpool.NewWithConfig(
		ctx,
		poolConfig,
	)
	if err != nil {
		adminPool.Close()
		t.Fatalf(
			"connect derived identity schema postgres: %v",
			err,
		)
	}

	applyAllDerivedIdentityMigrations(
		t,
		pool,
	)

	fixture := &derivedIdentityFixture{
		pool:                 pool,
		adminPool:            adminPool,
		schemaName:           schemaName,
		qualityRepository:    NewDataQualityRepository(pool),
		trajectoryRepository: NewTrajectoryRepository(pool),
		state: flightstate.FlightState{
			ID:                       "22222222-2222-2222-2222-222222222222",
			IngestionRunID:           "11111111-1111-1111-1111-111111111111",
			ICAO24:                   "ABC123",
			Callsign:                 "TEST123",
			Latitude:                 40.4675,
			Longitude:                50.0467,
			BarometricAltitudeM:      1000,
			BarometricAltitudeStatus: flightstate.AltitudeStatusObserved,
			GeometricAltitudeM:       1100,
			GeometricAltitudeStatus:  flightstate.AltitudeStatusObserved,
			VelocityMPS:              200,
			HeadingDegrees:           90,
			VerticalRateMPS:          0,
			OriginCountry:            "Azerbaijan",
			ObservedAt: time.Date(
				2026,
				time.July,
				11,
				17,
				0,
				0,
				0,
				time.UTC,
			),
			SourceName: "test",
		},
		qualityTaskID:    "33333333-3333-3333-3333-333333333333",
		trajectoryTaskID: "44444444-4444-4444-4444-444444444444",
	}

	prepareDerivedIdentityData(
		t,
		fixture,
	)

	return fixture
}

func (fixture *derivedIdentityFixture) close(
	t *testing.T,
) {
	t.Helper()

	fixture.pool.Close()

	_, err := fixture.adminPool.Exec(
		context.Background(),
		"DROP SCHEMA IF EXISTS "+
			pgx.Identifier{fixture.schemaName}.Sanitize()+
			" CASCADE",
	)
	fixture.adminPool.Close()

	if err != nil {
		t.Fatalf(
			"drop derived identity schema: %v",
			err,
		)
	}
}

func applyAllDerivedIdentityMigrations(
	t *testing.T,
	pool *pgxpool.Pool,
) {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal(
			"resolve derived identity integration test path",
		)
	}

	migrationDirectory := filepath.Clean(
		filepath.Join(
			filepath.Dir(currentFile),
			"../../../../../database/migrations",
		),
	)

	entries, err := os.ReadDir(
		migrationDirectory,
	)
	if err != nil {
		t.Fatalf(
			"read migration directory %s: %v",
			migrationDirectory,
			err,
		)
	}

	migrationNames := make(
		[]string,
		0,
		len(entries),
	)

	for _, entry := range entries {
		if entry.IsDir() ||
			!strings.HasSuffix(
				entry.Name(),
				".sql",
			) {
			continue
		}

		migrationNames = append(
			migrationNames,
			entry.Name(),
		)
	}

	sort.Strings(
		migrationNames,
	)

	for _, migrationName := range migrationNames {
		migrationPath := filepath.Join(
			migrationDirectory,
			migrationName,
		)

		migrationSQL, err := os.ReadFile(
			migrationPath,
		)
		if err != nil {
			t.Fatalf(
				"read migration %s: %v",
				migrationPath,
				err,
			)
		}

		if _, err := pool.Exec(
			context.Background(),
			string(migrationSQL),
		); err != nil {
			t.Fatalf(
				"execute migration %s: %v",
				migrationName,
				err,
			)
		}
	}
}

func prepareDerivedIdentityData(
	t *testing.T,
	fixture *derivedIdentityFixture,
) {
	t.Helper()

	_, err := fixture.pool.Exec(
		context.Background(),
		`
			INSERT INTO ingestion_runs (
				id,
				source_name,
				started_at,
				status
			)
			VALUES (
				$1,
				'test',
				$2,
				'success'
			);
		`,
		fixture.state.IngestionRunID,
		fixture.state.ObservedAt,
	)
	if err != nil {
		t.Fatalf(
			"insert derived identity ingestion run: %v",
			err,
		)
	}

	_, err = fixture.pool.Exec(
		context.Background(),
		`
			INSERT INTO flight_states (
				id,
				ingestion_run_id,
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
				source_name
			)
			VALUES (
				$1,
				$2,
				$3,
				$4,
				$5,
				$6,
				$7,
				$8,
				$9,
				$10,
				$11,
				$12,
				$13,
				$14,
				$15,
				$16,
				$17
			);
		`,
		fixture.state.ID,
		fixture.state.IngestionRunID,
		fixture.state.ICAO24,
		fixture.state.Callsign,
		fixture.state.Latitude,
		fixture.state.Longitude,
		fixture.state.BarometricAltitudeM,
		string(fixture.state.BarometricAltitudeStatus),
		fixture.state.GeometricAltitudeM,
		string(fixture.state.GeometricAltitudeStatus),
		fixture.state.VelocityMPS,
		fixture.state.HeadingDegrees,
		fixture.state.VerticalRateMPS,
		fixture.state.OnGround,
		fixture.state.OriginCountry,
		fixture.state.ObservedAt,
		fixture.state.SourceName,
	)
	if err != nil {
		t.Fatalf(
			"insert derived identity flight state: %v",
			err,
		)
	}

	_, err = fixture.pool.Exec(
		context.Background(),
		`
			INSERT INTO derived_reconciliation_tasks (
				id,
				deduplication_key,
				ingestion_run_id,
				icao24,
				derivation_type,
				status,
				observed_from,
				observed_to,
				attempt_count,
				processing_started_at,
				claimed_signal_version
			)
			VALUES (
				$1,
				'quality-identity',
				$2,
				'abc123',
				'flight_state_quality',
				'processing',
				$3,
				$3,
				1,
				now(),
				1
			);
		`,
		fixture.qualityTaskID,
		fixture.state.IngestionRunID,
		fixture.state.ObservedAt,
	)
	if err != nil {
		t.Fatalf(
			"insert quality reconciliation task: %v",
			err,
		)
	}

	_, err = fixture.pool.Exec(
		context.Background(),
		`
			INSERT INTO derived_reconciliation_tasks (
				id,
				deduplication_key,
				ingestion_run_id,
				icao24,
				derivation_type,
				status,
				observed_from,
				observed_to,
				attempt_count,
				processing_started_at,
				claimed_signal_version
			)
			VALUES (
				$1,
				'trajectory-identity',
				$2,
				'abc123',
				'trajectory',
				'processing',
				$3,
				$4,
				1,
				now(),
				1
			);
		`,
		fixture.trajectoryTaskID,
		fixture.state.IngestionRunID,
		fixture.state.ObservedAt,
		fixture.state.ObservedAt.Add(time.Minute),
	)
	if err != nil {
		t.Fatalf(
			"insert trajectory reconciliation task: %v",
			err,
		)
	}
}

func makeDerivedIdentityQuality(
	score float64,
	warningCode string,
) dataquality.DataQuality {
	return dataquality.DataQuality{
		ValidationStatus: dataquality.ValidationStatusValid,
		Completeness:     dataquality.CompletenessLevelComplete,
		Confidence:       dataquality.ConfidenceLevelHigh,
		Score:            score,
		MissingFields:    []string{},
		Warnings: []dataquality.Warning{
			{
				Code:    warningCode,
				Message: warningCode,
				Field:   "icao24",
			},
		},
	}
}

func makeDerivedIdentityTrajectory(
	segmentCount int,
) trajectory.FlightTrajectory {
	startTime := time.Date(
		2026,
		time.July,
		11,
		17,
		0,
		0,
		0,
		time.UTC,
	)

	segments := make(
		[]trajectory.TrajectorySegment,
		0,
		segmentCount,
	)

	for index := 0; index < segmentCount; index++ {
		segmentStart := startTime.Add(
			time.Duration(index) * time.Minute,
		)
		segmentEnd := segmentStart.Add(
			time.Minute,
		)

		segments = append(
			segments,
			trajectory.TrajectorySegment{
				ICAO24:          "ABC123",
				Callsign:        "TEST123",
				SequenceNumber:  index + 1,
				Status:          trajectory.SegmentStatusObserved,
				QualityScore:    0.9,
				StartTime:       segmentStart,
				EndTime:         segmentEnd,
				DurationSeconds: 60,
				StartLatitude:   40.4675 + float64(index)*0.01,
				StartLongitude:  50.0467 + float64(index)*0.01,
				EndLatitude:     40.4775 + float64(index)*0.01,
				EndLongitude:    50.0567 + float64(index)*0.01,
				PointCount:      2,
				SourceName:      "test",
			},
		)
	}

	return trajectory.FlightTrajectory{
		ICAO24:           "ABC123",
		Callsign:         "TEST123",
		StartTime:        startTime,
		EndTime:          startTime.Add(time.Duration(segmentCount) * time.Minute),
		DurationSeconds:  int64(segmentCount * 60),
		SegmentCount:     segmentCount,
		PointCount:       segmentCount * 2,
		CoverageGapCount: 0,
		QualityScore:     0.9,
		SourceName:       "test",
		Segments:         segments,
	}
}
