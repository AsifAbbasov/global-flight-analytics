package migrationaudit

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Auditor struct {
	migrationsDir string
	stateLoader   StateLoader
	now           func() time.Time
}

func New(config Config) (*Auditor, error) {
	migrationsDir := strings.TrimSpace(
		config.MigrationsDir,
	)
	if migrationsDir == "" {
		return nil, ErrMigrationsDirRequired
	}
	if config.StateLoader == nil {
		return nil, ErrStateLoaderRequired
	}

	absoluteDir, err := filepath.Abs(
		migrationsDir,
	)
	if err != nil {
		return nil, &LocalScanError{
			Path: migrationsDir,
			Err:  err,
		}
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}

	return &Auditor{
		migrationsDir: filepath.Clean(
			absoluteDir,
		),
		stateLoader: config.StateLoader,
		now:         now,
	}, nil
}

func (auditor *Auditor) Audit(
	ctx context.Context,
) (Report, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return Report{}, err
	}

	local, err := scanLocalMigrations(
		auditor.migrationsDir,
	)
	if err != nil {
		return Report{}, err
	}

	state, err := auditor.stateLoader.Load(ctx)
	if err != nil {
		return Report{}, err
	}
	if err := ctx.Err(); err != nil {
		return Report{}, err
	}

	report := Report{
		Version:                     Version,
		GeneratedAt:                 auditor.now().UTC(),
		MigrationsDir:               auditor.migrationsDir,
		SchemaMigrationsTableExists: state.SchemaMigrationsTableExists,
		LocalMigrationCount:         len(local.migrations),
		InvalidLocalFileCount:       len(local.invalid),
		AppliedMigrationCount:       len(state.AppliedMigrations),
		LocalMigrations: append(
			[]LocalMigration(nil),
			local.migrations...,
		),
		InvalidLocalFiles: append(
			[]InvalidLocalFile(nil),
			local.invalid...,
		),
		AppliedMigrations: append(
			[]AppliedMigration(nil),
			state.AppliedMigrations...,
		),
		Findings: make(
			[]Finding,
			0,
		),
	}

	for _, invalid := range local.invalid {
		report.Findings = append(
			report.Findings,
			Finding{
				Severity: SeverityBlocker,
				Code:     FindingInvalidLocalFileName,
				Message: fmt.Sprintf(
					"Local migration file %q is not canonical: %s.",
					invalid.FileName,
					invalid.Reason,
				),
				LocalFiles: []string{invalid.FileName},
			},
		)
	}

	localByVersion := groupLocalByVersion(
		local.migrations,
	)
	appliedByVersion := groupAppliedByVersion(
		state.AppliedMigrations,
	)

	versions := unionVersions(
		localByVersion,
		appliedByVersion,
	)

	for _, version := range versions {
		localCandidates := localByVersion[version]
		appliedRecords := appliedByVersion[version]

		if len(localCandidates) > 1 {
			report.DuplicateLocalVersionCount++
			report.Findings = append(
				report.Findings,
				Finding{
					Severity: SeverityBlocker,
					Code:     FindingDuplicateLocalVersion,
					Version:  version,
					Message: fmt.Sprintf(
						"Migration version %s is assigned to %d local SQL files; schema_migrations can store only one row for this version.",
						version,
						len(localCandidates),
					),
					LocalFiles: localFileNames(localCandidates),
				},
			)
		}

		if len(appliedRecords) > 1 {
			report.Findings = append(
				report.Findings,
				Finding{
					Severity: SeverityBlocker,
					Code:     FindingDuplicateAppliedVersion,
					Version:  version,
					Message: fmt.Sprintf(
						"Database history contains %d rows for migration version %s.",
						len(appliedRecords),
						version,
					),
					LocalFiles: localFileNames(localCandidates),
				},
			)
		}

		if len(appliedRecords) == 0 {
			for _, candidate := range localCandidates {
				report.Findings = append(
					report.Findings,
					Finding{
						Severity: SeverityInfo,
						Code:     FindingPendingMigration,
						Version:  version,
						Message: fmt.Sprintf(
							"Local migration %s is not recorded as applied.",
							candidate.FileName,
						),
						LocalFiles: []string{
							candidate.FileName,
						},
					},
				)
			}
			continue
		}

		for _, applied := range appliedRecords {
			auditAppliedRecord(
				&report,
				applied,
				localCandidates,
			)
		}
	}

	if !state.SchemaMigrationsTableExists {
		report.Findings = append(
			report.Findings,
			Finding{
				Severity: SeverityBlocker,
				Code:     FindingSchemaMigrationsMissing,
				Message:  "The schema_migrations table does not exist, so deployed migration history cannot be reconciled safely.",
			},
		)
	}

	sortReport(&report)
	countFindings(&report)

	return report.Clone(), nil
}

