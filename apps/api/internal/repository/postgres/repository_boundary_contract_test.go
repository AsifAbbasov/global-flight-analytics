package postgres

import (
	"os"
	"strings"
	"testing"
)

func TestRepositoryOperationsDoNotInventCallerContext(t *testing.T) {
	t.Parallel()

	for _, fileName := range []string{
		"airport_import_repository.go",
		"airport_repository.go",
		"airport_pagination_read.go",
		"flightstate_repository.go",
		"trajectory_write_repository.go",
	} {
		source := readRepositoryBoundarySource(t, fileName)
		if !strings.Contains(source, "requireRepositoryContext(ctx)") {
			t.Fatalf("%s does not enforce the caller-owned context contract", fileName)
		}
		if strings.Contains(source, "ctx = context.Background()") {
			t.Fatalf("%s still invents a replacement caller context", fileName)
		}
	}
}

func TestTrajectoryWriteCoordinatorUsesExplicitMode(t *testing.T) {
	t.Parallel()

	source := readRepositoryBoundarySource(t, "trajectory_write_repository.go")
	for _, required := range []string{
		"newLiveTrajectoryWriteRequest(item)",
		"newReconciledTrajectoryWriteRequest(",
		"request.isReconciled()",
		"switch request.mode",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("trajectory write coordinator is missing %q", required)
		}
	}
	for _, forbidden := range []string{
		`saveTrajectory(ctx, "",`,
		`reconciliationTaskID == ""`,
		`reconciliationTaskID != ""`,
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("trajectory write coordinator still uses hidden mode sentinel %q", forbidden)
		}
	}
}

func readRepositoryBoundarySource(t *testing.T, fileName string) string {
	t.Helper()
	content, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatalf("read %s: %v", fileName, err)
	}
	return string(content)
}
