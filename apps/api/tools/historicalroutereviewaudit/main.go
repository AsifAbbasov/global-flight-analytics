package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type requirement struct {
	path      string
	fragments []string
	patterns  []*regexp.Regexp
	forbidden []string
}

func main() {
	strict := flag.Bool(
		"strict",
		false,
		"fail when a Historical Route review contract is absent",
	)
	flag.Parse()

	requirements := []requirement{
		{
			path: "internal/historicalintelligence/historicalroute/contracts.go",
			fragments: []string{
				"historical-route-intelligence-v2",
			},
		},
		{
			path: "internal/historicalintelligence/historicalroute/builder.go",
			fragments: []string{
				"CanonicalizePlan",
				"definition.AllowsScope",
				"validateSnapshotWindow",
				"decodeRouteEvidenceSet",
				"routeDatasetCoverage",
				"routeSourceNames",
				"latestRouteUpdate",
			},
			forbidden: []string{
				"matchesRouteScope",
				"SourceNames: []string",
				"routeCoverage(",
			},
		},
		{
			path: "internal/historicalintelligence/historicalroute/evidence.go",
			fragments: []string{
				"ValidatedResultAt",
				"RouteStatusComplete",
				"greatCircleDistanceKM",
				"meanEarthRadiusKM",
			},
			forbidden: []string{
				"json.Unmarshal",
				"Summary.GreatCircleDistanceKM",
			},
		},
		{
			path: "internal/historicalintelligence/historicalroute/selection.go",
			fragments: []string{
				"ErrRouteRecordIdentityRequired",
				"newerRouteRecord",
				"candidate.StoredAt",
			},
			forbidden: []string{
				"key = strings.TrimSpace(record.ID)",
			},
		},
		{
			path: "internal/historicalintelligence/historicalroute/metrics.go",
			fragments: []string{
				"type compensatedSum struct",
				"routeMetricCalculators",
				"activeRoutePairsValue",
				"ratioMetricValue",
				"greatCircleDistanceValue",
			},
		},
		{
			path: "internal/historicalintelligence/historicalroute/fingerprint.go",
			fragments: []string{
				"item.record.StoredAt.UTC()",
				"route_limit_reached|%t",
				"route_byte_limit_reached|%t",
				"validatedRouteResultDigest",
			},
		},
		{
			path: "internal/historicalintelligence/historicalroute/snapshot.go",
			fragments: []string{
				"windowContains",
				"ErrSnapshotWindowMismatch",
				"ErrRouteScopeCoverageUnavailable",
				"ErrRouteMatchedCountRequired",
				"ErrRouteMatchedCountInvalid",
			},
			forbidden: []string{
				"TotalForSource",
			},
		},
		{
			path: "internal/historicalintelligence/historicalread/route_decode.go",
			fragments: []string{
				"func (record RouteRecord) ValidatedResultAt",
				"routecontract.Validate(result)",
				"validateRoutePersistenceMetadata",
				"trajectory_identity_mismatch",
				"route_status_mismatch",
				"confidence_level_mismatch",
				"input_fingerprint_mismatch",
				"as_of_time_mismatch",
				"validation_warning_count_mismatch",
				"payload_fingerprint_mismatch",
			},
		},
		{
			path: "internal/historicalintelligence/historicalcontract/metric_catalog.go",
			fragments: []string{
				"Unit: \"route_pairs\"",
			},
			patterns: []*regexp.Regexp{
				regexp.MustCompile(`MetricNameCompleteRouteRatio[^\n]+AllowedScopes: \[\]ScopeType\{ScopeTypeGlobal\}`),
				regexp.MustCompile(`MetricNamePartialRouteRatio[^\n]+AllowedScopes: \[\]ScopeType\{ScopeTypeGlobal\}`),
				regexp.MustCompile(`MetricNameUnavailableRouteRatio[^\n]+AllowedScopes: \[\]ScopeType\{ScopeTypeGlobal\}`),
			},
		},
		{
			path: "internal/historicalintelligence/historicalroute/builder_test.go",
			fragments: []string{
				"TestBuildRejectsRoutePairScopeForStatusRatios",
				"TestBuildRejectsInvalidContract",
				"TestBuildRejectsInvalidJSON",
				"TestBuildRejectsPersistenceMetadataMismatch",
				"TestBuildValidatesSnapshotWindowContainment",
				"TestBuildRejectsIncompleteRoutePairCoverage",
				"TestRouteFingerprintBindsStoredAtAndLimitsOnlyWhenRelevant",
				"TestBucketBoundariesAndZeroDenominatorRemainExplicit",
				"TestSameAirportDistanceIsRecomputedAsZero",
			},
		},
		{
			path: "internal/historicalintelligence/historicalread/route_decode_contract_test.go",
			fragments: []string{
				"TestValidatedResultAtAcceptsCompleteContractAndMetadata",
				"TestValidatedResultAtRejectsInvalidRouteContract",
				"TestValidatedResultAtRejectsUnsupportedSchema",
				"TestValidatedResultAtRejectsPersistenceMetadataMismatch",
				"TestValidatedResultAtRejectsFutureEvidence",
				"TestValidatedResultAtRejectsInvalidJSONAndPayloadFingerprint",
			},
		},
		{
			path: "../../docs/128_HISTORICAL_ROUTE_REVIEW_HARDENING.md",
			fragments: []string{
				"Status: closed",
				"ROUTE_PAIR_STATUS_RATIOS=GLOBAL_ONLY",
				"ROUTE_CONTRACT_VALIDATION=ENFORCED",
				"PERSISTENCE_METADATA_RECONCILIATION=ENFORCED",
				"ROUTE_SCOPE_INCOMPLETE_COVERAGE=REJECTED",
				"ROUTE_FINGERPRINT_STORED_AT=BOUND",
				"HISTORICAL_ROUTE_ENGINEERING_REMEDIATION=IMPLEMENTED",
				"9741c4fce04e2b2c06ee0236cf13b5c384f38ffd",
				"513fa1efc7f3b81b895cdc5f881e294d80362e2e",
				"30334131538",
				"Backend Quality=SUCCESS",
				"Backend Quality Job=90195300495",
				"PostgreSQL 16 Integration=SUCCESS",
				"PostgreSQL 16 Integration Job=90195300516",
				"Backend Race Safety=SUCCESS",
				"Backend Race Safety Job=90195300546",
				"Backend Container=SUCCESS",
				"Backend Container Job=90195525282",
				"OPEN_CONFIRMED_FINDINGS=0",
				"UNCLASSIFIED_FINDINGS=0",
				"DEFERRED_FINDINGS=0",
				"HISTORICAL_ROUTE_ENGINEERING_DEBT=CLOSED",
				"HISTORICAL_ROUTE_ADDITIONAL_CODE_FIXES_REQUIRED=NO",
				"FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO",
				"HISTORICAL_ROUTE_REVIEW_STATUS=CLOSED",
			},
		},
	}

	failures := make([]string, 0)
	for _, item := range requirements {
		content, err := os.ReadFile(filepath.Clean(item.path))
		if err != nil {
			failures = append(failures, fmt.Sprintf("read %s: %v", item.path, err))
			continue
		}
		text := string(content)
		for _, fragment := range item.fragments {
			if !strings.Contains(text, fragment) {
				failures = append(failures, fmt.Sprintf("%s misses %q", item.path, fragment))
			}
		}
		for _, pattern := range item.patterns {
			if !pattern.MatchString(text) {
				failures = append(failures, fmt.Sprintf("%s misses pattern %q", item.path, pattern.String()))
			}
		}
		for _, fragment := range item.forbidden {
			if strings.Contains(text, fragment) {
				failures = append(failures, fmt.Sprintf("%s retains forbidden %q", item.path, fragment))
			}
		}
	}

	if len(failures) == 0 {
		fmt.Println("Historical route review audit: PASS")
		return
	}
	for _, failure := range failures {
		fmt.Fprintf(os.Stderr, "Historical route review audit: %s\n", failure)
	}
	if *strict {
		os.Exit(1)
	}
}
