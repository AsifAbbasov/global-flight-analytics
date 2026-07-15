package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalmaterialization"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalreplay"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const (
	expectedMigrationVersion  = "015"
	expectedMigrationName     = "create_historical_aggregate_results"
	expectedMigrationChecksum = "1f6d0243ee42d57f377dfc9ec0b6af88f7c2512fd662691e75b72dbc681149a7"

	evidenceDatasetLimit       = 5_000
	evidenceMaximumBucketCount = 32
	evidenceMaximumWindowCount = 8
)

type evidenceSchedule struct {
	AsOfTime       time.Time
	GeneratedAt    time.Time
	ClosedBoundary time.Time

	PreviousStart time.Time
	PreviousEnd   time.Time
	CurrentStart  time.Time
	CurrentEnd    time.Time
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

	if err := verifyEvidenceSchema(
		ctx,
		pool,
	); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: verify historical evidence schema: %v\n",
			err,
		)
		return 1
	}

	schedule, err := buildEvidenceSchedule(
		time.Now().UTC(),
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: build evidence schedule: %v\n",
			err,
		)
		return 1
	}

	fixture, err := buildEvidenceFixture(
		schedule,
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: build deterministic evidence fixture: %v\n",
			err,
		)
		return 1
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: begin evidence transaction: %v\n",
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

	if err := insertEvidenceFixture(
		ctx,
		tx,
		fixture,
	); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: insert deterministic evidence: %v\n",
			err,
		)
		return 1
	}

	readRepository, err :=
		historicalread.NewPostgresWithExecutor(
			tx,
		)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: compose transactional Historical Read Repository: %v\n",
			err,
		)
		return 1
	}

	aggregateStore, err :=
		historicalaggregate.NewPostgresWithExecutor(
			tx,
			func() time.Time {
				return schedule.GeneratedAt
			},
		)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: compose transactional Historical Aggregate Store: %v\n",
			err,
		)
		return 1
	}

	materializer, err :=
		historicalmaterialization.New(
			historicalmaterialization.Config{
				Repository: readRepository,
				Store:      aggregateStore,
				Now: func() time.Time {
					return schedule.GeneratedAt
				},
			},
		)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: compose Historical Materializer: %v\n",
			err,
		)
		return 1
	}

	expectations :=
		evidenceMetricExpectations()
	outcomes := make(
		[]historicalmaterialization.Outcome,
		0,
		len(expectations),
	)

	for _, expectation := range expectations {
		outcome, materializeErr :=
			materializer.Materialize(
				ctx,
				historicalmaterialization.Request{
					StartTime: schedule.CurrentStart,
					EndTime:   schedule.CurrentEnd,
					AsOfTime:  schedule.AsOfTime,

					Granularity: historicalcontract.
						GranularityHour,
					MetricName: expectation.Name,
					Scope:      expectation.Scope,

					DatasetLimit:       evidenceDatasetLimit,
					MaximumBucketCount: evidenceMaximumBucketCount,
					GeneratedAt:        schedule.GeneratedAt,
				},
			)
		if materializeErr != nil {
			fmt.Fprintf(
				stderr,
				"ERROR: materialize %s: %v\n",
				expectation.Name,
				materializeErr,
			)
			return 1
		}

		if err := validateMetricOutcome(
			outcome,
			expectation,
			schedule,
		); err != nil {
			fmt.Fprintf(
				stderr,
				"ERROR: validate %s evidence: %v\n",
				expectation.Name,
				err,
			)
			return 1
		}

		if err := reloadAggregate(
			ctx,
			aggregateStore,
			outcome,
		); err != nil {
			fmt.Fprintf(
				stderr,
				"ERROR: reload %s aggregate: %v\n",
				expectation.Name,
				err,
			)
			return 1
		}

		outcomes = append(
			outcomes,
			outcome.Clone(),
		)
	}

	replayRunner, err := historicalreplay.New(
		historicalreplay.Config{
			Materializer: materializer,
			Now: func() time.Time {
				return schedule.GeneratedAt
			},
		},
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: compose Historical Replay Runner: %v\n",
			err,
		)
		return 1
	}

	replayed, err := replayRunner.Run(
		ctx,
		historicalreplay.Request{
			StartTime: schedule.CurrentStart,
			EndTime:   schedule.CurrentEnd,
			AsOfTime:  schedule.AsOfTime,

			Granularity: historicalcontract.
				GranularityHour,
			MetricName: historicalcontract.
				MetricNameFlightCount,
			Scope: historicalcontract.Scope{
				Type: historicalcontract.
					ScopeTypeGlobal,
			},

			DatasetLimit:       evidenceDatasetLimit,
			MaximumBucketCount: evidenceMaximumBucketCount,
			MaximumWindowCount: evidenceMaximumWindowCount,
			GeneratedAt:        schedule.GeneratedAt,
		},
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: replay deterministic evidence: %v\n",
			err,
		)
		return 1
	}

	if err := validateReplayEvidence(
		replayed,
		schedule,
	); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: validate replay evidence: %v\n",
			err,
		)
		return 1
	}

	transactionalCounts, err := countEvidence(
		ctx,
		tx,
		fixture,
		schedule.AsOfTime,
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: count transactional evidence: %v\n",
			err,
		)
		return 1
	}
	expectedAggregateCount :=
		len(outcomes) +
			len(replayed.Windows)
	if err := validateTransactionalCounts(
		transactionalCounts,
		expectedAggregateCount,
	); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: validate transactional evidence counts: %v\n",
			err,
		)
		return 1
	}

	if err := tx.Rollback(ctx); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: rollback evidence transaction: %v\n",
			err,
		)
		return 1
	}
	transactionOpen = false

	rollbackCounts, err := countEvidence(
		ctx,
		pool,
		fixture,
		schedule.AsOfTime,
	)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: count evidence after rollback: %v\n",
			err,
		)
		return 1
	}
	if err := validateRollbackCounts(
		rollbackCounts,
	); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: validate evidence rollback: %v\n",
			err,
		)
		return 1
	}

	fmt.Fprintln(
		stdout,
		"PostgreSQL Transactional Historical Evidence Verification",
	)
	fmt.Fprintf(
		stdout,
		"Historical Read: %s\n",
		historicalread.Version,
	)
	fmt.Fprintf(
		stdout,
		"Materializer: %s\n",
		historicalmaterialization.Version,
	)
	fmt.Fprintf(
		stdout,
		"Replay: %s\n",
		historicalreplay.Version,
	)
	fmt.Fprintf(
		stdout,
		"Aggregate Store: %s\n",
		historicalaggregate.Version,
	)
	fmt.Fprintf(
		stdout,
		"Fixture: flights=%d trajectories=%d observations=%d routes=%d\n",
		len(fixture.FlightIDs),
		len(fixture.TrajectoryIDs),
		len(fixture.ObservationIDs),
		len(fixture.RouteRecordIDs),
	)
	fmt.Fprintln(
		stdout,
		"Migration identity: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Transactional source insertion: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Transactional Historical Read: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Flight count 5 vs 2: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Trajectory count 5 vs 2: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Observation count 10 vs 5: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Airport departures 5 vs 2: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Route observations 5 vs 2: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Exact period comparisons: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Aggregate persistence and reload: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Two-window replay totals 2 then 3: PASS",
	)
	fmt.Fprintln(
		stdout,
		"Source and aggregate rollback: PASS",
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

func buildEvidenceSchedule(
	now time.Time,
) (evidenceSchedule, error) {
	if now.IsZero() {
		return evidenceSchedule{},
			fmt.Errorf(
				"verification time is required",
			)
	}

	asOfTime := now.UTC()
	closedBoundary := asOfTime.Truncate(
		time.Hour,
	)
	currentStart := closedBoundary.Add(
		-2 * time.Hour,
	)
	previousStart := currentStart.Add(
		-2 * time.Hour,
	)

	return evidenceSchedule{
		AsOfTime:       asOfTime,
		GeneratedAt:    asOfTime,
		ClosedBoundary: closedBoundary,

		PreviousStart: previousStart,
		PreviousEnd:   currentStart,
		CurrentStart:  currentStart,
		CurrentEnd:    closedBoundary,
	}, nil
}

func verifyEvidenceSchema(
	ctx context.Context,
	pool *pgxpool.Pool,
) error {
	var migrationExact bool
	if err := pool.QueryRow(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM schema_migrations
				WHERE version = $1
				  AND name = $2
				  AND checksum = $3
			);
		`,
		expectedMigrationVersion,
		expectedMigrationName,
		expectedMigrationChecksum,
	).Scan(&migrationExact); err != nil {
		return fmt.Errorf(
			"query migration history: %w",
			err,
		)
	}
	if !migrationExact {
		return fmt.Errorf(
			"migration %s is not applied with the expected name and checksum",
			expectedMigrationVersion,
		)
	}

	requiredTables := []string{
		"flights",
		"flight_trajectories",
		"flight_states",
		"flight_route_results",
		"historical_aggregate_results",
	}
	for _, tableName := range requiredTables {
		var exists bool
		if err := pool.QueryRow(
			ctx,
			`
				SELECT to_regclass(
					'public.' || $1
				) IS NOT NULL;
			`,
			tableName,
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