func auditAppliedRecord(
	report *Report,
	applied AppliedMigration,
	localCandidates []LocalMigration,
) {
	if len(localCandidates) == 0 {
		report.Findings = append(
			report.Findings,
			Finding{
				Severity: SeverityBlocker,
				Code:     FindingAppliedMigrationMissingLocally,
				Version:  applied.Version,
				Message: fmt.Sprintf(
					"Database migration %s (%s) is applied but no local SQL file has that version.",
					applied.Version,
					applied.Name,
				),
				AppliedName:     applied.Name,
				AppliedChecksum: applied.Checksum,
			},
		)
		return
	}

	matches := make(
		[]LocalMigration,
		0,
		len(localCandidates),
	)
	for _, candidate := range localCandidates {
		if candidate.Checksum ==
			applied.Checksum {
			matches = append(matches, candidate)
		}
	}

	if len(matches) == 0 {
		report.Findings = append(
			report.Findings,
			Finding{
				Severity: SeverityBlocker,
				Code:     FindingAppliedChecksumMismatch,
				Version:  applied.Version,
				Message: fmt.Sprintf(
					"Applied migration version %s has checksum %s, which matches none of the local candidates.",
					applied.Version,
					applied.Checksum,
				),
				LocalFiles:      localFileNames(localCandidates),
				AppliedName:     applied.Name,
				AppliedChecksum: applied.Checksum,
			},
		)
		return
	}

	if len(matches) > 1 {
		report.Findings = append(
			report.Findings,
			Finding{
				Severity: SeverityBlocker,
				Code:     FindingAppliedDuplicateAmbiguous,
				Version:  applied.Version,
				Message: fmt.Sprintf(
					"Applied migration version %s matches multiple local files by checksum and cannot be attributed uniquely.",
					applied.Version,
				),
				LocalFiles:      localFileNames(matches),
				AppliedName:     applied.Name,
				AppliedChecksum: applied.Checksum,
			},
		)
		return
	}

	match := matches[0]
	if len(localCandidates) > 1 {
		report.Findings = append(
			report.Findings,
			Finding{
				Severity: SeverityInfo,
				Code:     FindingAppliedDuplicateResolved,
				Version:  applied.Version,
				Message: fmt.Sprintf(
					"Applied version %s matches local file %s by checksum; the sibling file with the same version remains unreachable until migration history is repaired.",
					applied.Version,
					match.FileName,
				),
				LocalFiles:      localFileNames(localCandidates),
				AppliedName:     applied.Name,
				AppliedChecksum: applied.Checksum,
			},
		)
	}

	if match.Name != applied.Name {
		report.Findings = append(
			report.Findings,
			Finding{
				Severity: SeverityWarning,
				Code:     FindingAppliedNameMismatch,
				Version:  applied.Version,
				Message: fmt.Sprintf(
					"Applied migration version %s matches %s by checksum but records name %q instead of %q.",
					applied.Version,
					match.FileName,
					applied.Name,
					match.Name,
				),
				LocalFiles:      []string{match.FileName},
				AppliedName:     applied.Name,
				AppliedChecksum: applied.Checksum,
			},
		)
	}
}

func groupLocalByVersion(
	migrations []LocalMigration,
) map[string][]LocalMigration {
	result := make(
		map[string][]LocalMigration,
	)
	for _, migration := range migrations {
		result[migration.Version] = append(
			result[migration.Version],
			migration,
		)
	}

	return result
}

func groupAppliedByVersion(
	migrations []AppliedMigration,
) map[string][]AppliedMigration {
	result := make(
		map[string][]AppliedMigration,
	)
	for _, migration := range migrations {
		result[migration.Version] = append(
			result[migration.Version],
			migration,
		)
	}

	return result
}

func unionVersions(
	local map[string][]LocalMigration,
	applied map[string][]AppliedMigration,
) []string {
	seen := make(
		map[string]struct{},
		len(local)+len(applied),
	)
	for version := range local {
		seen[version] = struct{}{}
	}
	for version := range applied {
		seen[version] = struct{}{}
	}

	versions := make(
		[]string,
		0,
		len(seen),
	)
	for version := range seen {
		versions = append(
			versions,
			version,
		)
	}
	sort.Strings(versions)

	return versions
}

func localFileNames(
	migrations []LocalMigration,
) []string {
	result := make(
		[]string,
		0,
		len(migrations),
	)
	for _, migration := range migrations {
		result = append(
			result,
			migration.FileName,
		)
	}
	sort.Strings(result)

	return result
}

func sortReport(report *Report) {
	sort.SliceStable(
		report.AppliedMigrations,
		func(left int, right int) bool {
			if report.AppliedMigrations[left].Version !=
				report.AppliedMigrations[right].Version {
				return report.AppliedMigrations[left].Version <
					report.AppliedMigrations[right].Version
			}
			if report.AppliedMigrations[left].Name !=
				report.AppliedMigrations[right].Name {
				return report.AppliedMigrations[left].Name <
					report.AppliedMigrations[right].Name
			}

			return report.AppliedMigrations[left].
				AppliedAt.Before(
				report.AppliedMigrations[right].
					AppliedAt,
			)
		},
	)

	severityRank := map[Severity]int{
		SeverityBlocker: 0,
		SeverityWarning: 1,
		SeverityInfo:    2,
	}
	sort.SliceStable(
		report.Findings,
		func(left int, right int) bool {
			leftFinding := report.Findings[left]
			rightFinding := report.Findings[right]

			if severityRank[leftFinding.Severity] !=
				severityRank[rightFinding.Severity] {
				return severityRank[leftFinding.Severity] <
					severityRank[rightFinding.Severity]
			}
			if leftFinding.Version !=
				rightFinding.Version {
				return leftFinding.Version <
					rightFinding.Version
			}
			if leftFinding.Code !=
				rightFinding.Code {
				return leftFinding.Code <
					rightFinding.Code
			}

			return leftFinding.Message <
				rightFinding.Message
		},
	)
}

func countFindings(report *Report) {
	for _, finding := range report.Findings {
		switch finding.Severity {
		case SeverityBlocker:
			report.BlockerCount++
		case SeverityWarning:
			report.WarningCount++
		case SeverityInfo:
			report.InfoCount++
		}
	}
}
