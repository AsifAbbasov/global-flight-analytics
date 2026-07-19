package dto

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/historicalcursor"
)

func TestToHistoricalIntelligenceAggregateRecordUsesStableJSONContract(
	t *testing.T,
) {
	record := historicalIntelligenceDTORecord()
	converted :=
		ToHistoricalIntelligenceAggregateRecord(
			record,
		)

	if converted.ID != record.ID ||
		converted.Result.SchemaVersion !=
			string(
				historicalcontract.
					SchemaVersionV1,
			) ||
		converted.Result.Metric.Name !=
			"flight_count" ||
		converted.Result.Scope.Type !=
			"global" ||
		len(converted.Result.Points) != 1 ||
		converted.Result.Comparison == nil ||
		converted.Result.Comparison.
			PercentageChange == nil ||
		*converted.Result.Comparison.
			PercentageChange != 100 {
		t.Fatalf(
			"unexpected converted record: %#v",
			converted,
		)
	}

	payload, err := json.Marshal(converted)
	if err != nil {
		t.Fatalf(
			"marshal converted record: %v",
			err,
		)
	}
	text := string(payload)
	for _, fragment := range []string{
		`"schema_version"`,
		`"input_fingerprint"`,
		`"coverage_ratio"`,
		`"previous_window"`,
		`"latest_source_updated_at"`,
	} {
		if !strings.Contains(
			text,
			fragment,
		) {
			t.Fatalf(
				"JSON does not contain %q: %s",
				fragment,
				text,
			)
		}
	}
}

func TestToHistoricalIntelligenceAggregateHistoryBuildsCursor(
	t *testing.T,
) {
	record := historicalIntelligenceDTORecord()
	cursor := &historicalaggregate.ListCursor{
		WindowEnd: record.Key.Window.EndTime,
		WindowStart: record.Key.Window.
			StartTime,
		AsOfTime: record.Key.Window.AsOfTime,
		ID:       record.ID,
	}
	history, err :=
		ToHistoricalIntelligenceAggregateHistory(
			historicalaggregate.Page{
				Records: []historicalaggregate.Record{
					record,
				},
				HasMore:    true,
				NextCursor: cursor,
			},
		)
	if err != nil {
		t.Fatalf(
			"convert history response: %v",
			err,
		)
	}

	decoded, err := historicalcursor.Decode(
		history.NextCursor,
	)
	if err != nil {
		t.Fatalf(
			"decode history cursor: %v",
			err,
		)
	}
	if len(history.Items) != 1 ||
		!history.HasMore ||
		history.NextCursor == "" ||
		decoded == nil ||
		!decoded.WindowEnd.Equal(
			cursor.WindowEnd,
		) ||
		!decoded.WindowStart.Equal(
			cursor.WindowStart,
		) ||
		!decoded.AsOfTime.Equal(
			cursor.AsOfTime,
		) ||
		decoded.ID != cursor.ID {
		t.Fatalf(
			"unexpected history response: %#v decoded=%#v",
			history,
			decoded,
		)
	}

	withoutMore, err :=
		ToHistoricalIntelligenceAggregateHistory(
			historicalaggregate.Page{
				Records: []historicalaggregate.Record{
					record,
				},
				HasMore: false,
			},
		)
	if err != nil {
		t.Fatalf(
			"convert terminal history response: %v",
			err,
		)
	}
	if withoutMore.NextCursor != "" {
		t.Fatalf(
			"unexpected cursor without more records: %q",
			withoutMore.NextCursor,
		)
	}
}

func TestHistoricalIntelligenceDTOCopiesMutableState(
	t *testing.T,
) {
	record := historicalIntelligenceDTORecord()
	converted :=
		ToHistoricalIntelligenceAggregateRecord(
			record,
		)

	converted.Result.Provenance.
		SourceNames[0] = "changed"
	*converted.Result.Comparison.
		PercentageChange = 999

	if record.Result.Provenance.
		SourceNames[0] == "changed" {
		t.Fatal(
			"source names share mutable state",
		)
	}
	if *record.Result.Comparison.
		PercentageChange == 999 {
		t.Fatal(
			"percentage comparison shares mutable state",
		)
	}
}

func historicalIntelligenceDTORecord() historicalaggregate.Record {
	percentage := 100.0
	endTime := time.Date(
		2026,
		time.July,
		15,
		12,
		0,
		0,
		0,
		time.UTC,
	)
	startTime := endTime.Add(
		-time.Hour,
	)
	asOfTime := endTime.Add(
		30 * time.Minute,
	)
	fingerprint := "sha256:" +
		strings.Repeat("a", 64)

	result := historicalcontract.Result{
		SchemaVersion: historicalcontract.SchemaVersionV1,
		Status:        historicalcontract.SeriesStatusComplete,
		Metric: historicalcontract.Metric{
			Name: historicalcontract.
				MetricNameFlightCount,
			Unit: "flights",
			Aggregation: historicalcontract.
				AggregationCount,
		},
		Scope: historicalcontract.Scope{
			Type: historicalcontract.
				ScopeTypeGlobal,
		},
		Window: historicalcontract.TimeWindow{
			StartTime: startTime,
			EndTime:   endTime,
			AsOfTime:  asOfTime,
		},
		Granularity: historicalcontract.GranularityHour,
		Points: []historicalcontract.Point{
			{
				StartTime: startTime,
				EndTime:   endTime,
				Status: historicalcontract.
					BucketStatusComplete,
				Value:         2,
				SampleCount:   2,
				CoverageRatio: 1,
				Confidence: historicalcontract.
					Confidence{
					Score: 1,
					Level: historicalcontract.
						ConfidenceLevelHigh,
					SampleCount: 2,
					Reasons: []historicalcontract.
						ConfidenceReason{
						{
							Code:         "coverage",
							Message:      "Coverage is complete.",
							Contribution: 1,
						},
					},
				},
				Limitations: []historicalcontract.Limitation{},
			},
		},
		Summary: historicalcontract.Summary{
			PointCount: 1,
			Total:      2,
			Minimum:    2,
			Maximum:    2,
			Average:    2,
			Median:     2,
		},
		Comparison: &historicalcontract.PeriodComparison{
			PreviousWindow: historicalcontract.TimeWindow{
				StartTime: startTime.Add(
					-time.Hour,
				),
				EndTime:  startTime,
				AsOfTime: asOfTime,
			},
			CurrentValue:     2,
			PreviousValue:    1,
			AbsoluteChange:   1,
			PercentageChange: &percentage,
			Direction: historicalcontract.
				TrendDirectionUp,
		},
		Confidence: historicalcontract.Confidence{
			Score: 1,
			Level: historicalcontract.
				ConfidenceLevelHigh,
			SampleCount: 2,
			Reasons: []historicalcontract.
				ConfidenceReason{},
		},
		Limitations: []historicalcontract.Limitation{},
		Provenance: historicalcontract.Provenance{
			BuilderVersion:   "historical-traffic-intelligence-v1",
			InputFingerprint: fingerprint,
			SourceNames: []string{
				"flights",
			},
			LatestSourceUpdatedAt: endTime,
		},
		GeneratedAt: asOfTime,
	}

	return historicalaggregate.Record{
		ID: "historical-aggregate-record-" +
			strings.Repeat("b", 64),
		Key: historicalaggregate.ResultKey{
			SchemaVersion: result.SchemaVersion,
			MetricName:    result.Metric.Name,
			Scope:         result.Scope,
			Granularity:   result.Granularity,
			Window:        result.Window,
		},
		InputFingerprint: fingerprint,
		Result:           result,
		StoredAt:         asOfTime,
	}
}
