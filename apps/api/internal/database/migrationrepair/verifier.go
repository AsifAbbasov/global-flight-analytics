package migrationrepair

import (
	"context"
	"fmt"
	"sort"
	"time"
)

type Verifier struct {
	inspector Inspector
	now       func() time.Time
}

func New(config Config) (*Verifier, error) {
	if config.Inspector == nil {
		return nil, ErrInspectorRequired
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}

	return &Verifier{
		inspector: config.Inspector,
		now:       now,
	}, nil
}

func (verifier *Verifier) Verify(
	ctx context.Context,
) (Report, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return Report{}, err
	}

	state, err := verifier.inspector.Load(ctx)
	if err != nil {
		return Report{}, err
	}
	if err := ctx.Err(); err != nil {
		return Report{}, err
	}

	report := evaluateState(
		state,
		verifier.now().UTC(),
	)

	return report.Clone(), nil
}

func evaluateState(
	state State,
	generatedAt time.Time,
) Report {
	report := Report{
		Version:     Version,
		GeneratedAt: generatedAt.UTC(),
		Checks:      make([]Check, 0, 9),
	}

	appendCheck(
		&report,
		CheckSchemaMigrationsTablePresent,
		state.SchemaMigrationsTableExists,
		"The schema_migrations table exists.",
		"The schema_migrations table is missing.",
	)

	version010 := appliedByVersion(
		state.AppliedMigrations,
		"010",
	)
	version010Exact :=
		len(version010) == 1 &&
			version010[0].Name ==
				ExpectedAppliedVersion010Name &&
			version010[0].Checksum ==
				ExpectedAppliedVersion010Checksum
	appendCheck(
		&report,
		CheckAppliedVersion010Exact,
		version010Exact,
		fmt.Sprintf(
			"Applied version 010 is exactly %s with the expected checksum.",
			ExpectedAppliedVersion010Name,
		),
		fmt.Sprintf(
			"Applied version 010 does not exactly match %s and checksum %s.",
			ExpectedAppliedVersion010Name,
			ExpectedAppliedVersion010Checksum,
		),
	)

	futureVersionsUnapplied :=
		len(appliedByVersion(
			state.AppliedMigrations,
			"011",
		)) == 0 &&
			len(appliedByVersion(
				state.AppliedMigrations,
				"012",
			)) == 0
	appendCheck(
		&report,
		CheckFutureVersionsUnapplied,
		futureVersionsUnapplied,
		"Migration versions 011 and 012 are not recorded as applied.",
		"Migration version 011 or 012 is already recorded as applied.",
	)

	reconciliationColumnsPresent :=
		state.
			FlightTrajectoryReconciliationTaskIDColumnExists &&
			state.
				DataQualityReconciliationTaskIDColumnExists
	appendCheck(
		&report,
		CheckReconciliationColumnsPresent,
		reconciliationColumnsPresent,
		"Both reconciliation_task_id columns are present.",
		"One or both reconciliation_task_id columns are missing.",
	)

	reconciliationConstraintsPresent :=
		state.
			FlightTrajectoryReconciliationForeignKeyExists &&
			state.
				DataQualityReconciliationForeignKeyExists
	appendCheck(
		&report,
		CheckReconciliationConstraintsPresent,
		reconciliationConstraintsPresent,
		"Both reconciliation foreign key constraints are present.",
		"One or both reconciliation foreign key constraints are missing.",
	)

	reconciliationIndexesPresent :=
		state.
			FlightTrajectoryReconciliationUniqueIndexExists &&
			state.
				DataQualityReconciliationUniqueIndexExists
	appendCheck(
		&report,
		CheckReconciliationIndexesPresent,
		reconciliationIndexesPresent,
		"Both reconciliation unique indexes are present.",
		"One or both reconciliation unique indexes are missing.",
	)

	identityColumnsAbsent :=
		!state.IdentityKeyColumnExists &&
			!state.IdentityBasisColumnExists &&
			!state.SplitReasonColumnExists
	appendCheck(
		&report,
		CheckIdentityColumnsAbsent,
		identityColumnsAbsent,
		"Identity columns are absent and migration 011 can create them.",
		"One or more identity columns already exist; migration 011 would not be safe.",
	)

	identityConstraintsAbsent :=
		!state.IdentityCompletenessCheckExists &&
			!state.IdentityKeyCheckExists &&
			!state.IdentityBasisCheckExists &&
			!state.SplitReasonCheckExists
	appendCheck(
		&report,
		CheckIdentityConstraintsAbsent,
		identityConstraintsAbsent,
		"Identity constraints are absent.",
		"One or more identity constraints already exist.",
	)

	appendCheck(
		&report,
		CheckIdentityIndexAbsent,
		!state.IdentityKeyTimeIndexExists,
		"Identity index is absent.",
		"Identity index already exists.",
	)

	sort.SliceStable(
		report.Checks,
		func(left int, right int) bool {
			return report.Checks[left].Code <
				report.Checks[right].Code
		},
	)

	report.Ready = report.BlockerCount == 0

	return report
}

func appendCheck(
	report *Report,
	code CheckCode,
	passed bool,
	successMessage string,
	failureMessage string,
) {
	check := Check{
		Code:   code,
		Passed: passed,
	}

	if passed {
		check.Severity = SeverityInfo
		check.Message = successMessage
		report.InfoCount++
	} else {
		check.Severity = SeverityBlocker
		check.Message = failureMessage
		report.BlockerCount++
	}

	report.Checks = append(
		report.Checks,
		check,
	)
}

func appliedByVersion(
	migrations []AppliedMigration,
	version string,
) []AppliedMigration {
	result := make(
		[]AppliedMigration,
		0,
	)
	for _, migration := range migrations {
		if migration.Version == version {
			result = append(
				result,
				migration,
			)
		}
	}

	return result
}
