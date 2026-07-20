package postgres

import (
	"os"
	"strings"
	"testing"
)

func TestTrajectoryRepositoryCoordinatorsRemainNarrow(t *testing.T) {
	t.Parallel()

	writeCoordinator := mustReadTrajectoryRepositorySource(
		t,
		"trajectory_repository.go",
	)
	if lines := sourceLineCount(writeCoordinator); lines > 40 {
		t.Fatalf(
			"trajectory_repository.go grew beyond its constructor responsibility: %d lines",
			lines,
		)
	}
	forbiddenWriteTokens := []string{
		"SaveTrajectory(",
		"SaveReconciledTrajectory(",
		"INSERT INTO",
		"DELETE FROM",
		"BeginTx(",
	}
	for _, token := range forbiddenWriteTokens {
		if strings.Contains(writeCoordinator, token) {
			t.Fatalf(
				"trajectory_repository.go regained write responsibility %q",
				token,
			)
		}
	}

	readCoordinator := mustReadTrajectoryRepositorySource(
		t,
		"trajectory_read_repository.go",
	)
	if lines := sourceLineCount(readCoordinator); lines > 110 {
		t.Fatalf(
			"trajectory_read_repository.go grew beyond its snapshot-coordinator responsibility: %d lines",
			lines,
		)
	}
	if count := strings.Count(
		readCoordinator,
		"repository.withTrajectoryReadSnapshot(",
	); count != 2 {
		t.Fatalf(
			"expected two public snapshot-bound trajectory reads, got %d",
			count,
		)
	}
	forbiddenReadTokens := []string{
		"SELECT",
		"FROM flight_trajectories",
		"FROM trajectory_segments",
		"FROM coverage_gaps",
		".Scan(",
	}
	for _, token := range forbiddenReadTokens {
		if strings.Contains(readCoordinator, token) {
			t.Fatalf(
				"trajectory_read_repository.go regained SQL or mapping responsibility %q",
				token,
			)
		}
	}
}

func TestTrajectoryRepositoryResponsibilitiesHaveDedicatedOwners(t *testing.T) {
	t.Parallel()

	expected := map[string][]string{
		"trajectory_write_repository.go": {
			") SaveTrajectory(",
			") SaveReconciledTrajectory(",
			") saveTrajectory(",
			"validateTrajectoryRelationalIntegrity(item)",
			"repository.db.BeginTx(",
		},
		"trajectory_reconciliation_write.go": {
			"func assertReconciliationAttemptOwned(",
			"func deleteExistingReconciledTrajectory(",
		},
		"trajectory_parent_write.go": {
			") insertFlightTrajectory(",
			") insertReconciledFlightTrajectory(",
		},
		"trajectory_segment_write.go": {
			") insertTrajectorySegments(",
		},
		"trajectory_gap_write.go": {
			") insertCoverageGaps(",
			"func inferredPreviousSegmentID(",
			"func inferredNextSegmentID(",
		},
		"trajectory_parent_read.go": {
			") getLatestTrajectoryByICAO24(",
			") getTrajectoryByID(",
			") queryTrajectory(",
			"func normalizeICAO24Lookup(",
		},
		"trajectory_child_read.go": {
			") loadTrajectoryChildren(",
		},
		"trajectory_segment_read.go": {
			") ListTrajectorySegments(",
		},
		"trajectory_gap_read.go": {
			") ListCoverageGaps(",
		},
	}

	for fileName, tokens := range expected {
		source := mustReadTrajectoryRepositorySource(t, fileName)
		for _, token := range tokens {
			if !strings.Contains(source, token) {
				t.Fatalf("%s does not own %q", fileName, token)
			}
		}
	}
}

func TestTrajectoryWriteValidationStillPrecedesTransaction(t *testing.T) {
	t.Parallel()

	source := mustReadTrajectoryRepositorySource(
		t,
		"trajectory_write_repository.go",
	)
	validationIndex := strings.Index(
		source,
		"validateTrajectoryRelationalIntegrity(item)",
	)
	transactionIndex := strings.Index(
		source,
		"repository.db.BeginTx(",
	)
	if validationIndex < 0 || transactionIndex < 0 {
		t.Fatal("write coordinator is missing validation or transaction boundary")
	}
	if validationIndex > transactionIndex {
		t.Fatal("trajectory validation moved after transaction creation")
	}
}

func mustReadTrajectoryRepositorySource(
	t *testing.T,
	fileName string,
) string {
	t.Helper()

	content, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatalf("read %s: %v", fileName, err)
	}

	return string(content)
}

func sourceLineCount(source string) int {
	trimmed := strings.TrimRight(source, "\n")
	if trimmed == "" {
		return 0
	}
	return strings.Count(trimmed, "\n") + 1
}
