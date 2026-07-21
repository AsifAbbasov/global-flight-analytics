package postgres

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestTrajectoryQueryProfilesUseExpectedIndexes(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping PostgreSQL query profiling")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
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
		t.Fatalf("begin trajectory query profile transaction: %v", err)
	}
	defer func() {
		rollbackCtx, rollbackCancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer rollbackCancel()
		_ = transaction.Rollback(rollbackCtx)
	}()

	for _, setting := range []string{
		"SET LOCAL enable_seqscan = off",
		"SET LOCAL enable_bitmapscan = off",
	} {
		if _, err := transaction.Exec(ctx, setting); err != nil {
			t.Fatalf("apply planner setting %q: %v", setting, err)
		}
	}

	profileCases := []struct {
		name          string
		query         string
		arguments     []any
		expectedIndex string
	}{
		{
			name:          "latest trajectory by ICAO24",
			query:         latestTrajectoryByICAO24Query,
			arguments:     []any{"A1B2C3"},
			expectedIndex: "flight_trajectories_icao24_latest_idx",
		},
		{
			name:  "analytical trajectories by end time",
			query: trajectoriesByEndTimeQuery,
			arguments: []any{
				time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2026, time.July, 31, 23, 59, 59, 0, time.UTC),
				100,
			},
			expectedIndex: "flight_trajectories_end_time_order_idx",
		},
		{
			name:          "trajectory segments by parent",
			query:         trajectorySegmentsByTrajectoryIDQuery,
			arguments:     []any{"11111111-1111-1111-1111-111111111111"},
			expectedIndex: "trajectory_segments_trajectory_sequence_unique",
		},
		{
			name:          "coverage gaps by parent",
			query:         coverageGapsByTrajectoryIDQuery,
			arguments:     []any{"11111111-1111-1111-1111-111111111111"},
			expectedIndex: "coverage_gaps_trajectory_time_idx",
		},
	}

	for _, profileCase := range profileCases {
		profileCase := profileCase
		t.Run(profileCase.name, func(t *testing.T) {
			plan := explainAnalyzeTrajectoryQuery(
				t,
				ctx,
				transaction,
				profileCase.query,
				profileCase.arguments...,
			)
			t.Logf("%s plan:\n%s", profileCase.name, plan)
			if !strings.Contains(plan, profileCase.expectedIndex) {
				t.Fatalf(
					"query profile did not use %s:\n%s",
					profileCase.expectedIndex,
					plan,
				)
			}
			for _, evidence := range []string{"Planning Time", "Execution Time"} {
				if !strings.Contains(plan, evidence) {
					t.Fatalf("query profile is missing %s evidence:\n%s", evidence, plan)
				}
			}
		})
	}
}

func explainAnalyzeTrajectoryQuery(
	t *testing.T,
	ctx context.Context,
	transaction pgx.Tx,
	query string,
	arguments ...any,
) string {
	t.Helper()

	rows, err := transaction.Query(
		ctx,
		"EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT) "+query,
		arguments...,
	)
	if err != nil {
		t.Fatalf("execute EXPLAIN ANALYZE: %v", err)
	}
	defer rows.Close()

	lines := make([]string, 0)
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatalf("scan EXPLAIN ANALYZE line: %v", err)
		}
		lines = append(lines, line)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate EXPLAIN ANALYZE output: %v", err)
	}
	if len(lines) == 0 {
		t.Fatal("EXPLAIN ANALYZE returned no plan lines")
	}
	return fmt.Sprintf("%s", strings.Join(lines, "\n"))
}
