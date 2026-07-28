package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type requirement struct {
	path      string
	fragments []string
	forbidden []string
}

func main() {
	strict := flag.Bool(
		"strict",
		false,
		"fail when a Historical Aggregate review contract is absent",
	)
	flag.Parse()

	requirements := []requirement{
		{
			path: "internal/historicalintelligence/historicalaggregatecontract/contracts.go",
			fragments: []string{
				"type Writer interface",
				"ErrResultPayloadConflict",
				"ErrContextRequired",
				"ErrStoredAtInvalid",
			},
		},
		{
			path: "internal/historicalintelligence/historicalaggregatecontract/pagination.go",
			fragments: []string{
				"WindowEnd",
				"WindowStart",
				"AsOfTime",
				"ID",
				"NormalizeListCursor",
			},
		},
		{
			path: "internal/historicalintelligence/historicalaggregate/contracts.go",
			fragments: []string{
				"type Writer =",
				"type ListCursor =",
			},
		},
		{
			path: "internal/historicalintelligence/historicalmaterialization/contracts.go",
			fragments: []string{
				"Store      historicalaggregate.Writer",
			},
			forbidden: []string{
				"Store      historicalaggregate.Store",
			},
		},
		{
			path: "internal/historicalintelligence/historicalaggregate/helpers.go",
			fragments: []string{
				"strings.ToLower(",
				"func normalizeResult(",
			},
			forbidden: []string{
				"func nonNilContext(",
			},
		},
		{
			path: "internal/historicalintelligence/historicalaggregate/postgres.go",
			fragments: []string{
				"storedRecordColumns",
				"scope_key",
				"series_status",
				"confidence_level",
				"listResultsAfterCursorSQL",
				"AND id > $8",
				"requireContext(ctx)",
				"resultsHaveSameCanonicalPayload",
				"ErrResultPayloadConflict",
				"validateStoredAt",
				"scanStoredRow",
				"recordFromStoredRow",
			},
			forbidden: []string{
				"ctx = nonNilContext(ctx)",
			},
		},
		{
			path: "internal/historicalintelligence/historicalaggregate/stored_record.go",
			fragments: []string{
				"type storedRow struct",
				"resultKeysEqual",
				"expectedScopeKey",
				"expectedID := makeRecordID",
				"series_status",
				"confidence_level",
				"resultsHaveSameCanonicalPayload",
				"ErrContextRequired",
				"ErrStoredAtInvalid",
			},
		},
		{
			path: "internal/historicalintelligence/historicalaggregate/postgres_test.go",
			fragments: []string{
				"TestPostgresStoreRejectsNilContext",
				"TestPostgresStoreValidatesBeforeCanonicalization",
				"TestPostgresStoreRejectsSameFingerprintDifferentPayload",
				"TestScanRecordRejectsStoredMetadataMismatch",
				"TestScanRecordRejectsStoredIdentifierMismatch",
				"TestPostgresStoreListUsesFullTupleCursorSQL",
				"TestPostgresStoreRejectsInvalidStoredAt",
			},
		},
		{
			path: "internal/historicalintelligence/historicalaggregate/migration_contract_test.go",
			fragments: []string{
				"029_harden_historical_aggregate_integrity.sql",
				"timestamp_mirror_check",
				"json_metadata_check",
			},
		},
		{
			path: "../../database/migrations/029_harden_historical_aggregate_integrity.sql",
			fragments: []string{
				"region_code ~ '^[a-z0-9][a-z0-9_-]{1,31}$'",
				"historical_aggregate_results_timestamp_mirror_check",
				"historical_aggregate_results_json_metadata_check",
				"historical_aggregate_results_stored_at_causality_check",
				"result_json #>> '{Metric,Name}'",
			},
			forbidden: []string{
				"region_code ~ '^[A-Z0-9_-]",
			},
		},
		{
			path: "cmd/verify-postgres-historical-aggregate-store/main.go",
			fragments: []string{
				`expectedMigrationVersion  = "029"`,
				`expectedMigrationName     = "harden_historical_aggregate_integrity"`,
				"ErrResultPayloadConflict",
				"Payload identity conflict detection: PASS",
			},
		},
		{
			path: "internal/database/migrationfile/production_catalog_regression_test.go",
			fragments: []string{
				`29: "029_harden_historical_aggregate_integrity.sql"`,
				"len(orderedVersions) != 29",
			},
		},
		{
			path: "internal/historicalintelligence/historicalcontract/metric_catalog.go",
			fragments: []string{
				"single production catalog for metric identity, unit",
				"func MetricSpecFor(",
			},
		},
		{
			path: "../../docs/131_HISTORICAL_AGGREGATE_REVIEW_HARDENING.md",
			fragments: []string{
				"Status: closed",
				"REGION_SCOPE_DATABASE_CONTRACT=LOWERCASE",
				"FULL_TUPLE_CURSOR=PREEXISTING_AND_VERIFIED",
				"STORED_METADATA_JSON_CONSISTENCY=ENFORCED",
				"STORED_RECORD_IDENTITY=ENFORCED",
				"IDEMPOTENCY_CANONICAL_PAYLOAD=ENFORCED",
				"MATERIALIZER_WRITER_INTERFACE=ENFORCED",
				"RAW_DOMAIN_VALIDATION_BEFORE_CANONICALIZATION=ENFORCED",
				"NIL_CONTEXT_REJECTED=YES",
				"STORED_AT_CAUSALITY=ENFORCED",
				"TIMESTAMP_MIRROR_DATABASE_CONSISTENCY=ENFORCED",
				"HISTORICAL_AGGREGATE_ENGINEERING_REMEDIATION=IMPLEMENTED",
				"82ebd68d0372c885d724308d2291c61dab2de378",
				"18dde73b2d122d00476ea21accb256b33fc23527",
				"30374964285",
				"Backend Quality=SUCCESS",
				"Backend Quality Job=90328145394",
				"Backend Race Safety=SUCCESS",
				"Backend Race Safety Job=90328145512",
				"PostgreSQL 16 Integration=SUCCESS",
				"PostgreSQL 16 Integration Job=90328145602",
				"Backend Container=SUCCESS",
				"Backend Container Job=90328492165",
				"OPEN_CONFIRMED_FINDINGS=0",
				"UNCLASSIFIED_FINDINGS=0",
				"DEFERRED_FINDINGS=0",
				"HISTORICAL_AGGREGATE_ENGINEERING_DEBT=CLOSED",
				"HISTORICAL_AGGREGATE_ADDITIONAL_CODE_FIXES_REQUIRED=NO",
				"FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO",
				"HISTORICAL_AGGREGATE_REVIEW_STATUS=CLOSED",
			},
		},
	}

	failures := make([]string, 0)
	for _, item := range requirements {
		content, err := os.ReadFile(
			filepath.Clean(item.path),
		)
		if err != nil {
			failures = append(
				failures,
				fmt.Sprintf(
					"read %s: %v",
					item.path,
					err,
				),
			)
			continue
		}
		text := string(content)
		for _, fragment := range item.fragments {
			if !strings.Contains(
				text,
				fragment,
			) {
				failures = append(
					failures,
					fmt.Sprintf(
						"%s misses %q",
						item.path,
						fragment,
					),
				)
			}
		}
		for _, fragment := range item.forbidden {
			if strings.Contains(
				text,
				fragment,
			) {
				failures = append(
					failures,
					fmt.Sprintf(
						"%s retains forbidden %q",
						item.path,
						fragment,
					),
				)
			}
		}
	}

	if len(failures) == 0 {
		fmt.Println(
			"Historical aggregate review audit: PASS",
		)
		return
	}
	for _, failure := range failures {
		fmt.Fprintf(
			os.Stderr,
			"Historical aggregate review audit: %s\n",
			failure,
		)
	}
	if *strict {
		os.Exit(1)
	}
}
