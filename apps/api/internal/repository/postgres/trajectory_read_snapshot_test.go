package postgres

import (
	"os"
	"strings"
	"testing"
)

func TestPublicTrajectoryReadsUseSnapshotBoundary(t *testing.T) {
	t.Parallel()

	source := mustReadTrajectorySnapshotSource(
		t,
		"trajectory_read_repository.go",
	)

	if count := strings.Count(
		source,
		"repository.withTrajectoryReadSnapshot(",
	); count != 2 {
		t.Fatalf(
			"expected both public trajectory aggregate reads to use the snapshot boundary, got %d uses",
			count,
		)
	}

	if !strings.Contains(
		source,
		"snapshotRepository.getLatestTrajectoryByICAO24",
	) {
		t.Fatal("latest trajectory read does not execute through the snapshot repository")
	}
	if !strings.Contains(
		source,
		"snapshotRepository.getTrajectoryByID",
	) {
		t.Fatal("trajectory-by-id read does not execute through the snapshot repository")
	}
}

func TestTrajectorySnapshotUsesReadOnlyRepeatableReadTransaction(t *testing.T) {
	t.Parallel()

	source := mustReadTrajectorySnapshotSource(
		t,
		"trajectory_read_snapshot.go",
	)

	requiredTokens := []string{
		"repository.db.BeginTx(",
		"IsoLevel:   pgx.RepeatableRead",
		"AccessMode: pgx.ReadOnly",
		"NewTrajectoryReadRepository(tx)",
		"tx.Commit(ctx)",
	}
	for _, token := range requiredTokens {
		if !strings.Contains(source, token) {
			t.Fatalf(
				"trajectory snapshot boundary is missing %q",
				token,
			)
		}
	}
}

func TestTransactionBoundTrajectoryRepositoryAvoidsNestedTransaction(t *testing.T) {
	t.Parallel()

	source := mustReadTrajectorySnapshotSource(
		t,
		"trajectory_read_snapshot.go",
	)

	if !strings.Contains(
		source,
		"if repository.db == nil {\n\t\treturn operation(repository)\n\t}",
	) {
		t.Fatal("transaction-bound trajectory repository does not reuse its caller-owned snapshot")
	}
}

func TestPoolBackedReadConstructorOwnsSnapshotBoundary(t *testing.T) {
	t.Parallel()

	source := mustReadTrajectorySnapshotSource(
		t,
		"trajectory_read_client.go",
	)

	if !strings.Contains(source, "client.(*pgxpool.Pool)") {
		t.Fatal("pool-backed read constructor does not identify the pool snapshot owner")
	}
	if !strings.Contains(source, "repository.db = pool") {
		t.Fatal("pool-backed read constructor does not retain the pool for snapshot creation")
	}
}

func mustReadTrajectorySnapshotSource(
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
