package postgres

import (
	"os"
	"strings"
	"testing"
)

func TestTrajectoryReadQueriesHaveOneCanonicalOwner(t *testing.T) {
	t.Parallel()

	querySource := mustReadTrajectoryRepositorySource(
		t,
		"trajectory_read_queries.go",
	)
	for _, required := range []string{
		"const flightTrajectorySelectColumns",
		"const latestTrajectoryByICAO24Query",
		"const trajectoryByIDQuery",
		"const trajectoriesByEndTimeQuery",
		"const trajectoriesByIDsQuery",
		"const trajectoriesByEndTimeAndBoundsQuery",
		"const trajectorySegmentsByTrajectoryIDQuery",
		"const coverageGapsByTrajectoryIDQuery",
		"WITH ORDINALITY",
		"FROM unnest($1::uuid[])",
		"ON trajectory.id = requested.id",
	} {
		if !strings.Contains(querySource, required) {
			t.Fatalf("canonical trajectory query owner is missing %q", required)
		}
	}

	for _, fileName := range []string{
		"trajectory_parent_read.go",
		"analytical_trajectory_repository.go",
		"analytical_trajectory_region_repository.go",
		"trajectory_segment_read.go",
		"trajectory_gap_read.go",
	} {
		source := mustReadTrajectoryRepositorySource(t, fileName)
		for _, forbidden := range []string{
			"SELECT\n",
			"FROM flight_trajectories",
			"FROM trajectory_segments",
			"FROM coverage_gaps",
			".Scan(",
			"scanAnalyticalTrajectories(",
		} {
			if strings.Contains(source, forbidden) {
				t.Fatalf("%s regained query or row-mapping ownership %q", fileName, forbidden)
			}
		}
	}
}

func TestTrajectoryRowMappingHasDedicatedOwners(t *testing.T) {
	t.Parallel()

	expected := map[string][]string{
		"trajectory_row_scan.go": {
			"type postgresRowScanner interface",
			"func scanFlightTrajectory(",
			") queryFlightTrajectory(",
			"func scanFlightTrajectoryRows(",
		},
		"trajectory_segment_row_scan.go": {
			"func scanTrajectorySegment(",
			"func scanTrajectorySegmentRows(",
		},
		"trajectory_gap_row_scan.go": {
			"func scanCoverageGap(",
			"func scanCoverageGapRows(",
		},
	}
	for fileName, requiredTokens := range expected {
		source := mustReadTrajectoryRepositorySource(t, fileName)
		if !strings.Contains(source, ".Scan(") {
			t.Fatalf("%s does not own row scanning", fileName)
		}
		for _, required := range requiredTokens {
			if !strings.Contains(source, required) {
				t.Fatalf("%s is missing %q", fileName, required)
			}
		}
	}
}

func TestAllTrajectoryReadBoundariesPreserveCallerContext(t *testing.T) {
	t.Parallel()

	for _, fileName := range []string{
		"trajectory_read_repository.go",
		"trajectory_read_snapshot.go",
		"analytical_trajectory_repository.go",
		"analytical_trajectory_region_repository.go",
		"trajectory_segment_read.go",
		"trajectory_gap_read.go",
	} {
		source := mustReadTrajectoryRepositorySource(t, fileName)
		if !strings.Contains(source, "requireRepositoryContext(ctx)") {
			t.Fatalf("%s does not require caller context", fileName)
		}
		if strings.Contains(source, "ctx = context.Background()") {
			t.Fatalf("%s still invents caller context", fileName)
		}
	}
}

func TestTrajectoryProfileIndexesMatchProductionOrdering(t *testing.T) {
	t.Parallel()

	migration, err := os.ReadFile(
		"../../../../../database/migrations/021_trajectory_query_profiles.sql",
	)
	if err != nil {
		t.Fatalf("read trajectory profile migration: %v", err)
	}
	migrationSource := string(migration)
	compactMigration := strings.Join(strings.Fields(migrationSource), " ")
	for _, required := range []string{
		"DROP INDEX trajectory_segments_trajectory_sequence_idx",
		"CREATE INDEX flight_trajectories_icao24_latest_idx ON flight_trajectories ( icao24, end_time DESC, start_time DESC, created_at DESC )",
		"CREATE INDEX flight_trajectories_end_time_order_idx ON flight_trajectories ( end_time DESC, start_time DESC, created_at DESC )",
	} {
		if !strings.Contains(compactMigration, required) {
			t.Fatalf("trajectory profile migration is missing %q", required)
		}
	}
	for _, forbidden := range []string{
		"coverage_gaps_trajectory_start_idx",
		"CREATE INDEX trajectory_segments_trajectory_sequence_idx",
	} {
		if strings.Contains(compactMigration, forbidden) {
			t.Fatalf("trajectory profile migration contains redundant index %q", forbidden)
		}
	}

	querySource := mustReadTrajectoryRepositorySource(t, "trajectory_read_queries.go")
	for _, required := range []string{
		"trajectory.end_time DESC",
		"trajectory.start_time DESC",
		"trajectory.created_at DESC",
		"segment.sequence_number ASC",
		"gap.gap_start_time ASC",
	} {
		if !strings.Contains(querySource, required) {
			t.Fatalf("production trajectory query ordering is missing %q", required)
		}
	}
}
