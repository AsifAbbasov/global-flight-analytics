package main

import (
	"context"
	"fmt"
	"os"
	"reflect"
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

	verificationDatasetLimit       = 5_000
	verificationMaximumBucketCount = 32
	verificationMaximumWindowCount = 8
)

type verificationSchedule struct {
	AsOfTime             time.Time
	GeneratedAt          time.Time
	MaterializationStart time.Time
	MaterializationEnd   time.Time
	ReplayStart          time.Time
	ReplayEnd            time.Time
}

func main() {
	os.Exit(run(os.Stdout, os.Stderr))
}

func run(stdout *os.File, stderr *os.File) int {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("apps/api/.env")

	cfg, err := config.LoadMigrationConfig()
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: load database configuration: %v\n", err)
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
		fmt.Fprintf(stderr, "ERROR: connect postgres: %v\n", err)
		return 1
	}
	defer pool.Close()

	if err := verifyMigration(ctx, pool); err != nil {
		fmt.Fprintf(stderr, "ERROR: verify migration 015: %v\n", err)
		return 1
	}

	schedule, err := buildVerificationSchedule(time.Now().UTC())
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: build verification schedule: %v\n", err)
		return 1
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: begin verification transaction: %v\n", err)
		return 1
	}
	transactionOpen := true
	defer func() {
		if transactionOpen {
			_ = tx.Rollback(context.Background())
		}
	}()

	readRepository, err := historicalread.NewPostgres(
		historicalread.PostgresConfig{Pool: pool},
	)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: compose Historical Read Repository: %v\n", err)
		return 1
	}

	aggregateStore, err := historicalaggregate.NewPostgresWithExecutor(
		tx,
		func() time.Time { return schedule.GeneratedAt },
	)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: compose Historical Aggregate Store: %v\n", err)
		return 1
	}

	materializer, err := historicalmaterialization.New(
		historicalmaterialization.Config{
			Repository: readRepository,
			Store:      aggregateStore,
			Now:        func() time.Time { return schedule.GeneratedAt },
		},
	)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: compose Historical Materializer: %v\n", err)
		return 1
	}

	materialized, err := materializer.Materialize(
		ctx,
		historicalmaterialization.Request{
			StartTime: schedule.MaterializationStart,
			EndTime:   schedule.MaterializationEnd,
			AsOfTime:  schedule.AsOfTime,
			Granularity: historicalcontract.
				GranularityHour,
			MetricName: historicalcontract.
				MetricNameFlightCount,
			Scope: historicalcontract.Scope{
				Type: historicalcontract.ScopeTypeGlobal,
			},
			DatasetLimit:       verificationDatasetLimit,
			MaximumBucketCount: verificationMaximumBucketCount,
			GeneratedAt:        schedule.GeneratedAt,
		},
	)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: materialize verification aggregate: %v\n", err)
		return 1
	}
	if err := validateMaterialization(materialized, schedule); err != nil {
		fmt.Fprintf(stderr, "ERROR: validate materialization: %v\n", err)
		return 1
	}

	loaded, err := aggregateStore.Get(ctx, materialized.Record.Key)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: reload materialized aggregate: %v\n", err)
		return 1
	}
	if !reflect.DeepEqual(materialized.Record, loaded) {
		fmt.Fprintln(
			stderr,
			"ERROR: reloaded aggregate differs from the materialized record",
		)
		return 1
	}

	replayRunner, err := historicalreplay.New(
		historicalreplay.Config{
			Materializer: materializer,
			Now:          func() time.Time { return schedule.GeneratedAt },
		},
	)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: compose Historical Replay Runner: %v\n", err)
		return 1
	}

	replayed, err := replayRunner.Run(
		ctx,
		historicalreplay.Request{
			StartTime: schedule.ReplayStart,
			EndTime:   schedule.ReplayEnd,
			AsOfTime:  schedule.AsOfTime,
			Granularity: historicalcontract.
				GranularityHour,
			MetricName: historicalcontract.
				MetricNameFlightCount,
			Scope: historicalcontract.Scope{
				Type: historicalcontract.ScopeTypeGlobal,
			},
			DatasetLimit:       verificationDatasetLimit,
			MaximumBucketCount: verificationMaximumBucketCount,
			MaximumWindowCount: verificationMaximumWindowCount,
			GeneratedAt:        schedule.GeneratedAt,
		},
	)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: run verification replay: %v\n", err)
		return 1
	}
	if err := validateReplay(replayed, schedule); err != nil {
		fmt.Fprintf(stderr, "ERROR: validate replay: %v\n", err)
		return 1
	}

	expectedRecordCount := 1 + len(replayed.Windows)
	var transactionalRecordCount int
	if err := tx.QueryRow(
		ctx,
		`
			SELECT count(*)
			FROM historical_aggregate_results
			WHERE as_of_time_unix_nano = $1
			  AND metric_name = 'flight_count'
			  AND scope_key = 'global';
		`,
		schedule.AsOfTime.UnixNano(),
	).Scan(&transactionalRecordCount); err != nil {
		fmt.Fprintf(stderr, "ERROR: count transactional aggregates: %v\n", err)
		return 1
	}
	if transactionalRecordCount != expectedRecordCount {
		fmt.Fprintf(
			stderr,
			"ERROR: transactional aggregate count = %d, want %d\n",
			transactionalRecordCount,
			expectedRecordCount,
		)
		return 1
	}

	if err := tx.Rollback(ctx); err != nil {
		fmt.Fprintf(stderr, "ERROR: rollback verification transaction: %v\n", err)
		return 1
	}
	transactionOpen = false

	var persistedRecordCount int
	if err := pool.QueryRow(
		ctx,
		`
			SELECT count(*)
			FROM historical_aggregate_results
			WHERE as_of_time_unix_nano = $1
			  AND metric_name = 'flight_count'
			  AND scope_key = 'global';
		`,
		schedule.AsOfTime.UnixNano(),
	).Scan(&persistedRecordCount); err != nil {
		fmt.Fprintf(stderr, "ERROR: verify transaction rollback: %v\n", err)
		return 1
	}
	if persistedRecordCount != 0 {
		fmt.Fprintf(
			stderr,
			"ERROR: %d verification records remained after rollback\n",
			persistedRecordCount,
		)
		return 1
	}

	fmt.Fprintln(
		stdout,
		"PostgreSQL Historical Materialization and Replay Verification",
	)
	fmt.Fprintf(stdout, "Materializer: %s\n", historicalmaterialization.Version)
	fmt.Fprintf(stdout, "Replay: %s\n", historicalreplay.Version)
	fmt.Fprintf(stdout, "Read repository: %s\n", historicalread.Version)
	fmt.Fprintf(stdout, "Aggregate store: %s\n", historicalaggregate.Version)
	fmt.Fprintf(stdout, "Schema: %s\n", historicalcontract.SchemaVersionV1)
	fmt.Fprintf(
		stdout,
		"Historical source rows: flights=%d trajectories=%d observations=%d routes=%d\n",
		materialized.ReadSummary.FlightCount,
		materialized.ReadSummary.TrajectoryCount,
		materialized.ReadSummary.ObservationCount,
		materialized.ReadSummary.RouteCount,
	)
	fmt.Fprintf(stdout, "Replay windows: %d\n", len(replayed.Windows))
	fmt.Fprintln(stdout, "Migration identity: PASS")
	fmt.Fprintln(stdout, "Historical read: PASS")
	fmt.Fprintln(stdout, "Window planning: PASS")
	fmt.Fprintln(stdout, "Current-period build: PASS")
	fmt.Fprintln(stdout, "Previous-period build: PASS")
	fmt.Fprintln(stdout, "Period comparison: PASS")
	fmt.Fprintln(stdout, "Combined comparison fingerprint: PASS")
	fmt.Fprintln(stdout, "Aggregate persistence: PASS")
	fmt.Fprintln(stdout, "Aggregate reload: PASS")
	fmt.Fprintln(stdout, "Chronological replay: PASS")
	fmt.Fprintln(stdout, "Replay aggregate persistence: PASS")
	fmt.Fprintln(stdout, "Transaction rollback: PASS")
	fmt.Fprintln(stdout, "Persistent verification rows: 0")
	fmt.Fprintln(stdout, "Result: PASS")

	return 0
}

