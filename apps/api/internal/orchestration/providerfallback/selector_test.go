package providerfallback

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

func TestSelectUsesPrimaryWhenPrimaryIsAllowed(
	t *testing.T,
) {
	decidedAt := time.Date(
		2026,
		time.July,
		12,
		19,
		0,
		0,
		0,
		time.UTC,
	)

	selector := New(
		func() time.Time {
			return decidedAt
		},
	)

	decision, err := selector.Select(
		[]Candidate{
			{
				Provider: providerpolicy.ProviderAirplanesLive,
				Allowed:  true,
			},
			{
				Provider: providerpolicy.ProviderOpenSky,
				Allowed:  true,
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"select provider: %v",
			err,
		)
	}

	if decision.Outcome != OutcomePrimarySelected {
		t.Fatalf(
			"expected %s, got %s",
			OutcomePrimarySelected,
			decision.Outcome,
		)
	}

	if decision.SelectedProvider !=
		providerpolicy.ProviderAirplanesLive {
		t.Fatalf(
			"unexpected selected provider: %s",
			decision.SelectedProvider,
		)
	}

	if decision.UsedFallback {
		t.Fatal(
			"expected primary provider without fallback",
		)
	}

	if len(decision.ConsideredProviders) != 1 {
		t.Fatalf(
			"expected one considered provider, got %d",
			len(decision.ConsideredProviders),
		)
	}

	if !decision.DecidedAt.Equal(
		decidedAt,
	) {
		t.Fatalf(
			"expected decision time %s, got %s",
			decidedAt,
			decision.DecidedAt,
		)
	}
}

func TestSelectUsesFirstAllowedFallbackInDeclaredOrder(
	t *testing.T,
) {
	primaryRetryAt := time.Date(
		2026,
		time.July,
		12,
		19,
		0,
		1,
		0,
		time.UTC,
	)

	selector := New(nil)

	decision, err := selector.Select(
		[]Candidate{
			{
				Provider: providerpolicy.ProviderAirplanesLive,
				Allowed:  false,
				DenialReason: providerbudget.
					DecisionReasonFixedWindowExhausted,
				RetryAt: primaryRetryAt,
			},
			{
				Provider: providerpolicy.ProviderOpenSky,
				Allowed:  true,
			},
			{
				Provider: providerpolicy.ProviderOpenMeteo,
				Allowed:  true,
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"select fallback provider: %v",
			err,
		)
	}

	if decision.Outcome != OutcomeFallbackSelected {
		t.Fatalf(
			"expected %s, got %s",
			OutcomeFallbackSelected,
			decision.Outcome,
		)
	}

	if decision.SelectedProvider !=
		providerpolicy.ProviderOpenSky {
		t.Fatalf(
			"expected first allowed fallback, got %s",
			decision.SelectedProvider,
		)
	}

	if !decision.UsedFallback {
		t.Fatal(
			"expected fallback usage",
		)
	}

	if decision.TriggerReason !=
		providerbudget.DecisionReasonFixedWindowExhausted {
		t.Fatalf(
			"unexpected fallback trigger: %s",
			decision.TriggerReason,
		)
	}

	if !decision.RetryAt.Equal(
		primaryRetryAt,
	) {
		t.Fatalf(
			"expected primary retry at %s, got %s",
			primaryRetryAt,
			decision.RetryAt,
		)
	}

	if len(decision.ConsideredProviders) != 2 {
		t.Fatalf(
			"expected two considered providers, got %d",
			len(decision.ConsideredProviders),
		)
	}
}

func TestSelectReportsNoProviderAndEarliestRetry(
	t *testing.T,
) {
	firstRetryAt := time.Date(
		2026,
		time.July,
		12,
		19,
		0,
		5,
		0,
		time.UTC,
	)
	earliestRetryAt := firstRetryAt.Add(
		-3 * time.Second,
	)

	selector := New(nil)

	decision, err := selector.Select(
		[]Candidate{
			{
				Provider: providerpolicy.ProviderAirplanesLive,
				Allowed:  false,
				DenialReason: providerbudget.
					DecisionReasonFixedWindowExhausted,
				RetryAt: firstRetryAt,
			},
			{
				Provider: providerpolicy.ProviderOpenSky,
				Allowed:  false,
				DenialReason: providerbudget.
					DecisionReasonProviderCooldown,
				RetryAt: earliestRetryAt,
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"select unavailable providers: %v",
			err,
		)
	}

	if decision.Outcome != OutcomeNoProviderAvailable {
		t.Fatalf(
			"expected %s, got %s",
			OutcomeNoProviderAvailable,
			decision.Outcome,
		)
	}

	if decision.SelectedProvider != "" {
		t.Fatalf(
			"expected no selected provider, got %s",
			decision.SelectedProvider,
		)
	}

	if !decision.RetryAt.Equal(
		earliestRetryAt,
	) {
		t.Fatalf(
			"expected earliest retry at %s, got %s",
			earliestRetryAt,
			decision.RetryAt,
		)
	}

	if len(decision.ConsideredProviders) != 2 {
		t.Fatalf(
			"expected two considered providers, got %d",
			len(decision.ConsideredProviders),
		)
	}
}

func TestSelectRejectsDuplicateProvider(
	t *testing.T,
) {
	selector := New(nil)

	_, err := selector.Select(
		[]Candidate{
			{
				Provider: providerpolicy.ProviderAirplanesLive,
				Allowed:  true,
			},
			{
				Provider: providerpolicy.ProviderAirplanesLive,
				Allowed:  true,
			},
		},
	)

	if !errors.Is(
		err,
		ErrDuplicateProvider,
	) {
		t.Fatalf(
			"expected ErrDuplicateProvider, got %v",
			err,
		)
	}
}

func TestSelectRejectsInconsistentCandidateDecision(
	t *testing.T,
) {
	selector := New(nil)

	_, err := selector.Select(
		[]Candidate{
			{
				Provider: providerpolicy.ProviderAirplanesLive,
				Allowed:  false,
				DenialReason: providerbudget.
					DecisionReasonAllowed,
			},
		},
	)

	if !errors.Is(
		err,
		ErrInconsistentCandidateDecision,
	) {
		t.Fatalf(
			"expected ErrInconsistentCandidateDecision, got %v",
			err,
		)
	}
}

func TestSelectRejectsMissingCandidatesAndNilSelector(
	t *testing.T,
) {
	selector := New(nil)

	_, err := selector.Select(nil)
	if !errors.Is(
		err,
		ErrCandidatesRequired,
	) {
		t.Fatalf(
			"expected ErrCandidatesRequired, got %v",
			err,
		)
	}

	var nilSelector *Selector

	_, err = nilSelector.Select(
		[]Candidate{
			{
				Provider: providerpolicy.ProviderAirplanesLive,
				Allowed:  true,
			},
		},
	)
	if !errors.Is(
		err,
		ErrSelectorRequired,
	) {
		t.Fatalf(
			"expected ErrSelectorRequired, got %v",
			err,
		)
	}
}
