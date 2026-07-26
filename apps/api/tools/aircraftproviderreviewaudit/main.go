package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	root := flag.String("root", "", "repository root")
	strict := flag.Bool("strict", true, "fail on contract violation")
	flag.Parse()

	resolved, err := resolveRoot(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	failures := make([]string, 0)
	require := func(path string, fragments ...string) {
		content, readErr := os.ReadFile(filepath.Join(resolved, path))
		if readErr != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", path, readErr))
			return
		}
		text := string(content)
		for _, fragment := range fragments {
			if !strings.Contains(text, fragment) {
				failures = append(failures, fmt.Sprintf("%s missing %q", path, fragment))
			}
		}
	}
	reject := func(path string, fragments ...string) {
		content, readErr := os.ReadFile(filepath.Join(resolved, path))
		if readErr != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", path, readErr))
			return
		}
		text := string(content)
		for _, fragment := range fragments {
			if strings.Contains(text, fragment) {
				failures = append(failures, fmt.Sprintf("%s contains forbidden %q", path, fragment))
			}
		}
	}

	require(
		"apps/api/internal/features/aircraftprovider/contracts.go",
		`const Version = "aircraft-feature-provider-v4"`,
		`DefaultNotFoundPolicyVersion = "aircraft-domain-not-found-v2"`,
		"DefaultMaxCacheEntries",
		"DefaultLookupTimeout",
	)
	require(
		"apps/api/internal/features/aircraftprovider/provider.go",
		"func (provider *Provider) acquire(",
		"context.WithoutCancel(ctx)",
		"context.WithTimeout(",
		"provider.pruneExpiredLocked(now)",
		"provider.evictOneLocked()",
		"ErrAircraftIdentityMissing",
		"errors.Is(err, aircraft.ErrNotFound)",
	)
	reject(
		"apps/api/internal/features/aircraftprovider/provider.go",
		"github.com/jackc/pgx/v5",
		"func (provider *Provider) cached(",
		"func (provider *Provider) beginCall(",
		"ctx = context.Background()",
	)
	require(
		"apps/api/internal/features/aircraftprovider/provider_review_hardening_test.go",
		"TestProviderRecognizesDomainAircraftNotFoundByDefault",
		"TestLeaderCancellationDoesNotCancelActiveWaiter",
		"TestAcquireAtomicallyUsesCompletedCache",
		"TestProviderBoundsCacheAndEvictsOldestExpiry",
		"TestProviderPrunesExpiredUniqueEntries",
		"TestProviderRejectsLookupWithoutAircraftIdentity",
		"TestProviderRejectsNilContext",
		"TestSharedLookupHasBoundedLifetime",
	)
	require(
		"apps/api/internal/features/extractorcomposition/composition_test.go",
		`AircraftProvider:    "aircraft-feature-provider-v4"`,
	)
	require(
		".github/workflows/backend-ci.yml",
		"Run aircraft provider review audit",
		"go run ./tools/aircraftproviderreviewaudit -strict",
	)
	require(
		"docs/116_AIRCRAFT_PROVIDER_REVIEW_HARDENING.md",
		"AIRCRAFT_PROVIDER_ACQUIRE=ATOMIC",
		"SHARED_LOOKUP_CANCELLATION=ISOLATED",
		"AIRCRAFT_CACHE_CAPACITY=BOUNDED",
		"AIRCRAFT_DOMAIN_NOT_FOUND=DEFAULT",
		"AIRCRAFT_LOOKUP_IDENTITY=REQUIRED",
		"AIRCRAFT_PROVIDER_REVIEW_STATUS=CLOSED",
	)

	if len(failures) == 0 {
		fmt.Println("Aircraft provider review audit: PASS")
		return
	}
	fmt.Fprintln(os.Stderr, "Aircraft provider review audit: FAIL")
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
