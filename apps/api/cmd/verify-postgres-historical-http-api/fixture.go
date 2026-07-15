package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalseries"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

const (
	verificationOriginICAO      = "UBBB"
	verificationDestinationICAO = "UGTB"
	verificationBuilderVersion  = "historical-http-runtime-verification-v1"
)

type verificationSchedule struct {
	AsOfTime       time.Time
	GeneratedAt    time.Time
	ClosedBoundary time.Time
}

func buildVerificationSchedule(
	now time.Time,
) (verificationSchedule, error) {
	if now.IsZero() {
		return verificationSchedule{},
			fmt.Errorf("verification time is required")
	}

	asOfTime := now.UTC()
	return verificationSchedule{
		AsOfTime:       asOfTime,
		GeneratedAt:    asOfTime,
		ClosedBoundary: asOfTime.Truncate(time.Hour),
	}, nil
}

func buildVerificationResults(
	schedule verificationSchedule,
) ([]historicalcontract.Result, error) {
	results := make([]historicalcontract.Result, 0, 4)

	globalValues := []float64{1, 2, 3}
	for index, value := range globalValues {
		endTime := schedule.ClosedBoundary.Add(
			time.Duration(index-2) * time.Hour,
		)
		startTime := endTime.Add(-time.Hour)
		previousValue := value - 1
		if previousValue < 0 {
			previousValue = 0
		}

		result, err := buildVerificationResult(
			historicalcontract.Metric{
				Name:        historicalcontract.MetricNameFlightCount,
				Unit:        "flights",
				Aggregation: historicalcontract.AggregationCount,
			},
			historicalcontract.Scope{
				Type: historicalcontract.ScopeTypeGlobal,
			},
			startTime,
			endTime,
			schedule,
			value,
			previousValue,
			fmt.Sprintf("global-%d", index),
		)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	routeEnd := schedule.ClosedBoundary
	routeResult, err := buildVerificationResult(
		historicalcontract.Metric{
			Name:        historicalcontract.MetricNameRouteObservations,
			Unit:        "routes",
			Aggregation: historicalcontract.AggregationCount,
		},
		historicalcontract.Scope{
			Type:                historicalcontract.ScopeTypeRoute,
			OriginICAOCode:      verificationOriginICAO,
			DestinationICAOCode: verificationDestinationICAO,
		},
		routeEnd.Add(-time.Hour),
		routeEnd,
		schedule,
		4,
		2,
		"route",
	)
	if err != nil {
		return nil, err
	}
	results = append(results, routeResult)

	return results, nil
}

func buildVerificationResult(
	metric historicalcontract.Metric,
	scope historicalcontract.Scope,
	startTime time.Time,
	endTime time.Time,
	schedule verificationSchedule,
	value float64,
	previousValue float64,
	seed string,
) (historicalcontract.Result, error) {
	window := historicalcontract.TimeWindow{
		StartTime: startTime,
		EndTime:   endTime,
		AsOfTime:  schedule.AsOfTime,
	}
	bucket := historicalwindow.Bucket{
		Key:       "historical-http-runtime-" + seed,
		Sequence:  1,
		StartTime: startTime,
		EndTime:   endTime,
	}

	result, err := historicalseries.Build(
		historicalseries.BuildRequest{
			Metric: metric,
			Scope:  scope,
			Plan: historicalwindow.Plan{
				Version:            historicalwindow.Version,
				Fingerprint:        fingerprint("plan|" + seed),
				RequestedStartTime: startTime,
				RequestedEndTime:   endTime,
				AsOfTime:           schedule.AsOfTime,
				Granularity:        historicalcontract.GranularityHour,
				EffectiveWindow:    &window,
				Buckets:            []historicalwindow.Bucket{bucket},
				MaximumBucketCount: 10,
			},
			Values: []historicalseries.BucketValue{
				{
					Bucket:      bucket,
					Value:       value,
					SampleCount: int(value),
				},
			},
			DataCoverageRatio:     1,
			BuilderVersion:        verificationBuilderVersion,
			InputFingerprint:      fingerprint("result|" + seed),
			SourceNames:           []string{"historical_http_runtime_verification"},
			LatestSourceUpdatedAt: endTime,
			GeneratedAt:           schedule.GeneratedAt,
		},
	)
	if err != nil {
		return historicalcontract.Result{},
			fmt.Errorf("build verification result %s: %w", seed, err)
	}

	absoluteChange := value - previousValue
	var percentageChange *float64
	if previousValue != 0 {
		percentage := absoluteChange / previousValue * 100
		percentageChange = &percentage
	}
	result.Comparison = &historicalcontract.PeriodComparison{
		PreviousWindow: historicalcontract.TimeWindow{
			StartTime: startTime.Add(-time.Hour),
			EndTime:   startTime,
			AsOfTime:  schedule.AsOfTime,
		},
		CurrentValue:     value,
		PreviousValue:    previousValue,
		AbsoluteChange:   absoluteChange,
		PercentageChange: percentageChange,
		Direction: historicalcontract.TrendDirectionForChange(
			absoluteChange,
		),
	}

	report := historicalcontract.Validate(result)
	if report.Status != historicalcontract.ValidationStatusValid {
		return historicalcontract.Result{},
			fmt.Errorf(
				"verification result %s is invalid: errors=%d warnings=%d",
				seed,
				report.ErrorCount,
				report.WarningCount,
			)
	}

	return result, nil
}

func fingerprint(seed string) string {
	digest := sha256.Sum256([]byte(seed))
	return "sha256:" + hex.EncodeToString(digest[:])
}
