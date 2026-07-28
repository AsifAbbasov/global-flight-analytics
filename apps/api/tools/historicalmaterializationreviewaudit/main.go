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
		"fail when a Historical Materialization review contract is absent",
	)
	flag.Parse()

	requirements := []requirement{
		{
			path: "internal/historicalintelligence/historicalread/contracts.go",
			fragments: []string{
				"type PeriodQueries struct",
				"type PeriodSnapshots struct",
				"type PeriodRepository interface",
				"func (query Query) Equal(",
			},
		},
		{
			path: "internal/historicalintelligence/historicalread/periods.go",
			fragments: []string{
				"func (repository *PostgresRepository) ReadPeriods(",
				"repository.beginner.BeginSnapshot(ctx)",
				"queries.Previous",
				"queries.Current",
				"SnapshotIsolationRepeatableRead",
				"ErrPeriodWindowsNotAdjacent",
				"ErrPeriodAsOfTimeMismatch",
			},
		},
		{
			path: "internal/historicalintelligence/historicalread/periods_test.go",
			fragments: []string{
				"TestPostgresRepositoryReadPeriodsUsesOneManagedSnapshotAndIndependentLimits",
				"query call count=%d want=10",
				"TestPostgresRepositoryReadPeriodsRollsBackCurrentPeriodFailure",
				"TestPostgresRepositoryReadPeriodsRejectsNilContext",
			},
		},
		{
			path: "internal/historicalintelligence/historicalmaterialization/contracts.go",
			fragments: []string{
				`const Version = "historical-materialization-v2"`,
				"Repository historicalread.Repository",
				"Store      historicalaggregate.Writer",
				"type PeriodReadSummaries struct",
				"CurrentPeriodResult historicalcontract.Result",
				"ReadSummaries PeriodReadSummaries",
				"Deprecated: ReadSummary is the aggregate",
			},
		},
		{
			path: "internal/historicalintelligence/historicalmaterialization/materializer.go",
			fragments: []string{
				"ErrContextRequired",
				"config.Repository.(",
				"historicalread.PeriodRepository",
				"ErrPeriodReadRepositoryRequired",
				"repository.ReadPeriods(",
				"validatePeriodSnapshots(",
				"queries.Previous",
				"queries.Current",
				"historicalcomparison.Attach(",
				"CurrentResult:",
				"record.Result.Clone()",
				"CurrentPeriodResult:",
				"validatePersistedRecord(",
				"StagePersistenceContract",
			},
			forbidden: []string{
				"context.Background()",
				"readWindow :=",
				"repository.Read(",
				"finalizeComparedResult(",
				"materializationFingerprint(",
				"func (request Request) DatasetLimitOr(",
			},
		},
		{
			path: "internal/historicalintelligence/historicalmaterialization/errors.go",
			fragments: []string{
				"type StageError struct",
				"ErrSnapshotVersionMismatch",
				"ErrSnapshotQueryMismatch",
				"ErrSnapshotIsolationMismatch",
				"ErrPersistedRecordMismatch",
			},
		},
		{
			path: "internal/historicalintelligence/historicalmaterialization/materializer_test.go",
			fragments: []string{
				"TestMaterializeReadsAdjacentPeriodsIndependentlyAndPersistsComparison",
				"TestMaterializeRejectsSnapshotContractViolations",
				"TestMaterializeRejectsNilContext",
				"TestMaterializeWrapsRepositoryAndPersistenceStages",
				"TestMaterializeUsesPersistedResultAsCanonicalOutcome",
				"TestMaterializeRejectsPersistedRecordIdentityMismatch",
				"TestMaterializeBindsGeneratedAtIntoFingerprint",
				"TestMaterializeUsesDefaultAndMaximumDatasetLimits",
				"TestOutcomeCloneIsolatesNestedResults",
			},
			forbidden: []string{
				"func (store *fakeAggregateStore) Get(",
				"unexpected combined read window",
			},
		},
		{
			path: "../../docs/132_HISTORICAL_MATERIALIZATION_REVIEW_HARDENING.md",
			fragments: []string{
				"Status: implemented review remediation",
				"INDEPENDENT_PERIOD_LIMITS=ENFORCED",
				"ATOMIC_TWO_PERIOD_READ=ENFORCED",
				"SNAPSHOT_QUERY_AND_VERSION=VERIFIED",
				"PERIOD_READ_SUMMARIES=EXPLICIT",
				"PERSISTED_RESULT_IS_CANONICAL=YES",
				"GENERATED_AT_FINGERPRINT_IDENTITY=BOUND",
				"NIL_CONTEXT_REJECTED=YES",
				"STAGE_ERRORS=EXPLICIT",
				"HISTORICAL_MATERIALIZATION_ENGINEERING_REMEDIATION=IMPLEMENTED",
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
			"Historical materialization review audit: PASS",
		)
		return
	}
	for _, failure := range failures {
		fmt.Fprintf(
			os.Stderr,
			"Historical materialization review audit: %s\n",
			failure,
		)
	}
	if *strict {
		os.Exit(1)
	}
}
