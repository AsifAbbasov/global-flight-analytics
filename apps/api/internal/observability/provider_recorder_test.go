package observability

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	providerhealthdomain "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/providerhealth"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerresponse"
)

type providerRecorderNextStub struct {
	httpCalls      int
	transportCalls int
	responseCalls  int
}

func (stub *providerRecorderNextStub) RecordHTTPResponse(
	providerresponse.Observation,
	time.Duration,
) error {
	stub.httpCalls++
	return nil
}

func (stub *providerRecorderNextStub) RecordTransportFailure(
	providerpolicy.Provider,
	providerhealthdomain.RequestOutcome,
	time.Duration,
) error {
	stub.transportCalls++
	return nil
}

func (stub *providerRecorderNextStub) RecordResponseFailure(
	providerpolicy.Provider,
	time.Duration,
) error {
	stub.responseCalls++
	return nil
}

func TestProviderRecorderRecordsMetricsAndPreservesNextRecorder(
	t *testing.T,
) {
	registry := NewRegistry(BuildInfo{})
	next := &providerRecorderNextStub{}
	recorder := NewProviderRecorder(registry, next)

	if err := recorder.RecordHTTPResponse(
		providerresponse.Observation{
			Provider:   providerpolicy.ProviderOpenSky,
			StatusCode: 200,
		},
		100*time.Millisecond,
	); err != nil {
		t.Fatalf("record HTTP response: %v", err)
	}
	if err := recorder.RecordTransportFailure(
		providerpolicy.ProviderOpenSky,
		providerhealthdomain.RequestOutcomeTimeout,
		200*time.Millisecond,
	); err != nil {
		t.Fatalf("record transport failure: %v", err)
	}
	if err := recorder.RecordResponseFailure(
		providerpolicy.ProviderOpenSky,
		300*time.Millisecond,
	); err != nil {
		t.Fatalf("record response failure: %v", err)
	}

	if next.httpCalls != 1 || next.transportCalls != 1 || next.responseCalls != 1 {
		t.Fatalf("unexpected next recorder calls: %+v", next)
	}

	output := registry.Render(context.Background())
	for _, expected := range []string{
		`provider="opensky",outcome="success"} 1`,
		`provider="opensky",outcome="timeout"} 1`,
		`provider="opensky",outcome="invalid_response"} 1`,
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected provider metric %q\n%s", expected, output)
		}
	}
}

func TestProviderRecorderDoesNotConvertNextFailureIntoMetricLabel(
	t *testing.T,
) {
	registry := NewRegistry(BuildInfo{})
	recorder := NewProviderRecorder(
		registry,
		providerHTTPRecorderFunc(
			func(providerresponse.Observation, time.Duration) error {
				return errors.New("sensitive provider failure")
			},
		),
	)

	err := recorder.RecordHTTPResponse(
		providerresponse.Observation{
			Provider:   providerpolicy.ProviderOpenSky,
			StatusCode: 500,
		},
		time.Second,
	)
	if err == nil {
		t.Fatal("expected next recorder failure")
	}
	if strings.Contains(registry.Render(context.Background()), "sensitive provider failure") {
		t.Fatal("provider error text leaked into metrics")
	}
}

type providerHTTPRecorderFunc func(
	providerresponse.Observation,
	time.Duration,
) error

func (function providerHTTPRecorderFunc) RecordHTTPResponse(
	observation providerresponse.Observation,
	latency time.Duration,
) error {
	return function(observation, latency)
}
