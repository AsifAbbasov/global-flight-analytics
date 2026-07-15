package main

import (
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

func TestVerificationResultProducesValidHistoricalContract(
	t *testing.T,
) {
	now := time.Date(
		2026,
		time.July,
		15,
		12,
		0,
		0,
		123456789,
		time.UTC,
	)

	result, err := verificationResult(now)
	if err != nil {
		t.Fatalf(
			"build verification result: %v",
			err,
		)
	}

	report := historicalcontract.Validate(result)
	if report.Status !=
		historicalcontract.ValidationStatusValid {
		t.Fatalf(
			"expected valid historical contract, got errors=%d warnings=%d issues=%#v",
			report.ErrorCount,
			report.WarningCount,
			report.Issues,
		)
	}
	if result.Status !=
		historicalcontract.SeriesStatusComplete {
		t.Fatalf(
			"expected complete result, got %s",
			result.Status,
		)
	}
	if result.Metric.Name !=
		historicalcontract.MetricNameFlightCount {
		t.Fatalf(
			"unexpected metric: %s",
			result.Metric.Name,
		)
	}
	if result.Scope.Type !=
		historicalcontract.ScopeTypeGlobal {
		t.Fatalf(
			"unexpected scope: %s",
			result.Scope.Type,
		)
	}
	if len(result.Points) != 2 ||
		result.Summary.Total != 5 ||
		result.Confidence.SampleCount != 5 {
		t.Fatalf(
			"unexpected verification evidence: points=%d total=%f samples=%d",
			len(result.Points),
			result.Summary.Total,
			result.Confidence.SampleCount,
		)
	}
	if result.Provenance.InputFingerprint !=
		"sha256:"+strings.Repeat("a", 64) {
		t.Fatalf(
			"unexpected fingerprint: %s",
			result.Provenance.InputFingerprint,
		)
	}
	if !result.GeneratedAt.Equal(now) {
		t.Fatalf(
			"generated time = %s, want %s",
			result.GeneratedAt,
			now,
		)
	}
}

func TestMigrationIdentityIsPinned(
	t *testing.T,
) {
	if expectedMigrationVersion != "015" {
		t.Fatalf(
			"unexpected migration version: %s",
			expectedMigrationVersion,
		)
	}
	if expectedMigrationName !=
		"create_historical_aggregate_results" {
		t.Fatalf(
			"unexpected migration name: %s",
			expectedMigrationName,
		)
	}
	if len(expectedMigrationChecksum) != 64 {
		t.Fatalf(
			"unexpected migration checksum length: %d",
			len(expectedMigrationChecksum),
		)
	}
}
