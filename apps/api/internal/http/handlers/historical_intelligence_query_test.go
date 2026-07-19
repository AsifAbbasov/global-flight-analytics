package handlers

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregatecontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

func TestParseHistoricalLatestQueryBuildsBaseQuery(
	t *testing.T,
) {
	query, err := parseHistoricalLatestQuery(
		historicalIntelligenceQueryValues{
			Metric:          "flight_count",
			Scope:           "global",
			Granularity:     "hour",
			Limit:           "invalid-for-history",
			BeforeWindowEnd: "invalid-for-history",
		},
	)
	if err != nil {
		t.Fatalf(
			"parse latest query: %v",
			err,
		)
	}

	if query.SchemaVersion !=
		historicalcontract.SchemaVersionV1 {
		t.Fatalf(
			"schema version = %q",
			query.SchemaVersion,
		)
	}
	if query.MetricName !=
		historicalcontract.MetricNameFlightCount {
		t.Fatalf(
			"metric = %q",
			query.MetricName,
		)
	}
	if query.Scope.Type !=
		historicalcontract.ScopeTypeGlobal {
		t.Fatalf(
			"scope = %q",
			query.Scope.Type,
		)
	}
	if query.Granularity !=
		historicalcontract.GranularityHour {
		t.Fatalf(
			"granularity = %q",
			query.Granularity,
		)
	}
	if query.Limit != 0 ||
		!query.BeforeWindowEnd.IsZero() {
		t.Fatalf(
			"latest query contains history pagination: %#v",
			query,
		)
	}
}

func TestParseHistoricalHistoryQueryAddsPagination(
	t *testing.T,
) {
	before := time.Date(
		2026,
		time.July,
		19,
		9,
		30,
		0,
		123,
		time.UTC,
	)

	query, err := parseHistoricalHistoryQuery(
		historicalIntelligenceQueryValues{
			Metric:          "flight_count",
			Scope:           "global",
			Granularity:     "hour",
			Limit:           "7",
			BeforeWindowEnd: before.Format(time.RFC3339Nano),
		},
	)
	if err != nil {
		t.Fatalf(
			"parse history query: %v",
			err,
		)
	}

	if query.Limit != 7 {
		t.Fatalf(
			"limit = %d, want 7",
			query.Limit,
		)
	}
	if !query.BeforeWindowEnd.Equal(before) {
		t.Fatalf(
			"cursor = %s, want %s",
			query.BeforeWindowEnd,
			before,
		)
	}
}

func TestParseHistoricalHistoryQueryRejectsInvalidPagination(
	t *testing.T,
) {
	_, err := parseHistoricalHistoryQuery(
		historicalIntelligenceQueryValues{
			Metric:      "flight_count",
			Scope:       "global",
			Granularity: "hour",
			Limit:       "0",
		},
	)
	if !errors.Is(
		err,
		errHistoricalLimitInvalid,
	) {
		t.Fatalf(
			"invalid limit error = %v",
			err,
		)
	}

	_, err = parseHistoricalHistoryQuery(
		historicalIntelligenceQueryValues{
			Metric:          "flight_count",
			Scope:           "global",
			Granularity:     "hour",
			Limit:           "1",
			BeforeWindowEnd: "not-a-timestamp",
		},
	)
	if !errors.Is(
		err,
		errHistoricalCursorInvalid,
	) {
		t.Fatalf(
			"invalid cursor error = %v",
			err,
		)
	}
}

func TestHistoricalAggregateContractLimitsRemainAuthoritative(
	t *testing.T,
) {
	if historicalaggregatecontract.
		DefaultListLimit != 20 {
		t.Fatalf(
			"default list limit = %d",
			historicalaggregatecontract.
				DefaultListLimit,
		)
	}
	if historicalaggregatecontract.
		MaximumListLimit != 100 {
		t.Fatalf(
			"maximum list limit = %d",
			historicalaggregatecontract.
				MaximumListLimit,
		)
	}
}
