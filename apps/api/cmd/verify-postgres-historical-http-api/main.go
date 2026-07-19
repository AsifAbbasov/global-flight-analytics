package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/historicalcursor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/server"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
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

	if err := verifySchema(ctx, pool); err != nil {
		fmt.Fprintf(stderr, "ERROR: verify Historical Intelligence schema: %v\n", err)
		return 1
	}

	schedule, err := buildVerificationSchedule(time.Now().UTC())
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: build verification schedule: %v\n", err)
		return 1
	}
	results, err := buildVerificationResults(schedule)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: build verification aggregates: %v\n", err)
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

	store, err := historicalaggregate.NewPostgresWithExecutor(
		tx,
		func() time.Time { return schedule.GeneratedAt },
	)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: compose transactional aggregate store: %v\n", err)
		return 1
	}

	fingerprints := make([]string, 0, len(results))
	for _, result := range results {
		record, putErr := store.Put(ctx, result)
		if putErr != nil {
			fmt.Fprintf(stderr, "ERROR: insert verification aggregate %s: %v\n", result.Metric.Name, putErr)
			return 1
		}
		fingerprints = append(fingerprints, record.InputFingerprint)
	}

	app := fiber.New()
	v1 := app.Group("/api/v1")
	if err := server.RegisterHistoricalIntelligenceReadRoutes(v1, store); err != nil {
		fmt.Fprintf(stderr, "ERROR: register Historical Intelligence routes: %v\n", err)
		return 1
	}

	if err := verifyLatestEndpoint(app, schedule); err != nil {
		fmt.Fprintf(stderr, "ERROR: verify latest endpoint: %v\n", err)
		return 1
	}
	if err := verifyHistoryPagination(app, schedule); err != nil {
		fmt.Fprintf(stderr, "ERROR: verify history pagination: %v\n", err)
		return 1
	}
	if err := verifyRouteScopeEndpoint(app); err != nil {
		fmt.Fprintf(stderr, "ERROR: verify route-scope endpoint: %v\n", err)
		return 1
	}
	if err := verifyHTTPErrorContracts(app); err != nil {
		fmt.Fprintf(stderr, "ERROR: verify HTTP errors: %v\n", err)
		return 1
	}

	transactionalCount, err := countVerificationRows(
		ctx,
		tx,
		fingerprints,
	)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: count transactional aggregates: %v\n", err)
		return 1
	}
	if transactionalCount != len(results) {
		fmt.Fprintf(
			stderr,
			"ERROR: transactional aggregate count = %d, want %d\n",
			transactionalCount,
			len(results),
		)
		return 1
	}

	if err := tx.Rollback(ctx); err != nil {
		fmt.Fprintf(stderr, "ERROR: rollback verification transaction: %v\n", err)
		return 1
	}
	transactionOpen = false

	persistentCount, err := countVerificationRows(
		ctx,
		pool,
		fingerprints,
	)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: count aggregates after rollback: %v\n", err)
		return 1
	}
	if persistentCount != 0 {
		fmt.Fprintf(stderr, "ERROR: %d verification rows remained after rollback\n", persistentCount)
		return 1
	}

	fmt.Fprintln(stdout, "PostgreSQL Historical Intelligence HTTP API Verification")
	fmt.Fprintf(stdout, "Aggregate Store: %s\n", historicalaggregate.Version)
	fmt.Fprintf(stdout, "Schema: %s\n", historicalcontract.SchemaVersionV1)
	fmt.Fprintf(stdout, "Fixture aggregates: %d\n", len(results))
	fmt.Fprintln(stdout, "Migration identity: PASS")
	fmt.Fprintln(stdout, "Production route registrar: PASS")
	fmt.Fprintln(stdout, "Transactional aggregate persistence: PASS")
	fmt.Fprintln(stdout, "Latest aggregate endpoint: PASS")
	fmt.Fprintln(stdout, "History first page and cursor: PASS")
	fmt.Fprintln(stdout, "History second page: PASS")
	fmt.Fprintln(stdout, "Route scope normalization: PASS")
	fmt.Fprintln(stdout, "Not-found contract: PASS")
	fmt.Fprintln(stdout, "Validation error contracts: PASS")
	fmt.Fprintln(stdout, "JSON response contract: PASS")
	fmt.Fprintln(stdout, "Transaction rollback: PASS")
	fmt.Fprintln(stdout, "Persistent verification rows: 0")
	fmt.Fprintln(stdout, "Result: PASS")

	return 0
}

