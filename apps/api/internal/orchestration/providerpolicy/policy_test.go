package providerpolicy

import (
	"errors"
	"testing"
)

func TestAllPoliciesAreValid(
	t *testing.T,
) {
	policies := All()

	if len(policies) == 0 {
		t.Fatal(
			"expected provider policies",
		)
	}

	for _, policy := range policies {
		if err := Validate(policy); err != nil {
			t.Fatalf(
				"validate provider %s policy: %v",
				policy.Provider,
				err,
			)
		}
	}
}

func TestAirplanesLiveUsesSourceBackedLimit(
	t *testing.T,
) {
	policy, err := Get(
		ProviderAirplanesLive,
	)
	if err != nil {
		t.Fatalf(
			"get airplanes.live policy: %v",
			err,
		)
	}

	if policy.BudgetMode != BudgetModeFixedWindow {
		t.Fatalf(
			"expected fixed-window budget mode, got %s",
			policy.BudgetMode,
		)
	}

	if len(policy.RequestLimits) != 2 {
		t.Fatalf(
			"expected two request limits, got %d",
			len(policy.RequestLimits),
		)
	}

	secondLimit := policy.RequestLimits[0]
	if secondLimit.MaxRequests != 1 ||
		secondLimit.Window != WindowSecond ||
		secondLimit.Provenance != ProvenanceSourceBacked {
		t.Fatalf(
			"unexpected per-second limit: %+v",
			secondLimit,
		)
	}

	dailyLimit := policy.RequestLimits[1]
	if dailyLimit.MaxRequests != 500 ||
		dailyLimit.Window != WindowDay ||
		dailyLimit.Provenance != ProvenanceSourceBacked {
		t.Fatalf(
			"unexpected daily free-tier limit: %+v",
			dailyLimit,
		)
	}
}

func TestOpenSkyUsesProviderReportedBudget(
	t *testing.T,
) {
	policy, err := Get(
		ProviderOpenSky,
	)
	if err != nil {
		t.Fatalf(
			"get OpenSky policy: %v",
			err,
		)
	}

	if policy.BudgetMode != BudgetModeProviderReported {
		t.Fatalf(
			"expected provider-reported mode, got %s",
			policy.BudgetMode,
		)
	}

	budget := policy.ProviderReportedBudget
	if budget == nil {
		t.Fatal(
			"expected provider-reported budget metadata",
		)
	}

	if budget.RemainingHeader != "X-Rate-Limit-Remaining" {
		t.Fatalf(
			"unexpected remaining header: %s",
			budget.RemainingHeader,
		)
	}

	if budget.RetryAfterSecondsHeader != "X-Rate-Limit-Retry-After-Seconds" {
		t.Fatalf(
			"unexpected retry-after header: %s",
			budget.RetryAfterSecondsHeader,
		)
	}
}

func TestOurAirportsUsesPublicationDrivenPolicy(
	t *testing.T,
) {
	policy, err := Get(
		ProviderOurAirports,
	)
	if err != nil {
		t.Fatalf(
			"get OurAirports policy: %v",
			err,
		)
	}

	if policy.BudgetMode != BudgetModePublicationDriven {
		t.Fatalf(
			"expected publication-driven mode, got %s",
			policy.BudgetMode,
		)
	}

	publication := policy.PublicationPolicy
	if publication == nil {
		t.Fatal(
			"expected publication policy",
		)
	}

	if publication.Cadence != "nightly" {
		t.Fatalf(
			"expected nightly cadence, got %s",
			publication.Cadence,
		)
	}
}

func TestGetRejectsUnknownProvider(
	t *testing.T,
) {
	_, err := Get(
		Provider("unknown"),
	)
	if err == nil {
		t.Fatal(
			"expected unknown provider error",
		)
	}

	if !errors.Is(
		err,
		ErrUnknownProvider,
	) {
		t.Fatalf(
			"expected ErrUnknownProvider, got %v",
			err,
		)
	}
}
