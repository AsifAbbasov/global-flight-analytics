package migrationrepair

import (
	"context"
	"fmt"
	"sort"
	"time"
)

type Verifier struct {
	inspector Inspector
	plan      Plan
	now       func() time.Time
}

func New(config Config) (*Verifier, error) {
	if config.Inspector == nil {
		return nil, ErrInspectorRequired
	}

	plan, err := LoadPlan(config.MigrationsDir, config.AnchorFileName)
	if err != nil {
		return nil, err
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}

	return &Verifier{
		inspector: config.Inspector,
		plan:      plan,
		now:       now,
	}, nil
}

func (verifier *Verifier) Verify(ctx context.Context) (Report, error) {
	if ctx == nil {
		return Report{}, ErrContextRequired
	}
	if err := ctx.Err(); err != nil {
		return Report{}, err
	}

	state, err := verifier.inspector.Load(ctx, verifier.plan)
	if err != nil {
		return Report{}, err
	}
	if err := ctx.Err(); err != nil {
		return Report{}, err
	}

	report := evaluateState(state, verifier.plan, verifier.now().UTC())
	return report.Clone(), nil
}

func evaluateState(
	state State,
	plan Plan,
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

	anchorRows := appliedByVersion(state.AppliedMigrations, plan.Anchor.Version)
	anchorExact := len(anchorRows) == 1 &&
		anchorRows[0].Name == plan.Anchor.Name &&
		anchorRows[0].Checksum == plan.AnchorChecksum
	appendCheck(
		&report,
		CheckAppliedMigrationExact,
		anchorExact,
		fmt.Sprintf(
			"Applied migration %s is exactly %s with the repository checksum.",
			plan.Anchor.Version,
			plan.Anchor.Name,
		),
		fmt.Sprintf(
			"Applied migration %s does not exactly match %s and repository checksum %s.",
			plan.Anchor.Version,
			plan.Anchor.Name,
			plan.AnchorChecksum,
		),
	)

	laterMigrationsUnapplied := true
	for _, migration := range state.AppliedMigrations {
		if plan.IsLaterVersion(migration.Version) {
			laterMigrationsUnapplied = false
			break
		}
	}
	appendCheck(
		&report,
		CheckLaterMigrationsUnapplied,
		laterMigrationsUnapplied,
		fmt.Sprintf(
			"No migration later than %s is recorded as applied.",
			plan.Anchor.Version,
		),
		fmt.Sprintf(
			"A migration later than %s is already recorded as applied.",
			plan.Anchor.Version,
		),
	)

	reconciliationColumnsPresent :=
		state.FlightTrajectoryReconciliationTaskIDColumnExists &&
			state.DataQualityReconciliationTaskIDColumnExists
	appendCheck(
		&report,
		CheckReconciliationColumnsPresent,
		reconciliationColumnsPresent,
		"Both reconciliation_task_id columns are present.",
		"One or both reconciliation_task_id columns are missing.",
	)

	reconciliationConstraintsPresent :=
		state.FlightTrajectoryReconciliationForeignKeyExists &&
			state.DataQualityReconciliationForeignKeyExists
	appendCheck(
		&report,
		CheckReconciliationConstraintsPresent,
		reconciliationConstraintsPresent,
		"Both reconciliation foreign key constraints are present.",
		"One or both reconciliation foreign key constraints are missing.",
	)

	reconciliationIndexesPresent :=
		state.FlightTrajectoryReconciliationUniqueIndexExists &&
			state.DataQualityReconciliationUniqueIndexExists
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
		"Identity columns are absent and the next migration can create them.",
		"One or more identity columns already exist; the repair sequence is not safe.",
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
			return report.Checks[left].Code < report.Checks[right].Code
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
	check := Check{Code: code, Passed: passed}
	if passed {
		check.Severity = SeverityInfo
		check.Message = successMessage
		report.InfoCount++
	} else {
		check.Severity = SeverityBlocker
		check.Message = failureMessage
		report.BlockerCount++
	}
	report.Checks = append(report.Checks, check)
}

func appliedByVersion(
	migrations []AppliedMigration,
	version string,
) []AppliedMigration {
	result := make([]AppliedMigration, 0)
	for _, migration := range migrations {
		if migration.Version == version {
			result = append(result, migration)
		}
	}
	return result
}
