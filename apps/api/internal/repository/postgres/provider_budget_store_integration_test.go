package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var providerBudgetSchemaCounter uint64

func TestProviderBudgetStoreSharesStateAcrossInstances(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping PostgreSQL integration test")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	bootstrap, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect PostgreSQL: %v", err)
	}

	schemaName := fmt.Sprintf(
		"provider_budget_%d_%d",
		time.Now().UnixNano(),
		atomic.AddUint64(&providerBudgetSchemaCounter, 1),
	)
	quotedSchema := pgx.Identifier{schemaName}.Sanitize()
	if _, err := bootstrap.Exec(
		ctx,
		"CREATE SCHEMA "+quotedSchema,
	); err != nil {
		_ = bootstrap.Close(ctx)
		t.Fatalf("create schema: %v", err)
	}

	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		_ = bootstrap.Close(ctx)
		t.Fatalf("parse pool config: %v", err)
	}
	if poolConfig.ConnConfig.RuntimeParams == nil {
		poolConfig.ConnConfig.RuntimeParams = make(map[string]string)
	}
	poolConfig.ConnConfig.RuntimeParams["search_path"] = schemaName
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		_ = bootstrap.Close(ctx)
		t.Fatalf("create pool: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
		cleanupContext, cleanupCancel := context.WithTimeout(
			context.Background(),
			30*time.Second,
		)
		defer cleanupCancel()
		if _, err := bootstrap.Exec(
			cleanupContext,
			"DROP SCHEMA IF EXISTS "+quotedSchema+" CASCADE",
		); err != nil {
			t.Errorf("drop schema: %v", err)
		}
		if err := bootstrap.Close(cleanupContext); err != nil {
			t.Errorf("close bootstrap connection: %v", err)
		}
	})

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve provider budget integration test path")
	}
	migrationPath := filepath.Clean(filepath.Join(
		filepath.Dir(currentFile),
		"../../../../../database/migrations/024_provider_budget_durability.sql",
	))
	migrationBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read provider budget migration: %v", err)
	}
	if _, err := pool.Exec(ctx, string(migrationBytes)); err != nil {
		t.Fatalf("apply provider budget migration: %v", err)
	}

	firstStore, err := NewProviderBudgetStore(pool, 5*time.Second)
	if err != nil {
		t.Fatalf("create first store: %v", err)
	}
	secondStore, err := NewProviderBudgetStore(pool, 5*time.Second)
	if err != nil {
		t.Fatalf("create second store: %v", err)
	}

	now := time.Date(
		2026,
		time.July,
		23,
		16,
		0,
		0,
		0,
		time.UTC,
	)
	reservations := []providerbudget.FixedWindowReservation{
		{
			LimitIndex:  0,
			WindowStart: now,
			WindowEnd:   now.Add(time.Second),
			MaxRequests: 1,
		},
	}

	firstDecision, err := firstStore.AcquireFixedWindow(
		providerpolicy.ProviderAirplanesLive,
		reservations,
		now,
	)
	if err != nil || !firstDecision.Allowed {
		t.Fatalf(
			"first fixed-window decision=%+v err=%v",
			firstDecision,
			err,
		)
	}
	secondDecision, err := secondStore.AcquireFixedWindow(
		providerpolicy.ProviderAirplanesLive,
		reservations,
		now,
	)
	if err != nil {
		t.Fatalf("second fixed-window acquire: %v", err)
	}
	if secondDecision.Allowed ||
		!secondDecision.RetryAt.Equal(now.Add(time.Second)) {
		t.Fatalf(
			"unexpected shared fixed-window decision: %+v",
			secondDecision,
		)
	}

	nextWindow := []providerbudget.FixedWindowReservation{
		{
			LimitIndex:  0,
			WindowStart: now.Add(time.Second),
			WindowEnd:   now.Add(2 * time.Second),
			MaxRequests: 1,
		},
	}
	nextDecision, err := secondStore.AcquireFixedWindow(
		providerpolicy.ProviderAirplanesLive,
		nextWindow,
		now.Add(time.Second),
	)
	if err != nil || !nextDecision.Allowed {
		t.Fatalf(
			"next-window decision=%+v err=%v",
			nextDecision,
			err,
		)
	}

	var fixedWindowRowCount int
	if err := pool.QueryRow(
		ctx,
		`
            SELECT COUNT(*)
            FROM provider_budget_fixed_windows
            WHERE provider_name = $1
                AND limit_index = 0
        `,
		string(providerpolicy.ProviderAirplanesLive),
	).Scan(&fixedWindowRowCount); err != nil {
		t.Fatalf("count fixed-window rows: %v", err)
	}
	if fixedWindowRowCount != 1 {
		t.Fatalf(
			"fixed-window row count = %d, want 1",
			fixedWindowRowCount,
		)
	}

	initialProbe, err := firstStore.AcquireProviderReported(
		providerpolicy.ProviderOpenSky,
		now,
		time.Minute,
	)
	if err != nil || !initialProbe.Allowed {
		t.Fatalf(
			"initial provider probe=%+v err=%v",
			initialProbe,
			err,
		)
	}
	duplicateProbe, err := secondStore.AcquireProviderReported(
		providerpolicy.ProviderOpenSky,
		now,
		time.Minute,
	)
	if err != nil {
		t.Fatalf("duplicate provider probe: %v", err)
	}
	if duplicateProbe.Allowed ||
		duplicateProbe.Reason != providerbudget.DecisionReasonProviderCooldown ||
		!duplicateProbe.RetryAt.Equal(now.Add(time.Minute)) {
		t.Fatalf(
			"unexpected duplicate provider probe: %+v",
			duplicateProbe,
		)
	}

	observedAt := now.Add(time.Second)
	if err := firstStore.ObserveProviderReportedBudget(
		providerpolicy.ProviderOpenSky,
		1,
		time.Time{},
		observedAt,
	); err != nil {
		t.Fatalf("observe provider-reported budget: %v", err)
	}

	reportedFirst, err := secondStore.AcquireProviderReported(
		providerpolicy.ProviderOpenSky,
		observedAt,
		time.Minute,
	)
	if err != nil || !reportedFirst.Allowed {
		t.Fatalf(
			"first reported decision=%+v err=%v",
			reportedFirst,
			err,
		)
	}
	reportedSecond, err := firstStore.AcquireProviderReported(
		providerpolicy.ProviderOpenSky,
		observedAt,
		time.Minute,
	)
	if err != nil {
		t.Fatalf("second reported acquire: %v", err)
	}
	if reportedSecond.Allowed ||
		reportedSecond.Reason != providerbudget.DecisionReasonProviderBudgetExhausted ||
		!reportedSecond.RetryAt.Equal(observedAt.Add(time.Minute)) {
		t.Fatalf(
			"unexpected shared reported decision: %+v",
			reportedSecond,
		)
	}
}