func buildVerificationSchedule(now time.Time) (verificationSchedule, error) {
	if now.IsZero() {
		return verificationSchedule{}, fmt.Errorf("verification time is required")
	}

	asOfTime := now.UTC()
	closedBoundary := asOfTime.Truncate(time.Hour)

	return verificationSchedule{
		AsOfTime:             asOfTime,
		GeneratedAt:          asOfTime,
		MaterializationStart: closedBoundary.Add(-2 * time.Hour),
		MaterializationEnd:   closedBoundary,
		ReplayStart:          closedBoundary.Add(-2 * time.Hour),
		ReplayEnd:            closedBoundary,
	}, nil
}

func verifyMigration(ctx context.Context, pool *pgxpool.Pool) error {
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
		return fmt.Errorf("query migration history: %w", err)
	}
	if !migrationExact {
		return fmt.Errorf(
			"migration %s is not applied with the expected name and checksum",
			expectedMigrationVersion,
		)
	}

	var tableExists bool
	if err := pool.QueryRow(
		ctx,
		`
			SELECT to_regclass(
				'public.historical_aggregate_results'
			) IS NOT NULL;
		`,
	).Scan(&tableExists); err != nil {
		return fmt.Errorf("query historical aggregate table: %w", err)
	}
	if !tableExists {
		return fmt.Errorf("historical_aggregate_results table is absent")
	}

	return nil
}

