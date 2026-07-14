package migrationaudit

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestWriteReportTextIsDeterministic(t *testing.T) {
	report := Report{
		Version: Version,
		GeneratedAt: time.Date(
			2026,
			time.July,
			14,
			16,
			0,
			0,
			0,
			time.UTC,
		),
		MigrationsDir:               "/project/database/migrations",
		SchemaMigrationsTableExists: true,
		LocalMigrationCount:         2,
		AppliedMigrationCount:       1,
		DuplicateLocalVersionCount:  1,
		BlockerCount:                1,
		InfoCount:                   1,
		Findings: []Finding{
			{
				Severity: SeverityBlocker,
				Code:     FindingDuplicateLocalVersion,
				Version:  "010",
				Message:  "duplicate version",
				LocalFiles: []string{
					"010_a.sql",
					"010_b.sql",
				},
			},
		},
	}

	var buffer bytes.Buffer
	if err := WriteReport(
		&buffer,
		report,
		OutputFormatText,
	); err != nil {
		t.Fatalf(
			"WriteReport() error = %v",
			err,
		)
	}

	output := buffer.String()
	for _, expected := range []string{
		"Migration History Audit",
		"schema_migrations: present",
		"Duplicate local versions: 1",
		"Blockers: 1",
		"[BLOCKER] duplicate_local_migration_version version=010: duplicate version",
		"local files: 010_a.sql, 010_b.sql",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf(
				"output does not contain %q:\n%s",
				expected,
				output,
			)
		}
	}
}

func TestWriteReportJSON(t *testing.T) {
	report := Report{
		Version:      Version,
		BlockerCount: 2,
		Findings: []Finding{
			{
				Severity: SeverityBlocker,
				Code:     FindingAppliedChecksumMismatch,
			},
		},
	}

	var buffer bytes.Buffer
	if err := WriteReport(
		&buffer,
		report,
		OutputFormatJSON,
	); err != nil {
		t.Fatalf(
			"WriteReport() error = %v",
			err,
		)
	}

	var decoded Report
	if err := json.Unmarshal(
		buffer.Bytes(),
		&decoded,
	); err != nil {
		t.Fatalf(
			"json.Unmarshal() error = %v",
			err,
		)
	}
	if decoded.Version != Version ||
		decoded.BlockerCount != 2 ||
		len(decoded.Findings) != 1 {
		t.Fatalf(
			"decoded report = %#v",
			decoded,
		)
	}
}

func TestWriteReportRejectsUnsupportedFormat(t *testing.T) {
	err := WriteReport(
		&bytes.Buffer{},
		Report{},
		"xml",
	)
	if err == nil {
		t.Fatal(
			"WriteReport() expected an error",
		)
	}
}
