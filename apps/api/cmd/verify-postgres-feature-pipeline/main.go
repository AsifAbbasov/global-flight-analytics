package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/aircraft"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractorcomposition"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/featurepipeline"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/featurestore"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/validator"
	"github.com/joho/godotenv"
)

const verificationSource = "postgres-feature-pipeline-verification"

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
	asOfTime := now
	identityKey := "flight-identity-" +
		strings.Repeat("a", 64)

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
				'VERIFY1',
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

	composition, err := featurepipeline.NewPostgres(
		featurepipeline.PostgresConfig{
			Extractor: extractorcomposition.Config{
				AircraftLookup: verificationAircraftLookup{},
				Now: func() time.Time {
					return now
				},
			},
			Executor: tx,
			Now: func() time.Time {
				return now
			},
		},
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: compose PostgreSQL feature pipeline: %v\n",
			err,
		)
		return 1
	}

	item := verificationTrajectory(
		trajectoryID,
		identityKey,
		startTime,
		endTime,
		now,
	)
	result, err := composition.Pipeline.Process(
		ctx,
		extractor.Request{
			Trajectory: item,
			AsOfTime:   asOfTime,
		},
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: process verification trajectory: %v\n",
			err,
		)
		return 1
	}

	replayed, err := composition.Pipeline.Process(
		ctx,
		extractor.Request{
			Trajectory: item,
			AsOfTime:   asOfTime,
		},
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: replay verification trajectory: %v\n",
			err,
		)
		return 1
	}
	if !reflect.DeepEqual(
		result.Record,
		replayed.Record,
	) {
		fmt.Fprintln(
			stderr,
			"ERROR: idempotent pipeline replay returned a different record",
		)
		return 1
	}

	alternateFeatures := result.Features()
	alternateFeatures.Provenance.ProcessingVersion =
		"flight-feature-processing-verification-v2"
	alternateRecord, err := composition.Store.Put(
		ctx,
		alternateFeatures,
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: store alternate processing version: %v\n",
			err,
		)
		return 1
	}
	if alternateRecord.ID == result.Record.ID {
		fmt.Fprintln(
			stderr,
			"ERROR: processing versions produced one record identifier",
		)
		return 1
	}
	loadedAlternate, err := composition.Store.Get(
		ctx,
		alternateRecord.Key,
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: get alternate processing snapshot: %v\n",
			err,
		)
		return 1
	}
	if !reflect.DeepEqual(
		alternateRecord,
		loadedAlternate,
	) {
		fmt.Fprintln(
			stderr,
			"ERROR: alternate processing snapshot differs",
		)
		return 1
	}

	loaded, err := composition.Store.Get(
		ctx,
		result.Record.Key,
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: get verification snapshot: %v\n",
			err,
		)
		return 1
	}
	if !reflect.DeepEqual(
		result.Record,
		loaded,
	) {
		fmt.Fprintln(
			stderr,
			"ERROR: loaded snapshot differs from stored record",
		)
		return 1
	}

	validationReport := loaded.Features.ValidationReport
	if validationReport.AuditState !=
		validator.AuditStateComplete ||
		validationReport.ValidatorVersion != validator.Version ||
		!validationReport.ValidatedAt.Equal(now) ||
		validationReport.Status != loaded.Features.Quality.Status {
		fmt.Fprintf(
			stderr,
			"ERROR: incomplete durable validation audit: %#v\n",
			validationReport,
		)
		return 1
	}
	if !reflect.DeepEqual(
		result.ValidationReport,
		validationReport,
	) {
		fmt.Fprintln(
			stderr,
			"ERROR: result validation report differs from stored audit",
		)
		return 1
	}

	latest, err := composition.Store.GetLatest(
		ctx,
		trajectoryID,
		flightfeatures.SchemaVersionV1,
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: get latest verification snapshot: %v\n",
			err,
		)
		return 1
	}
	if latest.ID != result.Record.ID {
		fmt.Fprintln(
			stderr,
			"ERROR: latest snapshot does not match stored record",
		)
		return 1
	}

	page, err := composition.Store.List(
		ctx,
		featurestore.ListQuery{
			TrajectoryID:  trajectoryID,
			SchemaVersion: flightfeatures.SchemaVersionV1,
			Limit:         1,
		},
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: list verification snapshots: %v\n",
			err,
		)
		return 1
	}
	if len(page.Records) != 1 ||
		page.HasMore ||
		page.Records[0].ID != result.Record.ID {
		fmt.Fprintf(
			stderr,
			"ERROR: unexpected verification history page: %#v\n",
			page,
		)
		return 1
	}

	var snapshotCount int
	if err := tx.QueryRow(
		ctx,
		`
			SELECT count(*)
			FROM flight_feature_snapshots
			WHERE trajectory_id = $1::uuid;
		`,
		trajectoryID,
	).Scan(&snapshotCount); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: count transactional verification snapshots: %v\n",
			err,
		)
		return 1
	}
	if snapshotCount != 2 {
		fmt.Fprintf(
			stderr,
			"ERROR: transactional snapshot count = %d, want 2\n",
			snapshotCount,
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

	var snapshotPersisted bool
	if err := pool.QueryRow(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM flight_feature_snapshots
				WHERE id = $1
			);
		`,
		result.Record.ID,
	).Scan(&snapshotPersisted); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: verify snapshot rollback: %v\n",
			err,
		)
		return 1
	}
	if snapshotPersisted {
		fmt.Fprintln(
			stderr,
			"ERROR: verification snapshot remained after rollback",
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
		"PostgreSQL Feature Pipeline Verification",
	)
	fmt.Fprintf(
		stdout,
		"Pipeline composition: %s\n",
		featurepipeline.PostgresCompositionVersion,
	)
	fmt.Fprintf(
		stdout,
		"Pipeline: %s\n",
		composition.Versions.Pipeline,
	)
	fmt.Fprintf(
		stdout,
		"Store: %s\n",
		composition.Versions.Store,
	)
	fmt.Fprintf(
		stdout,
		"Validation status: %s\n",
		result.Features().Quality.Status,
	)
	fmt.Fprintf(
		stdout,
		"Record identifier: %s\n",
		result.Record.ID,
	)
	fmt.Fprintln(
		stdout,
		"Put: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Idempotent replay: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Processing version isolation: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Validation audit trail: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Get: PASS",
	)
	fmt.Fprintln(
		stdout,
		"GetLatest: PASS",
	)
	fmt.Fprintln(
		stdout,
		"List: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Transaction rollback: PASS",
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

func verificationTrajectory(
	trajectoryID string,
	identityKey string,
	startTime time.Time,
	endTime time.Time,
	now time.Time,
) trajectory.FlightTrajectory {
	pointOneTime := startTime
	pointTwoTime := endTime

	return trajectory.FlightTrajectory{
		ID:          trajectoryID,
		IdentityKey: identityKey,
		IdentityBasis: trajectory.
			FlightIdentityBasisSourceFlightID,
		SplitReason: trajectory.
			FlightSplitReasonInitialObservation,
		ICAO24:          "ABC123",
		Callsign:        "VERIFY1",
		StartTime:       startTime,
		EndTime:         endTime,
		DurationSeconds: 60,
		SegmentCount:    1,
		PointCount:      2,
		QualityScore:    1,
		SourceName:      verificationSource,
		Points: []trajectory.TrackPoint4D{
			{
				ID:                  "verification-point-1",
				ICAO24:              "ABC123",
				Callsign:            "VERIFY1",
				Latitude:            40.4093,
				Longitude:           49.8671,
				BarometricAltitudeM: 1000,
				BarometricAltitudeStatus: flightstate.
					AltitudeStatusObserved,
				GeometricAltitudeM: 1020,
				GeometricAltitudeStatus: flightstate.
					AltitudeStatusObserved,
				VelocityMPS:     120,
				HeadingDegrees:  90,
				VerticalRateMPS: 2,
				ObservedAt:      pointOneTime,
				SourceName:      verificationSource,
			},
			{
				ID:                  "verification-point-2",
				ICAO24:              "ABC123",
				Callsign:            "VERIFY1",
				Latitude:            40.5093,
				Longitude:           50.0671,
				BarometricAltitudeM: 1200,
				BarometricAltitudeStatus: flightstate.
					AltitudeStatusObserved,
				GeometricAltitudeM: 1220,
				GeometricAltitudeStatus: flightstate.
					AltitudeStatusObserved,
				VelocityMPS:     130,
				HeadingDegrees:  100,
				VerticalRateMPS: 3,
				ObservedAt:      pointTwoTime,
				SourceName:      verificationSource,
			},
		},
		Segments: []trajectory.TrajectorySegment{
			{
				ID:             "verification-segment-1",
				TrajectoryID:   trajectoryID,
				ICAO24:         "ABC123",
				Callsign:       "VERIFY1",
				SequenceNumber: 1,
				Status: trajectory.
					SegmentStatusObserved,
				QualityScore:    1,
				StartTime:       startTime,
				EndTime:         endTime,
				DurationSeconds: 60,
				StartLatitude:   40.4093,
				StartLongitude:  49.8671,
				EndLatitude:     40.5093,
				EndLongitude:    50.0671,
				PointCount:      2,
				SourceName:      verificationSource,
				CreatedAt:       now,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type verificationAircraftLookup struct{}

func (verificationAircraftLookup) GetByICAO24(
	ctx context.Context,
	icao24 string,
) (aircraft.Aircraft, error) {
	if err := ctx.Err(); err != nil {
		return aircraft.Aircraft{}, err
	}
	if strings.TrimSpace(icao24) == "" {
		return aircraft.Aircraft{},
			errors.New("icao24 is required")
	}

	return aircraft.Aircraft{
		ICAO24:       strings.ToUpper(icao24),
		Registration: "VERIFY",
		Manufacturer: "Verification",
		Model:        "Pipeline",
		AircraftType: "test",
		Airline:      "Global Flight Analytics",
		Country:      "Azerbaijan",
	}, nil
}
