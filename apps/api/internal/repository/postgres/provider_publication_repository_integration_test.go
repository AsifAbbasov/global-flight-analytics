package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const providerPublicationTestDatabaseURL = "TEST_DATABASE_URL"

var providerPublicationSchemaCounter uint64

func TestProviderPublicationRepositoryLifecycle(t *testing.T) {
	fixture := newProviderPublicationFixture(t)
	currentTime := time.Date(2026, time.July, 23, 1, 0, 0, 0, time.UTC)
	repository := NewProviderPublicationRepository(
		fixture.pool,
		10*time.Minute,
		func() time.Time { return currentTime },
	)

	first, err := repository.ReservePublication(
		context.Background(),
		providerpolicy.ProviderOurAirports,
		"sha256:publication-a",
	)
	if err != nil {
		t.Fatalf("reserve first publication: %v", err)
	}
	if !first.Decision.Allowed {
		t.Fatal("expected first reservation to be allowed")
	}

	inProgress, err := repository.ReservePublication(
		context.Background(),
		providerpolicy.ProviderOurAirports,
		"sha256:publication-a",
	)
	if err != nil {
		t.Fatalf("reserve active publication: %v", err)
	}
	if inProgress.Decision.Allowed ||
		inProgress.Decision.Reason != providerbudget.DecisionReasonPublicationInProgress {
		t.Fatalf("unexpected active decision: %+v", inProgress.Decision)
	}
	if !inProgress.Decision.RetryAt.Equal(currentTime.Add(10 * time.Minute)) {
		t.Fatalf("retry at = %s", inProgress.Decision.RetryAt)
	}

	if err := repository.ReleasePublication(context.Background(), first); err != nil {
		t.Fatalf("release first publication: %v", err)
	}

	retry, err := repository.ReservePublication(
		context.Background(),
		providerpolicy.ProviderOurAirports,
		"sha256:publication-a",
	)
	if err != nil {
		t.Fatalf("reserve publication retry: %v", err)
	}
	if !retry.Decision.Allowed || retry.Token == first.Token {
		t.Fatalf("unexpected retry reservation: %+v", retry)
	}

	if err := repository.CommitPublication(context.Background(), retry); err != nil {
		t.Fatalf("commit publication: %v", err)
	}
	if err := repository.CommitPublication(context.Background(), retry); err != nil {
		t.Fatalf("repeat publication commit must be idempotent: %v", err)
	}

	duplicate, err := repository.ReservePublication(
		context.Background(),
		providerpolicy.ProviderOurAirports,
		"sha256:publication-a",
	)
	if err != nil {
		t.Fatalf("reserve committed publication: %v", err)
	}
	if duplicate.Decision.Allowed ||
		duplicate.Decision.Reason != providerbudget.DecisionReasonPublicationAlreadyProcessed {
		t.Fatalf("unexpected committed decision: %+v", duplicate.Decision)
	}
}

func TestProviderPublicationRepositoryReclaimsExpiredLease(t *testing.T) {
	fixture := newProviderPublicationFixture(t)
	currentTime := time.Date(2026, time.July, 23, 2, 0, 0, 0, time.UTC)
	repository := NewProviderPublicationRepository(
		fixture.pool,
		time.Minute,
		func() time.Time { return currentTime },
	)

	first, err := repository.ReservePublication(
		context.Background(),
		providerpolicy.ProviderOurAirports,
		"sha256:stale",
	)
	if err != nil {
		t.Fatalf("reserve publication: %v", err)
	}

	currentTime = currentTime.Add(time.Minute)
	reclaimed, err := repository.ReservePublication(
		context.Background(),
		providerpolicy.ProviderOurAirports,
		"sha256:stale",
	)
	if err != nil {
		t.Fatalf("reclaim publication: %v", err)
	}
	if !reclaimed.Decision.Allowed || reclaimed.Token == first.Token {
		t.Fatalf("unexpected reclaimed reservation: %+v", reclaimed)
	}

	if err := repository.CommitPublication(context.Background(), first); err == nil {
		t.Fatal("expected expired owner commit to be rejected")
	}
}

