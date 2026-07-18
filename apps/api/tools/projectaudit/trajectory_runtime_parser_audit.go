package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func auditTrajectoryRuntimeParser(
	repositoryRoot string,
	goEnums map[string][]string,
) error {
	path := filepath.Join(
		repositoryRoot,
		"apps/web/lib/api/trajectory.ts",
	)

	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf(
			"read trajectory runtime parser %s: %w",
			path,
			err,
		)
	}
	content := string(contentBytes)

	requiredFragments := []string{
		"const flightIdentityBases = new Set<FlightIdentityBasis>([",
		"const flightSplitReasons = new Set<FlightSplitReason>([",
		"identity_key: requireString(",
		"record.identity_key",
		"identity_basis: identityBasis as FlightIdentityBasis",
		"split_reason: splitReason as FlightSplitReason",
	}
	for _, fragment := range requiredFragments {
		if !strings.Contains(
			content,
			fragment,
		) {
			return fmt.Errorf(
				"trajectory runtime parser drifted: missing %q",
				fragment,
			)
		}
	}

	parserIdentityBases, err := parseTypeScriptSetValues(
		path,
		"flightIdentityBases",
	)
	if err != nil {
		return err
	}
	if err := compareStringSets(
		"FlightIdentityBasis",
		goEnums["FlightIdentityBasis"],
		"flightIdentityBases runtime parser set",
		parserIdentityBases,
	); err != nil {
		return err
	}

	parserSplitReasons, err := parseTypeScriptSetValues(
		path,
		"flightSplitReasons",
	)
	if err != nil {
		return err
	}
	if err := compareStringSets(
		"FlightSplitReason",
		goEnums["FlightSplitReason"],
		"flightSplitReasons runtime parser set",
		parserSplitReasons,
	); err != nil {
		return err
	}

	return nil
}

func parseTypeScriptSetValues(
	path string,
	constantName string,
) (
	[]string,
	error,
) {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	pattern := regexp.MustCompile(
		`(?s)const\s+` +
			regexp.QuoteMeta(constantName) +
			`\s*=\s*new\s+Set<[^>]+>\s*\(\s*\[(.*?)\]\s*\)`,
	)
	match := pattern.FindSubmatch(contentBytes)
	if len(match) != 2 {
		return nil, fmt.Errorf(
			"TypeScript Set %s is missing from %s",
			constantName,
			path,
		)
	}

	valuePattern := regexp.MustCompile(
		`'([^']+)'`,
	)
	valueMatches := valuePattern.FindAllSubmatch(
		match[1],
		-1,
	)
	if len(valueMatches) == 0 {
		return nil, fmt.Errorf(
			"TypeScript Set %s is empty",
			constantName,
		)
	}

	values := make(
		[]string,
		0,
		len(valueMatches),
	)
	seen := make(
		map[string]struct{},
	)
	for _, valueMatch := range valueMatches {
		value := string(valueMatch[1])
		if _, exists := seen[value]; exists {
			return nil, fmt.Errorf(
				"TypeScript Set %s contains duplicate %q",
				constantName,
				value,
			)
		}
		seen[value] = struct{}{}
		values = append(
			values,
			value,
		)
	}

	sort.Strings(values)
	return values, nil
}

// STAGE-14-1-AUDIT-FALSE-POSITIVE-FIX
