package providerfallback

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

func TestSelectorPreservesOrderedAttemptEvidence(
	t *testing.T,
) {
	selector := New(nil)
	decision, err := selector.Select(
		[]Candidate{
			{
				Provider: providerpolicy.ProviderAirplanesLive,
				Allowed:  false,
				DenialReason: providerbudget.
					DecisionReasonProviderUnavailable,
				Outcome:          AttemptOutcomeFailed,
				ErrorClass:       AttemptErrorClassProviderServer,
				RequestAttempted: true,
			},
			{
				Provider:         providerpolicy.ProviderOpenSky,
				Allowed:          true,
				Outcome:          AttemptOutcomeSuccess,
				RequestAttempted: true,
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"select fallback provider: %v",
			err,
		)
	}
	if len(decision.Attempts) != 2 {
		t.Fatalf(
			"attempt count = %d, want 2",
			len(decision.Attempts),
		)
	}
	if decision.Attempts[0].Provider !=
		providerpolicy.ProviderAirplanesLive ||
		decision.Attempts[0].ErrorClass !=
			AttemptErrorClassProviderServer ||
		!decision.Attempts[0].RequestAttempted {
		t.Fatalf(
			"unexpected primary attempt: %+v",
			decision.Attempts[0],
		)
	}
	if decision.Attempts[1].Provider !=
		providerpolicy.ProviderOpenSky ||
		decision.Attempts[1].Outcome !=
			AttemptOutcomeSuccess {
		t.Fatalf(
			"unexpected fallback attempt: %+v",
			decision.Attempts[1],
		)
	}
}

func TestSelectorMarksTerminalFailure(
	t *testing.T,
) {
	selector := New(nil)
	decision, err := selector.Select(
		[]Candidate{
			{
				Provider: providerpolicy.ProviderAirplanesLive,
				Allowed:  false,
				DenialReason: providerbudget.
					DecisionReasonFixedWindowExhausted,
				Outcome:          AttemptOutcomeDenied,
				ErrorClass:       AttemptErrorClassAccessDenied,
				RequestAttempted: false,
			},
			{
				Provider: providerpolicy.ProviderOpenSky,
				Allowed:  false,
				DenialReason: providerbudget.
					DecisionReasonProviderUnavailable,
				Outcome:          AttemptOutcomeTerminalFailure,
				ErrorClass:       AttemptErrorClassUnauthorized,
				RequestAttempted: true,
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"select terminal failure: %v",
			err,
		)
	}
	if decision.Outcome != OutcomeTerminalFailure {
		t.Fatalf(
			"outcome = %s, want terminal_failure",
			decision.Outcome,
		)
	}
}

func TestNoProviderAvailableErrorExposesRetryAndRequestEvidence(
	t *testing.T,
) {
	retryAt := time.Date(
		2026,
		time.July,
		23,
		1,
		0,
		0,
		0,
		time.UTC,
	)
	err := &NoProviderAvailableError{
		Decision: Decision{
			RetryAt: retryAt,
			Attempts: []AttemptEvidence{
				{
					Provider: providerpolicy.ProviderAirplanesLive,
					Outcome:  AttemptOutcomeDenied,
				},
			},
		},
	}
	if !err.RetryAtTime().Equal(retryAt) {
		t.Fatalf(
			"retry at = %s, want %s",
			err.RetryAtTime(),
			retryAt,
		)
	}
	if err.ExternalRequestAttempted() {
		t.Fatal(
			"expected no external request evidence",
		)
	}
}
