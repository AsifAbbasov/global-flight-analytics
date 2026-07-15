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
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalseries"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const (
	expectedMigrationVersion  = "015"
	expectedMigrationName     = "create_historical_aggregate_results"
	expectedMigrationChecksum = "1f6d0243ee42d57f377dfc9ec0b6af88f7c2512fd662691e75b72dbc681149a7"
)

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

	if err := verifyMigrationAndSchema(
		ctx,
		pool,
	); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: verify migration 015 and historical aggregate schema: %v\n",
			err,
		)
		return 1
	}

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
	result, err := verificationResult(now)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: build verification historical result: %v\n",
			err,
		)
		return 1
	}

	store, err := historicalaggregate.
		NewPostgresWithExecutor(
			tx,
			func() time.Time {
				return now
			},
		)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: compose PostgreSQL Historical Aggregate Store: %v\n",
			err,
		)
		return 1
	}

	record, err := store.Put(ctx, result)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: put verification historical aggregate: %v\n",
			err,
		)
		return 1
	}

	replayed, err := store.Put(ctx, result)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: replay verification historical aggregate: %v\n",
			err,
		)
		return 1
	}
	if !reflect.DeepEqual(record, replayed) {
		fmt.Fprintln(
			stderr,
			"ERROR: idempotent historical aggregate replay returned a different record",
		)
		return 1
	}

	loaded, err := store.Get(ctx, record.Key)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: get verification historical aggregate: %v\n",
			err,
		)
		return 1
	}
	if !reflect.DeepEqual(record, loaded) {
		fmt.Fprintln(
			stderr,
			"ERROR: loaded historical aggregate differs from the stored record",
		)
		return 1
	}

	query := historicalaggregate.ListQuery{
		SchemaVersion: historicalcontract.
			SchemaVersionV1,
		MetricName: historicalcontract.
			MetricNameFlightCount,
		Scope: historicalcontract.Scope{
			Type: historicalcontract.ScopeTypeGlobal,
		},
		Granularity: historicalcontract.
			GranularityHour,
		Limit: 1,
	}

	latest, err := store.GetLatest(ctx, query)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: get latest verification historical aggregate: %v\n",
			err,
		)
		return 1
	}
	if latest.ID != record.ID {
		fmt.Fprintln(
			stderr,
			"ERROR: latest historical aggregate does not match the stored record",
		)
		return 1
	}

	page, err := store.List(ctx, query)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: list verification historical aggregates: %v\n",
			err,
		)
		return 1
	}
	if len(page.Records) != 1 ||
		page.HasMore ||
		page.Records[0].ID != record.ID {
		fmt.Fprintf(
			stderr,
			"ERROR: unexpected historical aggregate page: %#v\n",
			page,
		)
		return 1
	}

	conflicting := result.Clone()
	conflicting.Provenance.InputFingerprint =
		"sha256:" + strings.Repeat("b", 64)
	_, err = store.Put(ctx, conflicting)
	if !errors.Is(
		err,
		historicalaggregate.ErrResultConflict,
	) {
		fmt.Fprintf(
			stderr,
			"ERROR: conflicting historical aggregate replay returned %v instead of ErrResultConflict\n",
			err,
		)
		return 1
	}

	var resultCount int
	if err := tx.QueryRow(
		ctx,
		`
			SELECT count(*)
			FROM historical_aggregate_results
			WHERE id = $1;
		`,
		record.ID,
	).Scan(&resultCount); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: count transactional verification aggregates: %v\n",
			err,
		)
		return 1
	}
	if resultCount != 1 {
		fmt.Fprintf(
			stderr,
			"ERROR: transactional historical aggregate count = %d, want 1\n",
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

	var recordPersisted bool
	if err := pool.QueryRow(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM historical_aggregate_results
				WHERE id = $1
			);
		`,
		record.ID,
	).Scan(&recordPersisted); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: verify historical aggregate rollback: %v\n",
			err,
		)
		return 1
	}
	if recordPersisted {
		fmt.Fprintln(
			stderr,
			"ERROR: verification historical aggregate remained after rollback",
		)
		return 1
	}

	fmt.Fprintln(
		stdout,
		"PostgreSQL Historical Aggregate Store Verification",
	)
	fmt.Fprintf(
		stdout,
		"Store: %s\n",
		historicalaggregate.Version,
	)
	fmt.Fprintf(
		stdout,
		"Schema: %s\n",
		historicalcontract.SchemaVersionV1,
	)
	fmt.Fprintf(
		stdout,
		"Migration: %s %s\n",
		expectedMigrationVersion,
		expectedMigrationName,
	)
	fmt.Fprintf(
		stdout,
		"Metric: %s\n",
		record.Result.Metric.Name,
	)
	fmt.Fprintf(
		stdout,
		"Scope: %s\n",
		record.Result.Scope.Type,
	)
	fmt.Fprintf(
		stdout,
		"Record identifier: %s\n",
		record.ID,
	)
	fmt.Fprintln(stdout, "Migration checksum: PASS")
	fmt.Fprintln(stdout, "Schema objects: PASS")
	fmt.Fprintln(stdout, "Put: PASS")
	fmt.Fprintln(stdout, "Idempotent replay: PASS")
	fmt.Fprintln(stdout, "Conflict detection: PASS")
	fmt.Fprintln(stdout, "Get: PASS")
	fmt.Fprintln(stdout, "GetLatest: PASS")
	fmt.Fprintln(stdout, "List: PASS")
	fmt.Fprintln(stdout, "JSON contract round trip: PASS")
	fmt.Fprintln(stdout, "Transaction rollback: PASS")
	fmt.Fprintln(stdout, "Persistent verification rows: 0")
	fmt.Fprintln(stdout, "Result: PASS")

	return 0
}

func verifyMigrationAndSchema(
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

	var tableExists bool
	if err := pool.QueryRow(
		ctx,
		`
			SELECT to_regclass(
				'public.historical_aggregate_results'
			) IS NOT NULL;
		`,
	).Scan(&tableExists); err != nil {
		return fmt.Errorf(
			"query historical aggregate table: %w",
			err,
		)
	}
	if !tableExists {
		return fmt.Errorf(
			"historical_aggregate_results table is absent",
		)
	}

	var requiredConstraintCount int
	if err := pool.QueryRow(
		ctx,
		`
			SELECT count(*)
			FROM pg_constraint AS constraint_record
			JOIN pg_class AS table_record
			  ON table_record.oid =
			     constraint_record.conrelid
			JOIN pg_namespace AS namespace_record
			  ON namespace_record.oid =
			     table_record.relnamespace
			WHERE namespace_record.nspname = 'public'
			  AND table_record.relname =
			      'historical_aggregate_results'
			  AND constraint_record.conname IN (
			      'historical_aggregate_results_id_check',
			      'historical_aggregate_results_schema_version_check',
			      'historical_aggregate_results_metric_name_check',
			      'historical_aggregate_results_scope_check',
			      'historical_aggregate_results_granularity_check',
			      'historical_aggregate_results_window_check',
			      'historical_aggregate_results_input_fingerprint_check',
			      'historical_aggregate_results_series_status_check',
			      'historical_aggregate_results_confidence_level_check',
			      'historical_aggregate_results_json_check',
			      'historical_aggregate_results_key_unique'
			  );
		`,
	).Scan(&requiredConstraintCount); err != nil {
		return fmt.Errorf(
			"query historical aggregate constraints: %w",
			err,
		)
	}
	if requiredConstraintCount != 11 {
		return fmt.Errorf(
			"historical aggregate required constraint count = %d, want 11",
			requiredConstraintCount,
		)
	}

	var historyIndexExists bool
	var statusIndexExists bool
	if err := pool.QueryRow(
		ctx,
		`
			SELECT
				to_regclass(
					'public.historical_aggregate_results_history_idx'
				) IS NOT NULL,
				to_regclass(
					'public.historical_aggregate_results_status_time_idx'
				) IS NOT NULL;
		`,
	).Scan(
		&historyIndexExists,
		&statusIndexExists,
	); err != nil {
		return fmt.Errorf(
			"query historical aggregate indexes: %w",
			err,
		)
	}
	if !historyIndexExists ||
		!statusIndexExists {
		return fmt.Errorf(
			"historical aggregate indexes are incomplete: history=%t status=%t",
			historyIndexExists,
			statusIndexExists,
		)
	}

	return nil
}

func verificationResult(
	now time.Time,
) (historicalcontract.Result, error) {
	generatedAt := now.UTC()
	endTime := generatedAt.Truncate(time.Hour)
	startTime := endTime.Add(-2 * time.Hour)

	window := historicalcontract.TimeWindow{
		StartTime: startTime,
		EndTime:   endTime,
		AsOfTime:  endTime,
	}
	firstBucket := historicalwindow.Bucket{
		Key:       "verification-bucket-0",
		Sequence:  0,
		StartTime: startTime,
		EndTime: startTime.Add(
			time.Hour,
		),
	}
	secondBucket := historicalwindow.Bucket{
		Key:       "verification-bucket-1",
		Sequence:  1,
		StartTime: firstBucket.EndTime,
		EndTime:   endTime,
	}

	return historicalseries.Build(
		historicalseries.BuildRequest{
			Metric: historicalcontract.Metric{
				Name: historicalcontract.
					MetricNameFlightCount,
				Unit: "flights",
				Aggregation: historicalcontract.
					AggregationCount,
			},
			Scope: historicalcontract.Scope{
				Type: historicalcontract.
					ScopeTypeGlobal,
			},
			Plan: historicalwindow.Plan{
				Version: historicalwindow.Version,
				Fingerprint: "runtime-verification-" +
					generatedAt.Format(
						time.RFC3339Nano,
					),
				RequestedStartTime: startTime,
				RequestedEndTime:   endTime,
				AsOfTime:           endTime,
				Granularity: historicalcontract.
					GranularityHour,
				EffectiveWindow: &window,
				Buckets: []historicalwindow.Bucket{
					firstBucket,
					secondBucket,
				},
				MaximumBucketCount: 100,
			},
			Values: []historicalseries.BucketValue{
				{
					Bucket:      firstBucket,
					Value:       2,
					SampleCount: 2,
				},
				{
					Bucket:      secondBucket,
					Value:       3,
					SampleCount: 3,
				},
			},
			DataCoverageRatio: 1,
			BuilderVersion:    "historical-aggregate-runtime-verification-v1",
			InputFingerprint: "sha256:" +
				strings.Repeat("a", 64),
			SourceNames: []string{
				"runtime_verification",
			},
			LatestSourceUpdatedAt: endTime,
			GeneratedAt:           generatedAt,
		},
	)
}
