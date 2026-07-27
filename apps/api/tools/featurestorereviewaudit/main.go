package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type rule struct {
	path      string
	required  []string
	forbidden []string
}

func main() {
	root := flag.String("root", "", "repository root")
	strict := flag.Bool("strict", true, "fail on contract violation")
	flag.Parse()

	resolved, err := resolveRoot(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	rules := []rule{
		{
			path: "apps/api/internal/features/featurestore/contracts.go",
			required: []string{
				`const Version = "flight-feature-store-v2"`,
				"OutputFingerprint string",
				"type SnapshotWriter interface",
				"type SnapshotReader interface",
				"MaximumRecords int",
			},
		},
		{
			path: "apps/api/internal/features/featurestore/snapshot_payload_v1.go",
			required: []string{
				"snapshotPayloadVersionV1",
				"type snapshotPayloadV1 struct",
				"OutputFingerprint string",
				"strictDecodeJSON",
				"validateSnapshotPayloadShape",
				"fingerprintSnapshotOutput",
				"payload.ExtractedAt = time.Time{}",
				"payload.ValidationReport.ValidatedAt = time.Time{}",
			},
		},
		{
			path: "apps/api/internal/features/featurestore/store_contract_validation.go",
			required: []string{
				"canonicalTrajectoryID",
				"uuid.Parse",
				"snapshotFingerprintPattern",
				"validateWritableFeatures",
				"validateWritableValidationProof",
				"validateFiniteFeatureValues",
			},
		},
		{
			path: "apps/api/internal/features/featurestore/memory.go",
			required: []string{
				"requireStoreContext(ctx)",
				"fingerprintSnapshotOutput(normalized)",
				"existing.OutputFingerprint != outputFingerprint",
				"ErrMemoryCapacityExceeded",
			},
			forbidden: []string{"ctx = context.Background()"},
		},
		{
			path: "apps/api/internal/features/featurestore/postgres.go",
			required: []string{
				"encodeSnapshotPayload(normalized)",
				"decodeSnapshotPayload(payload)",
				"existing.OutputFingerprint != outputFingerprint",
				"requireStoreContext(ctx)",
			},
			forbidden: []string{
				"json.Marshal(normalized)",
				"json.Unmarshal(payload, &features)",
				"func nonNilContext(",
			},
		},
		{
			path: "apps/api/internal/features/featurestore/featurestore_review_hardening_test.go",
			required: []string{
				"TestMemoryStoreRejectsSameInputWithDifferentOutput",
				"TestPostgresStoreRejectsSameInputWithDifferentOutput",
				"TestVersionedPayloadRoundTripAndStrictDecoding",
				"TestStoresUseProductionIdentityContracts",
				"TestStoresRequireCompleteValidationProof",
				"TestMemoryStoreEnforcesCapacityWithoutEviction",
			},
		},
		{
			path: ".github/workflows/backend-ci.yml",
			required: []string{
				"Run feature store review audit",
				"go run ./tools/featurestorereviewaudit -strict",
				"./internal/features/featurestore",
				"Verify PostgreSQL feature pipeline",
			},
		},
		{
			path: "docs/117_FEATURE_STORE_REVIEW_HARDENING.md",
			required: []string{
				"FEATURE_STORE_OUTPUT_FINGERPRINT=ENFORCED",
				"FEATURE_STORE_PAYLOAD_VERSIONING=ENFORCED",
				"FEATURE_STORE_IMPLEMENTATION_CONFORMANCE=ENFORCED",
				"FEATURE_STORE_REVIEW_STATUS=CLOSED",
				"OPEN_CONFIRMED_FINDINGS=0",
			},
		},
	}

	failures := make([]string, 0)
	for _, item := range rules {
		content, readErr := os.ReadFile(filepath.Join(resolved, item.path))
		if readErr != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", item.path, readErr))
			continue
		}
		text := string(content)
		for _, fragment := range item.required {
			if !strings.Contains(text, fragment) {
				failures = append(failures, fmt.Sprintf("%s missing %q", item.path, fragment))
			}
		}
		for _, fragment := range item.forbidden {
			if strings.Contains(text, fragment) {
				failures = append(failures, fmt.Sprintf("%s contains forbidden %q", item.path, fragment))
			}
		}
	}

	if len(failures) == 0 {
		fmt.Println("Feature store review audit: PASS")
		return
	}
	fmt.Fprintln(os.Stderr, "Feature store review audit: FAIL")
	for _, failure := range failures {
		fmt.Fprintln(os.Stderr, "-", failure)
	}
	if *strict {
		os.Exit(1)
	}
}

func resolveRoot(explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return filepath.Abs(strings.TrimSpace(explicit))
	}
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, statErr := os.Stat(filepath.Join(current, "apps/api/go.mod")); statErr == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("repository root was not found")
		}
		current = parent
	}
}
