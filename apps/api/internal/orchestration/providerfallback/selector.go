package providerfallback

import (
	"errors"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

var (
	ErrSelectorRequired = errors.New(
		"provider fallback selector is required",
	)
	ErrCandidatesRequired = errors.New(
		"provider fallback candidates are required",
	)
	ErrProviderRequired = errors.New(
		"provider fallback candidate provider is required",
	)
	ErrDuplicateProvider = errors.New(
		"duplicate provider fallback candidate",
	)
	ErrInconsistentCandidateDecision = errors.New(
		"inconsistent provider fallback candidate decision",
	)
)

type Outcome string

const (
	OutcomePrimarySelected     Outcome = "primary_selected"
	OutcomeFallbackSelected    Outcome = "fallback_selected"
	OutcomeNoProviderAvailable Outcome = "no_provider_available"
	OutcomeTerminalFailure     Outcome = "terminal_failure"
)

type AttemptOutcome string

const (
	AttemptOutcomeSuccess         AttemptOutcome = "success"
	AttemptOutcomeDenied          AttemptOutcome = "denied"
	AttemptOutcomeFailed          AttemptOutcome = "failed"
	AttemptOutcomeTerminalFailure AttemptOutcome = "terminal_failure"
)

type AttemptErrorClass string

const (
	AttemptErrorClassAccessDenied     AttemptErrorClass = "access_denied"
	AttemptErrorClassRateLimited      AttemptErrorClass = "rate_limited"
	AttemptErrorClassPollingCooldown  AttemptErrorClass = "polling_cooldown"
	AttemptErrorClassProviderServer   AttemptErrorClass = "provider_server"
	AttemptErrorClassNetwork          AttemptErrorClass = "network"
	AttemptErrorClassTimeout          AttemptErrorClass = "timeout"
	AttemptErrorClassUnauthorized     AttemptErrorClass = "unauthorized"
	AttemptErrorClassResponseTooLarge AttemptErrorClass = "response_too_large"
	AttemptErrorClassCancelled        AttemptErrorClass = "cancelled"
	AttemptErrorClassUnknown          AttemptErrorClass = "unknown"
)

type AttemptEvidence struct {
	Provider         providerpolicy.Provider
	Outcome          AttemptOutcome
	Reason           providerbudget.DecisionReason
	RetryAt          time.Time
	ErrorClass       AttemptErrorClass
	RequestAttempted bool
}

type Candidate struct {
	Provider         providerpolicy.Provider
	Allowed          bool
	DenialReason     providerbudget.DecisionReason
	RetryAt          time.Time
	Outcome          AttemptOutcome
	ErrorClass       AttemptErrorClass
	RequestAttempted bool
}

type Decision struct {
	PrimaryProvider     providerpolicy.Provider
	SelectedProvider    providerpolicy.Provider
	UsedFallback        bool
	Outcome             Outcome
	TriggerReason       providerbudget.DecisionReason
	ConsideredProviders []providerpolicy.Provider
	Attempts            []AttemptEvidence
	RetryAt             time.Time
	DecidedAt           time.Time
}

type Selector struct {
	now func() time.Time
}

func New(
	now func() time.Time,
) *Selector {
	if now == nil {
		now = time.Now
	}

	return &Selector{
		now: now,
	}
}

func (
	selector *Selector,
) Select(
	candidates []Candidate,
) (Decision, error) {
	if selector == nil {
		return Decision{}, ErrSelectorRequired
	}

	if len(candidates) == 0 {
		return Decision{}, ErrCandidatesRequired
	}

	if err := validateCandidates(
		candidates,
	); err != nil {
		return Decision{}, err
	}

	consideredProviders := make(
		[]providerpolicy.Provider,
		0,
		len(candidates),
	)
	attempts := make(
		[]AttemptEvidence,
		0,
		len(candidates),
	)

	primaryCandidate := candidates[0]

	for index, candidate := range candidates {
		consideredProviders = append(
			consideredProviders,
			candidate.Provider,
		)
		attempts = append(
			attempts,
			attemptEvidenceFromCandidate(candidate),
		)

		if !candidate.Allowed {
			continue
		}

		outcome := OutcomePrimarySelected
		usedFallback := false
		triggerReason := providerbudget.DecisionReason("")

		if index > 0 {
			outcome = OutcomeFallbackSelected
			usedFallback = true
			triggerReason = primaryCandidate.DenialReason
		}

		return Decision{
			PrimaryProvider:     primaryCandidate.Provider,
			SelectedProvider:    candidate.Provider,
			UsedFallback:        usedFallback,
			Outcome:             outcome,
			TriggerReason:       triggerReason,
			ConsideredProviders: consideredProviders,
			Attempts:            attempts,
			RetryAt:             primaryCandidate.RetryAt.UTC(),
			DecidedAt:           selector.now().UTC(),
		}, nil
	}

	outcome := OutcomeNoProviderAvailable
	if containsTerminalFailure(attempts) {
		outcome = OutcomeTerminalFailure
	}

	return Decision{
		PrimaryProvider:     primaryCandidate.Provider,
		UsedFallback:        len(candidates) > 1,
		Outcome:             outcome,
		TriggerReason:       primaryCandidate.DenialReason,
		ConsideredProviders: consideredProviders,
		Attempts:            attempts,
		RetryAt:             earliestRetryAt(candidates),
		DecidedAt:           selector.now().UTC(),
	}, nil
}

func validateCandidates(
	candidates []Candidate,
) error {
	providers := make(
		map[providerpolicy.Provider]struct{},
		len(candidates),
	)

	for index, candidate := range candidates {
		if candidate.Provider == "" {
			return fmt.Errorf(
				"%w: index=%d",
				ErrProviderRequired,
				index,
			)
		}

		if _, exists := providers[candidate.Provider]; exists {
			return fmt.Errorf(
				"%w: %s",
				ErrDuplicateProvider,
				candidate.Provider,
			)
		}

		providers[candidate.Provider] = struct{}{}

		if candidate.Allowed {
			if candidate.DenialReason != "" &&
				candidate.DenialReason !=
					providerbudget.DecisionReasonAllowed {
				return fmt.Errorf(
					"%w: provider=%s allowed=true reason=%s",
					ErrInconsistentCandidateDecision,
					candidate.Provider,
					candidate.DenialReason,
				)
			}
			if candidate.Outcome != "" &&
				candidate.Outcome != AttemptOutcomeSuccess {
				return fmt.Errorf(
					"%w: provider=%s allowed=true outcome=%s",
					ErrInconsistentCandidateDecision,
					candidate.Provider,
					candidate.Outcome,
				)
			}

			continue
		}

		if candidate.DenialReason == "" ||
			candidate.DenialReason ==
				providerbudget.DecisionReasonAllowed {
			return fmt.Errorf(
				"%w: provider=%s allowed=false reason=%s",
				ErrInconsistentCandidateDecision,
				candidate.Provider,
				candidate.DenialReason,
			)
		}
		if candidate.Outcome == AttemptOutcomeSuccess {
			return fmt.Errorf(
				"%w: provider=%s allowed=false outcome=%s",
				ErrInconsistentCandidateDecision,
				candidate.Provider,
				candidate.Outcome,
			)
		}
	}

	return nil
}

func attemptEvidenceFromCandidate(
	candidate Candidate,
) AttemptEvidence {
	outcome := candidate.Outcome
	requestAttempted := candidate.RequestAttempted

	if outcome == "" {
		if candidate.Allowed {
			outcome = AttemptOutcomeSuccess
			requestAttempted = true
		} else {
			outcome = AttemptOutcomeDenied
		}
	}

	return AttemptEvidence{
		Provider:         candidate.Provider,
		Outcome:          outcome,
		Reason:           candidate.DenialReason,
		RetryAt:          candidate.RetryAt.UTC(),
		ErrorClass:       candidate.ErrorClass,
		RequestAttempted: requestAttempted,
	}
}

func containsTerminalFailure(
	attempts []AttemptEvidence,
) bool {
	for _, attempt := range attempts {
		if attempt.Outcome == AttemptOutcomeTerminalFailure {
			return true
		}
	}

	return false
}

func earliestRetryAt(
	candidates []Candidate,
) time.Time {
	var earliest time.Time

	for _, candidate := range candidates {
		retryAt := candidate.RetryAt.UTC()
		if retryAt.IsZero() {
			continue
		}

		if earliest.IsZero() ||
			retryAt.Before(earliest) {
			earliest = retryAt
		}
	}

	return earliest
}
