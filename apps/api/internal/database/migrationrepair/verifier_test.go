package migrationrepair

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestVerifierAcceptsExactRepairReadyState(t *testing.T) {
	generatedAt := time.Date(2026, time.July, 14, 16, 30, 0, 0, time.UTC)
	plan := newTestPlan(t)
	verifier := newTestVerifier(
		t,
		Config{
			Inspector: fakeInspector{state: readyState(plan)},
			Now:       func() time.Time { return generatedAt },
		},
		plan,
	)

	report, err := verifier.Verify(context.Background())
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if !report.Ready ||
		report.BlockerCount != 0 ||
		report.InfoCount != 9 ||
		len(report.Checks) != 9 ||
		!report.GeneratedAt.Equal(generatedAt) {
		t.Fatalf("unexpected report: %#v", report)
	}
	for _, check := range report.Checks {
		if !check.Passed || check.Severity != SeverityInfo {
			t.Fatalf("unexpected check: %#v", check)
		}
	}
}

func TestVerifierBlocksWrongAppliedAnchorMigration(t *testing.T) {
	plan := newTestPlan(t)
	state := readyState(plan)
	state.AppliedMigrations[0].Checksum = "wrong"

	report := evaluateState(state, plan, time.Time{})
	assertFailedCheck(t, report, CheckAppliedMigrationExact)
}

func TestVerifierBlocksAnyAppliedMigrationLaterThanAnchor(t *testing.T) {
	plan := newTestPlan(t)
	state := readyState(plan)
	state.AppliedMigrations = append(
		state.AppliedMigrations,
		AppliedMigration{
			Version:  "020",
			Name:     "later_migration",
			Checksum: "later-checksum",
		},
	)

	report := evaluateState(state, plan, time.Time{})
	assertFailedCheck(t, report, CheckLaterMigrationsUnapplied)
}

func TestVerifierBlocksMissingReconciliationObjects(t *testing.T) {
	plan := newTestPlan(t)
	state := readyState(plan)
	state.DataQualityReconciliationTaskIDColumnExists = false
	state.FlightTrajectoryReconciliationForeignKeyExists = false
	state.DataQualityReconciliationUniqueIndexExists = false

	report := evaluateState(state, plan, time.Time{})
	assertFailedCheck(t, report, CheckReconciliationColumnsPresent)
	assertFailedCheck(t, report, CheckReconciliationConstraintsPresent)
	assertFailedCheck(t, report, CheckReconciliationIndexesPresent)
}

func TestVerifierBlocksAnyExistingIdentityObject(t *testing.T) {
	plan := newTestPlan(t)
	state := readyState(plan)
	state.IdentityKeyColumnExists = true
	state.SplitReasonCheckExists = true
	state.IdentityKeyTimeIndexExists = true

	report := evaluateState(state, plan, time.Time{})
	assertFailedCheck(t, report, CheckIdentityColumnsAbsent)
	assertFailedCheck(t, report, CheckIdentityConstraintsAbsent)
	assertFailedCheck(t, report, CheckIdentityIndexAbsent)
}

func TestVerifierRequiresCallerContext(t *testing.T) {
	plan := newTestPlan(t)
	verifier := newTestVerifier(t, Config{Inspector: fakeInspector{}}, plan)

	_, err := verifier.Verify(nil)
	if !errors.Is(err, ErrContextRequired) {
		t.Fatalf("Verify(nil) error = %v", err)
	}
}

func TestVerifierPreservesContextCancellation(t *testing.T) {
	plan := newTestPlan(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	verifier := newTestVerifier(t, Config{Inspector: fakeInspector{}}, plan)

	_, err := verifier.Verify(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Verify() error = %v, want context.Canceled", err)
	}
}

func TestReportCloneDoesNotShareChecks(t *testing.T) {
	report := Report{Checks: []Check{{Code: CheckAppliedMigrationExact}}}
	cloned := report.Clone()
	cloned.Checks[0].Code = CheckIdentityColumnsAbsent

	if reflect.DeepEqual(report.Checks, cloned.Checks) {
		t.Fatal("Report.Clone() shared checks")
	}
	if report.Checks[0].Code != CheckAppliedMigrationExact {
		t.Fatal("Report.Clone() mutated original")
	}
}

func readyState(plan Plan) State {
	return State{
		SchemaMigrationsTableExists: true,
		AppliedMigrations: []AppliedMigration{
			{
				Version:  plan.Anchor.Version,
				Name:     plan.Anchor.Name,
				Checksum: plan.AnchorChecksum,
			},
		},
		FlightTrajectoryReconciliationTaskIDColumnExists: true,
		DataQualityReconciliationTaskIDColumnExists:      true,
		FlightTrajectoryReconciliationForeignKeyExists:   true,
		DataQualityReconciliationForeignKeyExists:        true,
		FlightTrajectoryReconciliationUniqueIndexExists:  true,
		DataQualityReconciliationUniqueIndexExists:       true,
	}
}

type fakeInspector struct {
	state State
	err   error
}

func (inspector fakeInspector) Load(
	ctx context.Context,
	plan Plan,
) (State, error) {
	if err := ctx.Err(); err != nil {
		return State{}, err
	}
	if err := plan.Validate(); err != nil {
		return State{}, err
	}
	if inspector.err != nil {
		return State{}, inspector.err
	}
	return inspector.state, nil
}

func newTestVerifier(
	t *testing.T,
	config Config,
	plan Plan,
) *Verifier {
	t.Helper()
	config.MigrationsDir = writeTestMigrationDirectory(
		t,
		plan.Anchor.FileName,
		"BEGIN;\nSELECT 10;\nCOMMIT;\n",
	)
	config.AnchorFileName = plan.Anchor.FileName
	verifier, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return verifier
}

func newTestPlan(t *testing.T) Plan {
	t.Helper()
	directory := writeTestMigrationDirectory(
		t,
		DefaultRepairAnchorFileName,
		"BEGIN;\nSELECT 10;\nCOMMIT;\n",
	)
	plan, err := LoadPlan(directory, DefaultRepairAnchorFileName)
	if err != nil {
		t.Fatalf("LoadPlan() error = %v", err)
	}
	return plan
}

func writeTestMigrationDirectory(
	t *testing.T,
	fileName string,
	content string,
) string {
	t.Helper()
	directory := t.TempDir()
	path := filepath.Join(directory, fileName)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write migration fixture: %v", err)
	}
	return directory
}

func assertFailedCheck(
	t *testing.T,
	report Report,
	code CheckCode,
) {
	t.Helper()
	for _, check := range report.Checks {
		if check.Code != code {
			continue
		}
		if check.Passed || check.Severity != SeverityBlocker {
			t.Fatalf("check %s = %#v", code, check)
		}
		if report.Ready {
			t.Fatalf("report remained ready after failed check %s", code)
		}
		return
	}
	t.Fatalf("check %s not found in %#v", code, report.Checks)
}
