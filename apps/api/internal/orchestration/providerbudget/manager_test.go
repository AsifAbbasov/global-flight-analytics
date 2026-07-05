package providerbudget

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

func TestAirplanesLiveFixedWindowBudget(
	t *testing.T,
) {
	currentTime := time.Date(
		2026,
		time.July,
		4,
		19,
		30,
		0,
		0,
		time.UTC,
	)

	manager, err := New(
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

	firstDecision, err := manager.Acquire(
		providerpolicy.ProviderAirplanesLive,
	)
	if err != nil {
		t.Fatalf(
			"first acquire: %v",
			err,
		)
	}

	if !firstDecision.Allowed {
		t.Fatal(
			"expected first request to be allowed",
		)
	}

	secondDecision, err := manager.Acquire(
		providerpolicy.ProviderAirplanesLive,
	)
	if err != nil {
		t.Fatalf(
			"second acquire: %v",
			err,
		)
	}

	if secondDecision.Allowed {
		t.Fatal(
			"expected second request in same source-backed window to be denied",
		)
	}

	expectedRetryAt := currentTime.Truncate(
		time.Second,
	).Add(
		time.Second,
	)

	if !secondDecision.RetryAt.Equal(
		expectedRetryAt,
	) {
		t.Fatalf(
			"expected retry at %s, got %s",
			expectedRetryAt,
			secondDecision.RetryAt,
		)
	}

	currentTime = currentTime.Add(
		time.Second,
	)

	thirdDecision, err := manager.Acquire(
		providerpolicy.ProviderAirplanesLive,
	)
	if err != nil {
		t.Fatalf(
			"third acquire: %v",
			err,
		)
	}

	if !thirdDecision.Allowed {
		t.Fatal(
			"expected request in next source-backed window to be allowed",
		)
	}
}

func TestProviderReportedBudgetAndCooldown(
	t *testing.T,
) {
	currentTime := time.Date(
		2026,
		time.July,
		4,
		19,
		30,
		0,
		0,
		time.UTC,
	)

	manager, err := New(
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

	err = manager.ObserveProviderReportedBudget(
		providerpolicy.ProviderOpenSky,
		1,
		0,
	)
	if err != nil {
		t.Fatalf(
			"observe provider budget: %v",
			err,
		)
	}

	firstDecision, err := manager.Acquire(
		providerpolicy.ProviderOpenSky,
	)
	if err != nil {
		t.Fatalf(
			"first provider-reported acquire: %v",
			err,
		)
	}

	if !firstDecision.Allowed {
		t.Fatal(
			"expected request with remaining provider budget to be allowed",
		)
	}

	secondDecision, err := manager.Acquire(
		providerpolicy.ProviderOpenSky,
	)
	if err != nil {
		t.Fatalf(
			"second provider-reported acquire: %v",
			err,
		)
	}

	if secondDecision.Allowed {
		t.Fatal(
			"expected exhausted provider budget to deny request",
		)
	}

	providerRetryAfter := 5 * time.Second

	err = manager.ObserveProviderReportedBudget(
		providerpolicy.ProviderOpenSky,
		0,
		providerRetryAfter,
	)
	if err != nil {
		t.Fatalf(
			"observe provider cooldown: %v",
			err,
		)
	}

	cooldownDecision, err := manager.Acquire(
		providerpolicy.ProviderOpenSky,
	)
	if err != nil {
		t.Fatalf(
			"acquire during provider cooldown: %v",
			err,
		)
	}

	if cooldownDecision.Allowed {
		t.Fatal(
			"expected provider cooldown to deny request",
		)
	}

	expectedRetryAt := currentTime.Add(
		providerRetryAfter,
	)

	if !cooldownDecision.RetryAt.Equal(
		expectedRetryAt,
	) {
		t.Fatalf(
			"expected provider retry at %s, got %s",
			expectedRetryAt,
			cooldownDecision.RetryAt,
		)
	}

	currentTime = expectedRetryAt

	postCooldownDecision, err := manager.Acquire(
		providerpolicy.ProviderOpenSky,
	)
	if err != nil {
		t.Fatalf(
			"acquire after provider cooldown: %v",
			err,
		)
	}

	if !postCooldownDecision.Allowed {
		t.Fatal(
			"expected one provider probe after provider-directed cooldown",
		)
	}
}

func TestPublicationDrivenBudgetRejectsDuplicatePublication(
	t *testing.T,
) {
	manager, err := New(nil)
	if err != nil {
		t.Fatalf(
			"create provider budget manager: %v",
			err,
		)
	}

	firstDecision, err := manager.AcquirePublication(
		providerpolicy.ProviderOurAirports,
		"publication-a",
	)
	if err != nil {
		t.Fatalf(
			"first publication acquire: %v",
			err,
		)
	}

	if !firstDecision.Allowed {
		t.Fatal(
			"expected first publication to be allowed",
		)
	}

	secondDecision, err := manager.AcquirePublication(
		providerpolicy.ProviderOurAirports,
		"publication-a",
	)
	if err != nil {
		t.Fatalf(
			"second publication acquire: %v",
			err,
		)
	}

	if secondDecision.Allowed {
		t.Fatal(
			"expected duplicate publication to be denied",
		)
	}

	thirdDecision, err := manager.AcquirePublication(
		providerpolicy.ProviderOurAirports,
		"publication-b",
	)
	if err != nil {
		t.Fatalf(
			"third publication acquire: %v",
			err,
		)
	}

	if !thirdDecision.Allowed {
		t.Fatal(
			"expected new publication to be allowed",
		)
	}
}

func TestAcquireRejectsPublicationDrivenProviderWithoutPublicationID(
	t *testing.T,
) {
	manager, err := New(nil)
	if err != nil {
		t.Fatalf(
			"create provider budget manager: %v",
			err,
		)
	}

	_, err = manager.Acquire(
		providerpolicy.ProviderOurAirports,
	)

	if err == nil {
		t.Fatal(
			"expected publication access error",
		)
	}

	if !errors.Is(
		err,
		ErrPublicationAccessRequired,
	) {
		t.Fatalf(
			"expected ErrPublicationAccessRequired, got %v",
			err,
		)
	}
}

func TestNewRejectsDuplicateProviderPolicies(
	t *testing.T,
) {
	policy, err := providerpolicy.Get(
		providerpolicy.ProviderAirplanesLive,
	)
	if err != nil {
		t.Fatalf(
			"get provider policy: %v",
			err,
		)
	}

	_, err = NewWithPolicies(
		[]providerpolicy.Policy{
			policy,
			policy,
		},
		nil,
	)

	if err == nil {
		t.Fatal(
			"expected duplicate provider policy error",
		)
	}

	if !errors.Is(
		err,
		ErrDuplicateProviderPolicy,
	) {
		t.Fatalf(
			"expected ErrDuplicateProviderPolicy, got %v",
			err,
		)
	}
}
