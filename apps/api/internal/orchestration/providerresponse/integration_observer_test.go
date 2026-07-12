package providerresponse

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	providerhealthdomain "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/providerhealth"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

type recordedTransportFailure struct {
	provider providerpolicy.Provider
	outcome  providerhealthdomain.RequestOutcome
	latency  time.Duration
}

type recordedResponseFailure struct {
	provider providerpolicy.Provider
	latency  time.Duration
}

type recordingObservationRecorder struct {
	observations      []Observation
	latencies         []time.Duration
	transportFailures []recordedTransportFailure
	responseFailures  []recordedResponseFailure
}

func (recorder *recordingObservationRecorder) RecordHTTPResponse(
	observation Observation,
	latency time.Duration,
) error {
	recorder.observations = append(
		recorder.observations,
		observation,
	)
	recorder.latencies = append(
		recorder.latencies,
		latency,
	)

	return nil
}

func (recorder *recordingObservationRecorder) RecordTransportFailure(
	provider providerpolicy.Provider,
	outcome providerhealthdomain.RequestOutcome,
	latency time.Duration,
) error {
	recorder.transportFailures = append(
		recorder.transportFailures,
		recordedTransportFailure{
			provider: provider,
			outcome:  outcome,
			latency:  latency,
		},
	)

	return nil
}

func (recorder *recordingObservationRecorder) RecordResponseFailure(
	provider providerpolicy.Provider,
	latency time.Duration,
) error {
	recorder.responseFailures = append(
		recorder.responseFailures,
		recordedResponseFailure{
			provider: provider,
			latency:  latency,
		},
	)

	return nil
}

