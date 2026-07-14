package migrationaudit

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestAuditIdentifiesWhichDuplicateVersionWasApplied(
	t *testing.T,
) {
	dir := t.TempDir()
	firstPath := filepath.Join(
		dir,
		"010_add_flight_identity_metadata.sql",
	)
	secondPath := filepath.Join(
		dir,
		"010_add_reconciliation_result_identity.sql",
	)
	if err := os.WriteFile(
		firstPath,
		[]byte("SELECT 'identity';\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		secondPath,
		[]byte("SELECT 'reconciliation';\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	firstChecksum, err :=
		calculateChecksum(firstPath)
	if err != nil {
		t.Fatal(err)
	}

	generatedAt := time.Date(
		2026,
		time.July,
		14,
		16,
		0,
		0,
		0,
		time.UTC,
	)
	auditor := newTestAuditor(
		t,
		Config{
			MigrationsDir: dir,
			StateLoader: fakeStateLoader{
				state: DatabaseState{
					SchemaMigrationsTableExists: true,
					AppliedMigrations: []AppliedMigration{
						{
							Version:   "010",
							Name:      "add_flight_identity_metadata",
							Checksum:  firstChecksum,
							AppliedAt: generatedAt.Add(-time.Hour),
						},
					},
				},
			},
			Now: func() time.Time {
				return generatedAt
			},
		},
	)

	report, err := auditor.Audit(
		context.Background(),
	)
	if err != nil {
		t.Fatalf("Audit() error = %v", err)
	}

	if report.DuplicateLocalVersionCount != 1 ||
		report.BlockerCount != 1 ||
		report.WarningCount != 0 ||
		report.InfoCount != 1 {
		t.Fatalf(
			"unexpected counts: %#v",
			report,
		)
	}
	assertFinding(
		t,
		report.Findings,
		FindingDuplicateLocalVersion,
		SeverityBlocker,
	)
	resolved := assertFinding(
		t,
		report.Findings,
		FindingAppliedDuplicateResolved,
		SeverityInfo,
	)
	if resolved.AppliedChecksum != firstChecksum ||
		!reflect.DeepEqual(
			resolved.LocalFiles,
			[]string{
				"010_add_flight_identity_metadata.sql",
				"010_add_reconciliation_result_identity.sql",
			},
		) {
		t.Fatalf(
			"unexpected resolved finding: %#v",
			resolved,
		)
	}
}

func TestAuditReportsAppliedChecksumMismatch(
	t *testing.T,
) {
	dir := t.TempDir()
	writeTestFile(
		t,
		dir,
		"001_first.sql",
		"SELECT 1;",
	)

	auditor := newTestAuditor(
		t,
		Config{
			MigrationsDir: dir,
			StateLoader: fakeStateLoader{
				state: DatabaseState{
					SchemaMigrationsTableExists: true,
					AppliedMigrations: []AppliedMigration{
						{
							Version:  "001",
							Name:     "first",
							Checksum: "different",
						},
					},
				},
			},
		},
	)

	report, err := auditor.Audit(
		context.Background(),
	)
	if err != nil {
		t.Fatalf("Audit() error = %v", err)
	}
	if report.BlockerCount != 1 {
		t.Fatalf(
			"blockers = %d, want 1",
			report.BlockerCount,
		)
	}
	assertFinding(
		t,
		report.Findings,
		FindingAppliedChecksumMismatch,
		SeverityBlocker,
	)
}

func TestAuditReportsMissingHistoryAndPendingMigrations(
	t *testing.T,
) {
	dir := t.TempDir()
	writeTestFile(
		t,
		dir,
		"001_first.sql",
		"SELECT 1;",
	)

	auditor := newTestAuditor(
		t,
		Config{
			MigrationsDir: dir,
			StateLoader: fakeStateLoader{
				state: DatabaseState{
					SchemaMigrationsTableExists: false,
				},
			},
		},
	)

	report, err := auditor.Audit(
		context.Background(),
	)
	if err != nil {
		t.Fatalf("Audit() error = %v", err)
	}
	if report.BlockerCount != 1 ||
		report.InfoCount != 1 {
		t.Fatalf(
			"unexpected counts: %#v",
			report,
		)
	}
	assertFinding(
		t,
		report.Findings,
		FindingSchemaMigrationsMissing,
		SeverityBlocker,
	)
	assertFinding(
		t,
		report.Findings,
		FindingPendingMigration,
		SeverityInfo,
	)
}

func TestAuditReportsAppliedMigrationMissingLocally(
	t *testing.T,
) {
	dir := t.TempDir()
	auditor := newTestAuditor(
		t,
		Config{
			MigrationsDir: dir,
			StateLoader: fakeStateLoader{
				state: DatabaseState{
					SchemaMigrationsTableExists: true,
					AppliedMigrations: []AppliedMigration{
						{
							Version:  "999",
							Name:     "removed",
							Checksum: "checksum",
						},
					},
				},
			},
		},
	)

	report, err := auditor.Audit(
		context.Background(),
	)
	if err != nil {
		t.Fatalf("Audit() error = %v", err)
	}
	assertFinding(
		t,
		report.Findings,
		FindingAppliedMigrationMissingLocally,
		SeverityBlocker,
	)
}

func TestAuditPreservesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	auditor := newTestAuditor(
		t,
		Config{
			MigrationsDir: t.TempDir(),
			StateLoader:   fakeStateLoader{},
		},
	)

	_, err := auditor.Audit(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Audit() error = %v, want context.Canceled",
			err,
		)
	}
}

func TestReportCloneDoesNotShareSlices(t *testing.T) {
	report := Report{
		LocalMigrations: []LocalMigration{
			{Version: "001"},
		},
		InvalidLocalFiles: []InvalidLocalFile{
			{FileName: "invalid.sql"},
		},
		AppliedMigrations: []AppliedMigration{
			{Version: "001"},
		},
		Findings: []Finding{
			{
				Code:       FindingPendingMigration,
				LocalFiles: []string{"001.sql"},
			},
		},
	}

	cloned := report.Clone()
	cloned.LocalMigrations[0].Version = "changed"
	cloned.InvalidLocalFiles[0].FileName =
		"changed"
	cloned.AppliedMigrations[0].Version =
		"changed"
	cloned.Findings[0].Code =
		FindingAppliedChecksumMismatch
	cloned.Findings[0].LocalFiles[0] =
		"changed"

	if report.LocalMigrations[0].Version != "001" ||
		report.InvalidLocalFiles[0].FileName !=
			"invalid.sql" ||
		report.AppliedMigrations[0].Version !=
			"001" ||
		report.Findings[0].Code !=
			FindingPendingMigration ||
		report.Findings[0].LocalFiles[0] !=
			"001.sql" {
		t.Fatal("Report.Clone() shared slices")
	}
}

type fakeStateLoader struct {
	state DatabaseState
	err   error
}

func (loader fakeStateLoader) Load(
	ctx context.Context,
) (DatabaseState, error) {
	if err := ctx.Err(); err != nil {
		return DatabaseState{}, err
	}
	if loader.err != nil {
		return DatabaseState{}, loader.err
	}

	return loader.state, nil
}

func newTestAuditor(
	t *testing.T,
	config Config,
) *Auditor {
	t.Helper()

	auditor, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return auditor
}

func assertFinding(
	t *testing.T,
	findings []Finding,
	code FindingCode,
	severity Severity,
) Finding {
	t.Helper()

	for _, finding := range findings {
		if finding.Code == code {
			if finding.Severity != severity {
				t.Fatalf(
					"finding %s severity = %s, want %s",
					code,
					finding.Severity,
					severity,
				)
			}

			return finding
		}
	}

	t.Fatalf(
		"finding %s was not found in %#v",
		code,
		findings,
	)

	return Finding{}
}