func verifySchema(ctx context.Context, pool *pgxpool.Pool) error {
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
		return fmt.Errorf("migration 015 identity does not match")
	}

	var tableExists bool
	if err := pool.QueryRow(
		ctx,
		`SELECT to_regclass('public.historical_aggregate_results') IS NOT NULL;`,
	).Scan(&tableExists); err != nil {
		return fmt.Errorf("query aggregate table: %w", err)
	}
	if !tableExists {
		return fmt.Errorf("historical_aggregate_results table is absent")
	}

	return nil
}

func verifyLatestEndpoint(
	app *fiber.App,
	schedule verificationSchedule,
) error {
	values := baseGlobalQuery(
		historicalcontract.MetricNameFlightCount,
	)
	payload, err := getSuccessRecord(
		app,
		"/api/v1"+server.HistoricalIntelligenceLatestPath+"?"+values.Encode(),
	)
	if err != nil {
		return err
	}

	if !payload.Success ||
		payload.Data.Result.Summary.Total != 3 ||
		payload.Data.Result.Metric.Name != "flight_count" ||
		payload.Data.Result.Scope.Type != "global" ||
		!payload.Data.Result.Window.EndTime.Equal(schedule.ClosedBoundary) {
		return fmt.Errorf("unexpected latest response: %#v", payload)
	}

	return nil
}

func verifyHistoryPagination(
	app *fiber.App,
	schedule verificationSchedule,
) error {
	firstValues := baseGlobalQuery(
		historicalcontract.MetricNameFlightCount,
	)
	firstValues.Set("limit", "2")

	first, err := getSuccessHistory(
		app,
		"/api/v1"+server.HistoricalIntelligenceHistoryPath+"?"+firstValues.Encode(),
	)
	if err != nil {
		return err
	}
	if !first.Success ||
		len(first.Data.Items) != 2 ||
		!first.Data.HasMore ||
		first.Data.NextCursor == "" ||
		first.Data.Items[0].Result.Summary.Total != 3 ||
		first.Data.Items[1].Result.Summary.Total != 2 {
		return fmt.Errorf(
			"unexpected first history page: %#v",
			first,
		)
	}

	decoded, err := historicalcursor.Decode(
		first.Data.NextCursor,
	)
	if err != nil {
		return fmt.Errorf(
			"decode first history cursor: %w",
			err,
		)
	}
	lastVisible := first.Data.Items[1]
	if decoded == nil ||
		decoded.ID != lastVisible.ID ||
		!decoded.WindowEnd.Equal(
			lastVisible.Result.Window.EndTime,
		) ||
		!decoded.WindowStart.Equal(
			lastVisible.Result.Window.StartTime,
		) ||
		!decoded.AsOfTime.Equal(
			lastVisible.Result.Window.AsOfTime,
		) {
		return fmt.Errorf(
			"history cursor does not match last visible record: cursor=%#v record=%#v",
			decoded,
			lastVisible,
		)
	}

	secondValues := baseGlobalQuery(
		historicalcontract.MetricNameFlightCount,
	)
	secondValues.Set("limit", "2")
	secondValues.Set(
		"cursor",
		first.Data.NextCursor,
	)

	second, err := getSuccessHistory(
		app,
		"/api/v1"+server.HistoricalIntelligenceHistoryPath+"?"+secondValues.Encode(),
	)
	if err != nil {
		return err
	}
	if !second.Success ||
		len(second.Data.Items) != 1 ||
		second.Data.HasMore ||
		second.Data.NextCursor != "" ||
		second.Data.Items[0].Result.Summary.Total != 1 {
		return fmt.Errorf(
			"unexpected second history page: %#v",
			second,
		)
	}

	return nil
}

func verifyRouteScopeEndpoint(app *fiber.App) error {
	values := url.Values{}
	values.Set("metric", "route_observations")
	values.Set("scope", "route")
	values.Set("granularity", "hour")
	values.Set("origin_icao", "ubbb")
	values.Set("destination_icao", "ugtb")

	payload, err := getSuccessRecord(
		app,
		"/api/v1"+server.HistoricalIntelligenceLatestPath+"?"+values.Encode(),
	)
	if err != nil {
		return err
	}
	if payload.Data.Result.Summary.Total != 4 ||
		payload.Data.Result.Scope.OriginICAOCode != verificationOriginICAO ||
		payload.Data.Result.Scope.DestinationICAOCode != verificationDestinationICAO {
		return fmt.Errorf("unexpected route response: %#v", payload)
	}

	return nil
}

