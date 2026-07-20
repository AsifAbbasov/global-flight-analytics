package postgres

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const trajectorySnapshotTestDatabaseURL = "TEST_DATABASE_URL"

var trajectorySnapshotSchemaCounter uint64

type trajectorySnapshotFixture struct {
	pool       *pgxpool.Pool
	repository *TrajectoryRepository
}

func TestTrajectoryReadSnapshotRemainsStableAcrossConcurrentCommit(t *testing.T) {
	fixture := newTrajectorySnapshotFixture(t)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	_, err := fixture.repository.withTrajectoryReadSnapshot(
		ctx,
		func(
			snapshotRepository *TrajectoryRepository,
		) (trajectory.FlightTrajectory, error) {
			var countBefore int
			if err := snapshotRepository.trajectoryReadExecutor().QueryRow(
				ctx,
				"SELECT COUNT(*) FROM trajectory_snapshot_probe",
			).Scan(&countBefore); err != nil {
				return trajectory.FlightTrajectory{}, fmt.Errorf(
					"read initial snapshot count: %w",
					err,
				)
			}
			if countBefore != 1 {
				return trajectory.FlightTrajectory{}, fmt.Errorf(
					"expected initial snapshot count 1, got %d",
					countBefore,
				)
			}

			if _, err := fixture.pool.Exec(
				ctx,
				"INSERT INTO trajectory_snapshot_probe (value) VALUES ('concurrent')",
			); err != nil {
				return trajectory.FlightTrajectory{}, fmt.Errorf(
					"commit concurrent snapshot mutation: %w",
					err,
				)
			}

			var countAfter int
			if err := snapshotRepository.trajectoryReadExecutor().QueryRow(
				ctx,
				"SELECT COUNT(*) FROM trajectory_snapshot_probe",
			).Scan(&countAfter); err != nil {
				return trajectory.FlightTrajectory{}, fmt.Errorf(
					"read repeated snapshot count: %w",
					err,
				)
			}
			if countAfter != 1 {
				return trajectory.FlightTrajectory{}, fmt.Errorf(
					"repeatable-read snapshot changed from 1 to %d after concurrent commit",
					countAfter,
				)
			}

			return trajectory.FlightTrajectory{}, nil
		},
	)
	if err != nil {
		t.Fatalf("execute trajectory read snapshot: %v", err)
	}

	var committedCount int
	if err := fixture.pool.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM trajectory_snapshot_probe",
	).Scan(&committedCount); err != nil {
		t.Fatalf("load committed probe count: %v", err)
	}
	if committedCount != 2 {
		t.Fatalf(
			"expected concurrent mutation to be committed outside the snapshot, got %d rows",
			committedCount,
		)
	}
}

func newTrajectorySnapshotFixture(
	t *testing.T,
) *trajectorySnapshotFixture {
	t.Helper()

	databaseURL := strings.TrimSpace(
		os.Getenv(trajectorySnapshotTestDatabaseURL),
	)
	if databaseURL == "" {
		t.Skipf(
			"%s is not set; skipping PostgreSQL integration test",
			trajectorySnapshotTestDatabaseURL,
		)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	bootstrap, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to PostgreSQL test database: %v", err)
	}

	schemaName := fmt.Sprintf(
		"trajectory_snapshot_test_%d_%d",
		time.Now().UnixNano(),
		atomic.AddUint64(&trajectorySnapshotSchemaCounter, 1),
	)
	quotedSchema := pgx.Identifier{schemaName}.Sanitize()

	if _, err := bootstrap.Exec(
		ctx,
		"CREATE SCHEMA "+quotedSchema,
	); err != nil {
		_ = bootstrap.Close(ctx)
		t.Fatalf("create trajectory snapshot test schema: %v", err)
	}

	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		_ = bootstrap.Close(ctx)
		t.Fatalf("parse PostgreSQL test pool config: %v", err)
	}
	if poolConfig.ConnConfig.RuntimeParams == nil {
		poolConfig.ConnConfig.RuntimeParams = make(map[string]string)
	}
	poolConfig.ConnConfig.RuntimeParams["search_path"] = schemaName
	poolConfig.MaxConns = 4

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		_ = bootstrap.Close(ctx)
		t.Fatalf("create trajectory snapshot test pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		_ = bootstrap.Close(ctx)
		t.Fatalf("ping trajectory snapshot test pool: %v", err)
	}

	if _, err := pool.Exec(
		ctx,
		`CREATE TABLE trajectory_snapshot_probe (
			id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			value text NOT NULL
		)`,
	); err != nil {
		pool.Close()
		_ = bootstrap.Close(ctx)
		t.Fatalf("create trajectory snapshot probe table: %v", err)
	}
	if _, err := pool.Exec(
		ctx,
		"INSERT INTO trajectory_snapshot_probe (value) VALUES ('initial')",
	); err != nil {
		pool.Close()
		_ = bootstrap.Close(ctx)
		t.Fatalf("insert initial trajectory snapshot probe: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()

		cleanupCtx, cleanupCancel := context.WithTimeout(
			context.Background(),
			30*time.Second,
		)
		defer cleanupCancel()

		if _, err := bootstrap.Exec(
			cleanupCtx,
			"DROP SCHEMA IF EXISTS "+quotedSchema+" CASCADE",
		); err != nil {
			t.Errorf("drop trajectory snapshot test schema: %v", err)
		}
		if err := bootstrap.Close(cleanupCtx); err != nil {
			t.Errorf("close trajectory snapshot bootstrap connection: %v", err)
		}
	})

	return &trajectorySnapshotFixture{
		pool:       pool,
		repository: NewTrajectoryRepository(pool),
	}
}
