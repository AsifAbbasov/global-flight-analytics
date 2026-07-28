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
		"fail when a Projection Contract review requirement is absent",
	)
	flag.Parse()

	requirements := []requirement{
		{
			path: "internal/domain/confidence/level.go",
			fragments: []string{
				"func (level Level) IsKnown() bool",
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontract/model.go",
			fragments: []string{
				`const Version = "projection-intelligence-contract-v2"`,
				"type ConfidenceLevel = domainconfidence.Level",
				"ConfidenceLevelNone   = domainconfidence.LevelNone",
				"ConfidenceLevelHigh   = domainconfidence.LevelHigh",
			},
			forbidden: []string{
				"type ConfidenceLevel string",
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontract/validation.go",
			fragments: []string{
				`const ValidationVersion = "projection-intelligence-contract-validation-v2"`,
				"IssueHorizonGridInvalid",
				"IssuePointGridInvalid",
				"duration%horizon.Step != 0",
				"expectedTime := result.Horizon.AsOfTime.Add(",
				"IssueLimitedContractInvalid",
				"hasLimitedStatusEvidence(",
				"IssueConfidenceReasonRequired",
				"reason.Contribution < -1",
				"IssueConfidenceReasonDuplicate",
				"IssueConfidenceExceedsEvidence",
				"fingerprintPattern",
				"`^sha256:[0-9a-f]{64}$`",
				"icao24Pattern",
				"`^[0-9A-Fa-f]{6}$`",
				"airportICAOPattern",
				"`^[A-Z]{4}$`",
				"IssueInputDuplicate",
				"IssueInputChronologyInvalid",
				"sourceRequired :=",
				"IssueLimitationDuplicate",
				"IssueExplanationDuplicate",
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontract/result_validation.go",
			fragments: []string{
				"var ErrResultInvalid",
				"type ResultValidationError struct",
				"func (result Result) Validate() error",
			},
		},
		{
			path: "internal/projectionintelligence/projectioncontract/projectioncontract_review_hardening_test.go",
			fragments: []string{
				"TestValidateRejectsNonDivisibleHorizonGrid",
				"TestValidateRejectsProjectionPointOffGrid",
				"TestValidateLimitedStatusRequiresExplicitEvidence",
				"TestValidateRejectsPositiveConfidenceWithoutReasons",
				"TestValidateRejectsOverallConfidenceAboveEvidence",
				"TestValidateDoesNotInventGlobalConfidenceScoreBands",
				"TestValidateRejectsMalformedFingerprint",
				"TestValidateRejectsInvalidICAO24",
				"TestValidateRejectsIncompleteObservedInputEvidence",
				"TestValidateRejectsDuplicateProvenanceInputs",
				"TestValidateAcceptsProducerQualifiedInputNames",
				"TestResultValidateReturnsTypedError",
			},
		},
		{
			path: "../../docs/134_PROJECTION_CONTRACT_REVIEW_HARDENING.md",
			fragments: []string{
				"Status: closed",
				"PROJECTION_GRID_ALIGNMENT=ENFORCED",
				"LIMITED_STATUS_EVIDENCE=REQUIRED",
				"POSITIVE_CONFIDENCE_REASONS=REQUIRED",
				"CONFIDENCE_CONTRIBUTIONS=BOUNDED",
				"RESULT_CONFIDENCE_EVIDENCE_BOUND=ENFORCED",
				"SHARED_CONFIDENCE_VOCABULARY=USED",
				"PROJECTION_FINGERPRINT_FORMAT=SHA256",
				"PROVENANCE_CHRONOLOGY=ENFORCED",
				"PROVENANCE_DUPLICATES=REJECTED",
				"ICAO24_FORMAT=HEX_24_BIT",
				"AIRPORT_ICAO_FORMAT=FOUR_LETTERS",
				"DOMAIN_OPTIONAL_POINTERS=RETAINED",
				"PUBLIC_RESULT_STRUCT=RETAINED_WITH_VALIDATE",
				"FIXED_POINT_CONFIDENCE_SCORE=REJECTED_FOR_NOW",
				"PHYSICAL_POLICY_LIMITS=PRODUCER_OWNED",
				"PROJECTION_CONTRACT_ENGINEERING_REMEDIATION=IMPLEMENTED",
				"405b141c431b7cd3e8b8150e88ac238924992e15",
				"964556d0ca8a1ce9aa74c37c55961cdd006b3de8",
				"30396070318",
				"Backend Quality=SUCCESS",
				"Backend Quality Job=90399157528",
				"PostgreSQL 16 Integration=SUCCESS",
				"PostgreSQL 16 Integration Job=90399157430",
				"Backend Race Safety=SUCCESS",
				"Backend Race Safety Job=90399157564",
				"Backend Container=SUCCESS",
				"Backend Container Job=90399476002",
				"OPEN_CONFIRMED_FINDINGS=0",
				"UNCLASSIFIED_FINDINGS=0",
				"DEFERRED_FINDINGS=0",
				"PROJECTION_CONTRACT_ENGINEERING_DEBT=CLOSED",
				"PROJECTION_CONTRACT_ADDITIONAL_CODE_FIXES_REQUIRED=NO",
				"FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO",
				"PROJECTION_CONTRACT_REVIEW_STATUS=CLOSED",
			},
		},
		{
			path: "../../.github/workflows/backend-ci.yml",
			fragments: []string{
				"go run ./tools/projectioncontractreviewaudit -strict",
			},
		},
	}

	failures := make([]string, 0)
	for _, item := range requirements {
		content, err := os.ReadFile(filepath.Clean(item.path))
		if err != nil {
			failures = append(
				failures,
				fmt.Sprintf("read %s: %v", item.path, err),
			)
			continue
		}
		text := string(content)
		for _, fragment := range item.fragments {
			if !strings.Contains(text, fragment) {
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
			if strings.Contains(text, fragment) {
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
		fmt.Println("Projection contract review audit: PASS")
		return
	}
	for _, failure := range failures {
		fmt.Fprintf(
			os.Stderr,
			"Projection contract review audit: %s\n",
			failure,
		)
	}
	if *strict {
		os.Exit(1)
	}
}
