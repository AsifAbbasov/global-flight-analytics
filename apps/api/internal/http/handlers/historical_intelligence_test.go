package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/gofiber/fiber/v2"
)

type historicalIntelligenceStoreStub struct {
	latest historicalaggregate.Record
	page   historicalaggregate.Page

	latestErr error
	listErr   error

	latestQuery historicalaggregate.ListQuery
	listQuery   historicalaggregate.ListQuery
}

func (stub *historicalIntelligenceStoreStub) GetLatest(
	_ context.Context,
	query historicalaggregate.ListQuery,
) (historicalaggregate.Record, error) {
	stub.latestQuery = query
	return stub.latest.Clone(),
		stub.latestErr
}

func (stub *historicalIntelligenceStoreStub) List(
	_ context.Context,
	query historicalaggregate.ListQuery,
) (historicalaggregate.Page, error) {
	stub.listQuery = query
	return stub.page.Clone(),
		stub.listErr
}

func TestHistoricalIntelligenceLatestEndpoint(
	t *testing.T,
) {
	record := historicalIntelligenceHandlerRecord()
	store := &historicalIntelligenceStoreStub{
		latest: record,
	}
	handler := NewHistoricalIntelligenceHandler(
		store,
	)
	app := fiber.New()
	app.Get(
		"/api/v1/historical-intelligence/aggregates/latest",
		handler.GetLatest,
	)

	requestURL :=
		"/api/v1/historical-intelligence/aggregates/latest" +
			"?metric=flight_count" +
			"&scope=global" +
			"&granularity=hour"
	result, err := app.Test(
		httptest.NewRequest(
			http.MethodGet,
			requestURL,
			nil,
		),
	)
	if err != nil {
		t.Fatalf(
			"execute latest request: %v",
			err,
		)
	}
	defer result.Body.Close()

	if result.StatusCode !=
		fiber.StatusOK {
		t.Fatalf(
			"status = %d, want 200",
			result.StatusCode,
		)
	}
	if store.latestQuery.SchemaVersion !=
		historicalcontract.SchemaVersionV1 ||
		store.latestQuery.MetricName !=
			historicalcontract.
				MetricNameFlightCount ||
		store.latestQuery.Scope.Type !=
			historicalcontract.ScopeTypeGlobal ||
		store.latestQuery.Granularity !=
			historicalcontract.GranularityHour {
		t.Fatalf(
			"unexpected latest query: %#v",
			store.latestQuery,
		)
	}

	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf(
			"read latest response: %v",
			err,
		)
	}
	text := string(body)
	for _, fragment := range []string{
		`"success":true`,
		`"schema_version":"historical-intelligence-v1"`,
		`"name":"flight_count"`,
		`"total":5`,
		`"direction":"up"`,
	} {
		if !strings.Contains(
			text,
			fragment,
		) {
			t.Fatalf(
				"response does not contain %q: %s",
				fragment,
				text,
			)
		}
	}
}

func TestHistoricalIntelligenceHistoryParsing(
	t *testing.T,
) {
	store := &historicalIntelligenceStoreStub{
		page: historicalaggregate.Page{},
	}
	handler := NewHistoricalIntelligenceHandler(
		store,
	)
	app := fiber.New()
	app.Get(
		"/api/v1/historical-intelligence/aggregates/history",
		handler.ListHistory,
	)

	before := time.Date(
		2026,
		time.July,
		15,
		10,
		0,
		0,
		123,
		time.UTC,
	)
	values := url.Values{}
	values.Set(
		"metric",
		"route_observations",
	)
	values.Set("scope", "route")
	values.Set("granularity", "hour")
	values.Set("origin_icao", "ubbb")
	values.Set("destination_icao", "ugtb")
	values.Set("limit", "7")
	values.Set(
		"before_window_end",
		before.Format(time.RFC3339Nano),
	)

	result, err := app.Test(
		httptest.NewRequest(
			http.MethodGet,
			"/api/v1/historical-intelligence/aggregates/history?"+
				values.Encode(),
			nil,
		),
	)
	if err != nil {
		t.Fatalf(
			"execute history request: %v",
			err,
		)
	}
	defer result.Body.Close()

	if result.StatusCode !=
		fiber.StatusOK {
		t.Fatalf(
			"status = %d, want 200",
			result.StatusCode,
		)
	}
	if store.listQuery.MetricName !=
		historicalcontract.
			MetricNameRouteObservations ||
		store.listQuery.Scope.Type !=
			historicalcontract.ScopeTypeRoute ||
		store.listQuery.Scope.OriginICAOCode !=
			"UBBB" ||
		store.listQuery.Scope.
			DestinationICAOCode != "UGTB" ||
		store.listQuery.Granularity !=
			historicalcontract.GranularityHour ||
		store.listQuery.Limit != 7 ||
		!store.listQuery.BeforeWindowEnd.Equal(
			before,
		) {
		t.Fatalf(
			"unexpected history query: %#v",
			store.listQuery,
		)
	}
}

