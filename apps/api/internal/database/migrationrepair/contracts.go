package migrationrepair

import (
	"context"
	"time"
)

const Version = "migration-sequence-repair-preflight-v1"

const (
	ExpectedAppliedVersion010Name     = "add_reconciliation_result_identity"
	ExpectedAppliedVersion010Checksum = "5c6481b807271a856654cdb1ada298dbac43b3b3aaab0bc8dbc62354916fae91"
)

type Severity string

const (
	SeverityBlocker Severity = "blocker"
	SeverityInfo    Severity = "info"
)

type CheckCode string

const (
	CheckSchemaMigrationsTablePresent     CheckCode = "schema_migrations_table_present"
	CheckAppliedVersion010Exact           CheckCode = "applied_version_010_exact"
	CheckFutureVersionsUnapplied          CheckCode = "future_versions_011_012_unapplied"
	CheckReconciliationColumnsPresent     CheckCode = "reconciliation_columns_present"
	CheckReconciliationConstraintsPresent CheckCode = "reconciliation_constraints_present"
	CheckReconciliationIndexesPresent     CheckCode = "reconciliation_indexes_present"
	CheckIdentityColumnsAbsent            CheckCode = "identity_columns_absent"
	CheckIdentityConstraintsAbsent        CheckCode = "identity_constraints_absent"
	CheckIdentityIndexAbsent              CheckCode = "identity_index_absent"
)

type AppliedMigration struct {
	Version  string
	Name     string
	Checksum string
}

type State struct {
	SchemaMigrationsTableExists bool
	AppliedMigrations           []AppliedMigration

	FlightTrajectoryReconciliationTaskIDColumnExists bool
	DataQualityReconciliationTaskIDColumnExists      bool
	FlightTrajectoryReconciliationForeignKeyExists   bool
	DataQualityReconciliationForeignKeyExists        bool
	FlightTrajectoryReconciliationUniqueIndexExists  bool
	DataQualityReconciliationUniqueIndexExists       bool

	IdentityKeyColumnExists         bool
	IdentityBasisColumnExists       bool
	SplitReasonColumnExists         bool
	IdentityCompletenessCheckExists bool
	IdentityKeyCheckExists          bool
	IdentityBasisCheckExists        bool
	SplitReasonCheckExists          bool
	IdentityKeyTimeIndexExists      bool
}

type Inspector interface {
	Load(ctx context.Context) (State, error)
}

type Config struct {
	Inspector Inspector
	Now       func() time.Time
}

type Check struct {
	Code     CheckCode `json:"code"`
	Severity Severity  `json:"severity"`
	Passed   bool      `json:"passed"`
	Message  string    `json:"message"`
}

type Report struct {
	Version      string    `json:"version"`
	GeneratedAt  time.Time `json:"generated_at"`
	Ready        bool      `json:"ready"`
	BlockerCount int       `json:"blocker_count"`
	InfoCount    int       `json:"info_count"`
	Checks       []Check   `json:"checks"`
}

func (report Report) Clone() Report {
	cloned := report
	cloned.Checks = append(
		[]Check(nil),
		report.Checks...,
	)

	return cloned
}
