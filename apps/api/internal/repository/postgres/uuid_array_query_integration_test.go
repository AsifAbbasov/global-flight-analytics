package postgres

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestUUIDArrayMembershipUsesTypedColumnComparison(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping PostgreSQL integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	connection, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to PostgreSQL: %v", err)
	}
	defer func() {
		if closeErr := connection.Close(ctx); closeErr != nil {
			t.Errorf("close PostgreSQL connection: %v", closeErr)
		}
	}()

	transaction, err := connection.Begin(ctx)
	if err != nil {
		t.Fatalf("begin transaction: %v", err)
	}
	defer func() {
		_ = transaction.Rollback(context.Background())
	}()

	if _, err := transaction.Exec(
		ctx,
		`CREATE TEMP TABLE stage14_uuid_membership (id uuid PRIMARY KEY) ON COMMIT DROP`,
	); err != nil {
		t.Fatalf("create temporary UUID table: %v", err)
	}
	if _, err := transaction.Exec(
		ctx,
		`INSERT INTO stage14_uuid_membership (id) VALUES
			('11111111-1111-1111-1111-111111111111'),
			('22222222-2222-2222-2222-222222222222')`,
	); err != nil {
		t.Fatalf("seed temporary UUID table: %v", err)
	}

	var count int
	if err := transaction.QueryRow(
		ctx,
		`SELECT COUNT(*)
		 FROM stage14_uuid_membership
		 WHERE id = ANY (
			SELECT candidate::uuid
			FROM unnest($1::text[]) AS candidates(candidate)
		 )`,
		[]string{
			"11111111-1111-1111-1111-111111111111",
			"22222222-2222-2222-2222-222222222222",
		},
	).Scan(&count); err != nil {
		t.Fatalf("query typed UUID membership: %v", err)
	}
	if count != 2 {
		t.Fatalf("matched row count = %d, want 2", count)
	}
}

func TestRepositoryDriverValuesFailClosedBeforePersistence(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping PostgreSQL integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	connection, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to PostgreSQL: %v", err)
	}
	defer func() {
		if closeErr := connection.Close(ctx); closeErr != nil {
			t.Errorf("close PostgreSQL connection: %v", closeErr)
		}
	}()

	var textValue string
	err = connection.QueryRow(
		ctx,
		`SELECT $1::text`,
		requiredSourceNameValue("   "),
	).Scan(&textValue)
	if !errors.Is(err, ErrRepositorySourceNameRequired) {
		t.Fatalf("expected source-name error, got %v", err)
	}

	err = connection.QueryRow(
		ctx,
		`SELECT $1::uuid::text`,
		nullableUUID("not-a-uuid"),
	).Scan(&textValue)
	if !errors.Is(err, ErrRepositoryUUIDArgumentInvalid) {
		t.Fatalf("expected UUID argument error, got %v", err)
	}
}