func validateMaterialization(
	outcome historicalmaterialization.Outcome,
	schedule verificationSchedule,
) error {
	if outcome.Version != historicalmaterialization.Version {
		return fmt.Errorf("materialization version = %q", outcome.Version)
	}
	if outcome.Plan.EffectiveWindow == nil ||
		outcome.PreviousPlan.EffectiveWindow == nil {
		return fmt.Errorf("materialization plans lack effective windows")
	}
	if len(outcome.Plan.Buckets) != 2 ||
		len(outcome.PreviousPlan.Buckets) != 2 {
		return fmt.Errorf(
			"unexpected bucket counts: current=%d previous=%d",
			len(outcome.Plan.Buckets),
			len(outcome.PreviousPlan.Buckets),
		)
	}
	if !outcome.Plan.EffectiveWindow.StartTime.Equal(
		schedule.MaterializationStart,
	) ||
		!outcome.Plan.EffectiveWindow.EndTime.Equal(
			schedule.MaterializationEnd,
		) {
		return fmt.Errorf(
			"unexpected materialization window: %#v",
			outcome.Plan.EffectiveWindow,
		)
	}
	if outcome.CurrentResult.Comparison == nil {
		return fmt.Errorf("materialized result has no period comparison")
	}
	if outcome.CurrentResult.Provenance.InputFingerprint ==
		outcome.PreviousResult.Provenance.InputFingerprint {
		return fmt.Errorf("combined comparison fingerprint was not created")
	}
	if outcome.Record.ID == "" {
		return fmt.Errorf("materialized record identifier is empty")
	}
	if !reflect.DeepEqual(outcome.Record.Result, outcome.CurrentResult) {
		return fmt.Errorf("persisted result differs from materialized result")
	}

	report := historicalcontract.Validate(outcome.CurrentResult)
	if report.Status != historicalcontract.ValidationStatusValid {
		return fmt.Errorf(
			"materialized contract is invalid: errors=%d warnings=%d",
			report.ErrorCount,
			report.WarningCount,
		)
	}

	return nil
}

func validateReplay(
	result historicalreplay.Result,
	schedule verificationSchedule,
) error {
	if result.Version != historicalreplay.Version {
		return fmt.Errorf("replay version = %q", result.Version)
	}
	if len(result.Plan.Buckets) != 2 ||
		len(result.Windows) != 2 {
		return fmt.Errorf(
			"unexpected replay counts: plan=%d windows=%d",
			len(result.Plan.Buckets),
			len(result.Windows),
		)
	}

	expectedStart := schedule.ReplayStart
	for index, window := range result.Windows {
		expectedEnd := expectedStart.Add(time.Hour)

		if window.Bucket.Sequence != index+1 {
			return fmt.Errorf(
				"replay sequence[%d] = %d, want %d",
				index,
				window.Bucket.Sequence,
				index+1,
			)
		}
		if !window.Bucket.StartTime.Equal(expectedStart) ||
			!window.Bucket.EndTime.Equal(expectedEnd) {
			return fmt.Errorf(
				"unexpected replay window[%d]: %#v",
				index,
				window.Bucket,
			)
		}
		if window.Record.ID == "" ||
			window.Record.Result.Comparison == nil {
			return fmt.Errorf(
				"replay record[%d] lacks persisted comparison evidence",
				index,
			)
		}

		report := historicalcontract.Validate(window.Record.Result)
		if report.Status != historicalcontract.ValidationStatusValid {
			return fmt.Errorf(
				"replay record[%d] is invalid: errors=%d warnings=%d",
				index,
				report.ErrorCount,
				report.WarningCount,
			)
		}

		expectedStart = expectedEnd
	}

	if !expectedStart.Equal(schedule.ReplayEnd) {
		return fmt.Errorf(
			"replay ended at %s, want %s",
			expectedStart,
			schedule.ReplayEnd,
		)
	}

	return nil
}