func verifyHTTPErrorContracts(app *fiber.App) error {
	notFoundValues := baseGlobalQuery(
		historicalcontract.MetricNameObservationCount,
	)
	if err := expectError(
		app,
		"/api/v1"+server.HistoricalIntelligenceLatestPath+"?"+notFoundValues.Encode(),
		fiber.StatusNotFound,
		"HISTORICAL_INTELLIGENCE_NOT_FOUND",
	); err != nil {
		return err
	}

	invalidMetric := baseGlobalQuery(
		historicalcontract.MetricName("not_supported"),
	)
	if err := expectError(
		app,
		"/api/v1"+server.HistoricalIntelligenceLatestPath+"?"+invalidMetric.Encode(),
		fiber.StatusBadRequest,
		"INVALID_HISTORICAL_INTELLIGENCE_METRIC",
	); err != nil {
		return err
	}

	invalidLimit := baseGlobalQuery(
		historicalcontract.MetricNameFlightCount,
	)
	invalidLimit.Set("limit", "101")
	if err := expectError(
		app,
		"/api/v1"+server.HistoricalIntelligenceHistoryPath+"?"+invalidLimit.Encode(),
		fiber.StatusBadRequest,
		"INVALID_HISTORICAL_INTELLIGENCE_LIMIT",
	); err != nil {
		return err
	}

	invalidCursor := baseGlobalQuery(
		historicalcontract.MetricNameFlightCount,
	)
	invalidCursor.Set("cursor", "not-a-cursor")
	return expectError(
		app,
		"/api/v1"+server.HistoricalIntelligenceHistoryPath+"?"+invalidCursor.Encode(),
		fiber.StatusBadRequest,
		"INVALID_HISTORICAL_INTELLIGENCE_CURSOR",
	)
}

func baseGlobalQuery(
	metric historicalcontract.MetricName,
) url.Values {
	values := url.Values{}
	values.Set("metric", string(metric))
	values.Set("scope", "global")
	values.Set("granularity", "hour")
	return values
}

func getSuccessRecord(
	app *fiber.App,
	requestURL string,
) (response.SuccessResponse[dto.HistoricalIntelligenceAggregateRecord], error) {
	result, body, err := executeGET(app, requestURL)
	if err != nil {
		return response.SuccessResponse[dto.HistoricalIntelligenceAggregateRecord]{}, err
	}
	if result.StatusCode != fiber.StatusOK {
		return response.SuccessResponse[dto.HistoricalIntelligenceAggregateRecord]{},
			fmt.Errorf("status = %d body = %s", result.StatusCode, body)
	}

	var payload response.SuccessResponse[dto.HistoricalIntelligenceAggregateRecord]
	if err := json.Unmarshal(body, &payload); err != nil {
		return payload, fmt.Errorf("decode record response: %w", err)
	}
	return payload, nil
}

func getSuccessHistory(
	app *fiber.App,
	requestURL string,
) (response.SuccessResponse[dto.HistoricalIntelligenceAggregateHistory], error) {
	result, body, err := executeGET(app, requestURL)
	if err != nil {
		return response.SuccessResponse[dto.HistoricalIntelligenceAggregateHistory]{}, err
	}
	if result.StatusCode != fiber.StatusOK {
		return response.SuccessResponse[dto.HistoricalIntelligenceAggregateHistory]{},
			fmt.Errorf("status = %d body = %s", result.StatusCode, body)
	}

	var payload response.SuccessResponse[dto.HistoricalIntelligenceAggregateHistory]
	if err := json.Unmarshal(body, &payload); err != nil {
		return payload, fmt.Errorf("decode history response: %w", err)
	}
	return payload, nil
}

func expectError(
	app *fiber.App,
	requestURL string,
	expectedStatus int,
	expectedCode string,
) error {
	result, body, err := executeGET(app, requestURL)
	if err != nil {
		return err
	}
	if result.StatusCode != expectedStatus {
		return fmt.Errorf(
			"error status = %d, want %d, body = %s",
			result.StatusCode,
			expectedStatus,
			body,
		)
	}

	var payload response.ErrorResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("decode error response: %w", err)
	}
	if payload.Success || payload.Error.Code != expectedCode {
		return fmt.Errorf("unexpected error response: %#v", payload)
	}

	return nil
}

func executeGET(
	app *fiber.App,
	requestURL string,
) (*http.Response, []byte, error) {
	result, err := app.Test(
		httptest.NewRequest(http.MethodGet, requestURL, nil),
		-1,
	)
	if err != nil {
		return nil, nil, err
	}
	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, nil, err
	}
	return result, body, nil
}

type rowQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func countVerificationRows(
	ctx context.Context,
	querier rowQuerier,
	fingerprints []string,
) (int, error) {
	var count int
	err := querier.QueryRow(
		ctx,
		`
            SELECT count(*)::integer
            FROM historical_aggregate_results
            WHERE input_fingerprint = ANY($1::text[]);
        `,
		fingerprints,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
