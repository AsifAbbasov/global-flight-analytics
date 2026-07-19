package handlers

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregatecontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/historicalcursor"
)

func TestParseHistoricalLatestQueryBuildsBaseQuery(
	t *testing.T,
) {
	query, err := parseHistoricalLatestQuery(
		historicalIntelligenceQueryValues{
			Metric:      "flight_count",
			Scope:       "global",
			Granularity: "hour",
			Limit:       "invalid-for-history",
			Cursor:      "invalid-for-history",
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
		query.Cursor != nil {
		t.Fatalf(
			"latest query contains history pagination: %#v",
			query,
		)
	}
}

func TestParseHistoricalHistoryQueryAddsCompositePagination(
	t *testing.T,
) {
	cursor := historicalaggregatecontract.ListCursor{
		WindowEnd: time.Date(
			2026,
			time.July,
			19,
			9,
			30,
			0,
			123,
			time.UTC,
		),
		WindowStart: time.Date(
			2026,
			time.July,
			19,
			8,
			30,
			0,
			123,
			time.UTC,
		),
		AsOfTime: time.Date(
			2026,
			time.July,
			19,
			9,
			35,
			0,
			123,
			time.UTC,
		),
		ID: "historical-aggregate-record-a",
	}
	encoded, err := historicalcursor.Encode(cursor)
	if err != nil {
		t.Fatalf(
			"encode cursor: %v",
			err,
		)
	}

	query, err := parseHistoricalHistoryQuery(
		historicalIntelligenceQueryValues{
			Metric:      "flight_count",
			Scope:       "global",
			Granularity: "hour",
			Limit:       "7",
			Cursor:      encoded,
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
	if query.Cursor == nil ||
		!query.Cursor.WindowEnd.Equal(
			cursor.WindowEnd,
		) ||
		!query.Cursor.WindowStart.Equal(
			cursor.WindowStart,
		) ||
		!query.Cursor.AsOfTime.Equal(
			cursor.AsOfTime,
		) ||
		query.Cursor.ID != cursor.ID {
		t.Fatalf(
			"unexpected parsed cursor: %#v",
			query.Cursor,
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
			Metric:      "flight_count",
			Scope:       "global",
			Granularity: "hour",
			Limit:       "1",
			Cursor:      "not-a-cursor",
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