func TestHistoricalIntelligenceRequestValidation(
	t *testing.T,
) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "metric required",
			path: "?scope=global&granularity=hour",
		},
		{
			name: "airport code required",
			path: "?metric=airport_departures&scope=airport&granularity=hour",
		},
		{
			name: "granularity supported",
			path: "?metric=flight_count&scope=global&granularity=month",
		},
		{
			name: "limit bounded",
			path: "?metric=flight_count&scope=global&granularity=hour&limit=0",
		},
		{
			name: "cursor timestamp",
			path: "?metric=flight_count&scope=global&granularity=hour&before_window_end=bad",
		},
		{
			name: "global scope rejects identifiers",
			path: "?metric=flight_count&scope=global&granularity=hour&airport_icao=UBBB",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				handler :=
					NewHistoricalIntelligenceHandler(
						&historicalIntelligenceStoreStub{},
					)
				app := fiber.New()
				app.Get(
					"/history",
					handler.ListHistory,
				)

				result, err := app.Test(
					httptest.NewRequest(
						http.MethodGet,
						"/history"+test.path,
						nil,
					),
				)
				if err != nil {
					t.Fatalf(
						"execute invalid request: %v",
						err,
					)
				}
				defer result.Body.Close()

				if result.StatusCode !=
					fiber.StatusBadRequest {
					t.Fatalf(
						"status = %d, want 400",
						result.StatusCode,
					)
				}
			},
		)
	}
}

func TestHistoricalIntelligenceNotFoundAndUnavailable(
	t *testing.T,
) {
	query :=
		"?metric=flight_count" +
			"&scope=global" +
			"&granularity=hour"

	missingHandler :=
		NewHistoricalIntelligenceHandler(
			&historicalIntelligenceStoreStub{
				latestErr: historicalaggregate.
					ErrResultNotFound,
			},
		)
	missingApp := fiber.New()
	missingApp.Get(
		"/latest",
		missingHandler.GetLatest,
	)
	missing, err := missingApp.Test(
		httptest.NewRequest(
			http.MethodGet,
			"/latest"+query,
			nil,
		),
	)
	if err != nil {
		t.Fatalf(
			"execute missing request: %v",
			err,
		)
	}
	defer missing.Body.Close()
	if missing.StatusCode !=
		fiber.StatusNotFound {
		t.Fatalf(
			"missing status = %d, want 404",
			missing.StatusCode,
		)
	}

	unavailableHandler :=
		NewHistoricalIntelligenceHandler(nil)
	unavailableApp := fiber.New()
	unavailableApp.Get(
		"/latest",
		unavailableHandler.GetLatest,
	)
	unavailable, err := unavailableApp.Test(
		httptest.NewRequest(
			http.MethodGet,
			"/latest"+query,
			nil,
		),
	)
	if err != nil {
		t.Fatalf(
			"execute unavailable request: %v",
			err,
		)
	}
	defer unavailable.Body.Close()
	if unavailable.StatusCode !=
		fiber.StatusServiceUnavailable {
		t.Fatalf(
			"unavailable status = %d, want 503",
			unavailable.StatusCode,
		)
	}
}

func historicalIntelligenceHandlerRecord() historicalaggregate.Record {
	currentValue := 5.0
	previousValue := 2.0
	percentageChange := 150.0
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
		-2 * time.Hour,
	)
	asOfTime := endTime.Add(
		30 * time.Minute,
	)
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
				EndTime: startTime.Add(
					time.Hour,
				),
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
				},
			},
			{
				StartTime: startTime.Add(
					time.Hour,
				),
				EndTime: endTime,
				Status: historicalcontract.
					BucketStatusComplete,
				Value:         3,
				SampleCount:   3,
				CoverageRatio: 1,
				Confidence: historicalcontract.
					Confidence{
					Score: 1,
					Level: historicalcontract.
						ConfidenceLevelHigh,
					SampleCount: 3,
				},
			},
		},
		Summary: historicalcontract.Summary{
			PointCount: 2,
			Total:      currentValue,
			Minimum:    2,
			Maximum:    3,
			Average:    2.5,
			Median:     2.5,
		},
		Comparison: &historicalcontract.PeriodComparison{
			PreviousWindow: historicalcontract.TimeWindow{
				StartTime: startTime.Add(
					-2 * time.Hour,
				),
				EndTime:  startTime,
				AsOfTime: asOfTime,
			},
			CurrentValue:  currentValue,
			PreviousValue: previousValue,
			AbsoluteChange: currentValue -
				previousValue,
			PercentageChange: &percentageChange,
			Direction: historicalcontract.
				TrendDirectionUp,
		},
		Confidence: historicalcontract.Confidence{
			Score: 1,
			Level: historicalcontract.
				ConfidenceLevelHigh,
			SampleCount: 5,
		},
		Provenance: historicalcontract.Provenance{
			BuilderVersion: "historical-traffic-intelligence-v1",
			InputFingerprint: "sha256:" +
				strings.Repeat(
					"a",
					64,
				),
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
		InputFingerprint: result.
			Provenance.InputFingerprint,
		Result:   result,
		StoredAt: asOfTime,
	}
}