func TestIntegrationObserverRecordsObservedHTTPResponse(t *testing.T) {
	t.Parallel()

	currentTime := time.Date(
		2026,
		time.July,
		12,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	budgetManager, err := providerbudget.New(
		func() time.Time {
			return currentTime
		},
	)
	if err != nil {
		t.Fatalf("create provider budget manager: %v", err)
	}

	controller, err := New(
		Config{
			BudgetManager: budgetManager,
			Now: func() time.Time {
				return currentTime
			},
		},
	)
	if err != nil {
		t.Fatalf("create provider response controller: %v", err)
	}

	recorder := &recordingObservationRecorder{}
	observer, err := NewIntegrationObserverWithRecorder(
		controller,
		recorder,
	)
	if err != nil {
		t.Fatalf("create integration observer: %v", err)
	}

	expectedLatency := 175 * time.Millisecond
	err = observer.ObserveProviderResponse(
		string(providerpolicy.ProviderAirplanesLive),
		http.StatusOK,
		make(http.Header),
		expectedLatency,
	)
	if err != nil {
		t.Fatalf("observe provider response: %v", err)
	}

	if len(recorder.observations) != 1 {
		t.Fatalf(
			"recorded observations = %d, want 1",
			len(recorder.observations),
		)
	}

	observation := recorder.observations[0]
	if observation.Provider != providerpolicy.ProviderAirplanesLive {
		t.Fatalf(
			"provider = %q, want %q",
			observation.Provider,
			providerpolicy.ProviderAirplanesLive,
		)
	}
	if observation.StatusCode != http.StatusOK {
		t.Fatalf(
			"status code = %d, want %d",
			observation.StatusCode,
			http.StatusOK,
		)
	}
	if len(recorder.latencies) != 1 ||
		recorder.latencies[0] != expectedLatency {
		t.Fatalf(
			"latencies = %v, want [%s]",
			recorder.latencies,
			expectedLatency,
		)
	}
}

func TestIntegrationObserverClassifiesTimeoutFailure(t *testing.T) {
	t.Parallel()

	observer, recorder := newTestIntegrationObserver(
		t,
	)
	expectedLatency := 2 * time.Second

	err := observer.ObserveProviderTransportFailure(
		string(providerpolicy.ProviderAirplanesLive),
		context.DeadlineExceeded,
		expectedLatency,
	)
	if err != nil {
		t.Fatalf("observe provider transport failure: %v", err)
	}

	if len(recorder.transportFailures) != 1 {
		t.Fatalf(
			"recorded failures = %d, want 1",
			len(recorder.transportFailures),
		)
	}

	failure := recorder.transportFailures[0]
	if failure.outcome != providerhealthdomain.RequestOutcomeTimeout {
		t.Fatalf(
			"outcome = %q, want %q",
			failure.outcome,
			providerhealthdomain.RequestOutcomeTimeout,
		)
	}
	if failure.latency != expectedLatency {
		t.Fatalf(
			"latency = %s, want %s",
			failure.latency,
			expectedLatency,
		)
	}
}

func TestIntegrationObserverClassifiesNetworkFailure(t *testing.T) {
	t.Parallel()

	observer, recorder := newTestIntegrationObserver(
		t,
	)

	err := observer.ObserveProviderTransportFailure(
		string(providerpolicy.ProviderAirplanesLive),
		errors.New("connection refused"),
		50*time.Millisecond,
	)
	if err != nil {
		t.Fatalf("observe provider transport failure: %v", err)
	}

	if len(recorder.transportFailures) != 1 {
		t.Fatalf(
			"recorded failures = %d, want 1",
			len(recorder.transportFailures),
		)
	}
	if recorder.transportFailures[0].outcome !=
		providerhealthdomain.RequestOutcomeNetworkError {
		t.Fatalf(
			"outcome = %q, want %q",
			recorder.transportFailures[0].outcome,
			providerhealthdomain.RequestOutcomeNetworkError,
		)
	}
}

func TestIntegrationObserverIgnoresCanceledRequest(t *testing.T) {
	t.Parallel()

	observer, recorder := newTestIntegrationObserver(
		t,
	)

	err := observer.ObserveProviderTransportFailure(
		string(providerpolicy.ProviderAirplanesLive),
		context.Canceled,
		10*time.Millisecond,
	)
	if err != nil {
		t.Fatalf("observe canceled request: %v", err)
	}
	if len(recorder.transportFailures) != 0 {
		t.Fatalf(
			"recorded failures = %d, want 0",
			len(recorder.transportFailures),
		)
	}
}

func TestIntegrationObserverRecordsInvalidResponseFailure(t *testing.T) {
	t.Parallel()

	observer, recorder := newTestIntegrationObserver(
		t,
	)
	expectedLatency := 75 * time.Millisecond

	err := observer.ObserveProviderResponseFailure(
		string(providerpolicy.ProviderAirplanesLive),
		errors.New("invalid json"),
		expectedLatency,
	)
	if err != nil {
		t.Fatalf("observe provider response failure: %v", err)
	}

	if len(recorder.responseFailures) != 1 {
		t.Fatalf(
			"recorded response failures = %d, want 1",
			len(recorder.responseFailures),
		)
	}

	failure := recorder.responseFailures[0]
	if failure.provider != providerpolicy.ProviderAirplanesLive {
		t.Fatalf(
			"provider = %q, want %q",
			failure.provider,
			providerpolicy.ProviderAirplanesLive,
		)
	}
	if failure.latency != expectedLatency {
		t.Fatalf(
			"latency = %s, want %s",
			failure.latency,
			expectedLatency,
		)
	}
}

func TestIntegrationObserverRejectsNilResponseFailure(t *testing.T) {
	t.Parallel()

	observer, _ := newTestIntegrationObserver(
		t,
	)

	err := observer.ObserveProviderResponseFailure(
		string(providerpolicy.ProviderAirplanesLive),
		nil,
		0,
	)
	if !errors.Is(
		err,
		ErrResponseFailureRequired,
	) {
		t.Fatalf(
			"error = %v, want %v",
			err,
			ErrResponseFailureRequired,
		)
	}
}

func newTestIntegrationObserver(
	t *testing.T,
) (*IntegrationObserver, *recordingObservationRecorder) {
	t.Helper()

	budgetManager, err := providerbudget.New(nil)
	if err != nil {
		t.Fatalf("create provider budget manager: %v", err)
	}

	controller, err := New(
		Config{
			BudgetManager: budgetManager,
		},
	)
	if err != nil {
		t.Fatalf("create provider response controller: %v", err)
	}

	recorder := &recordingObservationRecorder{}
	observer, err := NewIntegrationObserverWithRecorder(
		controller,
		recorder,
	)
	if err != nil {
		t.Fatalf("create integration observer: %v", err)
	}

	return observer, recorder
}
