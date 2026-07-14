package migrationaudit

import (
	"context"
	"time"
)

const Version = "migration-history-audit-v1"

type Severity string

const (
	SeverityBlocker Severity = "blocker"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

type FindingCode string

const (
	FindingInvalidLocalFileName           FindingCode = "invalid_local_migration_file_name"
	FindingDuplicateLocalVersion          FindingCode = "duplicate_local_migration_version"
	FindingSchemaMigrationsMissing        FindingCode = "schema_migrations_table_missing"
	FindingDuplicateAppliedVersion        FindingCode = "duplicate_applied_migration_version"
	FindingAppliedMigrationMissingLocally FindingCode = "applied_migration_missing_locally"
	FindingAppliedChecksumMismatch        FindingCode = "applied_migration_checksum_mismatch"
	FindingAppliedNameMismatch            FindingCode = "applied_migration_name_mismatch"
	FindingAppliedDuplicateResolved       FindingCode = "applied_duplicate_version_resolved_by_checksum"
	FindingAppliedDuplicateAmbiguous      FindingCode = "applied_duplicate_version_ambiguous"
	FindingPendingMigration               FindingCode = "pending_local_migration"
)

type LocalMigration struct {
	Version  string `json:"version"`
	Name     string `json:"name"`
	FileName string `json:"file_name"`
	Path     string `json:"path"`
	Checksum string `json:"checksum"`
}

type InvalidLocalFile struct {
	FileName string `json:"file_name"`
	Path     string `json:"path"`
	Reason   string `json:"reason"`
}

type AppliedMigration struct {
	Version   string    `json:"version"`
	Name      string    `json:"name"`
	Checksum  string    `json:"checksum"`
	AppliedAt time.Time `json:"applied_at"`
}

type DatabaseState struct {
	SchemaMigrationsTableExists bool               `json:"schema_migrations_table_exists"`
	AppliedMigrations           []AppliedMigration `json:"applied_migrations"`
}

type StateLoader interface {
	Load(ctx context.Context) (DatabaseState, error)
}

type Config struct {
	MigrationsDir string
	StateLoader   StateLoader
	Now           func() time.Time
}

type Finding struct {
	Severity        Severity    `json:"severity"`
	Code            FindingCode `json:"code"`
	Version         string      `json:"version,omitempty"`
	Message         string      `json:"message"`
	LocalFiles      []string    `json:"local_files,omitempty"`
	AppliedName     string      `json:"applied_name,omitempty"`
	AppliedChecksum string      `json:"applied_checksum,omitempty"`
}

type Report struct {
	Version                     string             `json:"version"`
	GeneratedAt                 time.Time          `json:"generated_at"`
	MigrationsDir               string             `json:"migrations_dir"`
	SchemaMigrationsTableExists bool               `json:"schema_migrations_table_exists"`
	LocalMigrationCount         int                `json:"local_migration_count"`
	InvalidLocalFileCount       int                `json:"invalid_local_file_count"`
	AppliedMigrationCount       int                `json:"applied_migration_count"`
	DuplicateLocalVersionCount  int                `json:"duplicate_local_version_count"`
	BlockerCount                int                `json:"blocker_count"`
	WarningCount                int                `json:"warning_count"`
	InfoCount                   int                `json:"info_count"`
	LocalMigrations             []LocalMigration   `json:"local_migrations"`
	InvalidLocalFiles           []InvalidLocalFile `json:"invalid_local_files"`
	AppliedMigrations           []AppliedMigration `json:"applied_migrations"`
	Findings                    []Finding          `json:"findings"`
}

func (report Report) Clone() Report {
	cloned := report
	cloned.LocalMigrations = append(
		[]LocalMigration(nil),
		report.LocalMigrations...,
	)
	cloned.InvalidLocalFiles = append(
		[]InvalidLocalFile(nil),
		report.InvalidLocalFiles...,
	)
	cloned.AppliedMigrations = append(
		[]AppliedMigration(nil),
		report.AppliedMigrations...,
	)
	cloned.Findings = make(
		[]Finding,
		0,
		len(report.Findings),
	)
	for _, finding := range report.Findings {
		clonedFinding := finding
		clonedFinding.LocalFiles = append(
			[]string(nil),
			finding.LocalFiles...,
		)
		cloned.Findings = append(
			cloned.Findings,
			clonedFinding,
		)
	}

	return cloned
}
