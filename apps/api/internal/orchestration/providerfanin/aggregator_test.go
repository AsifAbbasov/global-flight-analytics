package providerfanin

import (
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanout"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

func TestAggregateCreatesSucceededEnvelope(
	t *testing.T,
) {
	results := []providerfanout.Result{
		{
			TaskID:     "traffic",
			Provider:   providerpolicy.ProviderAirplanesLive,
			RequestKey: "traffic:regional-snapshot",
			Value:      "traffic-snapshot",
			Shared:     true,
		},
		{
			TaskID:     "weather",
			Provider:   providerpolicy.ProviderOpenMeteo,
			RequestKey: "weather:regional-context",
			Value:      "weather-snapshot",
		},
	}

	envelope := Aggregate(
		results,
	)

	if envelope.Status != BatchStatusSucceeded {
		t.Fatalf(
			"expected succeeded status, got %s",
			envelope.Status,
		)
	}

	if envelope.TotalCount != 2 {
		t.Fatalf(
			"expected total count 2, got %d",
			envelope.TotalCount,
		)
	}

	if envelope.SuccessCount != 2 {
		t.Fatalf(
			"expected success count 2, got %d",
			envelope.SuccessCount,
		)
	}

	if envelope.FailureCount != 0 {
		t.Fatalf(
			"expected failure count 0, got %d",
			envelope.FailureCount,
		)
	}

	if len(envelope.Successes) != 2 {
		t.Fatalf(
			"expected two successes, got %d",
			len(envelope.Successes),
		)
	}

	if !envelope.Successes[0].Shared {
		t.Fatal(
			"expected shared-result metadata to be preserved",
		)
	}
}

func TestAggregateCreatesPartialEnvelope(
	t *testing.T,
) {
	providerFailure := errors.New(
		"traffic provider failure",
	)

	results := []providerfanout.Result{
		{
			TaskID:     "traffic",
			Provider:   providerpolicy.ProviderAirplanesLive,
			RequestKey: "traffic:regional-snapshot",
			Err:        providerFailure,
		},
		{
			TaskID:     "weather",
			Provider:   providerpolicy.ProviderOpenMeteo,
			RequestKey: "weather:regional-context",
			Value:      "weather-snapshot",
		},
	}

	envelope := Aggregate(
		results,
	)

	if envelope.Status != BatchStatusPartial {
		t.Fatalf(
			"expected partial status, got %s",
			envelope.Status,
		)
	}

	if envelope.SuccessCount != 1 {
		t.Fatalf(
			"expected success count 1, got %d",
			envelope.SuccessCount,
		)
	}

	if envelope.FailureCount != 1 {
		t.Fatalf(
			"expected failure count 1, got %d",
			envelope.FailureCount,
		)
	}

	if len(envelope.Failures) != 1 {
		t.Fatalf(
			"expected one failure, got %d",
			len(envelope.Failures),
		)
	}

	if !errors.Is(
		envelope.Failures[0].Err,
		providerFailure,
	) {
		t.Fatalf(
			"expected provider failure, got %v",
			envelope.Failures[0].Err,
		)
	}

	if envelope.Successes[0].Value != "weather-snapshot" {
		t.Fatalf(
			"unexpected successful value: %v",
			envelope.Successes[0].Value,
		)
	}
}

func TestAggregateCreatesFailedEnvelope(
	t *testing.T,
) {
	results := []providerfanout.Result{
		{
			TaskID:   "traffic",
			Provider: providerpolicy.ProviderAirplanesLive,
			Err: errors.New(
				"traffic failure",
			),
		},
		{
			TaskID:   "weather",
			Provider: providerpolicy.ProviderOpenMeteo,
			Err: errors.New(
				"weather failure",
			),
		},
	}

	envelope := Aggregate(
		results,
	)

	if envelope.Status != BatchStatusFailed {
		t.Fatalf(
			"expected failed status, got %s",
			envelope.Status,
		)
	}

	if envelope.SuccessCount != 0 {
		t.Fatalf(
			"expected success count 0, got %d",
			envelope.SuccessCount,
		)
	}

	if envelope.FailureCount != 2 {
		t.Fatalf(
			"expected failure count 2, got %d",
			envelope.FailureCount,
		)
	}
}

func TestAggregateCreatesEmptyEnvelope(
	t *testing.T,
) {
	envelope := Aggregate(
		nil,
	)

	if envelope.Status != BatchStatusEmpty {
		t.Fatalf(
			"expected empty status, got %s",
			envelope.Status,
		)
	}

	if envelope.TotalCount != 0 {
		t.Fatalf(
			"expected total count 0, got %d",
			envelope.TotalCount,
		)
	}
}

func TestAggregatePreservesInputOrderInsideOutcomeGroups(
	t *testing.T,
) {
	firstFailure := errors.New(
		"first failure",
	)

	secondFailure := errors.New(
		"second failure",
	)

	results := []providerfanout.Result{
		{
			TaskID:   "traffic",
			Provider: providerpolicy.ProviderAirplanesLive,
			Err:      firstFailure,
		},
		{
			TaskID:   "weather",
			Provider: providerpolicy.ProviderOpenMeteo,
			Value:    "weather",
		},
		{
			TaskID:   "airports",
			Provider: providerpolicy.ProviderOurAirports,
			Err:      secondFailure,
		},
	}

	envelope := Aggregate(
		results,
	)

	if envelope.Successes[0].TaskID != "weather" {
		t.Fatalf(
			"expected weather success, got %s",
			envelope.Successes[0].TaskID,
		)
	}

	if envelope.Failures[0].TaskID != "traffic" {
		t.Fatalf(
			"expected first traffic failure, got %s",
			envelope.Failures[0].TaskID,
		)
	}

	if envelope.Failures[1].TaskID != "airports" {
		t.Fatalf(
			"expected second airports failure, got %s",
			envelope.Failures[1].TaskID,
		)
	}
}
