package postgres

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var airportPaginationSchemaCounter uint64

func TestAirportListPageUsesStableDuplicateNameCursor(t *testing.T) {
	pool := newAirportPaginationPool(t)
	createAirportPaginationSchema(t, pool)
	seedAirportPaginationRows(t, pool)

	repository := NewAirportRepository(pool)
	request := airport.ListRequest{Limit: 2}
	collected := make([]string, 0, 5)

	for {
		page, err := repository.ListPage(context.Background(), request)
		if err != nil {
			t.Fatalf("list airport page: %v", err)
		}
		if len(page.Items) > request.Limit {
			t.Fatalf("page returned %d items, limit is %d", len(page.Items), request.Limit)
		}
		for _, item := range page.Items {
			collected = append(collected, item.ICAOCode)
		}
		if page.NextCursor == nil {
			break
		}
		request.Cursor = page.NextCursor
	}

	assertAirportCodes(t, collected, []string{"AAAA", "AAAB", "BBBB", "BBBC", "CCCC"})
}

func TestAirportListLegacyAdapterCollectsBoundedPages(t *testing.T) {
	pool := newAirportPaginationPool(t)
	createAirportPaginationSchema(t, pool)
	seedAirportPaginationRows(t, pool)

	items, err := NewAirportRepository(pool).List(context.Background())
	if err != nil {
		t.Fatalf("list airports: %v", err)
	}
	codes := make([]string, 0, len(items))
	for _, item := range items {
		codes = append(codes, item.ICAOCode)
	}
	assertAirportCodes(t, codes, []string{"AAAA", "AAAB", "BBBB", "BBBC", "CCCC"})
}

func newAirportPaginationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	databaseURL := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping PostgreSQL integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bootstrap, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to PostgreSQL: %v", err)
	}

	schemaName := fmt.Sprintf(
		"airport_pagination_test_%d_%d",
		time.Now().UnixNano(),
		atomic.AddUint64(&airportPaginationSchemaCounter, 1),
	)
	quotedSchema := pgx.Identifier{schemaName}.Sanitize()
	if _, err := bootstrap.Exec(ctx, "CREATE SCHEMA "+quotedSchema); err != nil {
		_ = bootstrap.Close(ctx)
		t.Fatalf("create schema: %v", err)
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		_ = bootstrap.Close(ctx)
		t.Fatalf("parse pool config: %v", err)
	}
	if config.ConnConfig.RuntimeParams == nil {
		config.ConnConfig.RuntimeParams = make(map[string]string)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schemaName

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		_ = bootstrap.Close(ctx)
		t.Fatalf("create pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		_ = bootstrap.Close(ctx)
		t.Fatalf("ping pool: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if _, err := bootstrap.Exec(cleanupCtx, "DROP SCHEMA IF EXISTS "+quotedSchema+" CASCADE"); err != nil {
			t.Errorf("drop schema: %v", err)
		}
		if err := bootstrap.Close(cleanupCtx); err != nil {
			t.Errorf("close bootstrap connection: %v", err)
		}
	})
	return pool
}

func createAirportPaginationSchema(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	mustExecAirportPaginationSQL(
		t,
		pool,
		`CREATE TABLE countries (
			id uuid PRIMARY KEY,
			iso2 text,
			name text NOT NULL
		)`,
		`CREATE TABLE airports (
			id uuid PRIMARY KEY,
			icao_code text,
			iata_code text,
			name text NOT NULL,
			city text,
			country_id uuid REFERENCES countries(id),
			latitude double precision NOT NULL,
			longitude double precision NOT NULL,
			elevation_ft integer,
			timezone text
		)`,
		`CREATE TABLE airport_profiles (
			airport_id uuid PRIMARY KEY REFERENCES airports(id),
			description text
		)`,
	)
}

func seedAirportPaginationRows(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	mustExecAirportPaginationSQL(
		t,
		pool,
		`INSERT INTO countries (id, iso2, name)
		VALUES ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'AZ', 'Azerbaijan')`,
		`INSERT INTO airports (
			id, icao_code, iata_code, name, city, country_id,
			latitude, longitude, elevation_ft, timezone
		) VALUES
			('11111111-1111-1111-1111-111111111111', 'AAAA', 'A01', 'Alpha Airport', 'A', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 40.1, 49.1, 10, 'Asia/Baku'),
			('22222222-2222-2222-2222-222222222222', 'AAAB', 'A02', 'Alpha Airport', 'B', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 40.2, 49.2, NULL, 'Asia/Baku'),
			('33333333-3333-3333-3333-333333333333', 'BBBB', 'B01', 'Bravo Airport', 'C', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 40.3, 49.3, 20, 'Asia/Baku'),
			('44444444-4444-4444-4444-444444444444', 'BBBC', 'B02', 'Bravo Airport', 'D', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 40.4, 49.4, 30, 'Asia/Baku'),
			('55555555-5555-5555-5555-555555555555', 'CCCC', 'C01', 'Charlie Airport', 'E', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 40.5, 49.5, 40, 'Asia/Baku')`,
		`INSERT INTO airport_profiles (airport_id, description)
		VALUES ('11111111-1111-1111-1111-111111111111', 'first alpha')`,
	)
}

func mustExecAirportPaginationSQL(
	t *testing.T,
	pool *pgxpool.Pool,
	queries ...string,
) {
	t.Helper()
	for _, query := range queries {
		if _, err := pool.Exec(context.Background(), query); err != nil {
			t.Fatalf("execute airport pagination SQL: %v", err)
		}
	}
}

func assertAirportCodes(t *testing.T, actual []string, expected []string) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Fatalf("airport codes = %#v, want %#v", actual, expected)
	}
	for index := range expected {
		if actual[index] != expected[index] {
			t.Fatalf("airport codes = %#v, want %#v", actual, expected)
		}
	}
}
