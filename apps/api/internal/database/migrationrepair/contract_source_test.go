package migrationrepair

import (
	"os"
	"strings"
	"testing"
)

func TestMigrationRepairDerivesHistoryBoundaryFromPlan(t *testing.T) {
	t.Parallel()

	postgresSource, err := os.ReadFile("postgres.go")
	if err != nil {
		t.Fatal(err)
	}
	verifierSource, err := os.ReadFile("verifier.go")
	if err != nil {
		t.Fatal(err)
	}
	contractsSource, err := os.ReadFile("contracts.go")
	if err != nil {
		t.Fatal(err)
	}
	planSource, err := os.ReadFile("plan.go")
	if err != nil {
		t.Fatal(err)
	}
	combined := string(postgresSource) + string(verifierSource) + string(contractsSource) + string(planSource)

	for _, forbidden := range []string{
		"WHERE version IN ('010', '011', '012')",
		"ExpectedAppliedVersion010Checksum",
		"CheckAppliedVersion010Exact",
		"CheckFutureVersionsUnapplied",
		"future_versions_011_012_unapplied",
	} {
		if strings.Contains(combined, forbidden) {
			t.Fatalf("migration repair still contains hard-coded sequence fragment %q", forbidden)
		}
	}
	for _, required := range []string{
		"LoadPlan(",
		"plan.Anchor.Version",
		"plan.IsLaterVersion(",
		"WHERE version >= $1",
	} {
		if !strings.Contains(combined, required) {
			t.Fatalf("migration repair is missing generalized fragment %q", required)
		}
	}
}
