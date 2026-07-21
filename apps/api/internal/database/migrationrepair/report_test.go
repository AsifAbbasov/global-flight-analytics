package migrationrepair

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestWriteReportText(t *testing.T) {
	report := Report{
		Version: Version,
		GeneratedAt: time.Date(
			2026,
			time.July,
			14,
			16,
			30,
			0,
			0,
			time.UTC,
		),
		Ready:     true,
		InfoCount: 1,
		Checks: []Check{
			{
				Code:     CheckAppliedMigrationExact,
				Severity: SeverityInfo,
				Passed:   true,
				Message:  "exact",
			},
		},
	}

	var buffer bytes.Buffer
	if err := WriteReport(
		&buffer,
		report,
		OutputFormatText,
	); err != nil {
		t.Fatalf("WriteReport() error = %v", err)
	}

	output := buffer.String()
	for _, expected := range []string{
		"Migration Sequence Repair Preflight",
		"Readiness: READY",
		"[PASS] INFO applied_migration_exact: exact",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf(
				"output missing %q:\n%s",
				expected,
				output,
			)
		}
	}
}

func TestWriteReportJSON(t *testing.T) {
	report := Report{
		Version:      Version,
		Ready:        false,
		BlockerCount: 1,
	}

	var buffer bytes.Buffer
	if err := WriteReport(
		&buffer,
		report,
		OutputFormatJSON,
	); err != nil {
		t.Fatalf("WriteReport() error = %v", err)
	}

	var decoded Report
	if err := json.Unmarshal(
		buffer.Bytes(),
		&decoded,
	); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if decoded.Version != Version ||
		decoded.Ready ||
		decoded.BlockerCount != 1 {
		t.Fatalf("decoded = %#v", decoded)
	}
}

func TestWriteReportRejectsUnsupportedFormat(
	t *testing.T,
) {
	if err := WriteReport(
		&bytes.Buffer{},
		Report{},
		"xml",
	); err == nil {
		t.Fatal("expected unsupported format error")
	}
}