func TestProviderPublicationRepositorySerializesConcurrentReservation(t *testing.T) {
	fixture := newProviderPublicationFixture(t)
	currentTime := time.Date(2026, time.July, 23, 3, 0, 0, 0, time.UTC)
	repository := NewProviderPublicationRepository(
		fixture.pool,
		10*time.Minute,
		func() time.Time { return currentTime },
	)

	const workers = 12
	start := make(chan struct{})
	results := make(chan providerbudget.PublicationReservation, workers)
	errorsChannel := make(chan error, workers)
	var waitGroup sync.WaitGroup
	waitGroup.Add(workers)
	for index := 0; index < workers; index++ {
		go func() {
			defer waitGroup.Done()
			<-start
			reservation, err := repository.ReservePublication(
				context.Background(),
				providerpolicy.ProviderOurAirports,
				"sha256:concurrent",
			)
			if err != nil {
				errorsChannel <- err
				return
			}
			results <- reservation
		}()
	}
	close(start)
	waitGroup.Wait()
	close(results)
	close(errorsChannel)

	for err := range errorsChannel {
		t.Fatalf("concurrent reserve publication: %v", err)
	}

	allowed := 0
	inProgress := 0
	for result := range results {
		if result.Decision.Allowed {
			allowed++
			continue
		}
		if result.Decision.Reason == providerbudget.DecisionReasonPublicationInProgress {
			inProgress++
		}
	}
	if allowed != 1 || inProgress != workers-1 {
		t.Fatalf(
			"allowed=%d in_progress=%d, want 1 and %d",
			allowed,
			inProgress,
			workers-1,
		)
	}
}

type providerPublicationFixture struct {
	pool *pgxpool.Pool
}

func newProviderPublicationFixture(t *testing.T) *providerPublicationFixture {
	t.Helper()

	databaseURL := strings.TrimSpace(os.Getenv(providerPublicationTestDatabaseURL))
	if databaseURL == "" {
		t.Skipf(
			"%s is not set; skipping PostgreSQL integration test",
			providerPublicationTestDatabaseURL,
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bootstrap, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect PostgreSQL test database: %v", err)
	}

	schemaName := fmt.Sprintf(
		"provider_publication_test_%d_%d",
		time.Now().UnixNano(),
		atomic.AddUint64(&providerPublicationSchemaCounter, 1),
	)
	quotedSchema := pgx.Identifier{schemaName}.Sanitize()
	if _, err := bootstrap.Exec(ctx, "CREATE SCHEMA "+quotedSchema); err != nil {
		_ = bootstrap.Close(ctx)
		t.Fatalf("create PostgreSQL test schema: %v", err)
	}

	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		_ = bootstrap.Close(ctx)
		t.Fatalf("parse PostgreSQL pool config: %v", err)
	}
	if poolConfig.ConnConfig.RuntimeParams == nil {
		poolConfig.ConnConfig.RuntimeParams = make(map[string]string)
	}
	poolConfig.ConnConfig.RuntimeParams["search_path"] = schemaName

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		_ = bootstrap.Close(ctx)
		t.Fatalf("create PostgreSQL test pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		_ = bootstrap.Close(ctx)
		t.Fatalf("ping PostgreSQL test pool: %v", err)
	}

	applyProviderPublicationMigration(t, pool)

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
			t.Errorf("drop PostgreSQL test schema: %v", err)
		}
		if err := bootstrap.Close(cleanupContext); err != nil {
			t.Errorf("close PostgreSQL bootstrap connection: %v", err)
		}
	})

	return &providerPublicationFixture{pool: pool}
}

func applyProviderPublicationMigration(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve provider publication integration test path")
	}
	migrationPath := filepath.Clean(filepath.Join(
		filepath.Dir(currentFile),
		"../../../../../database/migrations/022_provider_publication_lifecycle.sql",
	))
	migrationBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read provider publication migration: %v", err)
	}
	if _, err := pool.Exec(context.Background(), string(migrationBytes)); err != nil {
		t.Fatalf("apply provider publication migration: %v", err)
	}
}
