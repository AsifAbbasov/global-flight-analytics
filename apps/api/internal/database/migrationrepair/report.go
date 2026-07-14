package migrationrepair

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const (
	OutputFormatText = "text"
	OutputFormatJSON = "json"
)

func WriteReport(
	writer io.Writer,
	report Report,
	format string,
) error {
	switch strings.ToLower(
		strings.TrimSpace(format),
	) {
	case "", OutputFormatText:
		return writeTextReport(writer, report)
	case OutputFormatJSON:
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report.Clone())
	default:
		return fmt.Errorf(
			"unsupported migration repair report format %q",
			format,
		)
	}
}

func writeTextReport(
	writer io.Writer,
	report Report,
) error {
	readiness := "BLOCKED"
	if report.Ready {
		readiness = "READY"
	}

	if _, err := fmt.Fprintf(
		writer,
		"Migration Sequence Repair Preflight\n"+
			"Version: %s\n"+
			"Generated at: %s\n"+
			"Readiness: %s\n"+
			"Blockers: %d\n"+
			"Information: %d\n",
		report.Version,
		report.GeneratedAt.Format(
			"2006-01-02T15:04:05Z07:00",
		),
		readiness,
		report.BlockerCount,
		report.InfoCount,
	); err != nil {
		return err
	}

	for _, check := range report.Checks {
		status := "PASS"
		if !check.Passed {
			status = "FAIL"
		}

		if _, err := fmt.Fprintf(
			writer,
			"[%s] %s %s: %s\n",
			status,
			strings.ToUpper(
				string(check.Severity),
			),
			check.Code,
			check.Message,
		); err != nil {
			return err
		}
	}

	return nil
}
