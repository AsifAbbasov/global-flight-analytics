package main

import (
	"fmt"
	"path/filepath"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database/contextaudit"
)

func auditMigratorContextSyntax(root string) []auditFailure {
	directory := filepath.Join(
		root,
		"apps",
		"api",
		"internal",
		"database",
		"migrator",
	)

	violations, err := contextaudit.AuditDirectory(
		directory,
		contextaudit.MigratorPolicy(),
	)
	if err != nil {
		return []auditFailure{{
			Check:  "Migrator context syntax policy",
			Detail: fmt.Sprintf("run Go syntax-tree audit: %v", err),
		}}
	}

	failures := make([]auditFailure, 0, len(violations))
	for _, violation := range violations {
		failures = append(failures, auditFailure{
			Check:  "Migrator context syntax policy",
			Detail: violation.String(),
		})
	}
	return failures
}
