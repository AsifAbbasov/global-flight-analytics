package migrationrepair

import (
	"context"
	"time"
)

const Version = "migration-sequence-repair-preflight-v2"

const DefaultRepairAnchorFileName = "010_add_reconciliation_result_identity.sql"

type Severity string

const (
	SeverityBlocker Severity = "blocker"
	SeverityInfo    Severity = "info"
)

type CheckCode string

const (
	CheckSchemaMigrationsTablePresent     CheckCode = "schema_migrations_table_present"
	CheckAppliedMigrationExact            CheckCode = "applied_migration_exact"
	CheckLaterMigrationsUnapplied         CheckCode = "later_migrations_unapplied"
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
	Load(ctx context.Context, plan Plan) (State, error)
}

type Config struct {
	Inspector      Inspector
	MigrationsDir  string
	AnchorFileName string
	Now            func() time.Time
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
	cloned.Checks = append([]Check(nil), report.Checks...)
	return cloned
}
