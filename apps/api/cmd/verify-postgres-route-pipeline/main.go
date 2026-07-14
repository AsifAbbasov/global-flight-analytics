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
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routepipeline"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routestore"
	"github.com/joho/godotenv"
)

const verificationSource = "postgres-route-pipeline-verification"

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
			_ = tx.Rollback(
				context.Background(),
			)
		}
	}()

	now := time.Now().UTC()
	endTime := now.Add(-2 * time.Minute)
	startTime := endTime.Add(
		-90 * time.Minute,
	)
	updatedAt := endTime.Add(time.Second)
	identityKey := "flight-identity-" +
		strings.Repeat("f", 64)

	var trajectoryID string
	if err := tx.QueryRow(
		ctx,
		`
			INSERT INTO flight_trajectories (
				identity_key,
				identity_basis,
				split_reason,
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
				created_at,
				updated_at
			)
			VALUES (
				$1,
				'source_flight_id',
				'initial_observation',
				'ABC123',
				'VERIFY3',
				$2,
				$3,
				5400,
				2,
				10,
				0,
				0.9,
				$4,
				$5,
				$5
			)
			RETURNING id::text;
		`,
		identityKey,
		startTime,
		endTime,
		verificationSource,
		updatedAt,
	).Scan(&trajectoryID); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: create transactional verification trajectory: %v\n",
			err,
		)
		return 1
	}

	store, err :=
		routestore.NewPostgresWithExecutor(
			tx,
			func() time.Time {
				return now
			},
		)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: compose transactional Route Store: %v\n",
			err,
		)
		return 1
	}

	item := verificationTrajectory(
		trajectoryID,
		identityKey,
		startTime,
		endTime,
		updatedAt,
	)

	pipeline, err := routepipeline.New(
		routepipeline.Config{
			TrajectoryReader: staticTrajectoryReader{
				item: item,
			},
			AirportLister: staticAirportLister{
				items: verificationAirports(),
			},
			Store: store,
			Now: func() time.Time {
				return now
			},
		},
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: compose Route Intelligence pipeline: %v\n",
			err,
		)
		return 1
	}

	first, err := pipeline.Process(
		ctx,
		routepipeline.Request{
			TrajectoryID: trajectoryID,
		},
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: process first Route Intelligence pipeline run: %v\n",
			err,
		)
		return 1
	}

	second, err := pipeline.Process(
		ctx,
		routepipeline.Request{
			TrajectoryID: trajectoryID,
		},
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: process idempotent Route Intelligence pipeline run: %v\n",
			err,
		)
		return 1
	}

	if first.Resolution.Result.Status !=
		routecontract.RouteStatusComplete ||
		first.Resolution.Result.Origin == nil ||
		first.Resolution.Result.Destination == nil ||
		first.Resolution.Result.Origin.
			Airport.ICAOCode != "UBBB" ||
		first.Resolution.Result.Destination.
			Airport.ICAOCode != "UGTB" {
		fmt.Fprintf(
			stderr,
			"ERROR: unexpected resolved route: %#v\n",
			first.Resolution.Result,
		)
		return 1
	}
	if first.Resolution.Validation.Status !=
		routecontract.ValidationStatusValid {
		fmt.Fprintf(
			stderr,
			"ERROR: route validation status = %s\n",
			first.Resolution.Validation.Status,
		)
		return 1
	}
	if !reflect.DeepEqual(
		first.Record,
		second.Record,
	) {
		fmt.Fprintln(
			stderr,
			"ERROR: idempotent pipeline replay returned a different stored record",
		)
		return 1
	}

	loaded, err := store.Get(
		ctx,
		first.Record.Key,
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: load stored pipeline result: %v\n",
			err,
		)
		return 1
	}
	if !reflect.DeepEqual(
		first.Record,
		loaded,
	) {
		fmt.Fprintln(
			stderr,
			"ERROR: stored pipeline result differs from the pipeline record",
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
			"ERROR: count transactional pipeline results: %v\n",
			err,
		)
		return 1
	}
	if resultCount != 1 {
		fmt.Fprintf(
			stderr,
			"ERROR: transactional pipeline result count = %d, want 1\n",
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

	var resultPersisted bool
	if err := pool.QueryRow(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM flight_route_results
				WHERE id = $1
			);
		`,
		first.Record.ID,
	).Scan(&resultPersisted); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: verify pipeline result rollback: %v\n",
			err,
		)
		return 1
	}
	if resultPersisted {
		fmt.Fprintln(
			stderr,
			"ERROR: verification pipeline result remained after rollback",
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
			"ERROR: verify pipeline trajectory rollback: %v\n",
			err,
		)
		return 1
	}
	if trajectoryPersisted {
		fmt.Fprintln(
			stderr,
			"ERROR: verification pipeline trajectory remained after rollback",
		)
		return 1
	}

	fmt.Fprintln(
		stdout,
		"Production Route Intelligence Pipeline Verification",
	)
	fmt.Fprintf(
		stdout,
		"Pipeline: %s\n",
		first.PipelineVersion,
	)
	fmt.Fprintf(
		stdout,
		"Store: %s\n",
		routestore.PostgresVersion,
	)
	fmt.Fprintf(
		stdout,
		"Route status: %s\n",
		first.Resolution.Result.Status,
	)
	fmt.Fprintf(
		stdout,
		"Origin: %s\n",
		first.Resolution.Result.Origin.
			Airport.ICAOCode,
	)
	fmt.Fprintf(
		stdout,
		"Destination: %s\n",
		first.Resolution.Result.Destination.
			Airport.ICAOCode,
	)
	fmt.Fprintln(
		stdout,
		"Pipeline process: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Idempotent replay: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Contract validation: PASS",
	)
	fmt.Fprintln(
		stdout,
		"PostgreSQL round trip: PASS",
	)
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

type staticTrajectoryReader struct {
	item trajectory.FlightTrajectory
}

func (reader staticTrajectoryReader) GetTrajectoryByID(
	ctx context.Context,
	trajectoryID string,
) (trajectory.FlightTrajectory, error) {
	if err := ctx.Err(); err != nil {
		return trajectory.FlightTrajectory{},
			err
	}
	if trajectoryID != reader.item.ID {
		return trajectory.FlightTrajectory{},
			fmt.Errorf(
				"unexpected trajectory id %s",
				trajectoryID,
			)
	}

	return reader.item, nil
}

type staticAirportLister struct {
	items []airport.Airport
}

func (lister staticAirportLister) List(
	ctx context.Context,
) ([]airport.Airport, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return append(
		[]airport.Airport(nil),
		lister.items...,
	), nil
}

func verificationTrajectory(
	trajectoryID string,
	identityKey string,
	startTime time.Time,
	endTime time.Time,
	updatedAt time.Time,
) trajectory.FlightTrajectory {
	return trajectory.FlightTrajectory{
		ID:          trajectoryID,
		IdentityKey: identityKey,
		IdentityBasis: trajectory.
			FlightIdentityBasisSourceFlightID,
		SplitReason: trajectory.
			FlightSplitReasonInitialObservation,
		FlightID:         "",
		AircraftID:       "",
		ICAO24:           "ABC123",
		Callsign:         "VERIFY3",
		StartTime:        startTime,
		EndTime:          endTime,
		DurationSeconds:  5400,
		SegmentCount:     2,
		PointCount:       10,
		CoverageGapCount: 0,
		QualityScore:     0.9,
		SourceName:       verificationSource,
		CreatedAt:        updatedAt,
		UpdatedAt:        updatedAt,
		Segments: []trajectory.TrajectorySegment{
			{
				ID:             "verification-origin",
				TrajectoryID:   trajectoryID,
				FlightID:       "",
				AircraftID:     "",
				ICAO24:         "ABC123",
				Callsign:       "VERIFY3",
				SequenceNumber: 1,
				Status: trajectory.
					SegmentStatusObserved,
				QualityScore: 0.95,
				StartTime:    startTime,
				EndTime: startTime.Add(
					40 *
						time.Minute,
				),
				DurationSeconds: 2400,
				StartLatitude:   40.4675,
				StartLongitude:  50.0467,
				EndLatitude:     40.8,
				EndLongitude:    48,
				PointCount:      5,
				SourceName:      verificationSource,
				CreatedAt:       updatedAt,
			},
			{
				ID:             "verification-destination",
				TrajectoryID:   trajectoryID,
				FlightID:       "",
				AircraftID:     "",
				ICAO24:         "ABC123",
				Callsign:       "VERIFY3",
				SequenceNumber: 2,
				Status: trajectory.
					SegmentStatusObserved,
				QualityScore: 0.9,
				StartTime: startTime.Add(
					50 *
						time.Minute,
				),
				EndTime:         endTime,
				DurationSeconds: 2400,
				StartLatitude:   41,
				StartLongitude:  46,
				EndLatitude:     41.6692,
				EndLongitude:    44.9547,
				PointCount:      5,
				SourceName:      verificationSource,
				CreatedAt:       updatedAt,
			},
		},
	}
}

func verificationAirports() []airport.Airport {
	return []airport.Airport{
		{
			ICAOCode:   "UBBB",
			IATACode:   "GYD",
			Name:       "Heydar Aliyev International Airport",
			City:       "Baku",
			Country:    "Azerbaijan",
			Latitude:   40.4675,
			Longitude:  50.0467,
			ElevationM: 3,
			Timezone:   "Asia/Baku",
		},
		{
			ICAOCode:   "UGTB",
			IATACode:   "TBS",
			Name:       "Tbilisi International Airport",
			City:       "Tbilisi",
			Country:    "Georgia",
			Latitude:   41.6692,
			Longitude:  44.9547,
			ElevationM: 495,
			Timezone:   "Asia/Tbilisi",
		},
	}
}
