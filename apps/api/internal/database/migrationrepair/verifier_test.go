package migrationrepair

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestVerifierAcceptsExactRepairReadyState(t *testing.T) {
	generatedAt := time.Date(
		2026,
		time.July,
		14,
		16,
		30,
		0,
		0,
		time.UTC,
	)
	verifier := newTestVerifier(
		t,
		Config{
			Inspector: fakeInspector{
				state: readyState(),
			},
			Now: func() time.Time {
				return generatedAt
			},
		},
	)

	report, err := verifier.Verify(
		context.Background(),
	)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if !report.Ready ||
		report.BlockerCount != 0 ||
		report.InfoCount != 9 ||
		len(report.Checks) != 9 ||
		!report.GeneratedAt.Equal(generatedAt) {
		t.Fatalf(
			"unexpected report: %#v",
			report,
		)
	}
	for _, check := range report.Checks {
		if !check.Passed ||
			check.Severity != SeverityInfo {
			t.Fatalf(
				"unexpected check: %#v",
				check,
			)
		}
	}
}

func TestVerifierBlocksWrongAppliedVersion010(t *testing.T) {
	state := readyState()
	state.AppliedMigrations[0].Checksum = "wrong"

	report := evaluateState(state, time.Time{})
	assertFailedCheck(
		t,
		report,
		CheckAppliedVersion010Exact,
	)
}

func TestVerifierBlocksAppliedFutureVersion(t *testing.T) {
	state := readyState()
	state.AppliedMigrations = append(
		state.AppliedMigrations,
		AppliedMigration{
			Version:  "011",
			Name:     "unexpected",
			Checksum: "unexpected",
		},
	)

	report := evaluateState(state, time.Time{})
	assertFailedCheck(
		t,
		report,
		CheckFutureVersionsUnapplied,
	)
}

func TestVerifierBlocksMissingReconciliationObjects(
	t *testing.T,
) {
	state := readyState()
	state.DataQualityReconciliationTaskIDColumnExists =
		false
	state.FlightTrajectoryReconciliationForeignKeyExists =
		false
	state.DataQualityReconciliationUniqueIndexExists =
		false

	report := evaluateState(state, time.Time{})
	assertFailedCheck(
		t,
		report,
		CheckReconciliationColumnsPresent,
	)
	assertFailedCheck(
		t,
		report,
		CheckReconciliationConstraintsPresent,
	)
	assertFailedCheck(
		t,
		report,
		CheckReconciliationIndexesPresent,
	)
}

func TestVerifierBlocksAnyExistingIdentityObject(
	t *testing.T,
) {
	state := readyState()
	state.IdentityKeyColumnExists = true
	state.SplitReasonCheckExists = true
	state.IdentityKeyTimeIndexExists = true

	report := evaluateState(state, time.Time{})
	assertFailedCheck(
		t,
		report,
		CheckIdentityColumnsAbsent,
	)
	assertFailedCheck(
		t,
		report,
		CheckIdentityConstraintsAbsent,
	)
	assertFailedCheck(
		t,
		report,
		CheckIdentityIndexAbsent,
	)
}

func TestVerifierPreservesContextCancellation(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	verifier := newTestVerifier(
		t,
		Config{
			Inspector: fakeInspector{},
		},
	)

	_, err := verifier.Verify(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Verify() error = %v, want context.Canceled",
			err,
		)
	}
}

func TestReportCloneDoesNotShareChecks(t *testing.T) {
	report := Report{
		Checks: []Check{
			{
				Code: CheckAppliedVersion010Exact,
			},
		},
	}

	cloned := report.Clone()
	cloned.Checks[0].Code =
		CheckIdentityColumnsAbsent

	if reflect.DeepEqual(
		report.Checks,
		cloned.Checks,
	) {
		t.Fatal("Report.Clone() shared checks")
	}
	if report.Checks[0].Code !=
		CheckAppliedVersion010Exact {
		t.Fatal("Report.Clone() mutated original")
	}
}

func readyState() State {
	return State{
		SchemaMigrationsTableExists: true,
		AppliedMigrations: []AppliedMigration{
			{
				Version:  expectedAppliedVersion010.Version,
				Name:     expectedAppliedVersion010.Name,
				Checksum: ExpectedAppliedVersion010Checksum,
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
) (State, error) {
	if err := ctx.Err(); err != nil {
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
) *Verifier {
	t.Helper()

	verifier, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return verifier
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
		if check.Passed ||
			check.Severity != SeverityBlocker {
			t.Fatalf(
				"check %s = %#v",
				code,
				check,
			)
		}
		if report.Ready {
			t.Fatalf(
				"report remained ready after failed check %s",
				code,
			)
		}

		return
	}

	t.Fatalf(
		"check %s not found in %#v",
		code,
		report.Checks,
	)
}
