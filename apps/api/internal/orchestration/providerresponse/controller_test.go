package providerresponse

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

func TestTooManyRequestsCreatesProviderCooldown(
	t *testing.T,
) {
	currentTime := time.Date(
		2026,
		time.July,
		4,
		20,
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
		t.Fatalf(
			"create provider budget manager: %v",
			err,
		)
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
		t.Fatalf(
			"create provider response controller: %v",
			err,
		)
	}

	headers := make(
		http.Header,
	)

	headers.Set(
		"Retry-After",
		"7",
	)

	observation, err := controller.ObserveHTTPResponse(
		providerpolicy.ProviderAirplanesLive,
		http.StatusTooManyRequests,
		headers,
	)
	if err != nil {
		t.Fatalf(
			"observe HTTP response: %v",
			err,
		)
	}

	expectedRetryAt := currentTime.Add(
		7 * time.Second,
	)

	if !observation.CooldownUntil.Equal(
		expectedRetryAt,
	) {
		t.Fatalf(
			"expected cooldown until %s, got %s",
			expectedRetryAt,
			observation.CooldownUntil,
		)
	}

	decision, err := controller.Acquire(
		providerpolicy.ProviderAirplanesLive,
	)
	if err != nil {
		t.Fatalf(
			"acquire during cooldown: %v",
			err,
		)
	}

	if decision.Allowed {
		t.Fatal(
			"expected provider cooldown to deny access",
		)
	}

	if decision.Reason !=
		providerbudget.DecisionReasonProviderCooldown {
		t.Fatalf(
			"expected provider cooldown reason, got %s",
			decision.Reason,
		)
	}

	if !decision.RetryAt.Equal(
		expectedRetryAt,
	) {
		t.Fatalf(
			"expected retry at %s, got %s",
			expectedRetryAt,
			decision.RetryAt,
		)
	}
}

func TestOpenSkyResponseUpdatesProviderReportedBudget(
	t *testing.T,
) {
	currentTime := time.Date(
		2026,
		time.July,
		4,
		20,
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
		t.Fatalf(
			"create provider budget manager: %v",
			err,
		)
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
		t.Fatalf(
			"create provider response controller: %v",
			err,
		)
	}

	headers := make(
		http.Header,
	)

	headers.Set(
		"X-Rate-Limit-Remaining",
		"1",
	)

	observation, err := controller.ObserveHTTPResponse(
		providerpolicy.ProviderOpenSky,
		http.StatusOK,
		headers,
	)
	if err != nil {
		t.Fatalf(
			"observe OpenSky response: %v",
			err,
		)
	}

	if !observation.RemainingKnown {
		t.Fatal(
			"expected known provider remaining budget",
		)
	}

	if observation.Remaining != 1 {
		t.Fatalf(
			"expected remaining budget 1, got %d",
			observation.Remaining,
		)
	}

	firstDecision, err := controller.Acquire(
		providerpolicy.ProviderOpenSky,
	)
	if err != nil {
		t.Fatalf(
			"first OpenSky acquire: %v",
			err,
		)
	}

	if !firstDecision.Allowed {
		t.Fatal(
			"expected first OpenSky request to be allowed",
		)
	}

	secondDecision, err := controller.Acquire(
		providerpolicy.ProviderOpenSky,
	)
	if err != nil {
		t.Fatalf(
			"second OpenSky acquire: %v",
			err,
		)
	}

	if secondDecision.Allowed {
		t.Fatal(
			"expected exhausted OpenSky budget to deny access",
		)
	}

	if secondDecision.Reason !=
		providerbudget.DecisionReasonProviderBudgetExhausted {
		t.Fatalf(
			"expected provider budget exhausted reason, got %s",
			secondDecision.Reason,
		)
	}
}

func TestOpenSkyRetryHeaderCreatesProviderDirectedCooldown(
	t *testing.T,
) {
	currentTime := time.Date(
		2026,
		time.July,
		4,
		20,
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
		t.Fatalf(
			"create provider budget manager: %v",
			err,
		)
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
		t.Fatalf(
			"create provider response controller: %v",
			err,
		)
	}

	headers := make(
		http.Header,
	)

	headers.Set(
		"X-Rate-Limit-Remaining",
		"0",
	)

	headers.Set(
		"X-Rate-Limit-Retry-After-Seconds",
		"11",
	)

	observation, err := controller.ObserveHTTPResponse(
		providerpolicy.ProviderOpenSky,
		http.StatusTooManyRequests,
		headers,
	)
	if err != nil {
		t.Fatalf(
			"observe OpenSky rate limit response: %v",
			err,
		)
	}

	expectedRetryAt := currentTime.Add(
		11 * time.Second,
	)

	if !observation.CooldownUntil.Equal(
		expectedRetryAt,
	) {
		t.Fatalf(
			"expected cooldown until %s, got %s",
			expectedRetryAt,
			observation.CooldownUntil,
		)
	}

	decision, err := controller.Acquire(
		providerpolicy.ProviderOpenSky,
	)
	if err != nil {
		t.Fatalf(
			"acquire OpenSky during cooldown: %v",
			err,
		)
	}

	if decision.Allowed {
		t.Fatal(
			"expected OpenSky cooldown to deny access",
		)
	}
}

func TestStandardRetryAfterHTTPDateIsSupported(
	t *testing.T,
) {
	currentTime := time.Date(
		2026,
		time.July,
		4,
		20,
		0,
		0,
		0,
		time.UTC,
	)

	retryAt := currentTime.Add(
		13 * time.Second,
	)

	headers := make(
		http.Header,
	)

	headers.Set(
		"Retry-After",
		retryAt.Format(
			http.TimeFormat,
		),
	)

	policy, err := providerpolicy.Get(
		providerpolicy.ProviderAirplanesLive,
	)
	if err != nil {
		t.Fatalf(
			"get provider policy: %v",
			err,
		)
	}

	retryAfter, known, err := readRetryAfter(
		policy,
		headers,
		currentTime,
	)
	if err != nil {
		t.Fatalf(
			"read HTTP-date retry-after: %v",
			err,
		)
	}

	if !known {
		t.Fatal(
			"expected known retry-after value",
		)
	}

	if retryAfter != 13*time.Second {
		t.Fatalf(
			"expected retry-after 13s, got %s",
			retryAfter,
		)
	}
}

func TestMalformedRemainingBudgetHeaderIsRejected(
	t *testing.T,
) {
	budgetManager, err := providerbudget.New(nil)
	if err != nil {
		t.Fatalf(
			"create provider budget manager: %v",
			err,
		)
	}

	controller, err := New(
		Config{
			BudgetManager: budgetManager,
		},
	)
	if err != nil {
		t.Fatalf(
			"create provider response controller: %v",
			err,
		)
	}

	headers := make(
		http.Header,
	)

	headers.Set(
		"X-Rate-Limit-Remaining",
		"not-a-number",
	)

	_, err = controller.ObserveHTTPResponse(
		providerpolicy.ProviderOpenSky,
		http.StatusOK,
		headers,
	)

	if !errors.Is(
		err,
		ErrInvalidRemainingBudgetHeader,
	) {
		t.Fatalf(
			"expected ErrInvalidRemainingBudgetHeader, got %v",
			err,
		)
	}
}
