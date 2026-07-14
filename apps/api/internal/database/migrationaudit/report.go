package migrationaudit

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
			"unsupported migration audit output format %q",
			format,
		)
	}
}

func writeTextReport(
	writer io.Writer,
	report Report,
) error {
	tableState := "missing"
	if report.SchemaMigrationsTableExists {
		tableState = "present"
	}

	if _, err := fmt.Fprintf(
		writer,
		"Migration History Audit\n"+
			"Version: %s\n"+
			"Generated at: %s\n"+
			"Migrations directory: %s\n"+
			"schema_migrations: %s\n"+
			"Local migrations: %d\n"+
			"Invalid local files: %d\n"+
			"Applied migrations: %d\n"+
			"Duplicate local versions: %d\n"+
			"Blockers: %d\n"+
			"Warnings: %d\n"+
			"Information: %d\n",
		report.Version,
		report.GeneratedAt.Format(
			"2006-01-02T15:04:05Z07:00",
		),
		report.MigrationsDir,
		tableState,
		report.LocalMigrationCount,
		report.InvalidLocalFileCount,
		report.AppliedMigrationCount,
		report.DuplicateLocalVersionCount,
		report.BlockerCount,
		report.WarningCount,
		report.InfoCount,
	); err != nil {
		return err
	}

	if len(report.Findings) > 0 {
		if _, err := fmt.Fprintln(
			writer,
			"\nFindings:",
		); err != nil {
			return err
		}

		for _, finding := range report.Findings {
			version := ""
			if finding.Version != "" {
				version = " version=" +
					finding.Version
			}
			if _, err := fmt.Fprintf(
				writer,
				"[%s] %s%s: %s\n",
				strings.ToUpper(
					string(finding.Severity),
				),
				finding.Code,
				version,
				finding.Message,
			); err != nil {
				return err
			}
			if len(finding.LocalFiles) > 0 {
				if _, err := fmt.Fprintf(
					writer,
					"  local files: %s\n",
					strings.Join(
						finding.LocalFiles,
						", ",
					),
				); err != nil {
					return err
				}
			}
			if finding.AppliedName != "" {
				if _, err := fmt.Fprintf(
					writer,
					"  applied name: %s\n",
					finding.AppliedName,
				); err != nil {
					return err
				}
			}
			if finding.AppliedChecksum != "" {
				if _, err := fmt.Fprintf(
					writer,
					"  applied checksum: %s\n",
					finding.AppliedChecksum,
				); err != nil {
					return err
				}
			}
		}
	}

	if len(report.LocalMigrations) > 0 {
		if _, err := fmt.Fprintln(
			writer,
			"\nLocal migrations:",
		); err != nil {
			return err
		}
		for _, migration := range report.LocalMigrations {
			if _, err := fmt.Fprintf(
				writer,
				"%s %s %s %s\n",
				migration.Version,
				migration.Name,
				migration.Checksum,
				migration.FileName,
			); err != nil {
				return err
			}
		}
	}

	if len(report.AppliedMigrations) > 0 {
		if _, err := fmt.Fprintln(
			writer,
			"\nApplied migrations:",
		); err != nil {
			return err
		}
		for _, migration := range report.AppliedMigrations {
			if _, err := fmt.Fprintf(
				writer,
				"%s %s %s %s\n",
				migration.Version,
				migration.Name,
				migration.Checksum,
				migration.AppliedAt.Format(
					"2006-01-02T15:04:05Z07:00",
				),
			); err != nil {
				return err
			}
		}
	}

	return nil
}
